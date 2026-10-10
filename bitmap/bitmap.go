package bitmap

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/bits"
	"strings"

	"github.com/nqd2/igo8583/internal/diagnostic"
)

var (
	ErrInvalidField = errors.New("bitmap: field number must be between 1 and 192")
	ErrBufferShort  = errors.New("bitmap: buffer too short to unpack bitmap")
)

// Bitmap represents a 192-bit ISO-8583 bitmap (fields 1 to 192).
// bits[0] holds fields 1-64
// bits[1] holds fields 65-128
// bits[2] holds fields 129-192
type Bitmap struct {
	words [3]uint64
}

// New creates a new empty Bitmap.
func New() *Bitmap {
	return &Bitmap{}
}

// Set enables the given field number (1..192).
func (b *Bitmap) Set(field int) error {
	if field < 1 || field > 192 {
		return fmt.Errorf("%w: %d", ErrInvalidField, field)
	}
	idx := (field - 1) / 64
	bit := 63 - ((field - 1) % 64) // field 1 is MSB (bit 63)
	b.words[idx] |= (1 << bit)
	return nil
}

// Clear disables the given field number (1..192).
func (b *Bitmap) Clear(field int) error {
	if field < 1 || field > 192 {
		return fmt.Errorf("%w: %d", ErrInvalidField, field)
	}
	idx := (field - 1) / 64
	bit := 63 - ((field - 1) % 64)
	b.words[idx] &^= (1 << bit)
	return nil
}

// IsSet returns true if the specified field is set.
func (b *Bitmap) IsSet(field int) bool {
	if field < 1 || field > 192 {
		return false
	}
	idx := (field - 1) / 64
	bit := 63 - ((field - 1) % 64)
	return (b.words[idx] & (1 << bit)) != 0
}

// MaxField returns the highest field number set in this bitmap (0 if empty).
func (b *Bitmap) MaxField() int {
	for w := 2; w >= 0; w-- {
		val := b.words[w]
		if val != 0 {
			trailingZeros := bits.TrailingZeros64(val)
			bitPos := 63 - trailingZeros
			return w*64 + bitPos + 1
		}
	}
	return 0
}

// Fields returns a sorted slice of all active field numbers.
func (b *Bitmap) Fields() []int {
	var list []int
	maxF := b.MaxField()
	for f := 1; f <= maxF; f++ {
		// skip indicator bits when querying data fields if desired, or return all
		if b.IsSet(f) {
			list = append(list, f)
		}
	}
	return list
}

// PrepareIndicators automatically sets bit 1 if fields > 64 exist, and bit 65 if fields > 128 exist.
func (b *Bitmap) PrepareIndicators() {
	maxField := b.MaxField()
	if maxField > 128 {
		// Need tertiary: bit 1 and bit 65 must be set
		_ = b.Set(1)
		_ = b.Set(65)
	} else if maxField > 64 {
		// Need secondary: bit 1 must be set
		_ = b.Set(1)
		_ = b.Clear(65)
	} else {
		_ = b.Clear(1)
		_ = b.Clear(65)
	}
}

// PackBinary packs the bitmap into 8, 16, or 24 bytes based on whether secondary or tertiary bits are set.
func (b *Bitmap) PackBinary() []byte {
	b.PrepareIndicators()
	numBytes := 8
	if b.IsSet(65) || b.words[2] != 0 {
		numBytes = 24
	} else if b.IsSet(1) || b.words[1] != 0 {
		numBytes = 16
	}

	out := make([]byte, numBytes)
	for w := 0; w < numBytes/8; w++ {
		val := b.words[w]
		for i := 0; i < 8; i++ {
			out[w*8+i] = byte(val >> (56 - i*8))
		}
	}
	return out
}

// UnpackBinary unpacks the bitmap from wire bytes. Returns number of bytes consumed.
func (b *Bitmap) UnpackBinary(data []byte) (int, error) {
	if len(data) < 8 {
		return 0, fmt.Errorf("%w: need at least 8 bytes, got %d", ErrBufferShort, len(data))
	}
	b.words = [3]uint64{}

	// Word 0
	for i := 0; i < 8; i++ {
		b.words[0] = (b.words[0] << 8) | uint64(data[i])
	}
	consumed := 8

	// Check if secondary bitmap is present (bit 1)
	if b.IsSet(1) {
		if len(data) < 16 {
			return 0, fmt.Errorf("%w: secondary bitmap indicated but have %d bytes", ErrBufferShort, len(data))
		}
		for i := 0; i < 8; i++ {
			b.words[1] = (b.words[1] << 8) | uint64(data[8+i])
		}
		consumed = 16

		// Check if tertiary bitmap is present (bit 65)
		if b.IsSet(65) {
			if len(data) < 24 {
				return 0, fmt.Errorf("%w: tertiary bitmap indicated but have %d bytes", ErrBufferShort, len(data))
			}
			for i := 0; i < 8; i++ {
				b.words[2] = (b.words[2] << 8) | uint64(data[16+i])
			}
			consumed = 24
		}
	}

	return consumed, nil
}

// PackHex encodes the bitmap into 16, 32, or 48 uppercase Hex characters.
func (b *Bitmap) PackHex() []byte {
	bin := b.PackBinary()
	return []byte(strings.ToUpper(hex.EncodeToString(bin)))
}

// UnpackHex unpacks the bitmap from hex ASCII characters.
func (b *Bitmap) UnpackHex(data []byte) (int, error) {
	if len(data) < 16 {
		return 0, fmt.Errorf("%w: need at least 16 hex chars, got %d", ErrBufferShort, len(data))
	}

	// Read primary 16 hex chars (8 bytes)
	firstBlock, err := hex.DecodeString(string(data[:16]))
	if err != nil {
		return 0, diagnostic.Redact("bitmap: invalid hex", err)
	}

	b.words = [3]uint64{}
	for i := 0; i < 8; i++ {
		b.words[0] = (b.words[0] << 8) | uint64(firstBlock[i])
	}
	consumed := 16

	if b.IsSet(1) {
		if len(data) < 32 {
			return 0, fmt.Errorf("%w: secondary bitmap indicated but have %d hex chars", ErrBufferShort, len(data))
		}
		secBlock, err := hex.DecodeString(string(data[16:32]))
		if err != nil {
			return 0, diagnostic.Redact("bitmap: invalid secondary hex", err)
		}
		for i := 0; i < 8; i++ {
			b.words[1] = (b.words[1] << 8) | uint64(secBlock[i])
		}
		consumed = 32

		if b.IsSet(65) {
			if len(data) < 48 {
				return 0, fmt.Errorf("%w: tertiary bitmap indicated but have %d hex chars", ErrBufferShort, len(data))
			}
			tertBlock, err := hex.DecodeString(string(data[32:48]))
			if err != nil {
				return 0, diagnostic.Redact("bitmap: invalid tertiary hex", err)
			}
			for i := 0; i < 8; i++ {
				b.words[2] = (b.words[2] << 8) | uint64(tertBlock[i])
			}
			consumed = 48
		}
	}

	return consumed, nil
}

// String returns a hex representation of the bitmap.
func (b *Bitmap) String() string {
	return string(b.PackHex())
}

// Clone returns an independent deep copy of the Bitmap.
func (b *Bitmap) Clone() *Bitmap {
	clone := &Bitmap{}
	clone.words = b.words
	return clone
}

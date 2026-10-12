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

/*
Bitmap represents an ISO 8583 multi-bitmap structure supporting up to 192 fields:
  - words[0]: Primary Bitmap (fields 1–64)
  - words[1]: Secondary Bitmap (fields 65–128)
  - words[2]: Tertiary Bitmap (fields 129–192)

Bit 1 is the secondary bitmap presence indicator; Bit 65 is the tertiary bitmap presence indicator.
*/
type Bitmap struct {
	words [3]uint64
}

/* ============================================================================
 * Bitmap Representation & Bit Manipulation
 * ============================================================================ */

func New() *Bitmap {
	return &Bitmap{}
}

func (b *Bitmap) Set(field int) error {
	if field < 1 || field > 192 {
		return fmt.Errorf("%w: %d", ErrInvalidField, field)
	}
	idx := (field - 1) / 64
	bit := 63 - ((field - 1) % 64)
	b.words[idx] |= (1 << bit)
	return nil
}

func (b *Bitmap) Clear(field int) error {
	if field < 1 || field > 192 {
		return fmt.Errorf("%w: %d", ErrInvalidField, field)
	}
	idx := (field - 1) / 64
	bit := 63 - ((field - 1) % 64)
	b.words[idx] &^= (1 << bit)
	return nil
}

func (b *Bitmap) IsSet(field int) bool {
	if field < 1 || field > 192 {
		return false
	}
	idx := (field - 1) / 64
	bit := 63 - ((field - 1) % 64)
	return (b.words[idx] & (1 << bit)) != 0
}

/*
MaxField calculates the highest active data field index (1..192) in constant time
using CPU bitwise trailing zero intrinsics (0 if empty).
*/
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

/*
Fields returns a sorted slice of all active field numbers.
*/
func (b *Bitmap) Fields() []int {
	var list []int
	maxF := b.MaxField()
	for f := 1; f <= maxF; f++ {
		if b.IsSet(f) {
			list = append(list, f)
		}
	}
	return list
}

/*
PrepareIndicators automatically configures indicator bits:
  - Sets bit 1 if active fields exceed 64
  - Sets bit 65 if active fields exceed 128
  - Clears indicators when higher fields are not present
*/
func (b *Bitmap) PrepareIndicators() {
	maxField := b.MaxField()
	if maxField > 128 {
		_ = b.Set(1)
		_ = b.Set(65)
	} else if maxField > 64 {
		_ = b.Set(1)
		_ = b.Clear(65)
	} else {
		_ = b.Clear(1)
		_ = b.Clear(65)
	}
}

func (b *Bitmap) String() string {
	return string(b.PackHex())
}

func (b *Bitmap) Clone() *Bitmap {
	clone := &Bitmap{}
	clone.words = b.words
	return clone
}

/* ============================================================================
 * Binary Wire Packing & Unpacking
 * ============================================================================ */

/*
PackBinary encodes the bitmap into 8, 16, or 24 raw bytes depending on indicator bits.
*/
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

/*
UnpackBinary decodes 8, 16, or 24 bytes of binary bitmap data into the Bitmap struct.
Returns total consumed bytes.
*/
func (b *Bitmap) UnpackBinary(data []byte) (int, error) {
	if len(data) < 8 {
		return 0, fmt.Errorf("%w: need at least 8 bytes, got %d", ErrBufferShort, len(data))
	}
	b.words = [3]uint64{}

	for i := 0; i < 8; i++ {
		b.words[0] = (b.words[0] << 8) | uint64(data[i])
	}
	consumed := 8

	if b.IsSet(1) {
		if len(data) < 16 {
			return 0, fmt.Errorf("%w: secondary bitmap indicated but have %d bytes", ErrBufferShort, len(data))
		}
		for i := 0; i < 8; i++ {
			b.words[1] = (b.words[1] << 8) | uint64(data[8+i])
		}
		consumed = 16

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

/* ============================================================================
 * Hex ASCII Wire Packing & Unpacking
 * ============================================================================ */

/*
PackHex encodes the bitmap into 16, 32, or 48 uppercase Hex ASCII characters.
*/
func (b *Bitmap) PackHex() []byte {
	bin := b.PackBinary()
	return []byte(strings.ToUpper(hex.EncodeToString(bin)))
}

/*
UnpackHex decodes 16, 32, or 48 hexadecimal characters into the Bitmap struct.
Returns total consumed bytes.
*/
func (b *Bitmap) UnpackHex(data []byte) (int, error) {
	if len(data) < 16 {
		return 0, fmt.Errorf("%w: need at least 16 hex chars, got %d", ErrBufferShort, len(data))
	}

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

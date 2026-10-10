package prefixer

import (
	"fmt"
	"math"
)

type bcdPrefixer struct {
	nDigits int
	maxVal  int
}

// Bcd returns a packed BCD length prefixer with nDigits (e.g. LL = 2 digits -> 1 byte, LLL = 3 digits -> 2 bytes).
func Bcd(nDigits int) Prefixer {
	maxVal := int(math.Pow10(nDigits)) - 1
	return &bcdPrefixer{nDigits: nDigits, maxVal: maxVal}
}

var (
	BcdL      = Bcd(1)
	BcdLL     = Bcd(2)
	BcdLLL    = Bcd(3)
	BcdLLLL   = Bcd(4)
	BcdLLLLL  = Bcd(5)
	BcdLLLLLL = Bcd(6)
)

func (b *bcdPrefixer) PackedLength() int {
	return (b.nDigits + 1) / 2
}

func (b *bcdPrefixer) EncodeLength(length int) ([]byte, error) {
	if length < 0 || length > b.maxVal {
		return nil, fmt.Errorf("%w: length %d exceeds max %d for %d BCD digits", ErrLengthTooLong, length, b.maxVal, b.nDigits)
	}
	packedLen := b.PackedLength()
	buf := make([]byte, packedLen)

	rem := length
	for i := packedLen - 1; i >= 0; i-- {
		twoDigits := rem % 100
		rem /= 100
		hi := byte(twoDigits / 10)
		lo := byte(twoDigits % 10)
		buf[i] = (hi << 4) | lo
	}
	return buf, nil
}

func (b *bcdPrefixer) DecodeLength(data []byte) (int, int, error) {
	packedLen := b.PackedLength()
	if len(data) < packedLen {
		return 0, 0, fmt.Errorf("%w: need %d bytes, got %d", ErrBufferTooShort, packedLen, len(data))
	}

	var length int
	for i := 0; i < packedLen; i++ {
		val := data[i]
		hi := int(val >> 4)
		lo := int(val & 0x0F)
		if hi > 9 || lo > 9 {
			return 0, 0, fmt.Errorf("%w: invalid BCD nibble at byte %d", ErrInvalidDigit, i)
		}
		// If odd number of digits, first high nibble must be 0
		if i == 0 && (b.nDigits%2 != 0) && hi != 0 {
			return 0, 0, fmt.Errorf("%w: leading high nibble must be 0 for %d-digit BCD", ErrInvalidDigit, b.nDigits)
		}
		length = length*100 + hi*10 + lo
	}
	return length, packedLen, nil
}

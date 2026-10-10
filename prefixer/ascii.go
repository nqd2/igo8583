package prefixer

import (
	"fmt"
	"math"
	"strconv"
)

type asciiPrefixer struct {
	nDigits int
	maxVal  int
}

// Ascii returns an ASCII length prefixer with nDigits characters (e.g. LL = 2, LLL = 3).
func Ascii(nDigits int) Prefixer {
	maxVal := int(math.Pow10(nDigits)) - 1
	return &asciiPrefixer{nDigits: nDigits, maxVal: maxVal}
}

var (
	AsciiL      = Ascii(1)
	AsciiLL     = Ascii(2)
	AsciiLLL    = Ascii(3)
	AsciiLLLL   = Ascii(4)
	AsciiLLLLL  = Ascii(5)
	AsciiLLLLLL = Ascii(6)
)

func (a *asciiPrefixer) PackedLength() int {
	return a.nDigits
}

func (a *asciiPrefixer) EncodeLength(length int) ([]byte, error) {
	if length < 0 || length > a.maxVal {
		return nil, fmt.Errorf("%w: length %d exceeds max %d for %d digits", ErrLengthTooLong, length, a.maxVal, a.nDigits)
	}
	s := fmt.Sprintf("%0*d", a.nDigits, length)
	return []byte(s), nil
}

func (a *asciiPrefixer) DecodeLength(data []byte) (int, int, error) {
	if len(data) < a.nDigits {
		return 0, 0, fmt.Errorf("%w: need %d bytes, got %d", ErrBufferTooShort, a.nDigits, len(data))
	}
	str := string(data[:a.nDigits])
	for _, ch := range str {
		if ch < '0' || ch > '9' {
			return 0, 0, fmt.Errorf("%w: non-digit character", ErrInvalidDigit)
		}
	}
	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %v", ErrInvalidDigit, err)
	}
	return val, a.nDigits, nil
}

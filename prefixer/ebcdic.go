package prefixer

import (
	"fmt"
	"math"
)

type ebcdicPrefixer struct {
	nDigits int
	maxVal  int
}

/*
Ebcdic returns an EBCDIC length prefixer with nDigits characters.
*/
func Ebcdic(nDigits int) Prefixer {
	maxVal := int(math.Pow10(nDigits)) - 1
	return &ebcdicPrefixer{nDigits: nDigits, maxVal: maxVal}
}

/*
Pre-configured EBCDIC prefixer singletons.
*/
var (
	EbcdicL   = Ebcdic(1)
	EbcdicLL  = Ebcdic(2)
	EbcdicLLL = Ebcdic(3)
)

func (e *ebcdicPrefixer) PackedLength() int {
	return e.nDigits
}

func (e *ebcdicPrefixer) EncodeLength(length int) ([]byte, error) {
	if length < 0 || length > e.maxVal {
		return nil, fmt.Errorf("%w: length %d exceeds max %d for %d EBCDIC digits", ErrLengthTooLong, length, e.maxVal, e.nDigits)
	}
	buf := make([]byte, e.nDigits)
	rem := length
	for i := e.nDigits - 1; i >= 0; i-- {
		digit := rem % 10
		rem /= 10
		buf[i] = byte(0xF0 + digit)
	}
	return buf, nil
}

func (e *ebcdicPrefixer) DecodeLength(data []byte) (int, int, error) {
	if len(data) < e.nDigits {
		return 0, 0, fmt.Errorf("%w: need %d bytes, got %d", ErrBufferTooShort, e.nDigits, len(data))
	}
	var length int
	for i := 0; i < e.nDigits; i++ {
		b := data[i]
		if b < 0xF0 || b > 0xF9 {
			return 0, 0, fmt.Errorf("%w: invalid EBCDIC digit at byte %d", ErrInvalidDigit, i)
		}
		digit := int(b - 0xF0)
		length = length*10 + digit
	}
	return length, e.nDigits, nil
}

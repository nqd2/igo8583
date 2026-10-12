package interpreter

import (
	"fmt"
)

type asciiInterpreter struct{}

/*
ASCII returns an Interpreter for 7-bit US-ASCII string data.
*/
func ASCII() Interpreter {
	return asciiInterpreter{}
}

func (a asciiInterpreter) Pack(data string) ([]byte, error) {
	for _, b := range []byte(data) {
		if b > 0x7f {
			return nil, fmt.Errorf("%w: non-ASCII byte", ErrInvalidData)
		}
	}
	return []byte(data), nil
}

func (a asciiInterpreter) Unpack(data []byte, count int) (string, int, error) {
	if count < 0 {
		return "", 0, fmt.Errorf("%w: negative unit count", ErrInvalidData)
	}
	if len(data) < count {
		return "", 0, fmt.Errorf("%w: need %d bytes, have %d", ErrBufferTooShort, count, len(data))
	}
	for _, b := range data[:count] {
		if b > 0x7f {
			return "", 0, fmt.Errorf("%w: non-ASCII byte", ErrInvalidData)
		}
	}
	return string(data[:count]), count, nil
}

func (a asciiInterpreter) PackedLength(nUnits int) int {
	return nUnits
}

package interpreter

import (
	"fmt"
)

type binaryInterpreter struct{}

// Binary returns an Interpreter that treats bytes directly.
func Binary() Interpreter {
	return binaryInterpreter{}
}

func (b binaryInterpreter) PackedLength(nUnits int) int {
	return nUnits
}

func (b binaryInterpreter) Pack(data string) ([]byte, error) {
	return []byte(data), nil
}

func (b binaryInterpreter) Unpack(data []byte, count int) (string, int, error) {
	if count < 0 {
		return "", 0, fmt.Errorf("%w: negative unit count", ErrInvalidData)
	}
	if len(data) < count {
		return "", 0, fmt.Errorf("%w: Binary requires %d bytes, got %d", ErrBufferTooShort, count, len(data))
	}
	return string(data[:count]), count, nil
}

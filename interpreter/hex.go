package interpreter

import (
	"encoding/hex"
	"fmt"
	"strings"
)

type hexInterpreter struct{ rightPadded bool }

// Hex returns an Interpreter for hexadecimal strings.
// A hex string like "1A2B" packs to bytes [0x1A, 0x2B].
func Hex() Interpreter {
	return hexInterpreter{}
}

// HexRightPadded appends a zero nibble to odd-length hexadecimal values.
func HexRightPadded() Interpreter {
	return hexInterpreter{rightPadded: true}
}

func (h hexInterpreter) PackedLength(nUnits int) int {
	return nUnits/2 + nUnits%2
}

func (h hexInterpreter) Pack(data string) ([]byte, error) {
	s := data
	if len(s)%2 != 0 {
		if h.rightPadded {
			s += "0"
		} else {
			s = "0" + s
		}
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid hex string", ErrInvalidData)
	}
	return b, nil
}

func (h hexInterpreter) Unpack(data []byte, count int) (string, int, error) {
	if count < 0 {
		return "", 0, fmt.Errorf("%w: negative unit count", ErrInvalidData)
	}
	packedLen := h.PackedLength(count)
	if len(data) < packedLen {
		return "", 0, fmt.Errorf("%w: Hex requires %d bytes, got %d", ErrBufferTooShort, packedLen, len(data))
	}
	str := strings.ToUpper(hex.EncodeToString(data[:packedLen]))
	if len(str) > count {
		// handle odd length
		if h.rightPadded {
			if str[len(str)-1] != '0' {
				return "", 0, fmt.Errorf("%w: invalid trailing hex pad nibble", ErrInvalidData)
			}
			str = str[:count]
		} else {
			if str[0] != '0' {
				return "", 0, fmt.Errorf("%w: invalid leading hex pad nibble", ErrInvalidData)
			}
			str = str[len(str)-count:]
		}
	}
	return str, packedLen, nil
}

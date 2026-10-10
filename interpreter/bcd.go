package interpreter

import (
	"fmt"
	"strings"
)

type bcdInterpreter struct {
	leftPadded bool
	fPadded    bool
}

// BCDLeftPadded returns a BCD interpreter that pads an odd digit count with a leading zero (e.g. "123" -> 0x01, 0x23).
func BCDLeftPadded() Interpreter {
	return &bcdInterpreter{leftPadded: true, fPadded: false}
}

// BCDRightPadded returns a BCD interpreter that pads an odd digit count with a trailing zero (e.g. "123" -> 0x12, 0x30).
func BCDRightPadded() Interpreter {
	return &bcdInterpreter{leftPadded: false, fPadded: false}
}

// BCDRightPaddedF returns a BCD interpreter that pads an odd digit count with trailing 0xF (e.g. "123" -> 0x12, 0x3F).
func BCDRightPaddedF() Interpreter {
	return &bcdInterpreter{leftPadded: false, fPadded: true}
}

func (b *bcdInterpreter) PackedLength(nUnits int) int {
	return nUnits/2 + nUnits%2
}

func (b *bcdInterpreter) Pack(data string) ([]byte, error) {
	for _, ch := range data {
		if ch < '0' || ch > '9' {
			return nil, fmt.Errorf("%w: BCD data must only contain digits 0-9", ErrInvalidData)
		}
	}

	length := len(data)
	padded := data
	if length%2 != 0 {
		if b.leftPadded {
			if b.fPadded {
				padded = "F" + data
			} else {
				padded = "0" + data
			}
		} else {
			if b.fPadded {
				padded = data + "F"
			} else {
				padded = data + "0"
			}
		}
	}

	out := make([]byte, len(padded)/2)
	for i := 0; i < len(padded); i += 2 {
		hi := charToNibble(padded[i])
		lo := charToNibble(padded[i+1])
		out[i/2] = (hi << 4) | lo
	}
	return out, nil
}

func (b *bcdInterpreter) Unpack(data []byte, count int) (string, int, error) {
	if count < 0 {
		return "", 0, fmt.Errorf("%w: negative unit count", ErrInvalidData)
	}
	packedLen := b.PackedLength(count)
	if len(data) < packedLen {
		return "", 0, fmt.Errorf("%w: BCD requires %d bytes, got %d", ErrBufferTooShort, packedLen, len(data))
	}

	var sb strings.Builder
	sb.Grow(packedLen * 2)
	for i := 0; i < packedLen; i++ {
		hi := (data[i] >> 4) & 0x0F
		lo := data[i] & 0x0F
		sb.WriteByte(nibbleToChar(hi))
		sb.WriteByte(nibbleToChar(lo))
	}

	full := sb.String()
	for i := 0; i < len(full); i++ {
		if count%2 != 0 && ((b.leftPadded && i == 0) || (!b.leftPadded && i == len(full)-1)) {
			expected := byte('0')
			if b.fPadded {
				expected = 'F'
			}
			if full[i] != expected {
				return "", 0, fmt.Errorf("%w: invalid BCD pad nibble", ErrInvalidData)
			}
			continue
		}
		if full[i] < '0' || full[i] > '9' {
			return "", 0, fmt.Errorf("%w: invalid BCD digit", ErrInvalidData)
		}
	}
	if count%2 != 0 {
		if b.leftPadded {
			// drop first padded nibble
			full = full[1 : count+1]
		} else {
			// drop last padded nibble
			full = full[:count]
		}
	} else {
		full = full[:count]
	}

	return full, packedLen, nil
}

func charToNibble(c byte) byte {
	if c >= '0' && c <= '9' {
		return c - '0'
	}
	if c >= 'a' && c <= 'f' {
		return c - 'a' + 10
	}
	if c >= 'A' && c <= 'F' {
		return c - 'A' + 10
	}
	return 0
}

func nibbleToChar(n byte) byte {
	if n < 10 {
		return '0' + n
	}
	return 'A' + (n - 10)
}

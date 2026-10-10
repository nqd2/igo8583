package interpreter

import (
	"errors"
	"testing"
)

func TestUnpackRejectsNegativeCount(t *testing.T) {
	for name, codec := range map[string]Interpreter{"ascii": ASCII(), "binary": Binary(), "ebcdic": EBCDIC(), "hex": Hex(), "bcd": BCDLeftPadded()} {
		t.Run(name, func(t *testing.T) {
			_, consumed, err := codec.Unpack([]byte{0}, -1)
			if !errors.Is(err, ErrInvalidData) || consumed != 0 {
				t.Fatalf("expected invalid data, got %d, %v", consumed, err)
			}
		})
	}
}

func TestUnpackRejectsOversizedCount(t *testing.T) {
	for name, codec := range map[string]Interpreter{"hex": Hex(), "bcd": BCDLeftPadded()} {
		t.Run(name, func(t *testing.T) {
			_, _, err := codec.Unpack(nil, int(^uint(0)>>1))
			if !errors.Is(err, ErrBufferTooShort) {
				t.Fatalf("expected short buffer, got %v", err)
			}
		})
	}
}

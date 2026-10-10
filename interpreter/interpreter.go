package interpreter

import "errors"

var (
	ErrInvalidData    = errors.New("interpreter: invalid data format")
	ErrBufferTooShort = errors.New("interpreter: buffer too short")
)

// Interpreter handles data encoding/decoding between string representation and wire bytes.
type Interpreter interface {
	// Pack converts a string into wire bytes.
	Pack(data string) ([]byte, error)
	// Unpack converts wire bytes into a string representation for a given number of units (characters or digits).
	// It returns the unpacked string, the number of bytes consumed from data, and any error.
	Unpack(data []byte, count int) (string, int, error)
	// PackedLength returns the number of wire bytes required to represent nUnits.
	PackedLength(nUnits int) int
}

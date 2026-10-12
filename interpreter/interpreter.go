package interpreter

import "errors"

/* ============================================================================
 * Errors
 * ============================================================================ */

var (
	ErrInvalidData    = errors.New("interpreter: invalid data format")
	ErrBufferTooShort = errors.New("interpreter: buffer too short")
)

/* ============================================================================
 * Interpreter Interface
 * ============================================================================ */

/*
Interpreter handles character set conversion and byte representation
between high-level Go strings and wire encoding (e.g., ASCII, BCD, Hex, EBCDIC).
*/
type Interpreter interface {
	Pack(data string) ([]byte, error)
	Unpack(data []byte, count int) (string, int, error)
	PackedLength(nUnits int) int
}

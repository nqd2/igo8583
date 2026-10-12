package prefixer

import "errors"

/* ============================================================================
 * Errors
 * ============================================================================ */

var (
	ErrLengthTooLong  = errors.New("prefixer: length exceeds maximum allowed digits")
	ErrBufferTooShort = errors.New("prefixer: buffer too short to read length prefix")
	ErrInvalidDigit   = errors.New("prefixer: invalid digit in length prefix")
)

/* ============================================================================
 * Prefixer Interface
 * ============================================================================ */

/*
Prefixer defines strategies for encoding and decoding field length headers
(e.g., LLVAR, LLLVAR, fixed length) on the wire. Implementations are stateless
and concurrency-safe.
*/
type Prefixer interface {
	EncodeLength(length int) ([]byte, error)
	DecodeLength(data []byte) (int, int, error)
	PackedLength() int
}

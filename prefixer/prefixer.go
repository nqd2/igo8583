package prefixer

import "errors"

var (
	ErrLengthTooLong  = errors.New("prefixer: length exceeds maximum allowed digits")
	ErrBufferTooShort = errors.New("prefixer: buffer too short to read length prefix")
	ErrInvalidDigit   = errors.New("prefixer: invalid digit in length prefix")
)

// Prefixer defines how the length of a field is encoded and decoded in the wire protocol.
type Prefixer interface {
	// EncodeLength encodes the integer length into prefix bytes.
	EncodeLength(length int) ([]byte, error)
	// DecodeLength decodes the integer length from raw bytes.
	// Returns the decoded length, number of prefix bytes consumed, and error.
	DecodeLength(data []byte) (int, int, error)
	// PackedLength returns the byte size of this prefix.
	PackedLength() int
}

package packager

import (
	"encoding/hex"
	"fmt"
	"strings"
)

/* ============================================================================
 * Diagnostic Errors
 * ============================================================================ */

/*
UnpackError records detailed diagnostic information when message decoding fails,
including the field position, stream byte offset, and a hex snippet of remaining bytes.
*/
type UnpackError struct {
	FieldID      int
	Offset       int
	Consumed     int
	RemainingHex string
	Err          error
}

func (e *UnpackError) Error() string {
	return fmt.Sprintf("unpack error at field %d (offset %d): decoding failed", e.FieldID, e.Offset)
}

func (e *UnpackError) Unwrap() error {
	return e.Err
}

/*
NewUnpackError constructs an UnpackError capturing up to 16 bytes of remaining data for hex diagnostics.
*/
func NewUnpackError(fieldID, offset int, data []byte, err error) *UnpackError {
	if len(data) > 16 {
		data = data[:16]
	}
	remHex := strings.ToUpper(hex.EncodeToString(data))
	return &UnpackError{
		FieldID:      fieldID,
		Offset:       offset,
		RemainingHex: remHex,
		Err:          err,
	}
}

/*
PackError records a failure during message serialization, pinpointing the failing field ID.
*/
type PackError struct {
	FieldID int
	Err     error
}

func (e *PackError) Error() string {
	return fmt.Sprintf("pack error at field %d: %v", e.FieldID, e.Err)
}

func (e *PackError) Unwrap() error {
	return e.Err
}

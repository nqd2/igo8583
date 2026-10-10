package packager

import (
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/nqd2/igo8583/bitmap"
	"github.com/nqd2/igo8583/field"
	"github.com/nqd2/igo8583/interpreter"
	"github.com/nqd2/igo8583/prefixer"
	"github.com/nqd2/igo8583/tlv"
)

func TestUnpackErrorDoesNotPrintPayload(t *testing.T) {
	cause := errors.New("invalid field")
	err := NewUnpackError(2, 20, []byte("private-value"), cause)
	if strings.Contains(err.Error(), err.RemainingHex) {
		t.Fatal("error prints payload")
	}
	if err.RemainingHex == "" || !errors.Is(err, cause) {
		t.Fatal("debug context or error chain lost")
	}
	var unpackErr *UnpackError
	if !errors.As(err, &unpackErr) || unpackErr.Offset != 20 {
		t.Fatal("typed error lost")
	}
}

func TestHexErrorsHideInvalidBytesAndPreserveCause(t *testing.T) {
	_, bitmapErr := bitmap.New().UnpackHex([]byte("x000000000000000"))
	_, _, fieldErr := field.NewASCIIBinaryFieldPackager(1, "test", nil).Unpack([]byte("x0"), 0)
	valueErr := tlv.NewTLVMap(nil).SetHex("82", "x0")
	for _, err := range []error{bitmapErr, fieldErr, valueErr} {
		if err == nil || strings.Contains(err.Error(), "'x'") || strings.Contains(err.Error(), "U+0078") {
			t.Fatalf("error exposes input: %v", err)
		}
		var invalid hex.InvalidByteError
		if !errors.As(err, &invalid) {
			t.Fatal("hex error chain lost")
		}
	}
}

func TestBuiltinErrorsDoNotPrintInvalidBytes(t *testing.T) {
	_, _, err := prefixer.AsciiLL.DecodeLength([]byte("x1"))
	if err == nil || strings.Contains(err.Error(), "'x'") {
		t.Fatal("prefix error prints input byte")
	}
	_, err = interpreter.BCDLeftPadded().Pack("123x")
	if err == nil || strings.Contains(err.Error(), "'x'") {
		t.Fatal("BCD error prints input byte")
	}
	_, err = interpreter.ASCII().Pack(string([]byte{0xAB}))
	if err == nil || strings.Contains(err.Error(), "AB") {
		t.Fatal("ASCII error prints input byte")
	}
}

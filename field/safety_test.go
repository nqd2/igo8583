package field

import (
	"errors"
	"testing"

	"github.com/nqd2/igo8583/prefixer"
)

func TestUnpackRejectsInvalidOffset(t *testing.T) {
	codecs := map[string]FieldPackager{
		"string":    NewStringFieldPackager(4, "test", nil, nil, nil),
		"binary":    NewBinaryFieldPackager(4, "test", nil),
		"composite": NewCompositeFieldPackager(4, "test", prefixer.AsciiLL, nil),
		"tlv":       NewTLVFieldPackager(4, "test", prefixer.AsciiLL),
		"amount":    NewAmountFieldPackager(4, "test", false),
	}
	for name, codec := range codecs {
		t.Run(name, func(t *testing.T) {
			for _, offset := range []int{-1, 5} {
				_, consumed, err := codec.Unpack([]byte("0000"), offset)
				if !errors.Is(err, ErrBufferShort) || consumed != 0 {
					t.Fatalf("offset %d: got %d, %v", offset, consumed, err)
				}
			}
		})
	}
}

func TestAmountRejectsNonASCIISign(t *testing.T) {
	for _, binary := range []bool{false, true} {
		codec := NewAmountFieldPackager(3, "test", binary)
		data := []byte{0xFF, '0', '1'}
		if binary {
			data = []byte{0xFF, 0x01}
		}
		if _, consumed, err := codec.Unpack(data, 0); err == nil || consumed != 0 {
			t.Fatal("accepted sign that cannot be packed")
		}
	}
}

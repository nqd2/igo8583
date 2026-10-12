package field

import (
	"encoding/hex"
	"fmt"

	"github.com/nqd2/igo8583/internal/diagnostic"
	"github.com/nqd2/igo8583/prefixer"
)

type binaryFieldPackager struct {
	length      int
	description string
	prefixer    prefixer.Prefixer
	asciiHex    bool
}

/*
NewASCIIBinaryFieldPackager stores binary bytes as uppercase ASCII hex representation.
*/
func NewASCIIBinaryFieldPackager(length int, description string, pref prefixer.Prefixer) FieldPackager {
	b := NewBinaryFieldPackager(length, description, pref).(*binaryFieldPackager)
	b.asciiHex = true
	return b
}

/*
NewBinaryFieldPackager creates a FieldPackager for raw binary data elements.
*/
func NewBinaryFieldPackager(length int, description string, pref prefixer.Prefixer) FieldPackager {
	if pref == nil {
		pref = prefixer.Fixed(length)
	}
	return &binaryFieldPackager{
		length:      length,
		description: description,
		prefixer:    pref,
	}
}

func (b *binaryFieldPackager) Length() int {
	return b.length
}

func (b *binaryFieldPackager) Description() string {
	return b.description
}

func (b *binaryFieldPackager) Validate() error {
	if b.length <= 0 {
		return fmt.Errorf("field %s: length must be positive", b.description)
	}
	if _, err := b.prefixer.EncodeLength(b.length); err != nil {
		return fmt.Errorf("field %s: maximum length: %w", b.description, err)
	}
	return nil
}

func (b *binaryFieldPackager) Pack(c *Component) ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("field %s: nil component", b.description)
	}
	bytesVal := c.GetBytes()
	if len(bytesVal) > b.length {
		return nil, fmt.Errorf("%w: binary length %d > max %d (%s)", ErrLengthExceeded, len(bytesVal), b.length, b.description)
	}

	prefixBytes, err := b.prefixer.EncodeLength(len(bytesVal))
	if err != nil {
		return nil, fmt.Errorf("binary prefix encode error (%s): %w", b.description, err)
	}

	value := bytesVal
	if b.asciiHex {
		value = make([]byte, hex.EncodedLen(len(bytesVal)))
		hex.Encode(value, bytesVal)
		for i := range value {
			if value[i] >= 'a' && value[i] <= 'f' {
				value[i] -= 'a' - 'A'
			}
		}
	}
	out := make([]byte, len(prefixBytes)+len(value))
	copy(out, prefixBytes)
	copy(out[len(prefixBytes):], value)
	return out, nil
}

func (b *binaryFieldPackager) Unpack(data []byte, offset int) (*Component, int, error) {
	if offset < 0 || offset > len(data) {
		return nil, 0, fmt.Errorf("%w: offset %d exceeds buffer %d", ErrBufferShort, offset, len(data))
	}

	subData := data[offset:]
	byteLen, prefixBytes, err := b.prefixer.DecodeLength(subData)
	if err != nil {
		return nil, 0, fmt.Errorf("binary length decode error (%s): %w", b.description, err)
	}

	if byteLen > b.length {
		return nil, 0, fmt.Errorf("%w: decoded length %d > max %d (%s)", ErrLengthExceeded, byteLen, b.length, b.description)
	}

	wireLen := byteLen
	if b.asciiHex {
		wireLen = hex.EncodedLen(byteLen)
	}
	if len(subData) < prefixBytes+wireLen {
		return nil, 0, fmt.Errorf("%w: need %d bytes, have %d for (%s)",
			ErrBufferShort, prefixBytes+wireLen, len(subData), b.description)
	}

	val := make([]byte, byteLen)
	if b.asciiHex {
		if _, err := hex.Decode(val, subData[prefixBytes:prefixBytes+wireLen]); err != nil {
			return nil, 0, diagnostic.Redact(fmt.Sprintf("binary hex decode error (%s)", b.description), err)
		}
	} else {
		copy(val, subData[prefixBytes:prefixBytes+wireLen])
	}

	return NewBytesComponent(0, val), prefixBytes + wireLen, nil
}

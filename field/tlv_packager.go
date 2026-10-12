package field

import (
	"fmt"

	"github.com/nqd2/igo8583/prefixer"
	"github.com/nqd2/igo8583/tlv"
)

type tlvFieldPackager struct {
	length      int
	description string
	prefixer    prefixer.Prefixer
}

/* NewTLVFieldPackager creates a packager for BER-TLV formatted fields such as EMV data. */
func NewTLVFieldPackager(length int, description string, pref prefixer.Prefixer) FieldPackager {
	if pref == nil {
		pref = prefixer.Fixed(length)
	}
	return &tlvFieldPackager{
		length:      length,
		description: description,
		prefixer:    pref,
	}
}

func (t *tlvFieldPackager) Length() int {
	return t.length
}

func (t *tlvFieldPackager) Description() string {
	return t.description
}

func (t *tlvFieldPackager) Validate() error {
	if t.length <= 0 {
		return fmt.Errorf("field %s: length must be positive", t.description)
	}
	if _, err := t.prefixer.EncodeLength(t.length); err != nil {
		return fmt.Errorf("field %s: maximum length: %w", t.description, err)
	}
	return nil
}

func (t *tlvFieldPackager) Pack(c *Component) ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("field %s: nil component", t.description)
	}
	var body []byte
	var err error

	if c.TLV != nil {
		body, err = c.TLV.Encode()
		if err != nil {
			return nil, fmt.Errorf("tlv field %s encode error: %w", t.description, err)
		}
	} else if len(c.BytesVal) > 0 {
		body = c.BytesVal
	} else if c.StringVal != "" {
		return nil, fmt.Errorf("tlv field %s requires BER-TLV bytes or a TLV map", t.description)
	}
	if _, err := tlv.ParseTLV(body); err != nil {
		return nil, fmt.Errorf("tlv field %s contains invalid BER-TLV: %w", t.description, err)
	}

	if len(body) > t.length {
		return nil, fmt.Errorf("%w: tlv field %s size %d > max %d", ErrLengthExceeded, t.description, len(body), t.length)
	}

	prefixBytes, err := t.prefixer.EncodeLength(len(body))
	if err != nil {
		return nil, fmt.Errorf("tlv prefix encode error (%s): %w", t.description, err)
	}

	out := make([]byte, len(prefixBytes)+len(body))
	copy(out, prefixBytes)
	copy(out[len(prefixBytes):], body)
	return out, nil
}

func (t *tlvFieldPackager) Unpack(data []byte, offset int) (*Component, int, error) {
	if offset < 0 || offset > len(data) {
		return nil, 0, fmt.Errorf("%w: offset %d exceeds buffer %d", ErrBufferShort, offset, len(data))
	}

	subData := data[offset:]
	bodyLen, prefixBytes, err := t.prefixer.DecodeLength(subData)
	if err != nil {
		return nil, 0, fmt.Errorf("tlv length decode error (%s): %w", t.description, err)
	}

	if bodyLen > t.length {
		return nil, 0, fmt.Errorf("%w: decoded length %d > max %d (%s)", ErrLengthExceeded, bodyLen, t.length, t.description)
	}

	if len(subData) < prefixBytes+bodyLen {
		return nil, 0, fmt.Errorf("%w: need %d bytes, have %d for (%s)",
			ErrBufferShort, prefixBytes+bodyLen, len(subData), t.description)
	}

	body := subData[prefixBytes : prefixBytes+bodyLen]
	tlvMap, err := tlv.ParseTLV(body)
	if err != nil {
		return nil, 0, fmt.Errorf("tlv field %s decode error: %w", t.description, err)
	}

	comp := NewTLVComponent(0, tlvMap)
	comp.BytesVal = body
	return comp, prefixBytes + bodyLen, nil
}

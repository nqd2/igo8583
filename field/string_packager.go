package field

import (
	"errors"
	"fmt"

	"github.com/nqd2/igo8583/interpreter"
	"github.com/nqd2/igo8583/padder"
	"github.com/nqd2/igo8583/prefixer"
)

var (
	ErrLengthExceeded = errors.New("field: length exceeds maximum defined in spec")
	ErrBufferShort    = errors.New("field: buffer too short to unpack field")
)

type stringFieldPackager struct {
	length      int
	description string
	padder      padder.Padder
	interpreter interpreter.Interpreter
	prefixer    prefixer.Prefixer
}

/*
NewStringFieldPackager creates a new FieldPackager for string data elements.
*/
func NewStringFieldPackager(

	length int,
	description string,
	p padder.Padder,
	i interpreter.Interpreter,
	pref prefixer.Prefixer,
) FieldPackager {
	if p == nil {
		p = padder.NoPadding
	}
	if i == nil {
		i = interpreter.ASCII()
	}
	if pref == nil {
		pref = prefixer.Fixed(length)
	}
	return &stringFieldPackager{
		length:      length,
		description: description,
		padder:      p,
		interpreter: i,
		prefixer:    pref,
	}
}

func (s *stringFieldPackager) Length() int {
	return s.length
}

func (s *stringFieldPackager) Description() string {
	return s.description
}

func (s *stringFieldPackager) Validate() error {
	if s.length <= 0 {
		return fmt.Errorf("field %s: length must be positive", s.description)
	}
	if _, err := s.prefixer.EncodeLength(s.length); err != nil {
		return fmt.Errorf("field %s: maximum length: %w", s.description, err)
	}
	if s.prefixer.PackedLength() > 0 {
		probe, err := s.padder.Pad("X", s.length)
		if err != nil || probe != "X" {
			return fmt.Errorf("field %s: variable length cannot use fixed-width padding", s.description)
		}
	}
	return nil
}

func (s *stringFieldPackager) Pack(c *Component) ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("field %s: nil component", s.description)
	}
	str := c.GetString()
	if len(str) > s.length {
		return nil, fmt.Errorf("%w: field length %d > max %d (%s)", ErrLengthExceeded, len(str), s.length, s.description)
	}

	padded := str
	if s.prefixer.PackedLength() == 0 {
		var err error
		padded, err = s.padder.Pad(str, s.length)
		if err != nil {
			return nil, fmt.Errorf("field pack error (%s): %w", s.description, err)
		}
	}

	// Calculate units count: for variable length fields, prefix length reflects actual padded length
	prefixLenBytes, err := s.prefixer.EncodeLength(len(padded))
	if err != nil {
		return nil, fmt.Errorf("field prefix encode error (%s): %w", s.description, err)
	}

	interpBytes, err := s.interpreter.Pack(padded)
	if err != nil {
		return nil, fmt.Errorf("field interpret error (%s): %w", s.description, err)
	}

	out := make([]byte, len(prefixLenBytes)+len(interpBytes))
	copy(out, prefixLenBytes)
	copy(out[len(prefixLenBytes):], interpBytes)
	return out, nil
}

func (s *stringFieldPackager) Unpack(data []byte, offset int) (*Component, int, error) {
	if offset < 0 || offset > len(data) {
		return nil, 0, fmt.Errorf("%w: offset %d exceeds buffer %d", ErrBufferShort, offset, len(data))
	}

	subData := data[offset:]
	contentUnits, prefixBytes, err := s.prefixer.DecodeLength(subData)
	if err != nil {
		return nil, 0, fmt.Errorf("field length decode error (%s): %w", s.description, err)
	}

	if contentUnits > s.length {
		return nil, 0, fmt.Errorf("%w: decoded length %d > max %d (%s)", ErrLengthExceeded, contentUnits, s.length, s.description)
	}

	neededBytes := s.interpreter.PackedLength(contentUnits)
	if len(subData) < prefixBytes+neededBytes {
		return nil, 0, fmt.Errorf("%w: need %d bytes, have %d for field (%s)",
			ErrBufferShort, prefixBytes+neededBytes, len(subData), s.description)
	}

	valueData := subData[prefixBytes : prefixBytes+neededBytes]
	unpackedStr, consumedUnits, err := s.interpreter.Unpack(valueData, contentUnits)
	if err != nil {
		return nil, 0, fmt.Errorf("field unpack error (%s): %w", s.description, err)
	}

	comp := NewStringComponent(0, unpackedStr)
	_ = consumedUnits // equal to neededBytes

	return comp, prefixBytes + neededBytes, nil
}

package field

import (
	"fmt"
	"strings"

	"github.com/nqd2/igo8583/interpreter"
)

// amountFieldPackager encodes a sign byte followed by zero-padded digits.
type amountFieldPackager struct {
	length      int
	description string
	binary      bool
}

// NewAmountFieldPackager creates an ISO87 signed amount field packager.
func NewAmountFieldPackager(length int, description string, binary bool) FieldPackager {
	return &amountFieldPackager{length: length, description: description, binary: binary}
}

func (a *amountFieldPackager) Length() int         { return a.length }
func (a *amountFieldPackager) Description() string { return a.description }

func (a *amountFieldPackager) Validate() error {
	if a.length < 2 || (a.binary && (a.length-1)%2 != 0) {
		return fmt.Errorf("amount %s: invalid length %d", a.description, a.length)
	}
	return nil
}

func (a *amountFieldPackager) Pack(c *Component) ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("amount %s: nil component", a.description)
	}
	s := c.GetString()
	if len(s) < 1 || len(s) > a.length || s[0] > 0x7F {
		return nil, fmt.Errorf("amount %s: invalid signed value length", a.description)
	}
	digits := s[1:]
	for i := range digits {
		if digits[i] < '0' || digits[i] > '9' {
			return nil, fmt.Errorf("amount %s: nonnumeric digit", a.description)
		}
	}
	digits = strings.Repeat("0", a.length-1-len(digits)) + digits
	if !a.binary {
		return []byte(s[:1] + digits), nil
	}
	packed, err := interpreter.BCDLeftPadded().Pack(digits)
	if err != nil {
		return nil, err
	}
	return append([]byte{s[0]}, packed...), nil
}

func (a *amountFieldPackager) Unpack(data []byte, offset int) (*Component, int, error) {
	if offset < 0 || offset > len(data) {
		return nil, 0, fmt.Errorf("%w: amount offset outside buffer", ErrBufferShort)
	}
	needed := a.length
	if a.binary {
		needed = 1 + (a.length-1)/2
	}
	if len(data)-offset < needed {
		return nil, 0, fmt.Errorf("%w: amount %s needs %d bytes", ErrBufferShort, a.description, needed)
	}
	if data[offset] > 0x7F {
		return nil, 0, fmt.Errorf("amount %s: non-ASCII sign", a.description)
	}
	sign := string(data[offset : offset+1])
	var digits string
	if a.binary {
		var err error
		digits, _, err = interpreter.BCDLeftPadded().Unpack(data[offset+1:offset+needed], a.length-1)
		if err != nil {
			return nil, 0, err
		}
	} else {
		digits = string(data[offset+1 : offset+needed])
		for i := range digits {
			if digits[i] < '0' || digits[i] > '9' {
				return nil, 0, fmt.Errorf("amount %s: nonnumeric digit", a.description)
			}
		}
	}
	return NewStringComponent(0, sign+digits), needed, nil
}

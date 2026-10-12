package field

import (
	"fmt"
	"sort"

	"github.com/nqd2/igo8583/prefixer"
)

type compositeFieldPackager struct {
	length       int
	description  string
	prefixer     prefixer.Prefixer
	subPackagers map[int]FieldPackager
}

/*
NewCompositeFieldPackager creates a packager for composite fields with nested subfields.
*/
func NewCompositeFieldPackager(

	length int,
	description string,
	pref prefixer.Prefixer,
	subPackagers map[int]FieldPackager,
) FieldPackager {
	if pref == nil {
		pref = prefixer.Fixed(length)
	}
	copied := make(map[int]FieldPackager, len(subPackagers))
	for id, fp := range subPackagers {
		copied[id] = fp
	}
	return &compositeFieldPackager{
		length:       length,
		description:  description,
		prefixer:     pref,
		subPackagers: copied,
	}
}

func (c *compositeFieldPackager) Length() int {
	return c.length
}

func (c *compositeFieldPackager) Description() string {
	return c.description
}

func (c *compositeFieldPackager) Validate() error {
	if c.length <= 0 {
		return fmt.Errorf("field %s: length must be positive", c.description)
	}
	if _, err := c.prefixer.EncodeLength(c.length); err != nil {
		return fmt.Errorf("field %s: maximum length: %w", c.description, err)
	}
	for id, sub := range c.subPackagers {
		if id < 1 || id > 192 || sub == nil || sub.Length() <= 0 {
			return fmt.Errorf("field %s: invalid subfield %d", c.description, id)
		}
		if v, ok := sub.(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return fmt.Errorf("field %s subfield %d: %w", c.description, id, err)
			}
		}
	}
	return nil
}

func (c *compositeFieldPackager) Pack(comp *Component) ([]byte, error) {
	if comp == nil {
		return nil, fmt.Errorf("composite field %s: nil component", c.description)
	}
	var body []byte

	if comp.Subfields != nil && len(comp.Subfields) > 0 {
		if len(comp.Subfields) != len(c.subPackagers) {
			return nil, fmt.Errorf("composite field %s: sequential layout requires every declared subfield", c.description)
		}
		// Collect and sort subfield IDs
		var subIDs []int
		for id := range comp.Subfields {
			subIDs = append(subIDs, id)
		}
		sort.Ints(subIDs)

		for _, id := range subIDs {
			subComp := comp.Subfields[id]
			if subComp == nil {
				return nil, fmt.Errorf("composite field %s: nil subfield %d", c.description, id)
			}
			packager, ok := c.subPackagers[id]
			if !ok {
				return nil, fmt.Errorf("composite field %s: missing packager for subfield %d", c.description, id)
			}
			packedSub, err := packager.Pack(subComp)
			if err != nil {
				return nil, fmt.Errorf("composite field %s subfield %d pack error: %w", c.description, id, err)
			}
			body = append(body, packedSub...)
		}
	} else if len(c.subPackagers) > 0 {
		return nil, fmt.Errorf("composite field %s: declared subfields are required", c.description)
	} else if len(comp.BytesVal) > 0 {
		body = comp.BytesVal
	} else if comp.StringVal != "" {
		body = []byte(comp.StringVal)
	}

	if len(body) > c.length {
		return nil, fmt.Errorf("%w: composite field %s size %d > max %d", ErrLengthExceeded, c.description, len(body), c.length)
	}

	prefixBytes, err := c.prefixer.EncodeLength(len(body))
	if err != nil {
		return nil, fmt.Errorf("composite prefix encode error (%s): %w", c.description, err)
	}

	out := make([]byte, len(prefixBytes)+len(body))
	copy(out, prefixBytes)
	copy(out[len(prefixBytes):], body)
	return out, nil
}

func (c *compositeFieldPackager) Unpack(data []byte, offset int) (*Component, int, error) {
	if offset < 0 || offset > len(data) {
		return nil, 0, fmt.Errorf("%w: offset %d exceeds buffer %d", ErrBufferShort, offset, len(data))
	}

	subData := data[offset:]
	bodyLen, prefixBytes, err := c.prefixer.DecodeLength(subData)
	if err != nil {
		return nil, 0, fmt.Errorf("composite length decode error (%s): %w", c.description, err)
	}

	if bodyLen > c.length {
		return nil, 0, fmt.Errorf("%w: decoded length %d > max %d (%s)", ErrLengthExceeded, bodyLen, c.length, c.description)
	}

	if len(subData) < prefixBytes+bodyLen {
		return nil, 0, fmt.Errorf("%w: need %d bytes, have %d for (%s)",
			ErrBufferShort, prefixBytes+bodyLen, len(subData), c.description)
	}

	body := subData[prefixBytes : prefixBytes+bodyLen]
	comp := NewCompositeComponent(0)
	comp.BytesVal = make([]byte, len(body))
	copy(comp.BytesVal, body)

	// Unpack known subfields sequentially
	if len(c.subPackagers) > 0 {
		var sortedIDs []int
		for id := range c.subPackagers {
			sortedIDs = append(sortedIDs, id)
		}
		sort.Ints(sortedIDs)

		subOffset := 0
		for _, id := range sortedIDs {
			packager := c.subPackagers[id]
			subComp, consumed, err := packager.Unpack(body, subOffset)
			if err != nil {
				return nil, 0, fmt.Errorf("composite subfield %d unpack error (%s): %w", id, c.description, err)
			}
			subComp.FieldID = id
			comp.SetSubfield(id, subComp)
			subOffset += consumed
		}
		if subOffset != len(body) {
			return nil, 0, fmt.Errorf("composite field %s: %d trailing bytes", c.description, len(body)-subOffset)
		}
	}

	return comp, prefixBytes + bodyLen, nil
}

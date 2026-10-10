package field

import (
	"encoding/hex"
	"strings"

	"github.com/nqd2/igo8583/tlv"
)

// Component represents a field in the message tree.
// Byte slices, subfields, and TLV maps are mutable references; use Clone for an independent copy.
type Component struct {
	FieldID   int
	StringVal string
	BytesVal  []byte
	Subfields map[int]*Component
	TLV       *tlv.TLVMap
}

// NewStringComponent creates a component with a string value.
func NewStringComponent(id int, val string) *Component {
	return &Component{
		FieldID:   id,
		StringVal: val,
	}
}

// NewBytesComponent creates a component with binary byte values.
func NewBytesComponent(id int, val []byte) *Component {
	return &Component{
		FieldID:  id,
		BytesVal: val,
	}
}

// NewCompositeComponent creates a component that contains nested subfields.
func NewCompositeComponent(id int) *Component {
	return &Component{
		FieldID:   id,
		Subfields: make(map[int]*Component),
	}
}

// NewTLVComponent creates a component with a TLV map.
func NewTLVComponent(id int, tlvMap *tlv.TLVMap) *Component {
	return &Component{
		FieldID: id,
		TLV:     tlvMap,
	}
}

// GetString returns the string value of this component.
func (c *Component) GetString() string {
	if c.StringVal != "" {
		return c.StringVal
	}
	if len(c.BytesVal) > 0 {
		return string(c.BytesVal)
	}
	return ""
}

// GetBytes returns the byte slice value of this component.
func (c *Component) GetBytes() []byte {
	if len(c.BytesVal) > 0 {
		return c.BytesVal
	}
	if c.StringVal != "" {
		return []byte(c.StringVal)
	}
	return nil
}

// GetHex returns the value encoded in uppercase hex.
func (c *Component) GetHex() string {
	if len(c.BytesVal) > 0 {
		return strings.ToUpper(hex.EncodeToString(c.BytesVal))
	}
	return strings.ToUpper(hex.EncodeToString([]byte(c.StringVal)))
}

// GetSubfield retrieves a subfield by ID.
func (c *Component) GetSubfield(subID int) (*Component, bool) {
	if c.Subfields == nil {
		return nil, false
	}
	sub, ok := c.Subfields[subID]
	return sub, ok
}

// SetSubfield sets a nested subfield component.
func (c *Component) SetSubfield(subID int, sub *Component) {
	if c.Subfields == nil {
		c.Subfields = make(map[int]*Component)
	}
	sub.FieldID = subID
	c.Subfields[subID] = sub
}

// Clone creates a deep copy of the component.
func (c *Component) Clone() *Component {
	if c == nil {
		return nil
	}
	clone := &Component{
		FieldID:   c.FieldID,
		StringVal: c.StringVal,
	}
	if c.BytesVal != nil {
		clone.BytesVal = make([]byte, len(c.BytesVal))
		copy(clone.BytesVal, c.BytesVal)
	}
	if c.Subfields != nil {
		clone.Subfields = make(map[int]*Component, len(c.Subfields))
		for k, v := range c.Subfields {
			clone.Subfields[k] = v.Clone()
		}
	}
	clone.TLV = c.TLV.Clone()
	return clone
}

// FieldPackager defines the interface for packing and unpacking a single field.
type FieldPackager interface {
	Pack(c *Component) ([]byte, error)
	Unpack(data []byte, offset int) (*Component, int, error)
	Length() int
	Description() string
}

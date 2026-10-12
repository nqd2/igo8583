package field

import (
	"encoding/hex"
	"strings"

	"github.com/nqd2/igo8583/tlv"
)

/* ============================================================================
 * Component Representation
 * ============================================================================ */

/*
Component represents an individual data element in the ISO 8583 message tree.
Supports scalar string values, binary byte slices, nested composite subfields,
or parsed EMV BER-TLV tag structures.
*/
type Component struct {
	FieldID   int
	StringVal string
	BytesVal  []byte
	Subfields map[int]*Component
	TLV       *tlv.TLVMap
}

func NewStringComponent(id int, val string) *Component {
	return &Component{
		FieldID:   id,
		StringVal: val,
	}
}

func NewBytesComponent(id int, val []byte) *Component {
	return &Component{
		FieldID:  id,
		BytesVal: val,
	}
}

func NewCompositeComponent(id int) *Component {
	return &Component{
		FieldID:   id,
		Subfields: make(map[int]*Component),
	}
}

func NewTLVComponent(id int, tlvMap *tlv.TLVMap) *Component {
	return &Component{
		FieldID: id,
		TLV:     tlvMap,
	}
}

/* ============================================================================
 * Component Accessors & Lifecycle
 * ============================================================================ */

func (c *Component) GetString() string {
	if c.StringVal != "" {
		return c.StringVal
	}
	if len(c.BytesVal) > 0 {
		return string(c.BytesVal)
	}
	return ""
}

func (c *Component) GetBytes() []byte {
	if len(c.BytesVal) > 0 {
		return c.BytesVal
	}
	if c.StringVal != "" {
		return []byte(c.StringVal)
	}
	return nil
}

func (c *Component) GetHex() string {
	if len(c.BytesVal) > 0 {
		return strings.ToUpper(hex.EncodeToString(c.BytesVal))
	}
	return strings.ToUpper(hex.EncodeToString([]byte(c.StringVal)))
}

func (c *Component) GetSubfield(subID int) (*Component, bool) {
	if c.Subfields == nil {
		return nil, false
	}
	sub, ok := c.Subfields[subID]
	return sub, ok
}

func (c *Component) SetSubfield(subID int, sub *Component) {
	if c.Subfields == nil {
		c.Subfields = make(map[int]*Component)
	}
	sub.FieldID = subID
	c.Subfields[subID] = sub
}

/*
Clone creates an independent deep copy of the component, including raw byte slices,
nested subfield maps, and TLV structures.
*/
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

/* ============================================================================
 * FieldPackager Interface
 * ============================================================================ */

/*
FieldPackager defines the strategy for packing and unpacking a single data element.
Implementations must be stateless and immutable for safe concurrent sharing.
*/
type FieldPackager interface {
	Pack(c *Component) ([]byte, error)
	Unpack(data []byte, offset int) (*Component, int, error)
	Length() int
	Description() string
}

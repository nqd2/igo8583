package tlv

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/nqd2/igo8583/internal/diagnostic"
)

var (
	ErrInvalidTag     = errors.New("tlv: invalid tag format")
	ErrBufferTooShort = errors.New("tlv: unexpected end of buffer")
	ErrInvalidLength  = errors.New("tlv: invalid length encoding")
)

// Node represents a single BER-TLV data element.
type Node struct {
	Tag      []byte
	Value    []byte
	Children []*Node
}

// TagHex returns the hex-encoded string of the tag in uppercase (e.g. "9F02", "82").
func (n *Node) TagHex() string {
	return strings.ToUpper(hex.EncodeToString(n.Tag))
}

// ValueHex returns the hex-encoded string of the value in uppercase.
func (n *Node) ValueHex() string {
	return strings.ToUpper(hex.EncodeToString(n.Value))
}

// IsConstructed returns true if the tag represents a constructed TLV object.
func (n *Node) IsConstructed() bool {
	if len(n.Tag) == 0 {
		return false
	}
	return (n.Tag[0] & 0x20) != 0
}

// NewNode creates a primitive TLV node without copying val.
func NewNode(tagHex string, val []byte) (*Node, error) {
	tagBytes, err := hex.DecodeString(tagHex)
	if err != nil || len(tagBytes) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidTag, tagHex)
	}
	return &Node{Tag: tagBytes, Value: val}, nil
}

// Decode parses a raw byte slice into a slice of TLV Nodes.
func Decode(data []byte) ([]*Node, error) {
	return decode(data, 0)
}

func decode(data []byte, depth int) ([]*Node, error) {
	if depth > 16 {
		return nil, fmt.Errorf("%w: nesting exceeds 16 levels", ErrInvalidTag)
	}
	var nodes []*Node
	offset := 0

	for offset < len(data) {
		// Skip padding bytes (0x00 or 0xFF) commonly found in EMV data
		if data[offset] == 0x00 || data[offset] == 0xFF {
			offset++
			continue
		}
		tagStart := offset
		firstByte := data[offset]
		offset++

		if (firstByte & 0x1F) == 0x1F {
			// Multi-byte tag
			for {
				if offset >= len(data) {
					return nil, fmt.Errorf("%w: truncated tag at offset %d", ErrBufferTooShort, tagStart)
				}
				b := data[offset]
				offset++
				if (b & 0x80) == 0 {
					// Last tag byte
					break
				}
			}
		}
		tag := make([]byte, offset-tagStart)
		copy(tag, data[tagStart:offset])
		if offset >= len(data) {
			return nil, fmt.Errorf("%w: missing length at offset %d", ErrBufferTooShort, offset)
		}

		firstLenByte := data[offset]
		offset++
		var length int

		if (firstLenByte & 0x80) == 0 {
			// Single byte length (0..127)
			length = int(firstLenByte)
		} else {
			numLenBytes := int(firstLenByte & 0x7F)
			if numLenBytes == 0 || numLenBytes > 4 {
				return nil, fmt.Errorf("%w: unsupported length byte count %d", ErrInvalidLength, numLenBytes)
			}
			if offset+numLenBytes > len(data) {
				return nil, fmt.Errorf("%w: truncated multi-byte length at %d", ErrBufferTooShort, offset)
			}
			for i := 0; i < numLenBytes; i++ {
				length = (length << 8) | int(data[offset])
				offset++
			}
		}
		if offset+length > len(data) {
			return nil, fmt.Errorf("%w: value truncated (expected %d bytes, got %d) for tag %X",
				ErrBufferTooShort, length, len(data)-offset, tag)
		}
		val := make([]byte, length)
		copy(val, data[offset:offset+length])
		offset += length

		node := &Node{
			Tag:   tag,
			Value: val,
		}

		// If constructed, optionally parse children
		if node.IsConstructed() {
			children, err := decode(val, depth+1)
			if err != nil {
				return nil, fmt.Errorf("tlv: constructed tag %X: %w", tag, err)
			}
			node.Children = children
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// Encode encodes a list of Nodes into wire BER-TLV bytes.
func Encode(nodes []*Node) ([]byte, error) {
	var out []byte
	for _, n := range nodes {
		b, err := EncodeNode(n)
		if err != nil {
			return nil, err
		}
		out = append(out, b...)
	}
	return out, nil
}

// EncodeNode serializes a single TLV Node.
func EncodeNode(n *Node) ([]byte, error) {
	if len(n.Tag) == 0 {
		return nil, ErrInvalidTag
	}

	val := n.Value
	if n.IsConstructed() && len(n.Children) > 0 {
		encodedChildren, err := Encode(n.Children)
		if err != nil {
			return nil, err
		}
		val = encodedChildren
	}

	var buf []byte
	buf = append(buf, n.Tag...)

	// Encode length
	valLen := len(val)
	if valLen <= 0x7F {
		buf = append(buf, byte(valLen))
	} else if valLen <= 0xFF {
		buf = append(buf, 0x81, byte(valLen))
	} else if valLen <= 0xFFFF {
		buf = append(buf, 0x82, byte(valLen>>8), byte(valLen&0xFF))
	} else {
		buf = append(buf, 0x83, byte(valLen>>16), byte((valLen>>8)&0xFF), byte(valLen&0xFF))
	}

	buf = append(buf, val...)
	return buf, nil
}

// TLVMap is a convenient wrapper around TLV nodes for key/value lookups.
type TLVMap struct {
	nodes []*Node
	index map[string]*Node
}

// NewTLVMap creates a TLVMap from a list of nodes.
func NewTLVMap(nodes []*Node) *TLVMap {
	m := &TLVMap{
		nodes: nodes,
		index: make(map[string]*Node),
	}
	for _, n := range nodes {
		m.index[n.TagHex()] = n
	}
	return m
}

// ParseTLV parses raw TLV data and returns a TLVMap.
func ParseTLV(data []byte) (*TLVMap, error) {
	nodes, err := Decode(data)
	if err != nil {
		return nil, err
	}
	return NewTLVMap(nodes), nil
}

// Nodes returns the stored slice and nodes by reference.
func (m *TLVMap) Nodes() []*Node {
	return m.nodes
}

// Clone returns an independent copy, including nested constructed tags.
func (m *TLVMap) Clone() *TLVMap {
	if m == nil {
		return nil
	}
	var cloneNode func(*Node) *Node
	cloneNode = func(n *Node) *Node {
		out := &Node{Tag: append([]byte(nil), n.Tag...), Value: append([]byte(nil), n.Value...)}
		for _, child := range n.Children {
			out.Children = append(out.Children, cloneNode(child))
		}
		return out
	}
	nodes := make([]*Node, 0, len(m.nodes))
	for _, n := range m.nodes {
		nodes = append(nodes, cloneNode(n))
	}
	return NewTLVMap(nodes)
}

// Get returns a tag value by reference.
func (m *TLVMap) Get(tagHex string) ([]byte, bool) {
	tagUpper := strings.ToUpper(tagHex)
	if node, ok := m.index[tagUpper]; ok {
		return node.Value, true
	}
	return nil, false
}

// GetHex returns the hex string value of a tag.
func (m *TLVMap) GetHex(tagHex string) (string, bool) {
	val, ok := m.Get(tagHex)
	if !ok {
		return "", false
	}
	return strings.ToUpper(hex.EncodeToString(val)), true
}

// Set adds or updates a tag without copying val.
func (m *TLVMap) Set(tagHex string, val []byte) error {
	tagUpper := strings.ToUpper(tagHex)
	if node, ok := m.index[tagUpper]; ok {
		node.Value = val
		return nil
	}
	node, err := NewNode(tagUpper, val)
	if err != nil {
		return err
	}
	m.nodes = append(m.nodes, node)
	m.index[tagUpper] = node
	return nil
}

// SetHex adds or updates a tag with a hex string value.
func (m *TLVMap) SetHex(tagHex string, valHex string) error {
	valBytes, err := hex.DecodeString(valHex)
	if err != nil {
		return diagnostic.Redact("tlv: invalid hex value", err)
	}
	return m.Set(tagHex, valBytes)
}

// Encode serializes all tags in the TLVMap.
func (m *TLVMap) Encode() ([]byte, error) {
	return Encode(m.nodes)
}

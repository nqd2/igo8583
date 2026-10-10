package message

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/nqd2/igo8583/bitmap"
	"github.com/nqd2/igo8583/field"
	"github.com/nqd2/igo8583/tlv"
)

// Message represents an in-memory ISO 8583 message.
// Its methods synchronize map access, but components and TLV values can be shared.
// Do not mutate a message or its referenced values while packing or inspecting it.
type Message struct {
	mu     sync.RWMutex
	header []byte
	fields map[int]*field.Component
	bitmap *bitmap.Bitmap
}

// New creates a new empty Message.
func New() *Message {
	return &Message{
		fields: make(map[int]*field.Component),
		bitmap: bitmap.New(),
	}
}

// SetHeader sets the optional ISO header (or TPDU).
func (m *Message) SetHeader(hdr []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.header = make([]byte, len(hdr))
	copy(m.header, hdr)
}

// GetHeader returns the ISO header.
func (m *Message) GetHeader() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.header == nil {
		return nil
	}
	cpy := make([]byte, len(m.header))
	copy(cpy, m.header)
	return cpy
}

// SetMTI sets the Message Type Indicator (field 0).
func (m *Message) SetMTI(mti string) {
	m.Set(0, mti)
}

// GetMTI retrieves the Message Type Indicator (field 0).
func (m *Message) GetMTI() string {
	s, _ := m.GetString(0)
	return s
}

// Set stores a string field into the message and marks the bitmap (if field > 0).
func (m *Message) Set(fieldID int, val string) {
	comp := field.NewStringComponent(fieldID, val)
	m.SetComponent(fieldID, comp)
}

// SetBytes stores a raw byte field without copying val.
func (m *Message) SetBytes(fieldID int, val []byte) {
	comp := field.NewBytesComponent(fieldID, val)
	m.SetComponent(fieldID, comp)
}

// SetTLV stores a TLV map without copying it.
func (m *Message) SetTLV(fieldID int, tlvMap *tlv.TLVMap) {
	comp := field.NewTLVComponent(fieldID, tlvMap)
	m.SetComponent(fieldID, comp)
}

// SetComponent stores comp by reference and updates its FieldID.
// comp must be non-nil.
func (m *Message) SetComponent(fieldID int, comp *field.Component) {
	m.mu.Lock()
	defer m.mu.Unlock()
	comp.FieldID = fieldID
	m.fields[fieldID] = comp
	if fieldID > 0 {
		_ = m.bitmap.Set(fieldID)
	}
}

// Get returns the stored component by reference.
func (m *Message) Get(fieldID int) (*field.Component, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	comp, ok := m.fields[fieldID]
	return comp, ok
}

// GetString retrieves the string value of a field.
func (m *Message) GetString(fieldID int) (string, bool) {
	comp, ok := m.Get(fieldID)
	if !ok {
		return "", false
	}
	return comp.GetString(), true
}

// GetBytes returns field bytes, which may reference stored data.
func (m *Message) GetBytes(fieldID int) ([]byte, bool) {
	comp, ok := m.Get(fieldID)
	if !ok {
		return nil, false
	}
	return comp.GetBytes(), true
}

// GetTLV returns the stored TLV map by reference.
func (m *Message) GetTLV(fieldID int) (*tlv.TLVMap, bool) {
	comp, ok := m.Get(fieldID)
	if !ok || comp.TLV == nil {
		return nil, false
	}
	return comp.TLV, true
}

// Unset removes a field from the message and clears the bitmap.
func (m *Message) Unset(fieldID int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.fields, fieldID)
	if fieldID > 0 {
		_ = m.bitmap.Clear(fieldID)
	}
}

// HasField returns true if the field is present.
func (m *Message) HasField(fieldID int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.fields[fieldID]
	return ok
}

// Bitmap returns a copy of the message's bitmap.
func (m *Message) Bitmap() *bitmap.Bitmap {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bitmap.Clone()
}

// MaxField returns the highest field ID present in the message.
func (m *Message) MaxField() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bitmap.MaxField()
}

// Fields returns a copy of the internal fields map.
func (m *Message) Fields() map[int]*field.Component {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cpy := make(map[int]*field.Component, len(m.fields))
	for k, v := range m.fields {
		cpy[k] = v.Clone()
	}
	return cpy
}

// SetSubfield sets a nested subfield inside a parent composite field.
func (m *Message) SetSubfield(parentID int, subID int, val string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	parent, ok := m.fields[parentID]
	if !ok {
		parent = field.NewCompositeComponent(parentID)
		m.fields[parentID] = parent
		_ = m.bitmap.Set(parentID)
	}
	parent.SetSubfield(subID, field.NewStringComponent(subID, val))
}

// GetSubfield traverses a dot-notated path like "48.1" or "127.2.1".
func (m *Message) GetSubfield(path string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return "", false
	}

	rootID, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", false
	}

	curr, ok := m.fields[rootID]
	if !ok {
		return "", false
	}

	for i := 1; i < len(parts); i++ {
		subID, err := strconv.Atoi(parts[i])
		if err != nil {
			return "", false
		}
		curr, ok = curr.GetSubfield(subID)
		if !ok {
			return "", false
		}
	}

	return curr.GetString(), true
}

// SetSubfieldByPath sets a subfield using a dot-notated path like "48.1".
func (m *Message) SetSubfieldByPath(path string, val string) error {
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return fmt.Errorf("message: path must have at least one subfield (e.g. '48.1'): %s", path)
	}

	rootID, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("message: invalid root field in path %s: %w", path, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	curr, ok := m.fields[rootID]
	if !ok {
		curr = field.NewCompositeComponent(rootID)
		m.fields[rootID] = curr
		_ = m.bitmap.Set(rootID)
	}

	for i := 1; i < len(parts)-1; i++ {
		subID, err := strconv.Atoi(parts[i])
		if err != nil {
			return fmt.Errorf("message: invalid intermediate subfield in path %s: %w", path, err)
		}
		next, ok := curr.GetSubfield(subID)
		if !ok {
			next = field.NewCompositeComponent(subID)
			curr.SetSubfield(subID, next)
		}
		curr = next
	}

	lastID, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return fmt.Errorf("message: invalid target subfield in path %s: %w", path, err)
	}
	curr.SetSubfield(lastID, field.NewStringComponent(lastID, val))
	return nil
}

// Clone creates an independent deep copy of the message.
func (m *Message) Clone() *Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clone := New()
	if m.header != nil {
		clone.header = make([]byte, len(m.header))
		copy(clone.header, m.header)
	}
	clone.bitmap = m.bitmap.Clone()
	for k, v := range m.fields {
		clone.fields[k] = v.Clone()
	}
	return clone
}

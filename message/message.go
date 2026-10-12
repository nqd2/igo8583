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

/*
Message represents an in-memory ISO 8583 financial message.
Internal map and bitmap modifications are synchronized using sync.RWMutex.
Referenced field components and TLV trees should not be mutated concurrently
during packing or inspection.
*/
type Message struct {
	mu     sync.RWMutex
	header []byte
	fields map[int]*field.Component
	bitmap *bitmap.Bitmap
}

/* ============================================================================
 * Constructors & Lifecycle
 * ============================================================================ */

/*
New creates an initialized empty Message with an empty bitmap and fields map.
*/
func New() *Message {
	return &Message{
		fields: make(map[int]*field.Component),
		bitmap: bitmap.New(),
	}
}

/*
Clone returns an independent deep copy of the message, duplicating header bytes,
bitmap state, and all stored field components.
*/
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

/* ============================================================================
 * Header & MTI Operations
 * ============================================================================ */

func (m *Message) SetHeader(hdr []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.header = make([]byte, len(hdr))
	copy(m.header, hdr)
}

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

func (m *Message) SetMTI(mti string) {
	m.Set(0, mti)
}

func (m *Message) GetMTI() string {
	s, _ := m.GetString(0)
	return s
}

/* ============================================================================
 * Data Element Accessors (Typed)
 * ============================================================================ */

func (m *Message) Set(fieldID int, val string) {
	comp := field.NewStringComponent(fieldID, val)
	m.SetComponent(fieldID, comp)
}

func (m *Message) GetString(fieldID int) (string, bool) {
	comp, ok := m.Get(fieldID)
	if !ok {
		return "", false
	}
	return comp.GetString(), true
}

func (m *Message) SetBytes(fieldID int, val []byte) {
	comp := field.NewBytesComponent(fieldID, val)
	m.SetComponent(fieldID, comp)
}

func (m *Message) GetBytes(fieldID int) ([]byte, bool) {
	comp, ok := m.Get(fieldID)
	if !ok {
		return nil, false
	}
	return comp.GetBytes(), true
}

func (m *Message) SetTLV(fieldID int, tlvMap *tlv.TLVMap) {
	comp := field.NewTLVComponent(fieldID, tlvMap)
	m.SetComponent(fieldID, comp)
}

func (m *Message) GetTLV(fieldID int) (*tlv.TLVMap, bool) {
	comp, ok := m.Get(fieldID)
	if !ok || comp.TLV == nil {
		return nil, false
	}
	return comp.TLV, true
}

/* ============================================================================
 * Generic Component & Map Operations
 * ============================================================================ */

/*
SetComponent stores a component reference and marks the field in the bitmap (if fieldID > 0).
*/
func (m *Message) SetComponent(fieldID int, comp *field.Component) {
	if comp == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	comp.FieldID = fieldID
	m.fields[fieldID] = comp
	if fieldID > 0 {
		_ = m.bitmap.Set(fieldID)
	}
}

func (m *Message) Get(fieldID int) (*field.Component, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	comp, ok := m.fields[fieldID]
	return comp, ok
}

func (m *Message) HasField(fieldID int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.fields[fieldID]
	return ok
}

/*
Fields returns a snapshot copy of the internal fields map with cloned components.
Safe for inspection without locking the original message.
*/
func (m *Message) Fields() map[int]*field.Component {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cpy := make(map[int]*field.Component, len(m.fields))
	for k, v := range m.fields {
		cpy[k] = v.Clone()
	}
	return cpy
}

/* ============================================================================
 * Composite & Subfield Traversal
 * ============================================================================ */

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

/*
GetSubfield traverses a dot-notated hierarchy path (e.g. "48.1" or "127.2.1").
Returns the string value and true if found, or empty string and false otherwise.
*/
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

/*
SetSubfieldByPath sets a subfield using a dot-notated path (e.g. "48.1" or "127.2.1").
Automatically initializes intermediate composite components as needed.
*/
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

/* ============================================================================
 * Bitmap Operations
 * ============================================================================ */

func (m *Message) Unset(fieldID int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.fields, fieldID)
	if fieldID > 0 {
		_ = m.bitmap.Clear(fieldID)
	}
}

func (m *Message) Bitmap() *bitmap.Bitmap {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bitmap.Clone()
}

func (m *Message) MaxField() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bitmap.MaxField()
}

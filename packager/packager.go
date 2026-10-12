package packager

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/nqd2/igo8583/bitmap"
	"github.com/nqd2/igo8583/field"
	"github.com/nqd2/igo8583/message"
)

var (
	ErrMissingMTI = errors.New("packager: message is missing MTI (field 0)")
)

/* ============================================================================
 * Interfaces & Compiled Codec Snapshot
 * ============================================================================ */

/*
Packager defines the high-level contract for serializing and deserializing
complete ISO 8583 messages to and from wire bytes.
*/
type Packager interface {
	Pack(msg *message.Message) ([]byte, error)
	Unpack(data []byte) (*message.Message, error)
}

/*
Codec represents an immutable, compiled snapshot of an ISO 8583 packager configuration.
It contains no mutator methods, guaranteeing safe concurrent use across multiple goroutines
as long as registered custom field packagers are themselves immutable.
*/
type Codec struct {
	p *BasePackager
}

func (c *Codec) Pack(msg *message.Message) ([]byte, error)    { return c.p.Pack(msg) }
func (c *Codec) Unpack(data []byte) (*message.Message, error) { return c.p.Unpack(data) }
func (c *Codec) HeaderLength() int                            { return c.p.HeaderLength() }
func (c *Codec) BitmapFormat() string                         { return c.p.BitmapFormat() }

/*
PackField encodes a single data element using this codec's exact field packager definition.
Convenient for cryptographic protocols (such as MAC calculation or PIN block encryption)
that require exact wire representation of selected fields without full message framing.
*/
func (c *Codec) PackField(fieldID int, comp *field.Component) ([]byte, error) {
	if fieldID <= 0 {
		return nil, fmt.Errorf("packager: field ID must be positive")
	}
	if comp == nil {
		return nil, fmt.Errorf("packager: nil component for field %d", fieldID)
	}
	c.p.mu.RLock()
	defer c.p.mu.RUnlock()
	fp, ok := c.p.fields[fieldID]
	if !ok {
		return nil, fmt.Errorf("packager: no field packager defined for field %d", fieldID)
	}
	return fp.Pack(comp)
}

/* ============================================================================
 * BasePackager Configuration & Compilation
 * ============================================================================ */

/*
BasePackager implements a configurable ISO 8583 message packager.
Field packagers, framing lengths, and bitmap formats are configured here
before calling Compile to create an immutable Codec.
*/
type BasePackager struct {
	mu           sync.RWMutex
	headerLength int
	lengthPrefix int
	bitmapFormat string
	fields       map[int]field.FieldPackager
}

func NewBasePackager() *BasePackager {
	return &BasePackager{
		bitmapFormat: "hex",
		fields:       make(map[int]field.FieldPackager),
	}
}

/*
Compile validates the packager configuration and returns an immutable, concurrency-safe Codec.
Validates:
  - Header and length prefix constraints (0..9 digits)
  - Bitmap format ("hex" or "binary")
  - Mandatory MTI (field 0) presence and fixed 4-character length
  - Individual field packager validation rules
*/
func (p *BasePackager) Compile() (*Codec, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.headerLength < 0 || p.lengthPrefix < 0 || p.lengthPrefix > 9 {
		return nil, fmt.Errorf("packager: invalid header or length prefix width")
	}
	if p.bitmapFormat != "hex" && p.bitmapFormat != "binary" {
		return nil, fmt.Errorf("packager: unsupported bitmap format %q", p.bitmapFormat)
	}
	if _, ok := p.fields[0]; !ok {
		return nil, fmt.Errorf("packager: field 0 (MTI) is required")
	}
	if p.fields[0] == nil || p.fields[0].Length() != 4 {
		return nil, fmt.Errorf("packager: MTI must contain four characters/digits")
	}
	copyP := NewBasePackager()
	copyP.headerLength = p.headerLength
	copyP.lengthPrefix = p.lengthPrefix
	copyP.bitmapFormat = p.bitmapFormat
	for id, fp := range p.fields {
		if id < 0 || id > 192 || id == 1 || id == 65 || fp == nil || fp.Length() <= 0 {
			return nil, fmt.Errorf("packager: invalid field definition %d", id)
		}
		if v, ok := fp.(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return nil, fmt.Errorf("packager: field %d: %w", id, err)
			}
		}
		copyP.fields[id] = fp
	}
	return &Codec{p: copyP}, nil
}

/*
MustCompile compiles the BasePackager and panics on error.
Intended for package-level initialization of verified presets.
*/
func (p *BasePackager) MustCompile() *Codec {
	c, err := p.Compile()
	if err != nil {
		panic(err)
	}
	return c
}

func (p *BasePackager) SetHeaderLength(len int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.headerLength = len
}

func (p *BasePackager) SetLengthPrefixASCII(width int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lengthPrefix = width
}

func (p *BasePackager) HeaderLength() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.headerLength
}

func (p *BasePackager) SetBitmapFormat(format string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.bitmapFormat = format
}

func (p *BasePackager) BitmapFormat() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.bitmapFormat
}

func (p *BasePackager) SetFieldPackager(fieldID int, fp field.FieldPackager) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.fields[fieldID] = fp
}

func (p *BasePackager) GetFieldPackager(fieldID int) (field.FieldPackager, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	fp, ok := p.fields[fieldID]
	return fp, ok
}

/* ============================================================================
 * Wire Serialization (Pack & Unpack)
 * ============================================================================ */

/*
Pack serializes an in-memory Message into wire bytes in standard ISO 8583 sequence:
 1. Header (if configured)
 2. MTI (field 0)
 3. Primary and secondary/tertiary bitmaps
 4. Data elements in ascending numeric order
 5. Transport length prefix (if configured)
*/
func (p *BasePackager) Pack(msg *message.Message) ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if msg == nil {
		return nil, fmt.Errorf("packager: nil message")
	}
	var out []byte
	msgHdr := msg.GetHeader()
	if len(msgHdr) != p.headerLength {
		return nil, fmt.Errorf("packager: header length %d, expected %d", len(msgHdr), p.headerLength)
	}
	if len(msgHdr) > 0 {
		out = append(out, msgHdr...)
	}
	mtiComp, ok := msg.Get(0)
	if !ok {
		return nil, &PackError{FieldID: 0, Err: ErrMissingMTI}
	}
	fp0, ok := p.fields[0]
	if !ok {
		return nil, &PackError{FieldID: 0, Err: fmt.Errorf("packager: no packager defined for field 0 (MTI)")}
	}
	mtiBytes, err := fp0.Pack(mtiComp)
	if err != nil {
		return nil, &PackError{FieldID: 0, Err: err}
	}
	out = append(out, mtiBytes...)
	bm := bitmap.New()
	msgFields := msg.Fields()
	for fID := range msgFields {
		if fID < 0 {
			return nil, &PackError{FieldID: fID, Err: fmt.Errorf("invalid data field ID")}
		}
		if fID > 0 {
			if fID == 1 || fID == 65 || fID > 192 {
				return nil, &PackError{FieldID: fID, Err: fmt.Errorf("invalid data field ID")}
			}
			if _, ok := p.fields[fID]; !ok {
				return nil, &PackError{FieldID: fID, Err: fmt.Errorf("no field packager defined")}
			}
			_ = bm.Set(fID)
		}
	}
	bm.PrepareIndicators()

	var bmBytes []byte
	if p.bitmapFormat == "binary" {
		bmBytes = bm.PackBinary()
	} else {
		bmBytes = bm.PackHex()
	}
	out = append(out, bmBytes...)
	maxField := bm.MaxField()
	for i := 2; i <= maxField; i++ {
		if i == 65 && maxField > 128 {
			continue
		}

		if bm.IsSet(i) {
			comp, hasComp := msgFields[i]
			if !hasComp {
				continue
			}
			fp, hasFP := p.fields[i]
			if !hasFP {
				return nil, &PackError{FieldID: i, Err: fmt.Errorf("packager: no field packager defined for field %d", i)}
			}
			packedField, err := fp.Pack(comp)
			if err != nil {
				return nil, &PackError{FieldID: i, Err: err}
			}
			out = append(out, packedField...)
		}
	}

	if p.lengthPrefix > 0 {
		s := strconv.Itoa(len(out))
		if len(s) > p.lengthPrefix {
			return nil, fmt.Errorf("packager: framed message exceeds %d-digit length prefix", p.lengthPrefix)
		}
		prefix := fmt.Sprintf("%0*d", p.lengthPrefix, len(out))
		out = append([]byte(prefix), out...)
	}
	return out, nil
}

/*
Unpack parses wire bytes into an in-memory Message.
Strictly validates:
  - Transport length prefix count matching actual body size
  - Exact header length matching configuration
  - Valid MTI bytes
  - Active fields present in bitmaps have corresponding packagers
  - Rejection of any trailing unconsumed bytes
*/
func (p *BasePackager) Unpack(data []byte) (*message.Message, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	msg := message.New()
	offset := 0
	if p.lengthPrefix > 0 {
		if len(data) < p.lengthPrefix {
			return nil, NewUnpackError(-1, 0, data, fmt.Errorf("short length prefix"))
		}
		for _, b := range data[:p.lengthPrefix] {
			if b < '0' || b > '9' {
				return nil, NewUnpackError(-1, 0, data, fmt.Errorf("non-decimal length prefix"))
			}
		}
		declared, _ := strconv.Atoi(string(data[:p.lengthPrefix]))
		if declared != len(data)-p.lengthPrefix {
			return nil, NewUnpackError(-1, 0, data, fmt.Errorf("frame length %d does not match %d body bytes", declared, len(data)-p.lengthPrefix))
		}
		offset = p.lengthPrefix
	}
	if p.headerLength > 0 {
		if len(data)-offset < p.headerLength {
			return nil, NewUnpackError(-1, offset, data[offset:], fmt.Errorf("buffer too short for header (need %d, have %d)", p.headerLength, len(data)-offset))
		}
		msg.SetHeader(data[offset : offset+p.headerLength])
		offset += p.headerLength
	}
	fp0, ok := p.fields[0]
	if !ok {
		return nil, NewUnpackError(0, offset, data[offset:], fmt.Errorf("packager: no field packager defined for field 0 (MTI)"))
	}
	mtiComp, consumedMTI, err := fp0.Unpack(data, offset)
	if err != nil {
		return nil, NewUnpackError(0, offset, data[offset:], err)
	}
	mtiComp.FieldID = 0
	msg.SetComponent(0, mtiComp)
	offset += consumedMTI
	bm := bitmap.New()
	var consumedBM int
	if p.bitmapFormat == "binary" {
		consumedBM, err = bm.UnpackBinary(data[offset:])
	} else {
		consumedBM, err = bm.UnpackHex(data[offset:])
	}
	if err != nil {
		return nil, NewUnpackError(1, offset, data[offset:], err)
	}
	offset += consumedBM
	maxField := bm.MaxField()
	for i := 2; i <= maxField; i++ {
		if i == 65 && maxField > 128 {
			continue
		}

		if bm.IsSet(i) {
			fp, ok := p.fields[i]
			if !ok {
				return nil, NewUnpackError(i, offset, data[offset:], fmt.Errorf("packager: no field packager defined for active field %d", i))
			}

			comp, consumedField, err := fp.Unpack(data, offset)
			if err != nil {
				return nil, NewUnpackError(i, offset, data[offset:], err)
			}
			comp.FieldID = i
			msg.SetComponent(i, comp)
			offset += consumedField
		}
	}

	if offset != len(data) {
		return nil, NewUnpackError(-1, offset, data[offset:], fmt.Errorf("%d trailing bytes", len(data)-offset))
	}
	return msg, nil
}

package spec

import (
	"github.com/nqd2/igo8583/field"
	"github.com/nqd2/igo8583/interpreter"
	"github.com/nqd2/igo8583/packager"
	"github.com/nqd2/igo8583/padder"
	"github.com/nqd2/igo8583/prefixer"
)

// Builder provides a fluent API for programmatically constructing an ISO-8583 BasePackager.
type Builder struct {
	pkg *packager.BasePackager
}

// NewBuilder creates a packager builder. The name is descriptive and is not retained in the codec.
func NewBuilder(name string) *Builder {
	return &Builder{
		pkg: packager.NewBasePackager(),
	}
}

// WithBitmapFormat sets the bitmap encoding format ("hex" or "binary").
func (b *Builder) WithBitmapFormat(format string) *Builder {
	b.pkg.SetBitmapFormat(format)
	return b
}

// WithHeaderLength sets the fixed header length.
func (b *Builder) WithHeaderLength(len int) *Builder {
	b.pkg.SetHeaderLength(len)
	return b
}

// WithASCIILengthPrefix frames each message with a width-digit body byte count.
func (b *Builder) WithASCIILengthPrefix(width int) *Builder {
	b.pkg.SetLengthPrefixASCII(width)
	return b
}

// Field registers a generic FieldPackager for a field ID.
func (b *Builder) Field(id int, fp field.FieldPackager) *Builder {
	b.pkg.SetFieldPackager(id, fp)
	return b
}

// StringField registers a string field packager.
func (b *Builder) StringField(
	id int,
	length int,
	description string,
	p padder.Padder,
	i interpreter.Interpreter,
	pref prefixer.Prefixer,
) *Builder {
	fp := field.NewStringFieldPackager(length, description, p, i, pref)
	b.pkg.SetFieldPackager(id, fp)
	return b
}

// BinaryField registers a raw binary field packager.
func (b *Builder) BinaryField(id int, length int, description string, pref prefixer.Prefixer) *Builder {
	fp := field.NewBinaryFieldPackager(length, description, pref)
	b.pkg.SetFieldPackager(id, fp)
	return b
}

// CompositeField registers a composite field packager with nested subfields.
func (b *Builder) CompositeField(
	id int,
	length int,
	description string,
	pref prefixer.Prefixer,
	subfields map[int]field.FieldPackager,
) *Builder {
	fp := field.NewCompositeFieldPackager(length, description, pref, subfields)
	b.pkg.SetFieldPackager(id, fp)
	return b
}

// TLVField registers a BER-TLV field packager.
func (b *Builder) TLVField(id int, length int, description string, pref prefixer.Prefixer) *Builder {
	fp := field.NewTLVFieldPackager(length, description, pref)
	b.pkg.SetFieldPackager(id, fp)
	return b
}

// Build finalizes and returns the compiled codec.
func (b *Builder) Build() (*packager.Codec, error) {
	return b.pkg.Compile()
}

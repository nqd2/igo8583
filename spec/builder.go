package spec

import (
	"github.com/nqd2/igo8583/field"
	"github.com/nqd2/igo8583/interpreter"
	"github.com/nqd2/igo8583/packager"
	"github.com/nqd2/igo8583/padder"
	"github.com/nqd2/igo8583/prefixer"
)

/*
Builder provides a fluent builder API for programmatically constructing an ISO 8583 packager.
*/
type Builder struct {
	pkg *packager.BasePackager
}

/* ============================================================================
 * Builder Initialization & Framing Options
 * ============================================================================ */

/*
NewBuilder creates a fluent packager builder instance.
The name parameter is descriptive documentation.
*/
func NewBuilder(name string) *Builder {
	return &Builder{
		pkg: packager.NewBasePackager(),
	}
}

func (b *Builder) WithBitmapFormat(format string) *Builder {
	b.pkg.SetBitmapFormat(format)
	return b
}

func (b *Builder) WithHeaderLength(len int) *Builder {
	b.pkg.SetHeaderLength(len)
	return b
}

func (b *Builder) WithASCIILengthPrefix(width int) *Builder {
	b.pkg.SetLengthPrefixASCII(width)
	return b
}

/* ============================================================================
 * Field Registration Methods
 * ============================================================================ */

func (b *Builder) Field(id int, fp field.FieldPackager) *Builder {
	b.pkg.SetFieldPackager(id, fp)
	return b
}

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

func (b *Builder) BinaryField(id int, length int, description string, pref prefixer.Prefixer) *Builder {
	fp := field.NewBinaryFieldPackager(length, description, pref)
	b.pkg.SetFieldPackager(id, fp)
	return b
}

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

func (b *Builder) TLVField(id int, length int, description string, pref prefixer.Prefixer) *Builder {
	fp := field.NewTLVFieldPackager(length, description, pref)
	b.pkg.SetFieldPackager(id, fp)
	return b
}

/* ============================================================================
 * Codec Compilation
 * ============================================================================ */

/*
Build validates all configured constraints and returns a compiled immutable Codec.
*/
func (b *Builder) Build() (*packager.Codec, error) {
	return b.pkg.Compile()
}

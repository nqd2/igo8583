package specs

import (
	"github.com/nqd2/igo8583/field"
	"github.com/nqd2/igo8583/interpreter"
	"github.com/nqd2/igo8583/packager"
	"github.com/nqd2/igo8583/padder"
	"github.com/nqd2/igo8583/prefixer"
)

func newISO87(binary bool) *packager.Codec {
	p := packager.NewBasePackager()
	bitmapFormat := "hex"
	mtiInterpreter := interpreter.ASCII()
	if binary {
		bitmapFormat = "binary"
		mtiInterpreter = interpreter.BCDLeftPadded()
	}
	p.SetBitmapFormat(bitmapFormat)
	p.SetFieldPackager(0, field.NewStringFieldPackager(4, "Message type indicator", padder.NoPadding, mtiInterpreter, prefixer.Fixed(4)))

	for _, def := range iso87Fields {
		lengthPrefix := iso87Prefixer(def, binary)
		if def.id == 28 || def.id == 29 || def.id == 30 || def.id == 31 || def.id == 97 {
			p.SetFieldPackager(def.id, field.NewAmountFieldPackager(def.length, def.name, binary))
			continue
		}
		if def.id == 55 && binary {
			p.SetFieldPackager(def.id, field.NewBinaryFieldPackager(def.length, def.name, lengthPrefix))
			continue
		}
		if def.kind == 'b' {
			if binary {
				p.SetFieldPackager(def.id, field.NewBinaryFieldPackager(def.length, def.name, lengthPrefix))
			} else {
				p.SetFieldPackager(def.id, field.NewASCIIBinaryFieldPackager(def.length, def.name, lengthPrefix))
			}
			continue
		}

		valueInterpreter := interpreter.ASCII()
		valuePadder := padder.NoPadding
		switch def.kind {
		case 'n':
			if binary {
				if def.variable == 0 {
					valueInterpreter = interpreter.BCDLeftPadded()
				} else {
					valueInterpreter = interpreter.BCDRightPadded()
				}
			}
			if def.variable == 0 {
				valuePadder = padder.LeftZero
			}
		case 'z':
			if binary {
				valueInterpreter = interpreter.HexRightPadded()
			}
		case 's':
			if def.variable == 0 {
				valuePadder = padder.RightSpace
			}
		}
		p.SetFieldPackager(def.id, field.NewStringFieldPackager(def.length, def.name, valuePadder, valueInterpreter, lengthPrefix))
	}
	return p.MustCompile()
}

func iso87Prefixer(def iso87FieldDef, binary bool) prefixer.Prefixer {
	switch def.variable {
	case 2:
		if binary {
			return prefixer.BcdLL
		}
		return prefixer.AsciiLL
	case 3:
		if binary {
			return prefixer.BcdLLL
		}
		return prefixer.AsciiLLL
	default:
		return prefixer.Fixed(def.length)
	}
}

package specs

import (
	"github.com/nqd2/igo8583/field"
	"github.com/nqd2/igo8583/interpreter"
	"github.com/nqd2/igo8583/packager"
	"github.com/nqd2/igo8583/padder"
	"github.com/nqd2/igo8583/prefixer"
)

/* ISO93Binary creates the ISO 8583:1993 binary codec. */
func ISO93Binary() *packager.Codec {
	p := packager.NewBasePackager()
	p.SetBitmapFormat("binary")

	// Field 0: MTI (4 BCD digits -> 2 bytes)
	p.SetFieldPackager(0, field.NewStringFieldPackager(4, "MESSAGE TYPE INDICATOR", padder.LeftZero, interpreter.BCDLeftPadded(), prefixer.Fixed(4)))
	// Field 2: PAN (BCD LLVAR)
	p.SetFieldPackager(2, field.NewStringFieldPackager(19, "PRIMARY ACCOUNT NUMBER", padder.NoPadding, interpreter.BCDLeftPadded(), prefixer.BcdLL))
	// Field 3: Processing Code (6 BCD digits -> 3 bytes)
	p.SetFieldPackager(3, field.NewStringFieldPackager(6, "PROCESSING CODE", padder.LeftZero, interpreter.BCDLeftPadded(), prefixer.Fixed(6)))
	// Field 4: Amount, Transaction (12 BCD digits -> 6 bytes)
	p.SetFieldPackager(4, field.NewStringFieldPackager(12, "AMOUNT, TRANSACTION", padder.LeftZero, interpreter.BCDLeftPadded(), prefixer.Fixed(12)))
	// Field 7: Transmission Date & Time (10 BCD digits -> 5 bytes)
	p.SetFieldPackager(7, field.NewStringFieldPackager(10, "TRANSMISSION DATE AND TIME", padder.LeftZero, interpreter.BCDLeftPadded(), prefixer.Fixed(10)))
	// Field 11: STAN (6 BCD digits -> 3 bytes)
	p.SetFieldPackager(11, field.NewStringFieldPackager(6, "SYSTEM TRACE AUDIT NUMBER", padder.LeftZero, interpreter.BCDLeftPadded(), prefixer.Fixed(6)))
	// Field 12: Date/Time Local (12 BCD digits -> 6 bytes)
	p.SetFieldPackager(12, field.NewStringFieldPackager(12, "DATE/TIME LOCAL TRANSACTION", padder.LeftZero, interpreter.BCDLeftPadded(), prefixer.Fixed(12)))
	// Field 22: POS Data Code (12 chars ASCII)
	p.SetFieldPackager(22, field.NewStringFieldPackager(12, "POS DATA CODE", padder.NoPadding, interpreter.ASCII(), prefixer.Fixed(12)))
	// Field 24: Function Code (3 BCD digits)
	p.SetFieldPackager(24, field.NewStringFieldPackager(3, "FUNCTION CODE", padder.LeftZero, interpreter.BCDLeftPadded(), prefixer.Fixed(3)))
	// Field 25: Message Reason Code (4 BCD digits)
	p.SetFieldPackager(25, field.NewStringFieldPackager(4, "MESSAGE REASON CODE", padder.LeftZero, interpreter.BCDLeftPadded(), prefixer.Fixed(4)))
	// Field 32: Acquiring Inst ID Code (BCD LLVAR)
	p.SetFieldPackager(32, field.NewStringFieldPackager(11, "ACQUIRING INST ID CODE", padder.NoPadding, interpreter.BCDLeftPadded(), prefixer.BcdLL))
	// Field 35: Track 2 Data (BCD LLVAR)
	p.SetFieldPackager(35, field.NewStringFieldPackager(37, "TRACK 2 DATA", padder.NoPadding, interpreter.Hex(), prefixer.BcdLL))
	// Field 37: RRN (12 chars ASCII)
	p.SetFieldPackager(37, field.NewStringFieldPackager(12, "RETRIEVAL REFERENCE NUMBER", padder.NoPadding, interpreter.ASCII(), prefixer.Fixed(12)))
	// Field 38: Approval Code (6 chars ASCII)
	p.SetFieldPackager(38, field.NewStringFieldPackager(6, "APPROVAL CODE", padder.NoPadding, interpreter.ASCII(), prefixer.Fixed(6)))
	// Field 39: Action Code (3 BCD digits)
	p.SetFieldPackager(39, field.NewStringFieldPackager(3, "ACTION CODE", padder.LeftZero, interpreter.BCDLeftPadded(), prefixer.Fixed(3)))
	// Field 41: Terminal ID (8 chars ASCII)
	p.SetFieldPackager(41, field.NewStringFieldPackager(8, "TERMINAL ID", padder.RightSpace, interpreter.ASCII(), prefixer.Fixed(8)))
	// Field 42: Card Acceptor ID Code (15 chars ASCII)
	p.SetFieldPackager(42, field.NewStringFieldPackager(15, "CARD ACCEPTOR ID CODE", padder.RightSpace, interpreter.ASCII(), prefixer.Fixed(15)))
	// Field 48: Private Additional Data (BCD LLLVAR)
	p.SetFieldPackager(48, field.NewBinaryFieldPackager(999, "PRIVATE ADDITIONAL DATA", prefixer.BcdLLL))
	// Field 52: PIN Data (8 bytes binary)
	p.SetFieldPackager(52, field.NewBinaryFieldPackager(8, "PIN DATA", prefixer.Fixed(8)))
	// Field 55: EMV Chip Data (BCD LLLVAR binary)
	p.SetFieldPackager(55, field.NewTLVFieldPackager(999, "EMV CHIP DATA", prefixer.BcdLLL))
	// Field 70: Network Management Info Code (3 BCD digits)
	p.SetFieldPackager(70, field.NewStringFieldPackager(3, "NETWORK MANAGEMENT INFO CODE", padder.LeftZero, interpreter.BCDLeftPadded(), prefixer.Fixed(3)))
	// Field 128: MAC (8 bytes binary)
	p.SetFieldPackager(128, field.NewBinaryFieldPackager(8, "MAC", prefixer.Fixed(8)))

	return p.MustCompile()
}

package specs

import (
	"github.com/nqd2/igo8583/field"
	"github.com/nqd2/igo8583/interpreter"
	"github.com/nqd2/igo8583/packager"
	"github.com/nqd2/igo8583/padder"
	"github.com/nqd2/igo8583/prefixer"
)

/* ISO93ASCII creates the ISO 8583:1993 ASCII codec. */
func ISO93ASCII() *packager.Codec {
	p := packager.NewBasePackager()
	p.SetBitmapFormat("hex")

	// Field 0: MTI
	p.SetFieldPackager(0, field.NewStringFieldPackager(4, "MESSAGE TYPE INDICATOR", padder.NoPadding, interpreter.ASCII(), prefixer.Fixed(4)))
	// Field 2: PAN
	p.SetFieldPackager(2, field.NewStringFieldPackager(19, "PRIMARY ACCOUNT NUMBER", padder.NoPadding, interpreter.ASCII(), prefixer.AsciiLL))
	// Field 3: Processing Code
	p.SetFieldPackager(3, field.NewStringFieldPackager(6, "PROCESSING CODE", padder.LeftZero, interpreter.ASCII(), prefixer.Fixed(6)))
	// Field 4: Amount, Transaction
	p.SetFieldPackager(4, field.NewStringFieldPackager(12, "AMOUNT, TRANSACTION", padder.LeftZero, interpreter.ASCII(), prefixer.Fixed(12)))
	// Field 7: Transmission Date & Time
	p.SetFieldPackager(7, field.NewStringFieldPackager(10, "TRANSMISSION DATE AND TIME", padder.LeftZero, interpreter.ASCII(), prefixer.Fixed(10)))
	// Field 11: STAN
	p.SetFieldPackager(11, field.NewStringFieldPackager(6, "SYSTEM TRACE AUDIT NUMBER", padder.LeftZero, interpreter.ASCII(), prefixer.Fixed(6)))
	// Field 12: Date and Time, Local Transaction (12 chars YYMMDDhhmmss)
	p.SetFieldPackager(12, field.NewStringFieldPackager(12, "DATE/TIME LOCAL TRANSACTION", padder.LeftZero, interpreter.ASCII(), prefixer.Fixed(12)))
	// Field 22: Point of Service Data Code
	p.SetFieldPackager(22, field.NewStringFieldPackager(12, "POS DATA CODE", padder.NoPadding, interpreter.ASCII(), prefixer.Fixed(12)))
	// Field 24: Function Code (ISO93 specific)
	p.SetFieldPackager(24, field.NewStringFieldPackager(3, "FUNCTION CODE", padder.LeftZero, interpreter.ASCII(), prefixer.Fixed(3)))
	// Field 25: Message Reason Code (ISO93 specific)
	p.SetFieldPackager(25, field.NewStringFieldPackager(4, "MESSAGE REASON CODE", padder.LeftZero, interpreter.ASCII(), prefixer.Fixed(4)))
	// Field 32: Acquiring Inst ID Code
	p.SetFieldPackager(32, field.NewStringFieldPackager(11, "ACQUIRING INST ID CODE", padder.NoPadding, interpreter.ASCII(), prefixer.AsciiLL))
	// Field 35: Track 2 Data
	p.SetFieldPackager(35, field.NewStringFieldPackager(37, "TRACK 2 DATA", padder.NoPadding, interpreter.ASCII(), prefixer.AsciiLL))
	// Field 37: Retrieval Reference Number
	p.SetFieldPackager(37, field.NewStringFieldPackager(12, "RETRIEVAL REFERENCE NUMBER", padder.NoPadding, interpreter.ASCII(), prefixer.Fixed(12)))
	// Field 38: Approval Code
	p.SetFieldPackager(38, field.NewStringFieldPackager(6, "APPROVAL CODE", padder.NoPadding, interpreter.ASCII(), prefixer.Fixed(6)))
	// Field 39: Action Code (3 digits in ISO 8583:1993)
	p.SetFieldPackager(39, field.NewStringFieldPackager(3, "ACTION CODE", padder.LeftZero, interpreter.ASCII(), prefixer.Fixed(3)))
	// Field 41: Terminal ID
	p.SetFieldPackager(41, field.NewStringFieldPackager(8, "TERMINAL ID", padder.RightSpace, interpreter.ASCII(), prefixer.Fixed(8)))
	// Field 42: Card Acceptor ID Code
	p.SetFieldPackager(42, field.NewStringFieldPackager(15, "CARD ACCEPTOR ID CODE", padder.RightSpace, interpreter.ASCII(), prefixer.Fixed(15)))
	// Field 48: Private Additional Data
	p.SetFieldPackager(48, field.NewStringFieldPackager(999, "PRIVATE ADDITIONAL DATA", padder.NoPadding, interpreter.ASCII(), prefixer.AsciiLLL))
	// Field 49: Currency Code, Transaction
	p.SetFieldPackager(49, field.NewStringFieldPackager(3, "CURRENCY CODE, TRANSACTION", padder.LeftZero, interpreter.ASCII(), prefixer.Fixed(3)))
	// Field 52: PIN Data
	p.SetFieldPackager(52, field.NewBinaryFieldPackager(8, "PIN DATA", prefixer.Fixed(8)))
	// Field 55: EMV Chip Data
	p.SetFieldPackager(55, field.NewTLVFieldPackager(999, "EMV CHIP DATA", prefixer.AsciiLLL))
	// Field 70: Network Management Info Code
	p.SetFieldPackager(70, field.NewStringFieldPackager(3, "NETWORK MANAGEMENT INFO CODE", padder.LeftZero, interpreter.ASCII(), prefixer.Fixed(3)))
	// Field 128: MAC
	p.SetFieldPackager(128, field.NewBinaryFieldPackager(8, "MAC", prefixer.Fixed(8)))

	return p.MustCompile()
}

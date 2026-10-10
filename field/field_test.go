package field

import (
	"bytes"
	"testing"

	"github.com/nqd2/igo8583/interpreter"
	"github.com/nqd2/igo8583/padder"
	"github.com/nqd2/igo8583/prefixer"
	"github.com/nqd2/igo8583/tlv"
)

func TestStringFieldPackagerFixed(t *testing.T) {
	// Processing Code: Fixed 6 digits, ASCII, Left-padded with '0'
	fp := NewStringFieldPackager(6, "Processing Code", padder.LeftZero, interpreter.ASCII(), prefixer.Fixed(6))

	comp := NewStringComponent(3, "123")
	packed, err := fp.Pack(comp)
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}
	if string(packed) != "000123" {
		t.Fatalf("expected '000123', got '%s'", string(packed))
	}

	unpackedComp, consumed, err := fp.Unpack(packed, 0)
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}
	if consumed != 6 || unpackedComp.GetString() != "000123" {
		t.Fatalf("unexpected unpack: consumed %d, val '%s'", consumed, unpackedComp.GetString())
	}
}

func TestStringFieldPackagerLLVAR(t *testing.T) {
	// PAN: LLVAR, ASCII prefix, ASCII digits, max 19
	fp := NewStringFieldPackager(19, "PAN", padder.NoPadding, interpreter.ASCII(), prefixer.AsciiLL)

	comp := NewStringComponent(2, "4111111111111111")
	packed, err := fp.Pack(comp)
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}
	// Prefix is "16", value is "4111111111111111"
	if string(packed) != "164111111111111111" {
		t.Fatalf("expected '164111111111111111', got '%s'", string(packed))
	}

	unpackedComp, consumed, err := fp.Unpack(packed, 0)
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}
	if consumed != 18 || unpackedComp.GetString() != "4111111111111111" {
		t.Fatalf("unexpected unpack: %d, %s", consumed, unpackedComp.GetString())
	}
}

func TestBinaryFieldPackager(t *testing.T) {
	fp := NewBinaryFieldPackager(8, "MAC", prefixer.Fixed(8))
	raw := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	comp := NewBytesComponent(128, raw)

	packed, err := fp.Pack(comp)
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}
	if !bytes.Equal(packed, raw) {
		t.Fatalf("packed binary mismatch")
	}

	unpackedComp, consumed, err := fp.Unpack(packed, 0)
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}
	if consumed != 8 || !bytes.Equal(unpackedComp.GetBytes(), raw) {
		t.Fatalf("unpacked binary mismatch")
	}
}

func TestCompositeFieldPackager(t *testing.T) {
	// DE 48 containing subfields:
	// subfield 1: fixed 2 chars (e.g. "AB")
	// subfield 2: fixed 4 chars (e.g. "1234")
	sub1 := NewStringFieldPackager(2, "Sub 1", padder.NoPadding, interpreter.ASCII(), prefixer.Fixed(2))
	sub2 := NewStringFieldPackager(4, "Sub 2", padder.NoPadding, interpreter.ASCII(), prefixer.Fixed(4))

	subs := map[int]FieldPackager{
		1: sub1,
		2: sub2,
	}

	compositeFP := NewCompositeFieldPackager(999, "Field 48", prefixer.AsciiLLL, subs)

	comp := NewCompositeComponent(48)
	comp.SetSubfield(1, NewStringComponent(1, "AB"))
	comp.SetSubfield(2, NewStringComponent(2, "1234"))

	packed, err := compositeFP.Pack(comp)
	if err != nil {
		t.Fatalf("composite pack failed: %v", err)
	}
	// Prefix = "006" (3 chars), body = "AB1234" (6 chars)
	if string(packed) != "006AB1234" {
		t.Fatalf("expected '006AB1234', got '%s'", string(packed))
	}

	unpackedComp, consumed, err := compositeFP.Unpack(packed, 0)
	if err != nil {
		t.Fatalf("composite unpack failed: %v", err)
	}
	if consumed != 9 {
		t.Fatalf("expected consumed 9, got %d", consumed)
	}

	s1, ok1 := unpackedComp.GetSubfield(1)
	s2, ok2 := unpackedComp.GetSubfield(2)
	if !ok1 || s1.GetString() != "AB" || !ok2 || s2.GetString() != "1234" {
		t.Fatalf("subfields mismatch: %+v, %+v", s1, s2)
	}
}

func TestTLVFieldPackager(t *testing.T) {
	tlvFP := NewTLVFieldPackager(999, "DE 55 Chip Data", prefixer.AsciiLLL)

	tlvMap := tlv.NewTLVMap(nil)
	_ = tlvMap.SetHex("9F02", "000000001000")

	comp := NewTLVComponent(55, tlvMap)
	packed, err := tlvFP.Pack(comp)
	if err != nil {
		t.Fatalf("tlv pack failed: %v", err)
	}

	unpackedComp, consumed, err := tlvFP.Unpack(packed, 0)
	if err != nil {
		t.Fatalf("tlv unpack failed: %v", err)
	}
	if consumed != len(packed) {
		t.Fatalf("consumed %d != len %d", consumed, len(packed))
	}

	valHex, ok := unpackedComp.TLV.GetHex("9F02")
	if !ok || valHex != "000000001000" {
		t.Fatalf("tlv value mismatch: %s", valHex)
	}
}

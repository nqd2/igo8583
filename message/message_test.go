package message

import (
	"strings"
	"testing"

	"github.com/nqd2/igo8583/tlv"
)

func TestMessageBasic(t *testing.T) {
	msg := New()
	msg.SetMTI("0800")
	msg.Set(3, "000000")
	msg.Set(11, "000001")
	msg.Set(41, "TERM0001")

	if msg.GetMTI() != "0800" {
		t.Fatalf("expected MTI 0800, got %s", msg.GetMTI())
	}

	pcode, ok := msg.GetString(3)
	if !ok || pcode != "000000" {
		t.Fatalf("expected 000000, got %s", pcode)
	}

	// Verify bitmap
	bm := msg.Bitmap()
	if !bm.IsSet(3) || !bm.IsSet(11) || !bm.IsSet(41) {
		t.Fatalf("bitmap does not have expected fields set")
	}

	// Unset
	msg.Unset(11)
	if msg.HasField(11) {
		t.Fatalf("field 11 should have been unset")
	}
	bm2 := msg.Bitmap()
	if bm2.IsSet(11) {
		t.Fatalf("field 11 should be cleared from bitmap")
	}
}

func TestMessageSubfields(t *testing.T) {
	msg := New()
	msg.SetMTI("0100")
	msg.SetSubfield(48, 1, "PROMO_CODE_123")
	msg.SetSubfield(48, 2, "DISCOUNT_50")

	val1, ok1 := msg.GetSubfield("48.1")
	val2, ok2 := msg.GetSubfield("48.2")
	if !ok1 || val1 != "PROMO_CODE_123" {
		t.Fatalf("expected PROMO_CODE_123, got %s", val1)
	}
	if !ok2 || val2 != "DISCOUNT_50" {
		t.Fatalf("expected DISCOUNT_50, got %s", val2)
	}

	// Set deep path
	err := msg.SetSubfieldByPath("127.3.1", "DEEP_VAL")
	if err != nil {
		t.Fatalf("set subfield by path failed: %v", err)
	}
	deepVal, ok := msg.GetSubfield("127.3.1")
	if !ok || deepVal != "DEEP_VAL" {
		t.Fatalf("expected DEEP_VAL, got %s", deepVal)
	}
}

func TestMessageMaskingAndDump(t *testing.T) {
	msg := New()
	msg.SetMTI("0200")
	msg.Set(2, "4111111111111111")
	msg.Set(3, "000000")
	msg.Set(52, "F1E2D3C4B5A69788") // PIN Block
	msg.Set(35, "4111111111111111=25121010000000000000")

	tlvMap := tlv.NewTLVMap(nil)
	_ = tlvMap.SetHex("9F02", "000000001000")
	_ = tlvMap.SetHex("5A", "4111111111111111")
	msg.SetTLV(55, tlvMap)

	dumpStr := msg.Dump(true)

	// Check PAN masked: 411111******1111
	if !strings.Contains(dumpStr, "411111******1111") {
		t.Fatalf("expected masked PAN in dump:\n%s", dumpStr)
	}

	// Check PIN block masked
	if !strings.Contains(dumpStr, "****************") {
		t.Fatalf("expected masked PIN block in dump:\n%s", dumpStr)
	}

	// Plaintext PAN should NOT appear anywhere in the masked dump
	if strings.Contains(dumpStr, "4111111111111111") {
		t.Fatalf("unmasked PAN leaked in masked dump:\n%s", dumpStr)
	}
}

func TestMessageClone(t *testing.T) {
	msg := New()
	msg.SetMTI("0800")
	msg.Set(3, "000000")

	clone := msg.Clone()
	clone.Set(3, "111111")

	orig3, _ := msg.GetString(3)
	clone3, _ := clone.GetString(3)
	if orig3 != "000000" || clone3 != "111111" {
		t.Fatalf("cloning is not isolated: orig=%s, clone=%s", orig3, clone3)
	}
}

func TestMessageCloneCopiesTLV(t *testing.T) {
	tags := tlv.NewTLVMap(nil)
	if err := tags.SetHex("9F02", "000000000100"); err != nil {
		t.Fatal(err)
	}
	msg := New()
	msg.SetMTI("0200")
	msg.SetTLV(55, tags)
	copyMsg := msg.Clone()
	if err := tags.SetHex("9F02", "000000000200"); err != nil {
		t.Fatal(err)
	}
	copyTags, ok := copyMsg.GetTLV(55)
	if !ok {
		t.Fatal("cloned TLV missing")
	}
	if got, _ := copyTags.GetHex("9F02"); got != "000000000100" {
		t.Fatalf("cloned TLV changed: %s", got)
	}
}

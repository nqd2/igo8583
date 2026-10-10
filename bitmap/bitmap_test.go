package bitmap

import (
	"testing"
)

func TestBitmapPrimary(t *testing.T) {
	bm := New()
	_ = bm.Set(3)
	_ = bm.Set(4)
	_ = bm.Set(11)

	if !bm.IsSet(3) || !bm.IsSet(4) || !bm.IsSet(11) {
		t.Fatalf("fields not set properly")
	}
	if bm.IsSet(2) || bm.IsSet(12) {
		t.Fatalf("unexpected field set")
	}
	if bm.MaxField() != 11 {
		t.Fatalf("expected MaxField 11, got %d", bm.MaxField())
	}

	// Pack Binary (should be 8 bytes, since maxField <= 64 and bit 1 not set)
	bin := bm.PackBinary()
	if len(bin) != 8 {
		t.Fatalf("expected 8 bytes, got %d", len(bin))
	}

	bm2 := New()
	consumed, err := bm2.UnpackBinary(bin)
	if err != nil || consumed != 8 {
		t.Fatalf("unpack binary failed: %d, %v", consumed, err)
	}
	if !bm2.IsSet(3) || !bm2.IsSet(4) || !bm2.IsSet(11) {
		t.Fatalf("unpacked fields mismatch")
	}
}

func TestBitmapSecondary(t *testing.T) {
	bm := New()
	_ = bm.Set(3)
	_ = bm.Set(70) // forces secondary bitmap

	bin := bm.PackBinary()
	if len(bin) != 16 {
		t.Fatalf("expected 16 bytes for secondary bitmap, got %d", len(bin))
	}
	if !bm.IsSet(1) {
		t.Fatalf("bit 1 should be set automatically")
	}

	bm2 := New()
	consumed, err := bm2.UnpackBinary(bin)
	if err != nil || consumed != 16 {
		t.Fatalf("unpack secondary failed: %d, %v", consumed, err)
	}
	if !bm2.IsSet(3) || !bm2.IsSet(70) || !bm2.IsSet(1) {
		t.Fatalf("fields mismatch on unpack")
	}

	// Hex roundtrip
	hexBytes := bm.PackHex()
	if len(hexBytes) != 32 {
		t.Fatalf("expected 32 hex chars, got %d: %s", len(hexBytes), string(hexBytes))
	}
	bm3 := New()
	hexConsumed, err := bm3.UnpackHex(hexBytes)
	if err != nil || hexConsumed != 32 {
		t.Fatalf("unpack hex failed: %d, %v", hexConsumed, err)
	}
	if !bm3.IsSet(70) {
		t.Fatalf("field 70 missing in hex unpack")
	}
}

func TestBitmapTertiary(t *testing.T) {
	bm := New()
	_ = bm.Set(3)
	_ = bm.Set(70)
	_ = bm.Set(130) // forces tertiary bitmap

	bin := bm.PackBinary()
	if len(bin) != 24 {
		t.Fatalf("expected 24 bytes for tertiary bitmap, got %d", len(bin))
	}
	if !bm.IsSet(1) || !bm.IsSet(65) {
		t.Fatalf("bits 1 and 65 must be set for tertiary bitmap")
	}

	bm2 := New()
	consumed, err := bm2.UnpackBinary(bin)
	if err != nil || consumed != 24 {
		t.Fatalf("unpack tertiary binary failed: %d, %v", consumed, err)
	}
	if !bm2.IsSet(130) {
		t.Fatalf("field 130 missing after unpack")
	}

	// Hex roundtrip
	hexBytes := bm.PackHex()
	if len(hexBytes) != 48 {
		t.Fatalf("expected 48 hex chars, got %d", len(hexBytes))
	}
	bm3 := New()
	consumed, err = bm3.UnpackHex(hexBytes)
	if err != nil || consumed != 48 {
		t.Fatalf("unpack tertiary hex failed: %d, %v", consumed, err)
	}
	if !bm3.IsSet(130) || !bm3.IsSet(70) || !bm3.IsSet(3) {
		t.Fatalf("tertiary fields mismatch")
	}
}

func TestBitmapCloneAndClear(t *testing.T) {
	bm := New()
	_ = bm.Set(4)
	clone := bm.Clone()
	_ = bm.Clear(4)

	if bm.IsSet(4) {
		t.Fatalf("clear failed")
	}
	if !clone.IsSet(4) {
		t.Fatalf("clone modified unexpectedly")
	}
}

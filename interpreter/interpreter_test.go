package interpreter

import (
	"bytes"
	"testing"
)

func TestASCIIInterpreter(t *testing.T) {
	interp := ASCII()
	packed, err := interp.Pack("HELLO")
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}
	if string(packed) != "HELLO" {
		t.Fatalf("expected HELLO, got %s", string(packed))
	}

	unpacked, consumed, err := interp.Unpack(packed, 5)
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}
	if consumed != 5 || unpacked != "HELLO" {
		t.Fatalf("unexpected unpack: %s, %d", unpacked, consumed)
	}
}

func TestBCDInterpreter(t *testing.T) {
	// Even length
	left := BCDLeftPadded()
	packed, err := left.Pack("1234")
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}
	expected := []byte{0x12, 0x34}
	if !bytes.Equal(packed, expected) {
		t.Fatalf("expected %X, got %X", expected, packed)
	}

	unpacked, consumed, err := left.Unpack(packed, 4)
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}
	if consumed != 2 || unpacked != "1234" {
		t.Fatalf("expected 1234, got %s (consumed %d)", unpacked, consumed)
	}

	// Odd length - Left Padded with 0: "123" -> 0x01, 0x23
	packed, err = left.Pack("123")
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}
	expected = []byte{0x01, 0x23}
	if !bytes.Equal(packed, expected) {
		t.Fatalf("expected %X, got %X", expected, packed)
	}

	unpacked, consumed, err = left.Unpack(packed, 3)
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}
	if consumed != 2 || unpacked != "123" {
		t.Fatalf("expected 123, got %s (consumed %d)", unpacked, consumed)
	}

	// Odd length - Right Padded with 0: "123" -> 0x12, 0x30
	right := BCDRightPadded()
	packed, err = right.Pack("123")
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}
	expected = []byte{0x12, 0x30}
	if !bytes.Equal(packed, expected) {
		t.Fatalf("expected %X, got %X", expected, packed)
	}

	unpacked, consumed, err = right.Unpack(packed, 3)
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}
	if consumed != 2 || unpacked != "123" {
		t.Fatalf("expected 123, got %s (consumed %d)", unpacked, consumed)
	}

	// Odd length - Right Padded with F: "123" -> 0x12, 0x3F
	rightF := BCDRightPaddedF()
	packed, err = rightF.Pack("123")
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}
	expected = []byte{0x12, 0x3F}
	if !bytes.Equal(packed, expected) {
		t.Fatalf("expected %X, got %X", expected, packed)
	}

	unpacked, consumed, err = rightF.Unpack(packed, 3)
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}
	if consumed != 2 || unpacked != "123" {
		t.Fatalf("expected 123, got %s (consumed %d)", unpacked, consumed)
	}
}

func TestHexInterpreter(t *testing.T) {
	interp := Hex()
	packed, err := interp.Pack("A1B2C3")
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}
	expected := []byte{0xA1, 0xB2, 0xC3}
	if !bytes.Equal(packed, expected) {
		t.Fatalf("expected %X, got %X", expected, packed)
	}

	unpacked, consumed, err := interp.Unpack(packed, 6)
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}
	if consumed != 3 || unpacked != "A1B2C3" {
		t.Fatalf("expected A1B2C3, got %s (consumed %d)", unpacked, consumed)
	}
}

func TestBinaryInterpreter(t *testing.T) {
	interp := Binary()
	data := "test\x00\xff"
	packed, err := interp.Pack(data)
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}
	unpacked, consumed, err := interp.Unpack(packed, len(data))
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}
	if consumed != len(data) || unpacked != data {
		t.Fatalf("roundtrip binary failed")
	}
}

func TestEBCDICInterpreter(t *testing.T) {
	interp := EBCDIC()
	testStr := "0800Hello World"
	packed, err := interp.Pack(testStr)
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}

	// Verify '0' is 0xF0
	if packed[0] != 0xF0 {
		t.Fatalf("expected 0xF0 for '0', got 0x%X", packed[0])
	}

	unpacked, consumed, err := interp.Unpack(packed, len(testStr))
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}
	if consumed != len(testStr) || unpacked != testStr {
		t.Fatalf("expected %s, got %s", testStr, unpacked)
	}
}

func TestRejectInvalidPackedDigits(t *testing.T) {
	if _, _, err := BCDLeftPadded().Unpack([]byte{0x1A}, 2); err == nil {
		t.Fatal("invalid BCD digit accepted")
	}
	if _, _, err := BCDRightPaddedF().Unpack([]byte{0x12, 0x30}, 3); err == nil {
		t.Fatal("wrong BCD pad nibble accepted")
	}
	if _, _, err := Hex().Unpack([]byte{0xF1, 0x23}, 3); err == nil {
		t.Fatal("wrong hex pad nibble accepted")
	}
	if _, err := ASCII().Pack("é"); err == nil {
		t.Fatal("non-ASCII text accepted")
	}
}

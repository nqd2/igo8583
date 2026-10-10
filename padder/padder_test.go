package padder

import (
	"errors"
	"testing"
)

func TestLeftPadder(t *testing.T) {
	p := LeftZero
	padded, err := p.Pad("123", 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if padded != "000123" {
		t.Fatalf("expected '000123', got '%s'", padded)
	}

	unpadded := p.Unpad(padded)
	if unpadded != "123" {
		t.Fatalf("expected '123', got '%s'", unpadded)
	}

	// exact length
	padded, err = p.Pad("123456", 6)
	if err != nil || padded != "123456" {
		t.Fatalf("exact length failed: %s, %v", padded, err)
	}

	// too long
	_, err = p.Pad("1234567", 6)
	if !errors.Is(err, ErrDataTooLong) {
		t.Fatalf("expected ErrDataTooLong, got %v", err)
	}
}

func TestRightPadder(t *testing.T) {
	p := RightSpace
	padded, err := p.Pad("ABC", 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if padded != "ABC   " {
		t.Fatalf("expected 'ABC   ', got '%s'", padded)
	}

	unpadded := p.Unpad(padded)
	if unpadded != "ABC" {
		t.Fatalf("expected 'ABC', got '%s'", unpadded)
	}
}

func TestNonePadder(t *testing.T) {
	p := NoPadding
	padded, err := p.Pad("HELLO", 5)
	if err != nil || padded != "HELLO" {
		t.Fatalf("exact match failed: %s, %v", padded, err)
	}
	if p.Unpad(padded) != "HELLO" {
		t.Fatalf("unpad failed")
	}

	padded, err = p.Pad("HELLO", 6)
	if err != nil || padded != "HELLO" {
		t.Fatalf("expected HELLO when within max length")
	}

	_, err = p.Pad("HELLO", 4)
	if !errors.Is(err, ErrDataTooLong) {
		t.Fatalf("expected ErrDataTooLong when exceeding max length")
	}
}

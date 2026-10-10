package prefixer

import (
	"bytes"
	"testing"
)

func TestFixedPrefixer(t *testing.T) {
	p := Fixed(6)
	b, err := p.EncodeLength(6)
	if err != nil || len(b) != 0 {
		t.Fatalf("unexpected encode for fixed: %v, %v", b, err)
	}

	_, err = p.EncodeLength(5)
	if err == nil {
		t.Fatalf("expected error on mismatched fixed length")
	}

	l, consumed, err := p.DecodeLength([]byte{0x01, 0x02})
	if err != nil || l != 6 || consumed != 0 {
		t.Fatalf("unexpected decode for fixed: %d, %d, %v", l, consumed, err)
	}
}

func TestAsciiPrefixer(t *testing.T) {
	p := AsciiLL
	b, err := p.EncodeLength(19)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if string(b) != "19" {
		t.Fatalf("expected '19', got '%s'", string(b))
	}

	l, consumed, err := p.DecodeLength([]byte("19extra"))
	if err != nil || l != 19 || consumed != 2 {
		t.Fatalf("decode failed: %d, %d, %v", l, consumed, err)
	}

	// LLL
	p3 := AsciiLLL
	b, err = p3.EncodeLength(5)
	if err != nil || string(b) != "005" {
		t.Fatalf("expected '005', got '%s', err: %v", string(b), err)
	}

	l, consumed, err = p3.DecodeLength([]byte("005xyz"))
	if err != nil || l != 5 || consumed != 3 {
		t.Fatalf("decode failed: %d, %d, %v", l, consumed, err)
	}
}

func TestBcdPrefixer(t *testing.T) {
	// LL: 2 digits -> 1 byte
	p := BcdLL
	b, err := p.EncodeLength(19)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if !bytes.Equal(b, []byte{0x19}) {
		t.Fatalf("expected [0x19], got %X", b)
	}

	l, consumed, err := p.DecodeLength([]byte{0x19, 0x99})
	if err != nil || l != 19 || consumed != 1 {
		t.Fatalf("decode failed: %d, %d, %v", l, consumed, err)
	}

	// LLL: 3 digits -> 2 bytes, e.g. 123 -> [0x01, 0x23]
	p3 := BcdLLL
	b, err = p3.EncodeLength(123)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if !bytes.Equal(b, []byte{0x01, 0x23}) {
		t.Fatalf("expected [0x01, 0x23], got %X", b)
	}

	l, consumed, err = p3.DecodeLength([]byte{0x01, 0x23})
	if err != nil || l != 123 || consumed != 2 {
		t.Fatalf("decode failed: %d, %d, %v", l, consumed, err)
	}
}

func TestBinaryPrefixer(t *testing.T) {
	p := Binary2BE
	b, err := p.EncodeLength(300)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if !bytes.Equal(b, []byte{0x01, 0x2C}) {
		t.Fatalf("expected [0x01, 0x2C], got %X", b)
	}

	l, consumed, err := p.DecodeLength(b)
	if err != nil || l != 300 || consumed != 2 {
		t.Fatalf("decode failed: %d, %d, %v", l, consumed, err)
	}
}

func TestEbcdicPrefixer(t *testing.T) {
	p := EbcdicLL
	b, err := p.EncodeLength(25)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	// '2' is 0xF2, '5' is 0xF5
	if !bytes.Equal(b, []byte{0xF2, 0xF5}) {
		t.Fatalf("expected [0xF2, 0xF5], got %X", b)
	}

	l, consumed, err := p.DecodeLength(b)
	if err != nil || l != 25 || consumed != 2 {
		t.Fatalf("decode failed: %d, %d, %v", l, consumed, err)
	}
}

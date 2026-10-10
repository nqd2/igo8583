package tlv

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestTLVPrimitive(t *testing.T) {
	// EMV test vector:
	// 9F02 (Amount authorized) = 000000010000 (6 bytes)
	// 9F36 (ATC) = 0001 (2 bytes)
	// 82 (AIP) = 1C00 (2 bytes)
	tlvMap := NewTLVMap(nil)
	_ = tlvMap.SetHex("9F02", "000000010000")
	_ = tlvMap.SetHex("9F36", "0001")
	_ = tlvMap.SetHex("82", "1C00")

	encoded, err := tlvMap.Encode()
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	decodedMap, err := ParseTLV(encoded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	v9F02, ok := decodedMap.GetHex("9F02")
	if !ok || v9F02 != "000000010000" {
		t.Fatalf("expected 000000010000, got %s", v9F02)
	}

	v9F36, ok := decodedMap.GetHex("9F36")
	if !ok || v9F36 != "0001" {
		t.Fatalf("expected 0001, got %s", v9F36)
	}

	v82, ok := decodedMap.GetHex("82")
	if !ok || v82 != "1C00" {
		t.Fatalf("expected 1C00, got %s", v82)
	}
}

func TestTLVMultiByteLength(t *testing.T) {
	// Length > 127, e.g. 150 bytes
	longVal := make([]byte, 150)
	for i := range longVal {
		longVal[i] = byte(i)
	}

	node := &Node{
		Tag:   []byte{0xDF, 0x01},
		Value: longVal,
	}

	encoded, err := Encode([]*Node{node})
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// Verify length byte starts with 0x81 (multi-byte indicator) followed by 150 (0x96)
	if encoded[2] != 0x81 || encoded[3] != 150 {
		t.Fatalf("length encoding incorrect: %X %X", encoded[2], encoded[3])
	}

	nodes, err := Decode(encoded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(nodes) != 1 || !bytes.Equal(nodes[0].Value, longVal) {
		t.Fatalf("value mismatch after multi-byte decode")
	}
}

func TestTLVConstructed(t *testing.T) {
	// Tag 0xE1 (Constructed, application template)
	child, _ := NewNode("9F02", []byte{0x01, 0x02})
	parent := &Node{
		Tag:      []byte{0xE1},
		Children: []*Node{child},
	}

	encoded, err := EncodeNode(parent)
	if err != nil {
		t.Fatalf("encode constructed failed: %v", err)
	}

	nodes, err := Decode(encoded)
	if err != nil {
		t.Fatalf("decode constructed failed: %v", err)
	}
	if len(nodes) != 1 || !nodes[0].IsConstructed() {
		t.Fatalf("node should be constructed")
	}
	if len(nodes[0].Children) != 1 || nodes[0].Children[0].TagHex() != "9F02" {
		t.Fatalf("child node mismatch")
	}
}

func TestTLVWithPadding(t *testing.T) {
	// 0x00 and 0xFF padding
	raw, _ := hex.DecodeString("00FF82021C0000")
	nodes, err := Decode(raw)
	if err != nil {
		t.Fatalf("decode failed with padding: %v", err)
	}
	if len(nodes) != 1 || nodes[0].TagHex() != "82" {
		t.Fatalf("expected 1 node with tag 82, got %d", len(nodes))
	}
}

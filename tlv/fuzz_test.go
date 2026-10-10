package tlv

import (
	"bytes"
	"testing"
)

func FuzzDecode(f *testing.F) {
	f.Add([]byte{0x9F, 0x02, 0x02, 0x00, 0x01})
	f.Add([]byte{0xE1, 0x03, 0x82, 0x01, 0x00})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 4096 {
			t.Skip()
		}
		nodes, err := Decode(data)
		if err != nil {
			return
		}
		wire, err := Encode(nodes)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := Decode(wire)
		if err != nil {
			t.Fatal(err)
		}
		canonical, err := Encode(decoded)
		if err != nil || !bytes.Equal(wire, canonical) {
			t.Fatalf("unstable encoding: %v", err)
		}
	})
}

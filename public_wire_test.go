package igo8583_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/nqd2/igo8583"
)

/* Field 11 sets bit 11: 1 << (64 - 11). No other data fields are set. */
func TestSyntheticWireFixtures(t *testing.T) {
	binary, err := hex.DecodeString("02000020000000000000000042")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		codec *igo8583.Codec
		wire  []byte
	}{
		{"ascii", igo8583.ISO87ASCII(), []byte("02000020000000000000000042")},
		{"binary", igo8583.ISO87Binary(), binary},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := igo8583.NewMessage()
			msg.SetMTI("0200")
			msg.Set(11, "000042")
			packed, err := tc.codec.Pack(msg)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(packed, tc.wire) {
				t.Fatalf("wire = %X, want %X", packed, tc.wire)
			}
			decoded, err := tc.codec.Unpack(tc.wire)
			if err != nil {
				t.Fatal(err)
			}
			stan, ok := decoded.GetString(11)
			if !ok || stan != "000042" || decoded.GetMTI() != "0200" {
				t.Fatal("fixture did not decode")
			}
		})
	}
}

func FuzzUnpackISO87(f *testing.F) {
	f.Add([]byte("02000020000000000000000042"))
	f.Add([]byte{})
	ascii, binary := igo8583.ISO87ASCII(), igo8583.ISO87Binary()
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 4096 {
			t.Skip()
		}
		for _, codec := range []*igo8583.Codec{ascii, binary} {
			msg, err := codec.Unpack(data)
			if err != nil {
				continue
			}
			wire, err := codec.Pack(msg)
			if err != nil {
				t.Fatalf("decoded message cannot pack: %v", err)
			}
			if _, err := codec.Unpack(wire); err != nil {
				t.Fatalf("repacked message cannot decode: %v", err)
			}
		}
	})
}

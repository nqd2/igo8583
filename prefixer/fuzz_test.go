package prefixer

import "testing"

func FuzzLengthPrefixes(f *testing.F) {
	f.Add([]byte("123"))
	f.Add([]byte{0x12, 0x30})
	f.Add([]byte{})
	codecs := []Prefixer{AsciiLL, AsciiLLL, BcdLL, BcdLLL, Binary1, Binary2BE, EbcdicLL}
	f.Fuzz(func(t *testing.T, data []byte) {
		for _, codec := range codecs {
			length, consumed, err := codec.DecodeLength(data)
			if err != nil {
				continue
			}
			if consumed < 0 || consumed > len(data) || length < 0 {
				t.Fatal("invalid decode result")
			}
			encoded, err := codec.EncodeLength(length)
			if err != nil {
				t.Fatal(err)
			}
			got, _, err := codec.DecodeLength(encoded)
			if err != nil || got != length {
				t.Fatalf("prefix did not round trip: %v", err)
			}
		}
	})
}

package spec_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/nqd2/igo8583/message"
	"github.com/nqd2/igo8583/packager"
	"github.com/nqd2/igo8583/spec"
	"github.com/nqd2/igo8583/specs"
)

/*
These literal bytes are derived from ISO 8583 bitmap positions and ISO
specifications, rather than from another parser's output.
*/
func TestIndependentWireFixtures(t *testing.T) {
	t.Run("ISO87 ASCII newly covered fields", func(t *testing.T) {
		codec := specs.ISO87ASCII()
		m := message.New()
		m.SetMTI("0200")
		m.Set(8, "12345678")
		m.Set(28, "C12345678")
		m.Set(45, "TRACK")
		m.SetBytes(64, []byte{1, 2, 3, 4, 5, 6, 7, 8})
		m.Set(66, "7")
		m.Set(104, "PAY")
		m.Set(120, "NAME")
		m.SetBytes(128, []byte{8, 7, 6, 5, 4, 3, 2, 1})
		// Primary DE1,8,28,45,64; secondary DE66,104,120,128.
		want := []byte("02008100001000080001400000000100010112345678C1234567805TRACK")
		want = append(want, []byte("0102030405060708")...)
		want = append(want, []byte("7003PAY004NAME")...)
		want = append(want, []byte("0807060504030201")...)
		got, err := codec.Pack(m)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("wire mismatch\ngot  %X\nwant %X", got, want)
		}
		parsed, err := codec.Unpack(want)
		if err != nil {
			t.Fatal(err)
		}
		if value, _ := parsed.GetString(104); value != "PAY" {
			t.Fatalf("DE104 = %q", value)
		}
		if value, _ := parsed.GetBytes(128); !bytes.Equal(value, []byte{8, 7, 6, 5, 4, 3, 2, 1}) {
			t.Fatalf("DE128 = %X", value)
		}
	})

	t.Run("ASCII LLVAR LLLVAR TLV secondary bitmap", func(t *testing.T) {
		definition := []byte(`{"bitmap_format":"hex","fields":{
			"0":{"name":"MTI","len":4,"type":"string","prefixer":"fixed","interpreter":"ascii"},
			"2":{"name":"PAN","len":19,"type":"string","prefixer":"ascii_ll","interpreter":"ascii"},
			"3":{"name":"Processing code","len":6,"type":"string","prefixer":"fixed","interpreter":"ascii"},
			"4":{"name":"Amount","len":12,"type":"string","prefixer":"fixed","interpreter":"ascii"},
			"11":{"name":"STAN","len":6,"type":"string","prefixer":"fixed","interpreter":"ascii"},
			"48":{"name":"Private","len":999,"type":"string","prefixer":"ascii_lll","interpreter":"ascii"},
			"55":{"name":"ICC","len":255,"type":"tlv","prefixer":"ascii_lll"},
			"70":{"name":"Network","len":3,"type":"string","prefixer":"fixed","interpreter":"ascii"}
		}}`)
		codec, err := spec.FromJSON(definition)
		if err != nil {
			t.Fatal(err)
		}
		tlvBytes, _ := hex.DecodeString("9F0206000000005000")
		m := message.New()
		m.SetMTI("0200")
		m.Set(2, "4111111111111111")
		m.Set(3, "000000")
		m.Set(4, "000000001234")
		m.Set(11, "123456")
		m.Set(48, "ABC")
		m.SetBytes(55, tlvBytes)
		m.Set(70, "301")
		// DE 1,2,3,4,11,48,55 -> F020000000010200; DE70 -> 0400000000000000.
		want := []byte("0200F0200000000102000400000000000000164111111111111111000000000000001234123456003ABC009")
		want = append(want, tlvBytes...)
		want = append(want, []byte("301")...)
		got, err := codec.Pack(m)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("wire mismatch\ngot  %X\nwant %X", got, want)
		}
		parsed, err := codec.Unpack(want)
		if err != nil {
			t.Fatal(err)
		}
		if pan, _ := parsed.GetString(2); pan != "4111111111111111" {
			t.Fatalf("PAN = %q", pan)
		}
		if data, _ := parsed.GetBytes(55); !bytes.Equal(data, tlvBytes) {
			t.Fatalf("TLV bytes = %X", data)
		}
	})

	t.Run("BCD LLVAR and binary bitmap", func(t *testing.T) {
		codec := specs.ISO87Binary()
		m := message.New()
		m.SetMTI("0800")
		m.Set(2, "123456789012345")
		m.Set(3, "000000")
		m.Set(70, "301")
		// MTI 0800; primary indicator+DE2+DE3; secondary DE70; BCD PAN LL=15.
		want, _ := hex.DecodeString("0800E00000000000000004000000000000001512345678901234500000000301")
		got, err := codec.Pack(m)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("wire mismatch\ngot  %X\nwant %X", got, want)
		}
		parsed, err := codec.Unpack(want)
		if err != nil {
			t.Fatal(err)
		}
		if pan, _ := parsed.GetString(2); pan != "123456789012345" {
			t.Fatalf("PAN = %q", pan)
		}
	})

	t.Run("four digit length framed codec", func(t *testing.T) {
		codec, err := spec.FromJSON([]byte(`{
			"name": "Length Prefixed Spec",
			"bitmap_format": "hex",
			"length_prefix": {"encoding": "ascii", "width": 4, "counts": "body"},
			"fields": {
				"0": {"name": "MTI", "len": 4, "type": "string", "prefixer": "fixed"},
				"7": {"name": "DE7", "len": 10, "type": "string", "prefixer": "fixed"},
				"11": {"name": "DE11", "len": 6, "type": "string", "prefixer": "fixed"},
				"70": {"name": "DE70", "len": 3, "type": "string", "prefixer": "fixed"},
				"128": {"name": "DE128", "len": 16, "type": "string", "prefixer": "fixed"}
			}
		}`))
		if err != nil {
			t.Fatal(err)
		}
		m := message.New()
		m.SetMTI("0800")
		m.Set(7, "1006183000")
		m.Set(11, "000123")
		m.Set(70, "301")
		m.Set(128, "A1B2C3D4E5F60718")
		// Body: MTI 4 + two 16-char bitmaps + DE7 10 + DE11 6 + DE70 3 + DE128 16 = 71.
		want := []byte("00710800822000000000000004000000000000011006183000000123301A1B2C3D4E5F60718")
		got, err := codec.Pack(m)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("wire mismatch\ngot  %q\nwant %q", got, want)
		}
		parsed, err := codec.Unpack(want)
		if err != nil {
			t.Fatal(err)
		}
		if value, _ := parsed.GetString(128); value != "A1B2C3D4E5F60718" {
			t.Fatalf("DE128 = %q", value)
		}
	})
}

func TestStrictConfiguration(t *testing.T) {
	base := `{"fields":{"0":{"name":"MTI","len":4,"type":"string","prefixer":"fixed","interpreter":"ascii"}}}`
	for _, source := range []string{
		`{"bitmap_format":"base64","fields":{"0":{"len":4,"type":"string","prefixer":"fixed"}}}`,
		`{"bitmap_field":2,"fields":{"0":{"len":4,"type":"string","prefixer":"fixed"}}}`,
		`{"third_bitmap_field":66,"fields":{"0":{"len":4,"type":"string","prefixer":"fixed"}}}`,
		`{"fields":{"0":{"len":4,"type":"string","prefixer":"unknown"}}}`,
		`{"fields":{"0":{"len":4,"type":"unknown","prefixer":"fixed"}}}`,
		`{"fields":{"0":{"len":4,"type":"string","prefixer":"fixed","interpreter":"unknown"}}}`,
		`{"fields":{"0":{"len":4,"type":"string","prefixer":"fixed","padder":"unknown"}}}`,
		`{"fields":{"0":{"len":0,"type":"string","prefixer":"fixed"}}}`,
		`{"fields":{"0":{"len":4,"type":"string","prefixer":"fixed"},"193":{"len":2,"type":"string","prefixer":"fixed"}}}`,
		`{"fields":{"0":{"len":4,"type":"string","prefixer":"fixed"},"1":{"len":2,"type":"string","prefixer":"fixed"}}}`,
		`{"fields":{"0":{"len":4,"type":"string","prefixer":"fixed"},"2":{"len":100,"type":"string","prefixer":"ascii_ll"}}}`,
		`{"unknown_runtime_option":true,"fields":{"0":{"len":4,"type":"string","prefixer":"fixed"}}}`,
		base + base,
	} {
		if _, err := spec.FromJSON([]byte(source)); err == nil {
			t.Errorf("accepted invalid spec: %s", source)
		}
	}
	if _, err := spec.FromJSON([]byte(base)); err != nil {
		t.Fatalf("valid base rejected: %v", err)
	}
}

func TestStrictWireRejection(t *testing.T) {
	codec, err := spec.FromJSON([]byte(`{"fields":{
		"0":{"len":4,"type":"string","prefixer":"fixed","interpreter":"ascii"},
		"2":{"len":19,"type":"string","prefixer":"ascii_ll","interpreter":"ascii"},
		"55":{"len":255,"type":"tlv","prefixer":"ascii_lll"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	m := message.New()
	m.SetMTI("020")
	if _, err := codec.Pack(m); err == nil {
		t.Fatal("short fixed MTI accepted")
	}
	m.SetMTI("0200")
	m.Set(2, "123")
	packed, err := codec.Pack(m)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasSuffix(packed, []byte("03123")) {
		t.Fatalf("LLVAR padded to maximum: %q", packed)
	}
	if _, err := codec.Unpack(append(packed, '!')); err == nil {
		t.Fatal("trailing byte accepted")
	}
	m.Set(193, "X")
	if _, err := codec.Pack(m); err == nil {
		t.Fatal("field 193 accepted")
	}
	m.Unset(193)
	m.Set(55, "not BER-TLV bytes")
	if _, err := codec.Pack(m); err == nil {
		t.Fatal("TLV string accepted and silently discarded")
	}
	m.SetBytes(55, []byte{0x9F, 0x02, 0x06, 0x00})
	if _, err := codec.Pack(m); err == nil {
		t.Fatal("malformed TLV accepted on pack")
	}
	broken := append([]byte("02000000000000000200004"), []byte{0x9F, 0x02, 0x06, 0x00}...)
	_, err = codec.Unpack(broken)
	var unpackErr *packager.UnpackError
	if !errors.As(err, &unpackErr) || unpackErr.FieldID != 55 {
		t.Fatalf("expected DE55 error, got %v", err)
	}
	_, err = codec.Unpack([]byte("020040000000000000000312")) // DE2 says 3 but only 2 bytes
	if !errors.As(err, &unpackErr) || unpackErr.FieldID != 2 {
		t.Fatalf("expected DE2 truncation, got %v", err)
	}
	_, err = codec.Unpack([]byte("02002000000000000000")) // DE3 has no definition
	if !errors.As(err, &unpackErr) || unpackErr.FieldID != 3 {
		t.Fatalf("expected DE3 missing packager, got %v", err)
	}
	framedCodec, err := spec.FromJSON([]byte(`{"length_prefix":{"encoding":"ascii","width":4,"counts":"body"},"fields":{"0":{"len":4,"type":"string","prefixer":"fixed"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, frame := range []string{"999908000000000000000000", "ABCD08000000000000000000", "000108000000000000000000"} {
		if _, err := framedCodec.Unpack([]byte(frame)); err == nil {
			t.Errorf("invalid framed message accepted: %q", frame)
		}
	}
}

func TestTertiaryBitmapAndSequentialComposite(t *testing.T) {
	codec, err := spec.FromJSON([]byte(`{"fields":{
		"0":{"len":4,"type":"string","prefixer":"fixed","interpreter":"ascii"},
		"48":{"len":99,"type":"composite","prefixer":"ascii_lll","subfields":{
			"1":{"len":2,"type":"string","prefixer":"fixed","interpreter":"ascii"},
			"2":{"len":5,"type":"string","prefixer":"ascii_ll","interpreter":"ascii"}}},
		"130":{"len":2,"type":"string","prefixer":"fixed","interpreter":"ascii"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	m := message.New()
	m.SetMTI("0200")
	m.SetSubfield(48, 1, "AB")
	m.SetSubfield(48, 2, "XYZ")
	m.Set(130, "OK")
	// Primary bit 1 and DE48; secondary bit 65; tertiary DE130.
	want := []byte("0200800000000001000080000000000000004000000000000000007AB03XYZOK")
	got, err := codec.Pack(m)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("wire mismatch\ngot  %q\nwant %q", got, want)
	}
	parsed, err := codec.Unpack(want)
	if err != nil {
		t.Fatal(err)
	}
	if sub, _ := parsed.GetSubfield("48.2"); sub != "XYZ" {
		t.Fatalf("DE48.2 = %q", sub)
	}
	if data, _ := parsed.GetString(130); data != "OK" {
		t.Fatalf("DE130 = %q", data)
	}
	m.Unset(48)
	m.SetSubfield(48, 1, "AB")
	if _, err := codec.Pack(m); err == nil {
		t.Fatal("incomplete sequential composite accepted")
	}
}

func FuzzUnpackFramed(f *testing.F) {
	codec, err := spec.FromJSON([]byte(`{"bitmap_format":"hex","length_prefix":{"encoding":"ascii","width":4,"counts":"body"},"fields":{"0":{"len":4,"type":"string","prefixer":"fixed"},"7":{"len":10,"type":"string","prefixer":"fixed"},"11":{"len":6,"type":"string","prefixer":"fixed"},"70":{"len":3,"type":"string","prefixer":"fixed"},"128":{"len":16,"type":"string","prefixer":"fixed"}}}`))
	if err != nil {
		f.Fatal(err)
	}
	f.Add([]byte("00710800822000000000000004000000000000011006183000000123301A1B2C3D4E5F60718"))
	f.Add([]byte{0, 1, 2, 3})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 4096 {
			return
		}
		m, err := codec.Unpack(data)
		if err != nil {
			return
		}
		packed, err := codec.Pack(m)
		if err != nil {
			t.Fatalf("accepted message cannot be packed: %v", err)
		}
		if _, err := codec.Unpack(packed); err != nil {
			t.Fatalf("repacked message cannot be unpacked: %v", err)
		}
	})
}

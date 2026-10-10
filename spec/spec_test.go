package spec

import (
	"testing"

	"github.com/nqd2/igo8583/interpreter"
	"github.com/nqd2/igo8583/message"
	"github.com/nqd2/igo8583/padder"
	"github.com/nqd2/igo8583/prefixer"
)

func TestSpecFromJSON(t *testing.T) {
	jsonSpec := `
	{
		"name": "Sample ISO JSON Spec",
		"bitmap_format": "hex",
		"fields": {
			"0": { "name": "MTI", "len": 4, "type": "string", "prefixer": "fixed", "interpreter": "ascii" },
			"2": { "name": "PAN", "len": 19, "type": "string", "prefixer": "ascii_ll", "interpreter": "ascii" },
			"3": { "name": "Proc Code", "len": 6, "type": "string", "prefixer": "fixed", "interpreter": "ascii", "padder": "left_zero" },
			"4": { "name": "Amount", "len": 12, "type": "string", "prefixer": "fixed", "interpreter": "ascii", "padder": "left_zero" },
			"48": {
				"name": "Private Data",
				"len": 999,
				"type": "composite",
				"prefixer": "ascii_lll",
				"subfields": {
					"1": { "name": "Sub 1", "len": 2, "type": "string", "prefixer": "fixed", "interpreter": "ascii" },
					"2": { "name": "Sub 2", "len": 4, "type": "string", "prefixer": "fixed", "interpreter": "ascii" }
				}
			},
			"55": {
				"name": "EMV Chip Data",
				"len": 999,
				"type": "tlv",
				"prefixer": "ascii_lll"
			}
		}
	}
	`

	pkg, err := FromJSON([]byte(jsonSpec))
	if err != nil {
		t.Fatalf("failed to create packager from JSON: %v", err)
	}

	msg := message.New()
	msg.SetMTI("0800")
	msg.Set(2, "4111111111111111")
	msg.Set(3, "000000")
	msg.Set(4, "50000")
	msg.SetSubfield(48, 1, "XY")
	msg.SetSubfield(48, 2, "9999")

	packed, err := pkg.Pack(msg)
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}

	unpacked, err := pkg.Unpack(packed)
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}

	if unpacked.GetMTI() != "0800" {
		t.Fatalf("expected MTI 0800, got %s", unpacked.GetMTI())
	}
	pan, _ := unpacked.GetString(2)
	if pan != "4111111111111111" {
		t.Fatalf("expected PAN 4111111111111111, got %s", pan)
	}
	s1, _ := unpacked.GetSubfield("48.1")
	s2, _ := unpacked.GetSubfield("48.2")
	if s1 != "XY" || s2 != "9999" {
		t.Fatalf("subfields mismatch: %s, %s", s1, s2)
	}
}

func TestFluentBuilder(t *testing.T) {
	pkg, err := NewBuilder("Builder Spec").
		WithBitmapFormat("hex").
		StringField(0, 4, "MTI", padder.NoPadding, interpreter.ASCII(), prefixer.Fixed(4)).
		StringField(3, 6, "Proc Code", padder.LeftZero, interpreter.ASCII(), prefixer.Fixed(6)).
		Build()
	if err != nil {
		t.Fatalf("builder compile failed: %v", err)
	}

	msg := message.New()
	msg.SetMTI("0800")
	msg.Set(3, "000000")

	packed, err := pkg.Pack(msg)
	if err != nil {
		t.Fatalf("builder pack failed: %v", err)
	}

	unpacked, err := pkg.Unpack(packed)
	if err != nil {
		t.Fatalf("builder unpack failed: %v", err)
	}

	pcode, _ := unpacked.GetString(3)
	if pcode != "000000" {
		t.Fatalf("expected 000000, got %s", pcode)
	}
}

func TestBuilderRejectsInvalidFieldConfiguration(t *testing.T) {
	_, err := NewBuilder("bad").
		StringField(0, 4, "MTI", padder.NoPadding, interpreter.ASCII(), prefixer.Fixed(4)).
		StringField(2, 100, "PAN", padder.NoPadding, interpreter.ASCII(), prefixer.AsciiLL).
		Build()
	if err == nil {
		t.Fatal("builder accepted field maximum beyond LL prefix")
	}
	_, err = NewBuilder("bad").
		WithBitmapFormat("invalid").
		StringField(0, 4, "MTI", padder.NoPadding, interpreter.ASCII(), prefixer.Fixed(4)).
		Build()
	if err == nil {
		t.Fatal("builder accepted invalid bitmap format")
	}
}

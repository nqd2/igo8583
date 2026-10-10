package specs

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nqd2/igo8583/message"
	"github.com/nqd2/igo8583/spec"
)

func TestISO87FieldInventory(t *testing.T) {
	if len(iso87Fields) != 126 {
		t.Fatalf("expected 126 data fields, got %d", len(iso87Fields))
	}
	seen := make(map[int]bool, len(iso87Fields))
	for _, def := range iso87Fields {
		if def.id < 2 || def.id > 128 || def.id == 65 || seen[def.id] {
			t.Fatalf("invalid or duplicate DE%d", def.id)
		}
		seen[def.id] = true
	}
	for id := 2; id <= 128; id++ {
		if id != 65 && !seen[id] {
			t.Fatalf("missing DE%d", id)
		}
	}
}

func TestISO87AllConfiguredFieldsPackAndUnpack(t *testing.T) {
	for _, profile := range []struct {
		name   string
		binary bool
	}{
		{"ASCII", false},
		{"Binary", true},
	} {
		t.Run(profile.name, func(t *testing.T) {
			codec := newISO87(profile.binary)
			var jsonCodec interface {
				Pack(*message.Message) ([]byte, error)
			}
			if !profile.binary {
				loaded, err := spec.FromJSONFile("../examples/json_specs/iso87_ascii_128.json")
				if err != nil {
					t.Fatal(err)
				}
				jsonCodec = loaded
			}
			for _, def := range iso87Fields {
				msg := message.New()
				msg.SetMTI("0200")
				var wantString string
				var wantBytes []byte
				switch def.kind {
				case 'b':
					wantBytes = []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}
					if def.id == 96 {
						wantBytes = append(wantBytes, wantBytes...)
					}
					msg.SetBytes(def.id, wantBytes)
				case 'n':
					wantString = "123"
					if def.variable == 0 {
						wantString = strings.Repeat("1", def.length)
					}
					msg.Set(def.id, wantString)
				case 'z':
					wantString = "12D3"
					msg.Set(def.id, wantString)
				default:
					wantString = "AZ"
					if def.variable == 0 {
						wantString = strings.Repeat("A", def.length)
					}
					if def.id == 28 || def.id == 29 || def.id == 30 || def.id == 31 || def.id == 97 {
						wantString = "C" + strings.Repeat("1", def.length-1)
					}
					msg.Set(def.id, wantString)
				}
				wire, err := codec.Pack(msg)
				if err != nil {
					t.Errorf("DE%d pack: %v", def.id, err)
					continue
				}
				if jsonCodec != nil {
					fromJSON, err := jsonCodec.Pack(msg)
					if err != nil || !bytes.Equal(wire, fromJSON) {
						t.Errorf("DE%d JSON preset differs: %v", def.id, err)
					}
				}
				parsed, err := codec.Unpack(wire)
				if err != nil {
					t.Errorf("DE%d unpack: %v", def.id, err)
					continue
				}
				if def.kind == 'b' {
					got, _ := parsed.GetBytes(def.id)
					if !bytes.Equal(got, wantBytes) {
						t.Errorf("DE%d = %X, want %X", def.id, got, wantBytes)
					}
				} else {
					got, _ := parsed.GetString(def.id)
					if got != wantString {
						t.Errorf("DE%d = %q, want %q", def.id, got, wantString)
					}
				}
			}
		})
	}
}

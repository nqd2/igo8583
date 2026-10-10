package spec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/nqd2/igo8583/field"
	"github.com/nqd2/igo8583/interpreter"
	"github.com/nqd2/igo8583/packager"
	"github.com/nqd2/igo8583/padder"
	"github.com/nqd2/igo8583/prefixer"
)

// FieldDef defines a single field specification for JSON serialization.
type FieldDef struct {
	Name        string              `json:"name"`
	Length      int                 `json:"len"`
	Type        string              `json:"type"`        // "string", "binary", "amount", "composite", "tlv"
	Prefixer    string              `json:"prefixer"`    // "fixed", "ascii_ll", "bcd_lll", etc.
	Interpreter string              `json:"interpreter"` // "ascii", "bcd", "hex", "ebcdic", etc.
	Padder      string              `json:"padder"`      // "none", "left_zero", "right_space", etc.
	Subfields   map[string]FieldDef `json:"subfields,omitempty"`
}

// SpecDef defines the root JSON structure of a packager specification.
type SpecDef struct {
	Name             string              `json:"name"`
	HeaderLength     int                 `json:"header_length,omitempty"`
	LengthPrefix     *LengthPrefixDef    `json:"length_prefix,omitempty"`
	BitmapFormat     string              `json:"bitmap_format,omitempty"`      // "hex" (default) or "binary"
	BitmapField      int                 `json:"bitmap_field,omitempty"`       // default 1
	ThirdBitmapField int                 `json:"third_bitmap_field,omitempty"` // default -1
	Fields           map[string]FieldDef `json:"fields"`
	Reference        json.RawMessage     `json:"reference,omitempty"` // documentation only
}

type LengthPrefixDef struct {
	Encoding string `json:"encoding"`
	Width    int    `json:"width"`
	Counts   string `json:"counts"`
}

// FromJSON compiles a codec from a JSON byte slice.
func FromJSON(data []byte) (*packager.Codec, error) {
	var def SpecDef
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&def); err != nil {
		return nil, fmt.Errorf("spec: failed to parse JSON: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("spec: trailing JSON content")
	}
	if def.HeaderLength < 0 {
		return nil, fmt.Errorf("spec: negative header_length")
	}
	if def.BitmapField != 0 && def.BitmapField != 1 {
		return nil, fmt.Errorf("spec: only bitmap_field 1 is supported")
	}
	if def.ThirdBitmapField != 0 && def.ThirdBitmapField != 65 {
		return nil, fmt.Errorf("spec: only third_bitmap_field 65 is supported")
	}
	if len(def.Fields) == 0 {
		return nil, fmt.Errorf("spec: fields are required")
	}

	pkg := packager.NewBasePackager()
	pkg.SetHeaderLength(def.HeaderLength)
	if def.LengthPrefix != nil {
		if def.LengthPrefix.Encoding != "ascii" || def.LengthPrefix.Counts != "body" || def.LengthPrefix.Width < 1 || def.LengthPrefix.Width > 9 {
			return nil, fmt.Errorf("spec: length_prefix requires ASCII width 1..9 counting body bytes")
		}
		pkg.SetLengthPrefixASCII(def.LengthPrefix.Width)
	}

	if def.BitmapFormat != "" && def.BitmapFormat != "hex" && def.BitmapFormat != "binary" {
		return nil, fmt.Errorf("spec: unsupported bitmap_format %q", def.BitmapFormat)
	}
	if def.BitmapFormat != "" {
		pkg.SetBitmapFormat(def.BitmapFormat)
	}

	for fIDStr, fDef := range def.Fields {
		fID, err := strconv.Atoi(fIDStr)
		if err != nil {
			return nil, fmt.Errorf("spec: invalid field id '%s': %w", fIDStr, err)
		}
		if strconv.Itoa(fID) != fIDStr || fID < 0 || fID > 192 || fID == 1 || fID == 65 {
			return nil, fmt.Errorf("spec: invalid or reserved field id %q", fIDStr)
		}
		if fID == 0 && (fDef.Type != "string" || fDef.Length != 4 || fDef.Prefixer != "fixed") {
			return nil, fmt.Errorf("spec: MTI must be a fixed four-character string")
		}
		fp, err := buildFieldPackager(fDef)
		if err != nil {
			return nil, fmt.Errorf("spec: error building field %d (%s): %w", fID, fDef.Name, err)
		}
		pkg.SetFieldPackager(fID, fp)
	}

	return pkg.Compile()
}

// FromJSONFile loads and compiles a codec from a JSON file.
func FromJSONFile(filename string) (*packager.Codec, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("spec: cannot read file %s: %w", filename, err)
	}
	return FromJSON(data)
}

func buildFieldPackager(def FieldDef) (field.FieldPackager, error) {
	if def.Length <= 0 {
		return nil, fmt.Errorf("length must be positive")
	}
	fType := strings.ToLower(def.Type)
	pref, err := resolvePrefixer(def.Prefixer, def.Length)
	if err != nil {
		return nil, err
	}
	if _, err := pref.EncodeLength(def.Length); err != nil {
		return nil, fmt.Errorf("field maximum length cannot be represented by prefixer %q: %w", def.Prefixer, err)
	}

	switch fType {
	case "binary":
		if (def.Interpreter != "" && def.Interpreter != "ascii_hex") || def.Padder != "" || len(def.Subfields) > 0 {
			return nil, fmt.Errorf("binary field has unsupported string/subfield options")
		}
		if def.Interpreter == "ascii_hex" {
			return field.NewASCIIBinaryFieldPackager(def.Length, def.Name, pref), nil
		}
		return field.NewBinaryFieldPackager(def.Length, def.Name, pref), nil

	case "amount":
		if def.Prefixer != "fixed" || def.Padder != "" || len(def.Subfields) > 0 ||
			(def.Interpreter != "ascii" && def.Interpreter != "bcd") {
			return nil, fmt.Errorf("amount field requires fixed prefix, ascii or bcd interpreter, and no padding/subfields")
		}
		return field.NewAmountFieldPackager(def.Length, def.Name, def.Interpreter == "bcd"), nil

	case "composite":
		if def.Interpreter != "" || def.Padder != "" {
			return nil, fmt.Errorf("composite field has string options")
		}
		subs := make(map[int]field.FieldPackager)
		for subIDStr, subDef := range def.Subfields {
			subID, err := strconv.Atoi(subIDStr)
			if err != nil {
				return nil, fmt.Errorf("invalid subfield id '%s': %w", subIDStr, err)
			}
			if strconv.Itoa(subID) != subIDStr || subID < 1 || subID > 192 {
				return nil, fmt.Errorf("invalid subfield id %q", subIDStr)
			}
			subFP, err := buildFieldPackager(subDef)
			if err != nil {
				return nil, err
			}
			subs[subID] = subFP
		}
		return field.NewCompositeFieldPackager(def.Length, def.Name, pref, subs), nil

	case "tlv":
		if def.Interpreter != "" || def.Padder != "" || len(def.Subfields) > 0 {
			return nil, fmt.Errorf("TLV field has unsupported string/subfield options")
		}
		return field.NewTLVFieldPackager(def.Length, def.Name, pref), nil

	case "string":
		if len(def.Subfields) > 0 {
			return nil, fmt.Errorf("string field cannot have subfields")
		}
		if pref.PackedLength() > 0 && def.Padder != "" && def.Padder != "none" {
			return nil, fmt.Errorf("variable field cannot use fixed-width padding")
		}
		p, err := resolvePadder(def.Padder)
		if err != nil {
			return nil, err
		}
		interp, err := resolveInterpreter(def.Interpreter)
		if err != nil {
			return nil, err
		}
		return field.NewStringFieldPackager(def.Length, def.Name, p, interp, pref), nil
	default:
		return nil, fmt.Errorf("unsupported field type %q", def.Type)
	}
}

func resolvePrefixer(name string, length int) (prefixer.Prefixer, error) {
	switch strings.ToLower(name) {
	case "ascii_l":
		return prefixer.AsciiL, nil
	case "ascii_ll":
		return prefixer.AsciiLL, nil
	case "ascii_lll":
		return prefixer.AsciiLLL, nil
	case "ascii_llll":
		return prefixer.AsciiLLLL, nil
	case "ascii_lllll":
		return prefixer.AsciiLLLLL, nil
	case "bcd_l":
		return prefixer.BcdL, nil
	case "bcd_ll":
		return prefixer.BcdLL, nil
	case "bcd_lll":
		return prefixer.BcdLLL, nil
	case "bcd_llll":
		return prefixer.BcdLLLL, nil
	case "binary_1":
		return prefixer.Binary1, nil
	case "binary_2", "binary_2be":
		return prefixer.Binary2BE, nil
	case "binary_4", "binary_4be":
		return prefixer.Binary4BE, nil
	case "ebcdic_l":
		return prefixer.EbcdicL, nil
	case "ebcdic_ll":
		return prefixer.EbcdicLL, nil
	case "ebcdic_lll":
		return prefixer.EbcdicLLL, nil
	case "fixed":
		return prefixer.Fixed(length), nil
	default:
		return nil, fmt.Errorf("unsupported prefixer %q", name)
	}
}

func resolveInterpreter(name string) (interpreter.Interpreter, error) {
	switch strings.ToLower(name) {
	case "bcd", "bcd_left":
		return interpreter.BCDLeftPadded(), nil
	case "bcd_right":
		return interpreter.BCDRightPadded(), nil
	case "bcd_right_f":
		return interpreter.BCDRightPaddedF(), nil
	case "hex":
		return interpreter.Hex(), nil
	case "binary":
		return interpreter.Binary(), nil
	case "ebcdic":
		return interpreter.EBCDIC(), nil
	case "ascii", "":
		return interpreter.ASCII(), nil
	default:
		return nil, fmt.Errorf("unsupported interpreter %q", name)
	}
}

func resolvePadder(name string) (padder.Padder, error) {
	switch strings.ToLower(name) {
	case "left_zero":
		return padder.LeftZero, nil
	case "left_space":
		return padder.LeftSpace, nil
	case "right_space":
		return padder.RightSpace, nil
	case "right_zero":
		return padder.RightZero, nil
	case "none", "":
		return padder.NoPadding, nil
	default:
		return nil, fmt.Errorf("unsupported padder %q", name)
	}
}

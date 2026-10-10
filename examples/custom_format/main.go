package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/nqd2/igo8583"
	"github.com/nqd2/igo8583/field"
	"github.com/nqd2/igo8583/interpreter"
	"github.com/nqd2/igo8583/padder"
	"github.com/nqd2/igo8583/prefixer"
	"github.com/nqd2/igo8583/tlv"
)

func main() {
	fromJSON, err := igo8583.FromJSONFile("examples/json_specs/custom.json")
	if err != nil {
		log.Fatal(err)
	}
	codec, err := igo8583.NewBuilder("Synthetic framed format").
		WithBitmapFormat("hex").WithHeaderLength(2).WithASCIILengthPrefix(4).
		StringField(0, 4, "MTI", padder.NoPadding, interpreter.ASCII(), prefixer.Fixed(4)).
		StringField(11, 6, "STAN", padder.LeftZero, interpreter.ASCII(), prefixer.Fixed(6)).
		CompositeField(48, 99, "Private data", prefixer.AsciiLL, map[int]field.FieldPackager{
			1: field.NewStringFieldPackager(3, "Channel", nil, nil, nil),
		}).TLVField(55, 99, "Chip data", prefixer.AsciiLL).Build()
	if err != nil {
		log.Fatal(err)
	}
	msg := igo8583.NewMessage()
	msg.SetHeader([]byte("EX"))
	msg.SetMTI("0200")
	msg.Set(11, "000042")
	if err := msg.SetSubfieldByPath("48.1", "POS"); err != nil {
		log.Fatal(err)
	}
	chip := tlv.NewTLVMap(nil)
	if err := chip.SetHex("9F02", "000000001000"); err != nil {
		log.Fatal(err)
	}
	msg.SetTLV(55, chip)
	wire, err := codec.Pack(msg)
	if err != nil {
		log.Fatal(err)
	}
	jsonWire, err := fromJSON.Pack(msg)
	if err != nil {
		log.Fatal(err)
	}
	if !bytes.Equal(wire, jsonWire) {
		log.Fatal("builder and JSON formats differ")
	}
	for _, candidate := range []*igo8583.Codec{codec, fromJSON} {
		decoded, err := candidate.Unpack(wire)
		if err != nil {
			log.Fatal(err)
		}
		channel, ok := decoded.GetSubfield("48.1")
		if !ok || channel != "POS" {
			log.Fatal("composite field did not round trip")
		}
		tags, ok := decoded.GetTLV(55)
		if !ok {
			log.Fatal("TLV field missing")
		}
		amount, ok := tags.GetHex("9F02")
		if !ok || amount != "000000001000" {
			log.Fatal("TLV value did not round trip")
		}
	}
	fmt.Println("builder and JSON agree; header, composite, and TLV round trip")
}

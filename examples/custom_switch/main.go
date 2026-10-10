package main

import (
	"fmt"
	"log"

	"github.com/nqd2/igo8583"
)

func main() {
	codec, err := igo8583.FromJSONFile("examples/json_specs/sample_iso87.json")
	if err != nil {
		log.Fatal(err)
	}
	msg := igo8583.NewMessage()
	msg.SetMTI("0200")
	msg.Set(3, "000000")
	msg.Set(4, "000000001000")
	msg.Set(11, "000001")
	msg.Set(41, "TERM0001")
	if err := msg.SetSubfieldByPath("48.1", "POS"); err != nil {
		log.Fatal(err)
	}
	if err := msg.SetSubfieldByPath("48.2", "0123456789ABCDEF0123456789ABCDEF"); err != nil {
		log.Fatal(err)
	}
	wire, err := codec.Pack(msg)
	if err != nil {
		log.Fatal(err)
	}
	parsed, err := codec.Unpack(wire)
	if err != nil {
		log.Fatal(err)
	}
	response := parsed.Clone()
	response.SetMTI("0210")
	response.Set(39, "00")
	reply, err := codec.Pack(response)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("request=%s response=%s bytes=%d\n", parsed.GetMTI(), response.GetMTI(), len(reply))
}

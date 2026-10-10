package main

import (
	"fmt"
	"log"

	"github.com/nqd2/igo8583"
)

type Request struct {
	MTI        string `iso:"0"`
	Processing string `iso:"3"`
	Amount     string `iso:"4"`
	STAN       string `iso:"11"`
	TerminalID string `iso:"41"`
	Channel    string `iso:"48.1"`
	Session    string `iso:"48.2"`
}

func main() {
	codec, err := igo8583.FromJSONFile("examples/json_specs/sample_iso87.json")
	if err != nil {
		log.Fatal(err)
	}
	req := Request{
		MTI: "0200", Processing: "000000", Amount: "000000001000",
		STAN: "000001", TerminalID: "TERM0001", Channel: "POS",
		Session: "0123456789ABCDEF0123456789ABCDEF",
	}
	wire, err := igo8583.PackStruct(req, codec)
	if err != nil {
		log.Fatal(err)
	}
	var decoded Request
	if err := igo8583.UnpackStruct(wire, &decoded, codec); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("MTI=%s STAN=%s channel=%s\n", decoded.MTI, decoded.STAN, decoded.Channel)
}

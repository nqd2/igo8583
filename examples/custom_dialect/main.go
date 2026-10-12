package main

import (
	"fmt"
	"log"

	"github.com/nqd2/igo8583"
	"github.com/nqd2/igo8583/examples/custom_dialect/specs/retail"
)

func main() {
	codec, err := retail.Codec()
	if err != nil {
		log.Fatal(err)
	}

	request := igo8583.NewMessage()
	request.SetMTI("0200")
	request.Set(3, "000000")
	request.Set(4, "000000100000")
	request.Set(11, "000123")
	request.Set(41, "POS00001")
	if err := retail.Validate(request); err != nil {
		log.Fatal(err)
	}

	wire, err := codec.Pack(request)
	if err != nil {
		log.Fatal(err)
	}
	parsed, err := codec.Unpack(wire)
	if err != nil {
		log.Fatal(err)
	}
	if err := retail.Validate(parsed); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("validated MTI=%s bytes=%d\n", parsed.GetMTI(), len(wire))
}

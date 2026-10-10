package main

import (
	"fmt"
	"log"

	"github.com/nqd2/igo8583"
)

func main() {
	codec := igo8583.ISO87ASCII()
	msg := igo8583.NewMessage()
	msg.SetMTI("0200")
	msg.Set(3, "000000")
	msg.Set(4, "000000001000")
	msg.Set(11, "000001")
	msg.Set(41, "TERM0001")

	wire, err := codec.Pack(msg)
	if err != nil {
		log.Fatal(err)
	}
	decoded, err := codec.Unpack(wire)
	if err != nil {
		log.Fatal(err)
	}
	stan, _ := decoded.GetString(11)
	fmt.Printf("MTI=%s STAN=%s bytes=%d\n", decoded.GetMTI(), stan, len(wire))
}

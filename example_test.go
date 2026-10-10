package igo8583_test

import (
	"errors"
	"fmt"

	"github.com/nqd2/igo8583"
	"github.com/nqd2/igo8583/packager"
)

func Example() {
	codec := igo8583.ISO87ASCII()
	msg := igo8583.NewMessage()
	msg.SetMTI("0200")
	msg.Set(11, "000042")
	wire, err := codec.Pack(msg)
	if err != nil {
		panic(err)
	}
	decoded, err := codec.Unpack(wire)
	if err != nil {
		panic(err)
	}
	stan, _ := decoded.GetString(11)
	fmt.Printf("MTI=%s STAN=%s\n", decoded.GetMTI(), stan)
	// Output: MTI=0200 STAN=000042
}

func ExampleFromJSON() {
	codec, err := igo8583.FromJSON([]byte(`{
  "fields": {
   "0": {"name":"MTI","len":4,"type":"string","prefixer":"fixed","interpreter":"ascii"},
   "11": {"name":"STAN","len":6,"type":"string","prefixer":"fixed","interpreter":"ascii","padder":"left_zero"}
  }
 }`))
	if err != nil {
		panic(err)
	}
	msg := igo8583.NewMessage()
	msg.SetMTI("0200")
	msg.Set(11, "42")
	wire, err := codec.Pack(msg)
	if err != nil {
		panic(err)
	}
	decoded, err := codec.Unpack(wire)
	if err != nil {
		panic(err)
	}
	stan, _ := decoded.GetString(11)
	fmt.Println(stan)
	// Output: 000042
}

func Example_errorHandling() {
	_, err := igo8583.ISO87ASCII().Unpack([]byte("02"))
	var detail *packager.UnpackError
	if errors.As(err, &detail) {
		fmt.Printf("field=%d offset=%d\n", detail.FieldID, detail.Offset)
	}
	// Output: field=0 offset=0
}

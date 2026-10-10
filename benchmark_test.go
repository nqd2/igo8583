package igo8583

import (
	"testing"
)

type BenchTransaction struct {
	MTI       string `iso:"0"`
	PAN       string `iso:"2"`
	ProcCode  string `iso:"3"`
	Amount    int64  `iso:"4"`
	STAN      int    `iso:"11"`
	LocalTime string `iso:"12"`
	LocalDate string `iso:"13"`
	TID       string `iso:"41"`
	MID       string `iso:"42"`
}

func BenchmarkPackISO87ASCII(b *testing.B) {
	pkg := ISO87ASCII()
	msg := NewMessage()
	msg.SetMTI("0200")
	msg.Set(2, "4111111111111111")
	msg.Set(3, "000000")
	msg.Set(4, "000000150000")
	msg.Set(11, "000042")
	msg.Set(12, "120000")
	msg.Set(13, "1006")
	msg.Set(41, "TERM0001")
	msg.Set(42, "MERCHANT0000001")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := pkg.Pack(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnpackISO87ASCII(b *testing.B) {
	pkg := ISO87ASCII()
	msg := NewMessage()
	msg.SetMTI("0200")
	msg.Set(2, "4111111111111111")
	msg.Set(3, "000000")
	msg.Set(4, "000000150000")
	msg.Set(11, "000042")
	msg.Set(12, "120000")
	msg.Set(13, "1006")
	msg.Set(41, "TERM0001")
	msg.Set(42, "MERCHANT0000001")

	raw, err := pkg.Pack(msg)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := pkg.Unpack(raw)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPackStruct(b *testing.B) {
	pkg := ISO87ASCII()
	txn := BenchTransaction{
		MTI:       "0200",
		PAN:       "4111111111111111",
		ProcCode:  "000000",
		Amount:    150000,
		STAN:      42,
		LocalTime: "120000",
		LocalDate: "1006",
		TID:       "TERM0001",
		MID:       "MERCHANT0000001",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := PackStruct(txn, pkg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnpackStruct(b *testing.B) {
	pkg := ISO87ASCII()
	txn := BenchTransaction{
		MTI:       "0200",
		PAN:       "4111111111111111",
		ProcCode:  "000000",
		Amount:    150000,
		STAN:      42,
		LocalTime: "120000",
		LocalDate: "1006",
		TID:       "TERM0001",
		MID:       "MERCHANT0000001",
	}

	raw, err := PackStruct(txn, pkg)
	if err != nil {
		b.Fatal(err)
	}

	var target BenchTransaction
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := UnpackStruct(raw, &target, pkg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

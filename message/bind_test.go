package message

import (
	"bytes"
	"testing"
)

type SampleTxn struct {
	MTI      string `iso:"0"`
	PAN      string `iso:"2"`
	ProcCode string `iso:"3"`
	Amount   int64  `iso:"4"`
	STAN     int    `iso:"11"`
	Data48_1 string `iso:"48.1"`
	MAC      []byte `iso:"128"`
}

func TestMarshalUnmarshalStruct(t *testing.T) {
	txn := SampleTxn{
		MTI:      "0200",
		PAN:      "4111111111111111",
		ProcCode: "000000",
		Amount:   150000,
		STAN:     42,
		Data48_1: "AUTH_TOKEN",
		MAC:      []byte{0xDE, 0xAD, 0xBE, 0xEF},
	}

	msg, err := Marshal(txn)
	if err != nil {
		t.Fatalf("marshal struct failed: %v", err)
	}

	if msg.GetMTI() != "0200" {
		t.Fatalf("expected MTI 0200, got %s", msg.GetMTI())
	}
	pan, _ := msg.GetString(2)
	if pan != "4111111111111111" {
		t.Fatalf("expected PAN 4111111111111111, got %s", pan)
	}
	sub, _ := msg.GetSubfield("48.1")
	if sub != "AUTH_TOKEN" {
		t.Fatalf("expected AUTH_TOKEN, got %s", sub)
	}
	mac, _ := msg.GetBytes(128)
	if !bytes.Equal(mac, []byte{0xDE, 0xAD, 0xBE, 0xEF}) {
		t.Fatalf("expected MAC bytes mismatch")
	}

	// Unmarshal back into a fresh struct
	var restored SampleTxn
	err = Unmarshal(msg, &restored)
	if err != nil {
		t.Fatalf("unmarshal struct failed: %v", err)
	}

	if restored.MTI != txn.MTI {
		t.Fatalf("MTI mismatch: %s vs %s", restored.MTI, txn.MTI)
	}
	if restored.PAN != txn.PAN {
		t.Fatalf("PAN mismatch")
	}
	if restored.Amount != txn.Amount {
		t.Fatalf("amount mismatch: %d vs %d", restored.Amount, txn.Amount)
	}
	if restored.STAN != txn.STAN {
		t.Fatalf("STAN mismatch: %d vs %d", restored.STAN, txn.STAN)
	}
	if restored.Data48_1 != txn.Data48_1 {
		t.Fatalf("subfield mismatch: %s", restored.Data48_1)
	}
	if !bytes.Equal(restored.MAC, txn.MAC) {
		t.Fatalf("MAC mismatch")
	}
}

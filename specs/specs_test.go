package specs

import (
	"bytes"
	"testing"

	"github.com/nqd2/igo8583/message"
)

func TestISO87ASCIIRoundtrip(t *testing.T) {
	pkg := ISO87ASCII()

	msg := message.New()
	msg.SetMTI("0200")
	msg.Set(2, "4111111111111111")
	msg.Set(3, "000000")
	msg.Set(4, "000000250000") // 2,500.00
	msg.Set(11, "123456")
	msg.Set(41, "TERM0001")
	msg.Set(70, "301") // secondary bitmap

	packed, err := pkg.Pack(msg)
	if err != nil {
		t.Fatalf("ISO87ASCII pack failed: %v", err)
	}
	// Primary: DE1,2,3,4,11,41; secondary: DE70.
	want := []byte("0200F0200000008000000400000000000000164111111111111111000000000000250000123456TERM0001301")
	if !bytes.Equal(packed, want) {
		t.Fatalf("ISO87ASCII wire mismatch\ngot  %q\nwant %q", packed, want)
	}

	unpacked, err := pkg.Unpack(want)
	if err != nil {
		t.Fatalf("ISO87ASCII unpack failed: %v", err)
	}

	if unpacked.GetMTI() != "0200" {
		t.Fatalf("expected MTI 0200, got %s", unpacked.GetMTI())
	}
	pan, _ := unpacked.GetString(2)
	if pan != "4111111111111111" {
		t.Fatalf("expected PAN 4111111111111111, got %s", pan)
	}
	tid, _ := unpacked.GetString(41)
	if tid != "TERM0001" {
		t.Fatalf("expected TID TERM0001, got %s", tid)
	}
}

func TestISO87BinaryRoundtrip(t *testing.T) {
	pkg := ISO87Binary()

	msg := message.New()
	msg.SetMTI("0800")
	msg.Set(3, "000000")
	msg.Set(11, "000042")
	msg.Set(70, "301")

	msg.Set(55, "ICC DATA")

	packed, err := pkg.Pack(msg)
	if err != nil {
		t.Fatalf("ISO87Binary pack failed: %v", err)
	}

	unpacked, err := pkg.Unpack(packed)
	if err != nil {
		t.Fatalf("ISO87Binary unpack failed: %v", err)
	}

	if unpacked.GetMTI() != "0800" {
		t.Fatalf("expected MTI 0800, got %s", unpacked.GetMTI())
	}
	stan, _ := unpacked.GetString(11)
	if stan != "000042" {
		t.Fatalf("expected STAN 000042, got %s", stan)
	}
	value55, _ := unpacked.GetString(55)
	if value55 != "ICC DATA" {
		t.Fatalf("expected ISO87 DE55 text, got %q", value55)
	}
}

func TestISO93ASCIIRoundtrip(t *testing.T) {
	pkg := ISO93ASCII()

	msg := message.New()
	msg.SetMTI("1200")
	msg.Set(2, "5200000000000001")
	msg.Set(3, "000000")
	msg.Set(4, "000000100000")
	msg.Set(11, "999888")
	msg.Set(24, "100")  // Function code
	msg.Set(25, "0001") // Message reason code
	msg.Set(39, "000")  // Action code 3 digits

	packed, err := pkg.Pack(msg)
	if err != nil {
		t.Fatalf("ISO93ASCII pack failed: %v", err)
	}

	unpacked, err := pkg.Unpack(packed)
	if err != nil {
		t.Fatalf("ISO93ASCII unpack failed: %v", err)
	}

	if unpacked.GetMTI() != "1200" {
		t.Fatalf("expected MTI 1200, got %s", unpacked.GetMTI())
	}
	f24, _ := unpacked.GetString(24)
	if f24 != "100" {
		t.Fatalf("expected function code 100, got %s", f24)
	}
	f39, _ := unpacked.GetString(39)
	if f39 != "000" {
		t.Fatalf("expected action code 000, got %s", f39)
	}
}

func TestISO93BinaryRoundtrip(t *testing.T) {
	pkg := ISO93Binary()

	msg := message.New()
	msg.SetMTI("1804")
	msg.Set(11, "123456")
	msg.Set(24, "801")
	msg.Set(70, "301")

	packed, err := pkg.Pack(msg)
	if err != nil {
		t.Fatalf("ISO93Binary pack failed: %v", err)
	}

	unpacked, err := pkg.Unpack(packed)
	if err != nil {
		t.Fatalf("ISO93Binary unpack failed: %v", err)
	}

	if unpacked.GetMTI() != "1804" {
		t.Fatalf("expected MTI 1804, got %s", unpacked.GetMTI())
	}
	f24, _ := unpacked.GetString(24)
	if f24 != "801" {
		t.Fatalf("expected function code 801, got %s", f24)
	}
}

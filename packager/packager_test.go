package packager

import (
	"errors"
	"sync"
	"testing"

	"github.com/nqd2/igo8583/field"
	"github.com/nqd2/igo8583/interpreter"
	"github.com/nqd2/igo8583/message"
	"github.com/nqd2/igo8583/padder"
	"github.com/nqd2/igo8583/prefixer"
)

func buildTestPackager() *BasePackager {
	pkg := NewBasePackager()
	pkg.SetBitmapFormat("hex")

	// Field 0: MTI (4 chars ASCII)
	pkg.SetFieldPackager(0, field.NewStringFieldPackager(4, "MTI", padder.NoPadding, interpreter.ASCII(), prefixer.Fixed(4)))
	// Field 2: PAN (LLVAR, max 19)
	pkg.SetFieldPackager(2, field.NewStringFieldPackager(19, "PAN", padder.NoPadding, interpreter.ASCII(), prefixer.AsciiLL))
	// Field 3: Processing Code (Fixed 6 digits)
	pkg.SetFieldPackager(3, field.NewStringFieldPackager(6, "Proc Code", padder.LeftZero, interpreter.ASCII(), prefixer.Fixed(6)))
	// Field 4: Amount (Fixed 12 digits)
	pkg.SetFieldPackager(4, field.NewStringFieldPackager(12, "Amount", padder.LeftZero, interpreter.ASCII(), prefixer.Fixed(12)))
	// Field 11: STAN (Fixed 6 digits)
	pkg.SetFieldPackager(11, field.NewStringFieldPackager(6, "STAN", padder.LeftZero, interpreter.ASCII(), prefixer.Fixed(6)))
	// Field 70: Network Code (Fixed 3 digits) - forces secondary bitmap
	pkg.SetFieldPackager(70, field.NewStringFieldPackager(3, "Net Code", padder.LeftZero, interpreter.ASCII(), prefixer.Fixed(3)))

	return pkg
}

func TestCodecPackFieldUsesConfiguredVariablePrefix(t *testing.T) {
	pkg := buildTestPackager()
	codec, err := pkg.Compile()
	if err != nil {
		t.Fatal(err)
	}
	got, err := codec.PackField(2, field.NewStringComponent(2, "12345"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "0512345" {
		t.Fatalf("PackField() = %q, want %q", got, "0512345")
	}
}

func TestPackUnpackRoundtrip(t *testing.T) {
	pkg := buildTestPackager()

	msg := message.New()
	msg.SetMTI("0800")
	msg.Set(2, "4111111111111111")
	msg.Set(3, "000000")
	msg.Set(4, "000000150000") // 1500.00
	msg.Set(11, "000042")
	msg.Set(70, "301")

	packed, err := pkg.Pack(msg)
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}

	unpackedMsg, err := pkg.Unpack(packed)
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}

	if unpackedMsg.GetMTI() != "0800" {
		t.Fatalf("expected MTI 0800, got %s", unpackedMsg.GetMTI())
	}

	pan, _ := unpackedMsg.GetString(2)
	if pan != "4111111111111111" {
		t.Fatalf("expected PAN 4111111111111111, got %s", pan)
	}

	amt, _ := unpackedMsg.GetString(4)
	if amt != "000000150000" {
		t.Fatalf("expected Amount 000000150000, got %s", amt)
	}

	net, _ := unpackedMsg.GetString(70)
	if net != "301" {
		t.Fatalf("expected Net Code 301, got %s", net)
	}
}

func TestUnpackDiagnosticError(t *testing.T) {
	pkg := buildTestPackager()

	msg := message.New()
	msg.SetMTI("0800")
	msg.Set(3, "000000")
	msg.Set(11, "000042")

	packed, err := pkg.Pack(msg)
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}

	// Truncate packed data to simulate communication cutoff
	truncated := packed[:len(packed)-3]

	_, err = pkg.Unpack(truncated)
	if err == nil {
		t.Fatalf("expected unpack error on truncated message")
	}

	var unpackErr *UnpackError
	if !errors.As(err, &unpackErr) {
		t.Fatalf("expected UnpackError type, got %T: %v", err, err)
	}

	// Offset and field ID should be reported
	if unpackErr.FieldID != 11 {
		t.Fatalf("expected failure at field 11, got field %d", unpackErr.FieldID)
	}
	t.Logf("Diagnostic error message: %s", unpackErr.Error())
}

func TestCompiledCodecSnapshotAndFraming(t *testing.T) {
	builder := NewBasePackager()
	builder.SetHeaderLength(2)
	builder.SetLengthPrefixASCII(3)
	builder.SetFieldPackager(0, field.NewStringFieldPackager(4, "MTI", padder.NoPadding, interpreter.ASCII(), prefixer.Fixed(4)))
	codec, err := builder.Compile()
	if err != nil {
		t.Fatal(err)
	}
	// Mutating the builder cannot change the compiled codec.
	builder.SetFieldPackager(0, field.NewStringFieldPackager(2, "Other", padder.NoPadding, interpreter.ASCII(), prefixer.Fixed(2)))
	m := message.New()
	m.SetHeader([]byte("AB"))
	m.SetMTI("0200")
	packed, err := codec.Pack(m)
	if err != nil {
		t.Fatal(err)
	}
	if string(packed) != "022AB02000000000000000000" {
		t.Fatalf("unexpected frame %q", packed)
	}
	decoded, err := codec.Unpack(packed)
	if err != nil {
		t.Fatal(err)
	}
	if string(decoded.GetHeader()) != "AB" {
		t.Fatalf("header = %q", decoded.GetHeader())
	}
	if _, err := codec.Pack(message.New()); err == nil {
		t.Fatal("missing MTI accepted")
	}
	m.SetHeader([]byte("A"))
	if _, err := codec.Pack(m); err == nil {
		t.Fatal("wrong header length accepted")
	}
}

func TestCompiledCodecConcurrentUse(t *testing.T) {
	codec, err := buildTestPackager().Compile()
	if err != nil {
		t.Fatal(err)
	}
	m := message.New()
	m.SetMTI("0200")
	m.Set(3, "000000")
	m.Set(11, "000042")
	var workers sync.WaitGroup
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for j := 0; j < 100; j++ {
				wire, err := codec.Pack(m)
				if err != nil {
					t.Errorf("pack: %v", err)
					return
				}
				if _, err := codec.Unpack(wire); err != nil {
					t.Errorf("unpack: %v", err)
					return
				}
			}
		}()
	}
	workers.Wait()
}

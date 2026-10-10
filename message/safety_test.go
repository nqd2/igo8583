package message

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestDumpShortTracks(t *testing.T) {
	for _, id := range []int{35, 45} {
		for length := 0; length <= 16; length++ {
			t.Run(fmt.Sprintf("field%d/length%d", id, length), func(t *testing.T) {
				msg := New()
				value := strings.Repeat("7", length)
				msg.Set(id, value)
				dump := msg.Dump(true)
				if length > 0 && strings.Contains(dump, value) {
					t.Fatal("dump contains unmasked track")
				}
			})
		}
	}
}

func TestUnmarshalNilMessage(t *testing.T) {
	var target struct {
		MTI string `iso:"0"`
	}
	if err := Unmarshal(nil, &target); err == nil {
		t.Fatal("expected error for nil message")
	}
}

func TestMarshalSkipsUnexportedFields(t *testing.T) {
	type request struct {
		MTI    string `iso:"0"`
		hidden bool   `iso:"3"`
	}
	msg, err := Marshal(request{MTI: "0200", hidden: true})
	if err != nil {
		t.Fatal(err)
	}
	if msg.HasField(3) {
		t.Fatal("unexported field was marshaled")
	}
	var decoded request
	if err := Unmarshal(msg, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.MTI != "0200" {
		t.Fatal("exported field did not round trip")
	}
}

func TestBindingNilAndNumericErrors(t *testing.T) {
	var request *struct {
		MTI string `iso:"0"`
	}
	if _, err := Marshal(request); !errors.Is(err, ErrNotStruct) {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := Unmarshal(New(), request); !errors.Is(err, ErrNotStructPtr) {
		t.Fatalf("unexpected error: %v", err)
	}
	msg := New()
	msg.Set(4, "private-value")
	var target struct {
		Amount int `iso:"4"`
	}
	err := Unmarshal(msg, &target)
	if err == nil || strings.Contains(err.Error(), "private-value") {
		t.Fatal("numeric binding error exposes input")
	}
	var numberErr *strconv.NumError
	if !errors.As(err, &numberErr) || !errors.Is(err, strconv.ErrSyntax) {
		t.Fatal("numeric error chain lost")
	}
}

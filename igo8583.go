package igo8583

import (
	"github.com/nqd2/igo8583/message"
	"github.com/nqd2/igo8583/packager"
	"github.com/nqd2/igo8583/spec"
	"github.com/nqd2/igo8583/specs"
)

// Message is an alias to message.Message.
type Message = message.Message

// Packager is an alias to packager.Packager.
type Packager = packager.Packager

// BasePackager is an alias to packager.BasePackager.
type BasePackager = packager.BasePackager

// Codec is a compiled, immutable packager configuration.
type Codec = packager.Codec

// Builder is an alias to spec.Builder.
type Builder = spec.Builder

// NewMessage creates a new empty Message.
func NewMessage() *Message {
	return message.New()
}

// NewBuilder creates a new fluent packager builder.
func NewBuilder(name string) *Builder {
	return spec.NewBuilder(name)
}

// FromJSON creates a packager from JSON specification data.
func FromJSON(data []byte) (*Codec, error) {
	return spec.FromJSON(data)
}

// FromJSONFile loads a packager from a JSON specification file.
func FromJSONFile(filename string) (*Codec, error) {
	return spec.FromJSONFile(filename)
}

// Marshal converts a struct tagged with `iso:"..."` into an ISO-8583 Message.
func Marshal(v any) (*Message, error) {
	return message.Marshal(v)
}

// Unmarshal populates a struct pointer from an ISO-8583 Message.
func Unmarshal(msg *Message, v any) error {
	return message.Unmarshal(msg, v)
}

// PackStruct is a convenience function that marshals a Go struct and packs it to wire bytes.
func PackStruct(v any, pkg Packager) ([]byte, error) {
	msg, err := Marshal(v)
	if err != nil {
		return nil, err
	}
	return pkg.Pack(msg)
}

// UnpackStruct is a convenience function that unmarshals wire bytes directly into a Go struct.
func UnpackStruct(data []byte, v any, pkg Packager) error {
	msg, err := pkg.Unpack(data)
	if err != nil {
		return err
	}
	return Unmarshal(msg, v)
}

// Standard Specs Presets
var (
	ISO87ASCII  = specs.ISO87ASCII
	ISO87Binary = specs.ISO87Binary
	ISO93ASCII  = specs.ISO93ASCII
	ISO93Binary = specs.ISO93Binary
)

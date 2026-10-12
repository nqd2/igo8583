package igo8583

import (
	"github.com/nqd2/igo8583/message"
	"github.com/nqd2/igo8583/packager"
	"github.com/nqd2/igo8583/spec"
	"github.com/nqd2/igo8583/specs"
)

/* ============================================================================
 * Type Aliases
 * ============================================================================ */

/* Message is the in-memory representation of an ISO 8583 message. */
type Message = message.Message

/* Packager packs complete messages and unpacks complete wire messages. */
type Packager = packager.Packager

/* BasePackager collects mutable settings before they are compiled into a Codec. */
type BasePackager = packager.BasePackager

/* Codec is an immutable compiled message layout. */
type Codec = packager.Codec

/* Builder fluently defines a message layout before Build compiles it. */
type Builder = spec.Builder

/* ============================================================================
 * Constructors & Codec Factories
 * ============================================================================ */

/* NewMessage creates an empty message with initialized field storage. */
func NewMessage() *Message {
	return message.New()
}

/* NewBuilder starts a fluent definition for a custom message layout. */
func NewBuilder(name string) *Builder {
	return spec.NewBuilder(name)
}

/*
FromJSON compiles a codec from a declarative JSON specification byte slice.
It validates field lengths, prefix types, and mandatory MTI presence.
*/
func FromJSON(data []byte) (*Codec, error) {
	return spec.FromJSON(data)
}

/*
FromJSONFile loads and compiles a codec from a JSON specification file path.
*/
func FromJSONFile(filename string) (*Codec, error) {
	return spec.FromJSONFile(filename)
}

/* ============================================================================
 * Struct Marshalling & Binding
 * ============================================================================ */

/*
Marshal serializes a Go struct with `iso:"..."` tags into an ISO 8583 Message.
Supports string, numeric, composite, and TLV field tags.
*/
func Marshal(v any) (*Message, error) {
	return message.Marshal(v)
}

/*
Unmarshal deserializes an ISO 8583 Message into a destination Go struct pointer.
*/
func Unmarshal(msg *Message, v any) error {
	return message.Unmarshal(msg, v)
}

/*
PackStruct serializes a Go struct and immediately encodes it to wire bytes.
*/
func PackStruct(v any, pkg Packager) ([]byte, error) {
	msg, err := Marshal(v)
	if err != nil {
		return nil, err
	}
	return pkg.Pack(msg)
}

/*
UnpackStruct decodes wire bytes and populates the destination struct pointer in one call.
*/
func UnpackStruct(data []byte, v any, pkg Packager) error {
	msg, err := pkg.Unpack(data)
	if err != nil {
		return err
	}
	return Unmarshal(msg, v)
}

/* ============================================================================
 * Standard Protocol Presets
 * ============================================================================ */

/* ============================================================================
 * Standard ISO Layout Factories
 * ============================================================================ */

/*
Standard ISO layout factories create compiled codecs for ASCII and binary
ISO 8583:1987 and ISO 8583:1993 field definitions.
*/
var (
	ISO87ASCII  = specs.ISO87ASCII
	ISO87Binary = specs.ISO87Binary
	ISO93ASCII  = specs.ISO93ASCII
	ISO93Binary = specs.ISO93Binary
)

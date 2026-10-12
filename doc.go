/*
Package igo8583 packs and unpacks complete ISO 8583 messages using compiled
field layouts. It provides ISO 8583:1987 and ISO 8583:1993 ASCII and binary
layouts, plus JSON and Go builder APIs for custom wire formats.

# Start here

Choose a standard layout when its field definitions match the connected
system. Use FromJSON or NewBuilder when the system changes field encodings,
bitmap representation, header length, or transport length prefix. Create one
Codec during startup and reuse it. Create a new Message for each transaction.

# Dialects and validation

A Codec validates whether fields can be represented by its wire layout. It
does not determine which fields a transaction requires or whether their values
are valid for a payment network. Keep those rules in the application that uses
the dialect. A useful layout is:

	/your-project/
	    specs/<dialect-name>/
	        spec.json
	        codec.go
	        validate.go

The package can load the JSON layout with FromJSON or FromJSONFile, expose a
Codec factory, and validate transaction-specific requirements. The library
does not include network-specific dialect packages or transaction validators.

# Package responsibilities

The root package provides common aliases, constructors, standard layout
factories, and struct binding. The message package stores field values. The
spec package compiles JSON or builder definitions. The packager package
encodes and decodes complete messages. Field encoders use the prefixer,
interpreter, and padder packages. The bitmap package handles bitmap blocks;
the tlv package handles BER-TLV values.

# Framing and streams

Codec.Pack emits one complete message, including a configured header and
length prefix. Codec.Unpack accepts exactly one complete message and rejects
trailing bytes. It does not read from a TCP connection or split a stream into
messages. The application must read the configured frame before calling
Unpack.

# Concurrency and ownership

Compiled codecs expose no configuration mutators and can be shared between
goroutines when custom field packagers are immutable. Message methods lock
their internal field map, but getters and setters can expose or retain mutable
components, byte slices, and TLV nodes. Do not mutate a message or referenced
value while another goroutine packs or inspects it. Clone a message before
independent edits.
*/
package igo8583

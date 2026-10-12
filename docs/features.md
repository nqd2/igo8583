# Features and Behavior

`igo8583` turns an in-memory message and a compiled field layout into wire
bytes, and decodes one complete wire message back into a message. The layout
controls encoding. Applications supply transaction and network rules.

## Message encoding

- The four built-in layouts cover ISO 8583:1987 and ISO 8583:1993 ASCII and
  binary profiles.
- JSON and a fluent builder define custom field positions, field types,
  prefixes, interpreters, padding, bitmaps, headers, and transport length
  prefixes.
- A compiled `Codec` validates and snapshots the field layout. `Pack` emits
  one message; `Unpack` rejects malformed frames and trailing bytes.
- The bitmap package supports up to 192 positions. The configured dialect
  determines whether each field is valid and how each bitmap is encoded.

## Field values

The field package provides string, binary, amount, composite, and TLV
packagers. Prefixers encode fixed or variable lengths. Interpreters handle
ASCII, EBCDIC, BCD, hex, and binary values. Padders align fixed-width values.
These strategies stay independent so a custom layout can combine them.

`message.Message` stores MTI, optional header bytes, field components, and
bitmap state. Composite subfields support nested paths. `tlv` decodes and
encodes BER-TLV nodes independently of any payment-network definition.

## Struct binding and diagnostics

`Marshal` and `Unmarshal` map Go structs to messages with `iso` tags.
`PackStruct` and `UnpackStruct` combine that mapping with a codec. `Dump`
formats message fields and can mask selected data; applications still need a
policy for sensitive fields and logs.

Pack and unpack errors retain field context and support `errors.Is` and
`errors.As`. Diagnostic errors avoid printing field values, but callers should
not log raw wire bytes or arbitrary nested error causes without review.

## Application responsibilities

Place a network dialect in the application, such as
`your-project/specs/<dialect-name>/`. Keep its specification, codec factory,
and transaction validator together. The library does not decide required
fields per MTI, response codes, transaction state, MAC keys, routing, retries,
or how TCP streams are split into frames.

See [integration](integration.md) for exact layout semantics and
[API reference](api.md) for public packages.

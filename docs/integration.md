# Integrating igo8583

## Choose a wire layout

Use `ISO87ASCII`, `ISO87Binary`, `ISO93ASCII`, or `ISO93Binary` only when their field definitions match the peer. The presets are field-format profiles, not transaction validators. In particular, the ISO87 ASCII preset encodes fixed binary fields as ASCII hex, while its binary preset uses raw bytes. DE55 in these presets is text or binary data, respectively; it is not automatically a TLV field.

Create one codec at startup. Fail startup if a custom specification does not compile. Build a new message for each operation. `Pack` emits the configured message or frame; `Unpack` requires exactly one complete message or frame and rejects trailing bytes. Reading a complete frame from a stream, enforcing an application size limit, timeouts, and retries are the caller's responsibility.

## Custom JSON and builder definitions

[custom.json](../examples/json_specs/custom.json) defines a small synthetic layout with a header, length prefix, composite field, and TLV field. [custom_format](../examples/custom_format/main.go) builds the same layout in Go, checks byte equality, and round-trips it through both codecs:

```sh
go run ./examples/custom_format
```

Load JSON with `FromJSON(data)` or `FromJSONFile(path)`. The JSON loader rejects unknown options, extra JSON content, reserved field IDs, and incompatible combinations. Both the loader and `Builder.Build()` return `(*Codec, error)`.

A minimal field definition looks like this:

```json
{"name":"STAN","len":6,"type":"string","prefixer":"fixed","interpreter":"ascii","padder":"left_zero"}
```

The root object requires `fields`, including field `0`: a fixed four-unit string MTI. `bitmap_format` is `hex` by default or `binary`. Positions 1 and 65 indicate additional bitmaps, so do not declare them as data fields. The bitmap engine supports up to 192 positions; the codec emits the required bitmap blocks from the active fields.

| Field type | Meaning of `len` and length prefix | Values |
| --- | --- | --- |
| `string` | Maximum logical units before interpretation | `Set` |
| `binary` | Maximum raw bytes, including when encoded as ASCII hex | `SetBytes` |
| `amount` | Total units: one sign character plus digits | `Set`, fixed prefix only |
| `composite` | Maximum encoded body bytes | Subfield setters |
| `tlv` | Maximum encoded BER-TLV body bytes | `SetTLV` or `SetBytes` |

String units are bytes for ASCII/EBCDIC/binary, decimal digits for BCD, and hex digits for the hex interpreter. Go strings are not counted in Unicode runes. For example, three BCD digits need two wire bytes, while an ASCII LL prefix describes the three logical digits. An ASCII-hex binary field uses two wire characters per byte but its length prefix counts bytes.

Fixed fields require the configured length after padding. Variable fields use their actual length and do not accept fixed-width padding. Common JSON prefix names are `fixed`, `ascii_ll`, `ascii_lll`, `bcd_ll`, `bcd_lll`, `binary_1`, `binary_2be`, and `binary_4be`. String interpreter names include `ascii`, `ebcdic`, `binary`, `hex`, `bcd_left`, `bcd_right`, and `bcd_right_f`. The default string interpreter is ASCII. Binary fields optionally use `ascii_hex`.

BCD left padding inserts a leading zero for odd digit counts; BCD right padding inserts a trailing zero or F nibble. Hex uses leading zero padding, or `HexRightPadded()` in Go for trailing zero padding. Choose the convention required by the peer. Padding nibbles do not increase the logical length.

## Header and framing

`header_length` specifies an exact number of bytes before the MTI. Set that many bytes with `SetHeader`; the library does not invent a TPDU or header. Header bytes are included in the body byte count.

For an ASCII frame length, configure:

```json
{"encoding":"ascii","width":4,"counts":"body"}
```

Place that object under the root `length_prefix` property. It counts all bytes after the prefix, including the header, MTI, bitmaps, and fields. `WithHeaderLength(n)` and `WithASCIILengthPrefix(width)` configure the same behavior in Go. Length widths must be 1–9. Presets have no transport length prefix or header unless the particular layout defines one; use a custom layout when your transport needs them.

## Composite fields and TLV

Composites are sequential layouts: every declared subfield must be present, in ascending numeric ID order, without a subfield bitmap. A field path such as `48.1` identifies a subfield; it does not define its wire encoding. `SetSubfieldByPath` returns an error for an invalid path. `SetSubfield` is convenient for a single nesting level.

A TLV field is configured explicitly with JSON `type: "tlv"` or `Builder.TLVField`. Use `tlv.NewTLVMap(nil)` and `SetHex` to construct tags, then `SetTLV` to attach the map. Check every `SetHex` error. BER-TLV decoding accepts 00/FF padding and has a nesting limit of 16; re-encoding may normalize padding and length representation. Application rules for tag presence or tag values remain separate.

## Struct binding

`Marshal` and `Unmarshal` convert between structs and messages. `PackStruct` and `UnpackStruct` add wire encoding. Use exported fields tagged `iso:"0"`, `iso:"11"`, or `iso:"48.1"`. Untagged fields, `iso:"-"`, and unexported fields are ignored.

Use strings for identifiers and fixed-width numbers whose leading zeros matter. Signed and unsigned integers are formatted as decimal; the codec supplies any configured padding. Byte slices represent binary fields. Marshal omits empty strings and empty byte slices, but includes numeric zero. Unmarshal leaves missing fields unchanged, so use a fresh target unless you want merge behavior. Supported target values are strings, signed/unsigned integers, and byte slices; unsupported kinds return an error. Struct tags do not override field definitions.

See [struct_binding](../examples/struct_binding/main.go) for a runnable example, including composite paths.

## Handle errors without logging payment data

`PackError` and `UnpackError` carry field context and implement `Unwrap`, so use `errors.As` to inspect them and `errors.Is` to classify underlying sentinel errors. The executable `ExampleFromJSON` shows a JSON codec; `Example_errorHandling` shows typed failure handling.

`UnpackError.Error()` prints only field and byte offset. `RemainingHex` stores up to 16 payload bytes for deliberate debugging; do not log it in production. The `Err` field and custom packager errors may also contain sensitive information. Built-in field value errors avoid printing the input, but configuration names and file paths are not redacted. Changes to error wording are not a parsing contract.

`Dump(true)` masks selected fields and primitive tags 5A/57. It is not a complete redaction policy for private fields, composite contexts, or constructed TLV data. Prefer logging selected metadata such as MTI and STAN. Never log raw wire bytes, track data, PIN blocks, or a full application struct.

## Ownership and concurrency

Compiled codecs have no mutators. Built-in field packagers are configured at construction; custom packagers must remain immutable for concurrent use. Compilation snapshots the configuration map but does not clone arbitrary packager implementations.

Message methods lock the field map. `Get`, `GetBytes`, and `GetTLV` can expose mutable references; `SetComponent`, `SetBytes`, and `SetTLV` can retain caller-owned data. Parsed binary/TLV values may alias the input buffer. Keep the input alive and unchanged while using the parsed message. `GetHeader`, `Bitmap`, `Fields`, and `Clone` copy their data; `Clone` includes nested components and TLV values.

Use one goroutine to own a message while building or packing it. Do not mutate shared components, slices, or TLV nodes during inspection or packing. Use `Clone` before independently changing a message. The library does not promise an atomic snapshot across simultaneous setters and serialization.

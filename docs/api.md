# API Reference

This guide groups public APIs by the task they perform. Package docs on
[pkg.go.dev](https://pkg.go.dev/github.com/nqd2/igo8583) remain the canonical
reference for signatures.

## Root package: `igo8583`

| API | Use |
|---|---|
| `NewMessage()` | Create an empty message |
| `ISO87ASCII()`, `ISO87Binary()` | Use an ISO 8583:1987 field layout |
| `ISO93ASCII()`, `ISO93Binary()` | Use an ISO 8583:1993 field layout |
| `FromJSON([]byte)` | Compile a layout from JSON data |
| `FromJSONFile(path)` | Load and compile a layout from a JSON file |
| `NewBuilder(name)` | Create a layout with Go builder calls |
| `Marshal`, `Unmarshal` | Convert between structs and messages |
| `PackStruct`, `UnpackStruct` | Combine struct binding and wire encoding |

The root package also aliases `Message`, `Codec`, `Packager`, `BasePackager`,
and `Builder` for convenience. Focused subpackages remain available for direct
use and extension.

## Message construction

`message.Message` exposes MTI and header access, string/byte/TLV field
operations, composite subfield helpers, `Fields`, `Bitmap`, `Clone`, and
`Unset`. `Get` and typed getters return components or nested values according
to the ownership rules in [integration](integration.md#ownership-and-concurrency).

## Layout definition

Use `spec.FromJSON` or `spec.FromJSONFile` for declarative layouts. Use
`spec.NewBuilder` to configure bitmap format, header length, length prefix,
and fields in Go. Both produce a compiled `packager.Codec` or an error.

The codec provides `Pack`, `Unpack`, `PackField`, `HeaderLength`, and
`BitmapFormat`. `PackField` encodes a single field using the same definition
as complete messages, which can help a dialect build authentication input.

## Extension interfaces

| Interface | Responsibility |
|---|---|
| `field.FieldPackager` | Encode and decode one component |
| `prefixer.Prefixer` | Encode and decode a length prefix |
| `interpreter.Interpreter` | Convert logical units to and from bytes |
| `padder.Padder` | Align a fixed-width value |
| `packager.Packager` | Pack and unpack complete messages |

The `spec` package resolves built-in strategy names for JSON. A builder can
accept a custom `FieldPackager` directly. Adding a custom JSON strategy name
requires a resolver in the library's `spec` package.

## Errors

`packager.PackError` and `packager.UnpackError` retain field context and
implement error unwrapping. Use `errors.As` to inspect typed details and
`errors.Is` for sentinel errors. Do not parse error strings as a stable API.

## Related guides

- [Integration](integration.md): layout semantics, framing, field units,
  struct binding, errors, and data ownership.
- [Maintaining](maintaining.md): package contracts, extension rules, and
  release verification.
- [Examples](../examples): runnable code for built-in and custom layouts.

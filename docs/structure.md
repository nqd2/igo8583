# Project Structure

`igo8583` is a Go module at `github.com/nqd2/igo8583`. It targets Go 1.22
and uses only the standard library. The module exposes focused packages for
message values and wire encoding. It does not open sockets or implement a
payment network protocol.

## Packages

| Path | Purpose |
|---|---|
| `.` | Convenient aliases, constructors, standard layout factories, and struct binding |
| `bitmap/` | Primary, secondary, and tertiary bitmap operations |
| `field/` | Component types and field-level packagers |
| `interpreter/` | Conversion between logical text/numbers and wire bytes |
| `message/` | Message state, field access, composite access, struct binding, and safe dumps |
| `packager/` | Layout compilation and complete message pack/unpack |
| `padder/` | Fixed-width value alignment |
| `prefixer/` | Fixed and variable field length encodings |
| `spec/` | JSON definitions and fluent builder |
| `specs/` | Built-in ISO 8583:1987 and ISO 8583:1993 layouts |
| `tlv/` | BER-TLV node and map operations |
| `internal/diagnostic/` | Redacted error wrapping for internal use |
| `examples/` | Runnable examples and JSON fixtures |
| `docs/` | Integration, maintenance, architecture, feature, API, and structure guides |

Custom network dialects belong to the consuming application's module, for
example `your-project/specs/napas`. That package can depend on `igo8583`, load
its own JSON field layout, and define network-specific validation. It should
not add a nested `go.mod` unless the dialect is intended for independent
versioning and reuse.

## Configuration and examples

- `go.mod` defines the public module path and minimum Go language version.
- `README.md` is the package landing page shown on pkg.go.dev.
- `examples/json_specs/` contains declarative field layouts.
- `examples/*/main.go` demonstrates message packing, custom layouts, and
  struct binding.
- `docs/integration.md` explains field units, framing, composite data,
  ownership, and error handling.
- `docs/maintaining.md` describes internal contracts and release checks.

Run examples and Go commands from this module directory. `igo8583` itself has
no executable entry point; applications import it as a dependency.

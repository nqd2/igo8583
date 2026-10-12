# igo8583

[![Go Reference](https://pkg.go.dev/badge/github.com/nqd2/igo8583.svg)](https://pkg.go.dev/github.com/nqd2/igo8583)
[![Go Report Card](https://goreportcard.com/badge/github.com/nqd2/igo8583)](https://goreportcard.com/report/github.com/nqd2/igo8583)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`igo8583` encodes and decodes complete ISO 8583 messages in Go. The library
provides ISO 8583:1987 and ISO 8583:1993 ASCII and binary layouts, configurable
field encoders, JSON and Go builder APIs, struct binding, composite fields, and
BER-TLV support. It uses only the Go standard library.

The package handles message representation and wire encoding. Your application
owns TCP framing, transaction state, network-specific required fields,
validation, routing, retries, and cryptographic key management.

## Install

Requires Go 1.22 or newer.

```sh
go get github.com/nqd2/igo8583
```

## Pack and unpack

```go
package main

import (
	"fmt"
	"log"

	"github.com/nqd2/igo8583"
)

func main() {
	codec := igo8583.ISO87ASCII()
	request := igo8583.NewMessage()
	request.SetMTI("0200")
	request.Set(3, "000000")
	request.Set(4, "000000100000") // 1,000.00 when DE4 has two implied decimals
	request.Set(11, "000123")

	wire, err := codec.Pack(request)
	if err != nil { log.Fatal(err) }

	parsed, err := codec.Unpack(wire)
	if err != nil { log.Fatal(err) }
	stan, _ := parsed.GetString(11)
	fmt.Printf("MTI=%s STAN=%s (%d bytes)\n", parsed.GetMTI(), stan, len(wire))
}
```

`Pack` produces one complete message. `Unpack` requires exactly one complete
message and rejects trailing bytes. If a transport sends a length header or
TPDU, configure it in the codec and read the complete frame from the stream
before calling `Unpack`.

## Custom dialects

Keep each network dialect in the application that uses it. A dialect package
can own its wire specification, codec factory, and transaction validation:

```text
/your-project/
├── cmd/server/
└── specs/
    └── napas/
        ├── napas_spec.json
        ├── napas.go
        └── validate.go
```

`igo8583` does not contain a NAPAS module or NAPAS transaction rules. The
`specs/napas` directory is an ordinary package inside the application's Go
module. The runnable [custom dialect example](examples/custom_dialect) shows
the same pattern without a nested Go module.

A dialect package loads and compiles its field definition with the library:

```go
package napas

import (
	_ "embed"

	"github.com/nqd2/igo8583"
)

//go:embed napas_spec.json
var specJSON []byte

/* Codec compiles and returns this application's NAPAS wire layout. */
func Codec() (*igo8583.Codec, error) {
	return igo8583.FromJSON(specJSON)
}
```

The JSON and builder APIs validate the layout: field IDs, lengths, encoders,
prefixes, padding, MTI definition, header size, and framing configuration.
They do not validate transaction requirements such as which fields an ATM
withdrawal needs or which DE 39 values NAPAS accepts. Implement those rules in
`specs/napas/validate.go`, next to the dialect definition.

See [custom JSON](examples/json_specs/custom.json),
[custom dialect with validation](examples/custom_dialect),
[custom format](examples/custom_format/main.go), and
[custom switch](examples/custom_switch/main.go) for runnable examples.

## Standard layouts

| Factory | Message format |
|---|---|
| `ISO87ASCII` | ISO 8583:1987 fields with ASCII bitmap and ASCII encoded fields |
| `ISO87Binary` | ISO 8583:1987 fields with binary bitmap and packed numeric fields |
| `ISO93ASCII` | ISO 8583:1993 fields with ASCII encoding |
| `ISO93Binary` | ISO 8583:1993 fields with binary encoding |

These layouts describe field encoding. Confirm the peer's field profile and
framing before using one. In particular, the ISO87 ASCII and binary layouts
encode binary fields differently. DE55 is not automatically parsed as TLV
unless the chosen definition declares it as a TLV field.

## API map

| Package | Responsibility |
|---|---|
| `igo8583` | Common types, layout factories, JSON loading, struct binding |
| `message` | In-memory MTI, header, fields, bitmap, composite fields |
| `spec` | JSON definitions and fluent builder |
| `packager` | Compile layouts; pack and unpack complete messages |
| `field` | String, binary, amount, composite, and TLV field encoders |
| `prefixer` | Fixed and variable length prefixes |
| `interpreter` | ASCII, BCD, binary, hex, and EBCDIC conversion |
| `padder` | Fixed-width field padding |
| `bitmap` | Primary, secondary, and tertiary bitmap blocks |
| `tlv` | BER-TLV decode, encode, and tree access |
| `specs` | Built-in ISO layout definitions |

See [API reference](docs/api.md), [architecture](docs/architecture.md),
[features](docs/features.md), and [project structure](docs/structure.md).
The [integration guide](docs/integration.md) covers JSON options, framing,
field lengths, ownership, and error handling. The [maintainer guide](docs/maintaining.md)
describes extension points and compatibility expectations.

## Concurrency and ownership

Compiled `Codec` values have no configuration mutators and can be shared across
goroutines when custom field packagers are immutable. `Message` synchronizes
its field map, but `Get`, `GetBytes`, and `GetTLV` can expose mutable references;
`SetComponent`, `SetBytes`, and `SetTLV` can retain caller-owned values. Use one
goroutine to modify a message at a time. Call `Clone` before making independent
changes to a message or its nested components.

## Development

Run commands from the module directory:

```sh
go test ./...
go test -race ./...
go vet ./...
go run ./examples/test_pack
go run ./examples/custom_format
```

The module has no third-party dependencies. See [maintaining](docs/maintaining.md)
for fuzz targets, benchmarks, CI checks, and release guidance.

## License

[MIT](LICENSE), copyright 2026 nqd2.

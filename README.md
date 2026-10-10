# igo8583

`igo8583` packs and unpacks ISO 8583 messages in Go. It provides ASCII and binary presets, custom field layouts, struct binding, composite fields, and BER-TLV support. It uses only the Go standard library.

## Install

Requires Go 1.22 or newer.

```sh
go get github.com/nqd2/igo8583
```

## Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/nqd2/igo8583"
)

func main() {
    codec := igo8583.ISO87ASCII()
    msg := igo8583.NewMessage()
    msg.SetMTI("0200")
    msg.Set(11, "000042")

    wire, err := codec.Pack(msg)
    if err != nil {
        log.Fatal(err)
    }
    decoded, err := codec.Unpack(wire)
    if err != nil {
        log.Fatal(err)
    }
    stan, _ := decoded.GetString(11)
    fmt.Printf("MTI=%s STAN=%s\n", decoded.GetMTI(), stan)
}
```

This is a serialization example using synthetic data. A network's required fields and transaction rules belong in your application.

## Documentation

- [Integration guide](docs/integration.md): presets, custom formats, binding, framing, and error handling.
- [Maintainer guide](docs/maintaining.md): codec contracts, extension points, and verification.
- [Examples](examples): run from the repository root, for example `go run ./examples/custom_format`.

Match your peer's field definitions and framing before using a preset. The library does not manage connections, routing, retries, or transaction state. Use separate messages per request; compiled codecs can be shared when custom packagers are immutable.

## Development

```sh
go test ./...
go vet ./...
go test -race ./...
go test -run '^$' -bench . -benchmem ./...
```

Benchmark results depend on the machine, message, and toolchain.

## License

[MIT](LICENSE), copyright 2026 nqd2.

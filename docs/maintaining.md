# Maintaining igo8583

## Package boundaries

The root package exposes convenience aliases, constructors, preset factories, and struct binding. `message` stores field components; it does not know the peer's field layouts. `spec` compiles JSON and builder definitions into `packager.Codec`. `specs` contains the built-in profiles.

The codec coordinates MTI, header, bitmap, and field order. Each `field.FieldPackager` defines a field's wire layout using a padder, interpreter, and prefixer. `bitmap` handles indicator bits and blocks. `tlv` handles BER-TLV values independently of the message format.

## Packing and unpacking

Packing checks the configured header and MTI, checks active field IDs, generates the bitmap, and encodes data fields in ascending order. Framing is added last and counts body bytes. It does not validate transaction requirements.

Unpacking validates framing, reads header and MTI, decodes the bitmap, then visits the indicated fields in ascending order. Each packager returns a component and bytes consumed. A failure becomes `UnpackError` with the field and absolute byte offset. Extra bytes are rejected.

Preserve valid wire bytes when refactoring. Use manually calculated synthetic fixtures as well as round-trip tests: two matching encoder/decoder mistakes can pass a round trip.

## Extending field encoding

Implement `field.FieldPackager`: `Pack(*Component)`, `Unpack(data, offset)`, `Length()`, and `Description()`. Unpack's offset is relative to the supplied buffer; its consumed count is the number of bytes starting at that offset, not an absolute position. Reject invalid offsets and truncated input with errors before slicing. Return zero consumed bytes on failure.

A packager may implement `Validate() error`; codec compilation calls it to check static configuration. Register it with `Builder.Field`. It must not change configuration during Pack/Unpack. Constructors for composites copy the subfield map but keep packager references.

For new interpreters, `PackedLength` maps nonnegative logical units to wire bytes. `Unpack` must reject a negative count and report the wire bytes consumed. For new prefixers, round-trip the length, check capacity, and distinguish prefix size from body units. Keep errors classifiable with `%w` and avoid including field values or invalid bytes in messages.

Adding a Go extension does not add a JSON name automatically. Update the resolver and its rejection tests if it should be available in JSON. Document its units and odd-length behavior.

## Local verification

```sh
go test ./...
go vet ./...
go test -race ./...
go test -run '^$' -bench . -benchmem ./...
```

Check formatting with `gofmt -l` on the public Go files. CI checks Go 1.22.12 and the current stable toolchain, tests and vet on Linux/Windows, and race tests on Linux.

Run the executable examples from the repository root:

```sh
go run ./examples/test_pack
go run ./examples/custom_switch
go run ./examples/struct_binding
go run ./examples/custom_format
```

Exercise the fuzz targets separately:

```sh
go test . -run '^$' -fuzz '^FuzzUnpackISO87$' -fuzztime 10s
go test ./prefixer -run '^$' -fuzz '^FuzzLengthPrefixes$' -fuzztime 10s
go test ./tlv -run '^$' -fuzz '^FuzzDecode$' -fuzztime 10s
```

Fuzzing checks malformed wire handling and stable re-encoding; it does not prove interoperability. Test any new preset against independently calculated field bytes. Add regression tests for the original failure before changing implementation.

## Compatibility and releases

Keep public signatures and valid wire representations stable. Rejecting malformed input and removing payloads from diagnostic strings are intentional hardening changes; do not use full error strings as a machine interface. Mutable getter ownership is part of the current API, so do not silently replace it with copying behavior.

Use synthetic data in tests and issues. Before a release, test a clean checkout containing exactly the files being published, confirm examples and documentation links, and run the minimum supported Go toolchain. The module has no external dependencies; adding one needs a concrete use case. Keep the license in source distributions.

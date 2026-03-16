# Changelog

## v0.1.0

Initial release.

- Generic `Hash[T]`, `MustHash[T]`, and `HashAny` functions with FNV-1a 64-bit default hasher
- Reflection-based walker supporting all Go types (structs, maps, slices, pointers, interfaces, etc.)
- Struct tags: `imprint:"ignore"`, `imprint:"set"`, `imprint:"string"`, `imprint:"omitempty"`
- Functional options: `WithHasher`, `WithTagName`, `WithSlicesAsSets`, `WithZeroNil`, `WithIgnoreZeroValue`, `WithUseStringer`, `WithTypeHasher`
- `Hasher` interface for custom hash implementations
- Compatibility layer for mitchellh/hashstructure v2 (`FromHashOptions`, `HashOptions`)
- Requires Go 1.22+, zero dependencies
- 96.4% test coverage with 48+ tests and 5 benchmarks

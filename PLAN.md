# imprint — Modern Go Struct Hashing

**Replaces:** [mitchellh/hashstructure](https://github.com/mitchellh/hashstructure) (archived July 2024)
**Package:** `github.com/agentine/imprint`
**Language:** Go 1.21+
**License:** MIT
**Dependencies:** Zero

## Why

mitchellh/hashstructure is archived (July 2024), single maintainer (mitchellh) no longer writing Go. 3,753 importers across v1 (2,961) and v2 (792). The "blessed fork" gohugoio/hashstructure has only 22 importers and 42 stars — negligible adoption. Most importers are stuck on the archived original.

### Problems with hashstructure

1. **Pre-generics API** — `func Hash(v interface{}, format Format, opts *HashOptions) (uint64, error)` loses type safety
2. **Options struct** — `HashOptions` with direct fields, not extensible without breaking changes
3. **Format versioning** — FormatV1 has known hash collision issues but remains available; users must remember to pick V2
4. **Global behavior** — `UseStringer`, `ZeroNil` etc. are all-or-nothing; no per-field control
5. **Requires source modification** — `Hashable`, `Includable`, `IncludableMap` interfaces require modifying types you may not own
6. **Slow paths** — uses `fmt.Sprintf` for some type hashing

## Architecture

### Package Structure

```
imprint/
├── hash.go          # Hash[T], MustHash[T], HashAny
├── walk.go          # Reflection-based value traversal
├── options.go       # Functional options (Option type, With* functions)
├── tags.go          # Struct tag parsing (imprint:"...")
├── compat.go        # hashstructure v2 compatibility layer
├── hash_test.go     # Unit tests
├── bench_test.go    # Benchmarks
├── go.mod
├── LICENSE
└── README.md
```

### Core API

```go
// Primary API — generic, type-safe
func Hash[T any](v T, opts ...Option) (uint64, error)
func MustHash[T any](v T, opts ...Option) uint64

// Compatibility — drop-in for hashstructure.Hash
func HashAny(v any, opts ...Option) (uint64, error)
```

### Options (Functional Pattern)

```go
type Option func(*config)

func WithHasher(h hash.Hash64) Option        // Default: FNV-1a
func WithTagName(name string) Option          // Default: "imprint"
func WithSlicesAsSets(b bool) Option          // Order-independent slice hashing
func WithZeroNil(b bool) Option              // nil and zero values hash the same
func WithIgnoreZeroValue(b bool) Option       // Skip zero-value fields
func WithUseStringer(b bool) Option           // Hash fmt.Stringer output
func WithTypeHasher(t reflect.Type, fn TypeHasherFunc) Option  // Per-type custom hasher
```

```go
type TypeHasherFunc func(v any, h hash.Hash64) error
```

### Struct Tags

Tag name: `imprint` (configurable via WithTagName)

```go
type Example struct {
    Name    string
    UUID    string    `imprint:"ignore"`   // Excluded from hash
    Tags    []string  `imprint:"set"`      // Order-independent
    Created time.Time `imprint:"string"`   // Hash via String()
    Temp    int       `imprint:"-"`        // Alias for ignore
}
```

Supported tag values: `ignore` / `-`, `set`, `string`, `omitempty` (skip if zero value).

### Interfaces

```go
// Hasher allows types to provide their own hash value.
type Hasher interface {
    ImprintHash() (uint64, error)
}
```

### Compatibility Layer (compat.go)

Drop-in migration path from hashstructure v2:

```go
// Wraps the generic API for hashstructure users
func HashAny(v any, opts ...Option) (uint64, error)

// Convert hashstructure.HashOptions to functional options
func FromHashOptions(hopts *HashOptions) []Option

// HashOptions mirrors hashstructure.HashOptions for migration
type HashOptions struct {
    Hasher          hash.Hash64
    TagName         string
    ZeroNil         bool
    IgnoreZeroValue bool
    SlicesAsSets    bool
    UseStringer     bool
}
```

Migration example:
```go
// Before (hashstructure)
hash, err := hashstructure.Hash(myStruct, hashstructure.FormatV2, nil)

// After (imprint)
hash, err := imprint.Hash(myStruct)

// Before (hashstructure with options)
hash, err := hashstructure.Hash(myStruct, hashstructure.FormatV2, &hashstructure.HashOptions{
    SlicesAsSets: true,
    TagName: "hash",
})

// After (imprint)
hash, err := imprint.Hash(myStruct, imprint.WithSlicesAsSets(true), imprint.WithTagName("hash"))
```

## Hashing Algorithm

- Default hasher: FNV-1a (64-bit), same as hashstructure v2 default
- Use hashstructure v2 algorithm (not v1) to avoid collision issues
- Type dispatch via `reflect.Kind`: primitives write binary encoding, pointers dereference and recurse, structs iterate exported fields, maps sort keys then hash k/v pairs, slices/arrays iterate or use set mode, interfaces unwrap underlying value
- `Hasher` interface checked first for custom hash override
- Unexported struct fields are skipped

## Deliverables

1. **Phase 1:** Module scaffolding — go.mod, package structure, LICENSE, README
2. **Phase 2:** Core Hash/MustHash functions with reflection walker, all Kind dispatch
3. **Phase 3:** Functional options, struct tag parsing, custom type hashers
4. **Phase 4:** Compatibility layer (HashAny, FromHashOptions, HashOptions struct)
5. **Phase 5:** Comprehensive tests and benchmarks (target >90% coverage, benchmark vs hashstructure v2)

# imprint

Modern Go struct hashing with generics. Drop-in replacement for [mitchellh/hashstructure](https://github.com/mitchellh/hashstructure).

## Features

- **Generic API** — `Hash[T]` provides type safety without `interface{}`
- **Functional options** — extensible configuration without breaking changes
- **Struct tags** — per-field control: `ignore`, `set`, `string`, `omitempty`
- **Custom hashers** — implement `Hasher` interface or register per-type functions
- **Zero dependencies** — only the Go standard library
- **Migration path** — compatibility layer for hashstructure v2 users

## Install

```bash
go get github.com/agentine/imprint
```

## Usage

```go
import "github.com/agentine/imprint"

type User struct {
    Name  string
    Email string
    Tags  []string `imprint:"set"`
    Token string   `imprint:"ignore"`
}

// Type-safe hashing
hash, err := imprint.Hash(User{Name: "alice", Email: "a@b.com"})

// Panic on error
hash := imprint.MustHash(User{Name: "alice"})

// With options
hash, err := imprint.Hash(user,
    imprint.WithSlicesAsSets(true),
    imprint.WithUseStringer(true),
)
```

## Options

| Function | Default | Description |
|---|---|---|
| `WithHasher(h)` | FNV-1a 64 | Custom `hash.Hash64` implementation |
| `WithTagName(s)` | `"imprint"` | Struct tag name to read |
| `WithSlicesAsSets(b)` | `false` | Order-independent slice hashing |
| `WithZeroNil(b)` | `false` | Treat nil and zero values the same |
| `WithIgnoreZeroValue(b)` | `false` | Skip zero-value fields |
| `WithUseStringer(b)` | `false` | Hash `fmt.Stringer` output |
| `WithTypeHasher(t, fn)` | — | Per-type custom hash function |

## Struct Tags

```go
type Example struct {
    Name    string
    UUID    string    `imprint:"ignore"`     // Excluded from hash
    Tags    []string  `imprint:"set"`        // Order-independent
    Created time.Time `imprint:"string"`     // Hash via String()
    Temp    int       `imprint:"-"`          // Alias for ignore
    Cache   string    `imprint:"omitempty"`  // Skip if zero value
}
```

## Migration from hashstructure

```go
// Before (hashstructure)
hash, err := hashstructure.Hash(myStruct, hashstructure.FormatV2, nil)

// After (imprint)
hash, err := imprint.Hash(myStruct)

// Before (hashstructure with options)
hash, err := hashstructure.Hash(myStruct, hashstructure.FormatV2, &hashstructure.HashOptions{
    SlicesAsSets: true,
    TagName:      "hash",
})

// After (imprint with options)
hash, err := imprint.Hash(myStruct, imprint.WithSlicesAsSets(true), imprint.WithTagName("hash"))

// Or convert existing HashOptions
opts := imprint.FromHashOptions(&imprint.HashOptions{SlicesAsSets: true})
hash, err := imprint.Hash(myStruct, opts...)
```

## License

MIT

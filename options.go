package imprint

import (
	"hash"
	"hash/fnv"
	"reflect"
)

// TypeHasherFunc is a function that writes a custom hash for a value into the hasher.
type TypeHasherFunc func(v any, h hash.Hash64) error

// Option configures hashing behavior.
type Option func(*config)

type config struct {
	hasher          hash.Hash64
	tagName         string
	slicesAsSets    bool
	zeroNil         bool
	ignoreZeroValue bool
	useStringer     bool
	typeHashers     map[reflect.Type]TypeHasherFunc
}

func defaultConfig() *config {
	return &config{
		hasher:  fnv.New64a(),
		tagName: "imprint",
	}
}

func applyOptions(opts []Option) *config {
	c := defaultConfig()
	for _, o := range opts {
		o(c)
	}
	return c
}

// WithHasher sets a custom hash.Hash64 implementation. Default is FNV-1a 64-bit.
func WithHasher(h hash.Hash64) Option {
	return func(c *config) { c.hasher = h }
}

// WithTagName sets the struct tag name to read. Default is "imprint".
func WithTagName(name string) Option {
	return func(c *config) { c.tagName = name }
}

// WithSlicesAsSets enables order-independent slice hashing.
func WithSlicesAsSets(b bool) Option {
	return func(c *config) { c.slicesAsSets = b }
}

// WithZeroNil treats nil and zero values as equivalent for hashing.
func WithZeroNil(b bool) Option {
	return func(c *config) { c.zeroNil = b }
}

// WithIgnoreZeroValue skips zero-value struct fields during hashing.
func WithIgnoreZeroValue(b bool) Option {
	return func(c *config) { c.ignoreZeroValue = b }
}

// WithUseStringer hashes fmt.Stringer output instead of the underlying value.
func WithUseStringer(b bool) Option {
	return func(c *config) { c.useStringer = b }
}

// WithTypeHasher registers a custom hash function for a specific type.
func WithTypeHasher(t reflect.Type, fn TypeHasherFunc) Option {
	return func(c *config) {
		if c.typeHashers == nil {
			c.typeHashers = make(map[reflect.Type]TypeHasherFunc)
		}
		c.typeHashers[t] = fn
	}
}

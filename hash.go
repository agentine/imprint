package imprint

import "reflect"

// Hash computes a deterministic uint64 hash for any value.
func Hash[T any](v T, opts ...Option) (uint64, error) {
	cfg := applyOptions(opts)
	w := newWalker(cfg)
	if err := w.walk(reflect.ValueOf(&v).Elem()); err != nil {
		return 0, err
	}
	return w.h.Sum64(), nil
}

// MustHash is like Hash but panics on error.
func MustHash[T any](v T, opts ...Option) uint64 {
	h, err := Hash(v, opts...)
	if err != nil {
		panic(err)
	}
	return h
}

// HashAny computes a hash for any value using an interface{} parameter.
// This is the compatibility entry point for hashstructure users.
func HashAny(v any, opts ...Option) (uint64, error) {
	cfg := applyOptions(opts)
	w := newWalker(cfg)
	if err := w.walk(reflect.ValueOf(v)); err != nil {
		return 0, err
	}
	return w.h.Sum64(), nil
}

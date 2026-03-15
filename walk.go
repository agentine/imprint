package imprint

import (
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"reflect"
	"sort"
)

// Hasher allows types to provide their own hash value.
type Hasher interface {
	ImprintHash() (uint64, error)
}

var hasherType = reflect.TypeOf((*Hasher)(nil)).Elem()

type walker struct {
	h   hash.Hash64
	cfg *config
	buf [8]byte
}

func newWalker(cfg *config) *walker {
	return &walker{h: cfg.hasher, cfg: cfg}
}

func (w *walker) writeUint64(v uint64) {
	binary.LittleEndian.PutUint64(w.buf[:], v)
	_, _ = w.h.Write(w.buf[:])
}

func (w *walker) writeString(s string) {
	w.writeUint64(uint64(len(s)))
	_, _ = w.h.Write([]byte(s))
}

func (w *walker) walk(val reflect.Value) error {
	// Handle invalid (zero) reflect.Value
	if !val.IsValid() {
		w.writeUint64(0)
		return nil
	}

	// Check for type-specific hasher
	if w.cfg.typeHashers != nil {
		if fn, ok := w.cfg.typeHashers[val.Type()]; ok {
			var iface any
			if val.CanInterface() {
				iface = val.Interface()
			}
			return fn(iface, w.h)
		}
	}

	// Check Hasher interface
	if val.Type().Implements(hasherType) && val.CanInterface() {
		h, err := val.Interface().(Hasher).ImprintHash()
		if err != nil {
			return err
		}
		w.writeUint64(h)
		return nil
	}
	if val.CanAddr() && val.Addr().Type().Implements(hasherType) && val.Addr().CanInterface() {
		h, err := val.Addr().Interface().(Hasher).ImprintHash()
		if err != nil {
			return err
		}
		w.writeUint64(h)
		return nil
	}

	// Check fmt.Stringer
	if w.cfg.useStringer {
		stringerType := reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
		if val.Type().Implements(stringerType) && val.CanInterface() {
			w.writeString(val.Interface().(fmt.Stringer).String())
			return nil
		}
		if val.CanAddr() && val.Addr().Type().Implements(stringerType) && val.Addr().CanInterface() {
			w.writeString(val.Addr().Interface().(fmt.Stringer).String())
			return nil
		}
	}

	return w.walkKind(val)
}

func (w *walker) walkKind(val reflect.Value) error {
	kind := val.Kind()

	switch kind {
	case reflect.Bool:
		if val.Bool() {
			w.writeUint64(1)
		} else {
			w.writeUint64(0)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		w.writeUint64(uint64(val.Int()))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		w.writeUint64(val.Uint())

	case reflect.Float32:
		w.writeUint64(uint64(math.Float32bits(float32(val.Float()))))

	case reflect.Float64:
		w.writeUint64(math.Float64bits(val.Float()))

	case reflect.Complex64:
		c := complex64(val.Complex())
		w.writeUint64(uint64(math.Float32bits(real(c))))
		w.writeUint64(uint64(math.Float32bits(imag(c))))

	case reflect.Complex128:
		c := val.Complex()
		w.writeUint64(math.Float64bits(real(c)))
		w.writeUint64(math.Float64bits(imag(c)))

	case reflect.String:
		w.writeString(val.String())

	case reflect.Ptr:
		if val.IsNil() {
			if w.cfg.zeroNil {
				return w.walk(reflect.Zero(val.Type().Elem()))
			}
			w.writeUint64(0)
			return nil
		}
		return w.walk(val.Elem())

	case reflect.Interface:
		if val.IsNil() {
			w.writeUint64(0)
			return nil
		}
		// Hash the type name so different types with same value hash differently
		w.writeString(val.Elem().Type().String())
		return w.walk(val.Elem())

	case reflect.Struct:
		return w.walkStruct(val)

	case reflect.Map:
		return w.walkMap(val)

	case reflect.Slice:
		if val.IsNil() {
			if w.cfg.zeroNil {
				w.writeUint64(0)
				return nil
			}
			w.writeUint64(0)
			return nil
		}
		return w.walkSlice(val, w.cfg.slicesAsSets)

	case reflect.Array:
		return w.walkSlice(val, w.cfg.slicesAsSets)

	default:
		return fmt.Errorf("imprint: unsupported kind %s", kind)
	}

	return nil
}

func (w *walker) walkStruct(val reflect.Value) error {
	t := val.Type()
	for i := range t.NumField() {
		f := t.Field(i)

		// Skip unexported fields
		if !f.IsExported() {
			continue
		}

		fieldVal := val.Field(i)

		// Parse struct tags
		tag := f.Tag.Get(w.cfg.tagName)
		opts := parseTag(tag)

		if opts.ignore {
			continue
		}

		if opts.omitempty && fieldVal.IsZero() {
			continue
		}

		if w.cfg.ignoreZeroValue && fieldVal.IsZero() {
			continue
		}

		// Hash field name for structural identity
		w.writeString(f.Name)

		if opts.str {
			stringerType := reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
			if fieldVal.Type().Implements(stringerType) && fieldVal.CanInterface() {
				w.writeString(fieldVal.Interface().(fmt.Stringer).String())
				continue
			}
			if fieldVal.CanAddr() && fieldVal.Addr().Type().Implements(stringerType) && fieldVal.Addr().CanInterface() {
				w.writeString(fieldVal.Addr().Interface().(fmt.Stringer).String())
				continue
			}
		}

		if opts.set && (fieldVal.Kind() == reflect.Slice || fieldVal.Kind() == reflect.Array) {
			if err := w.walkSlice(fieldVal, true); err != nil {
				return err
			}
			continue
		}

		if err := w.walk(fieldVal); err != nil {
			return err
		}
	}
	return nil
}

func (w *walker) walkMap(val reflect.Value) error {
	if val.IsNil() {
		w.writeUint64(0)
		return nil
	}

	w.writeUint64(uint64(val.Len()))

	// Sort map keys for deterministic ordering
	keys := val.MapKeys()
	sortedKeys := make([]mapEntry, len(keys))
	for i, k := range keys {
		kh := w.hashValue(k)
		sortedKeys[i] = mapEntry{key: k, hash: kh}
	}
	sort.Slice(sortedKeys, func(i, j int) bool {
		return sortedKeys[i].hash < sortedKeys[j].hash
	})

	for _, entry := range sortedKeys {
		if err := w.walk(entry.key); err != nil {
			return err
		}
		if err := w.walk(val.MapIndex(entry.key)); err != nil {
			return err
		}
	}
	return nil
}

type mapEntry struct {
	key  reflect.Value
	hash uint64
}

func (w *walker) hashValue(val reflect.Value) uint64 {
	sub := newWalker(w.cfg)
	_ = sub.walk(val)
	return sub.h.Sum64()
}

func (w *walker) walkSlice(val reflect.Value, asSet bool) error {
	n := val.Len()
	w.writeUint64(uint64(n))

	if asSet {
		// Hash each element independently and XOR for order independence
		var combined uint64
		for i := range n {
			combined ^= w.hashValue(val.Index(i))
		}
		w.writeUint64(combined)
		return nil
	}

	for i := range n {
		if err := w.walk(val.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

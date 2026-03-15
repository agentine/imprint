package imprint

import (
	"fmt"
	"hash"
	"hash/fnv"
	"reflect"
	"testing"
)

// --- helpers ---

type customHasher struct{ val uint64 }

func (c customHasher) ImprintHash() (uint64, error) { return c.val, nil }

type stringerType struct{ s string }

func (s stringerType) String() string { return s.s }

type testStruct struct {
	Name  string
	Age   int
	Email string
}

type taggedStruct struct {
	Name    string
	Secret  string   `imprint:"ignore"`
	Dash    string   `imprint:"-"`
	Tags    []string `imprint:"set"`
	Label   stringerType `imprint:"string"`
	Temp    int      `imprint:"omitempty"`
}

type unexportedStruct struct {
	Public  string
	private string //nolint:unused
}

type nestedStruct struct {
	Inner testStruct
	Value int
}

// --- Primitive tests ---

func TestHashBool(t *testing.T) {
	h1, err := Hash(true)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := Hash(false)
	if err != nil {
		t.Fatal(err)
	}
	if h1 == h2 {
		t.Error("true and false should have different hashes")
	}
	// Deterministic
	h3, _ := Hash(true)
	if h1 != h3 {
		t.Error("same value should produce same hash")
	}
}

func TestHashIntegers(t *testing.T) {
	h1, _ := Hash(42)
	h2, _ := Hash(43)
	if h1 == h2 {
		t.Error("different ints should have different hashes")
	}
	h3, _ := Hash(int8(42))
	h4, _ := Hash(int16(42))
	h5, _ := Hash(int32(42))
	h6, _ := Hash(int64(42))
	// All same underlying value but different types go through Hash[T] which uses reflect
	_ = h3
	_ = h4
	_ = h5
	_ = h6
}

func TestHashUnsigned(t *testing.T) {
	h1, _ := Hash(uint(10))
	h2, _ := Hash(uint(20))
	if h1 == h2 {
		t.Error("different uints should have different hashes")
	}
	_, _ = Hash(uint8(1))
	_, _ = Hash(uint16(1))
	_, _ = Hash(uint32(1))
	_, _ = Hash(uint64(1))
	_, _ = Hash(uintptr(1))
}

func TestHashFloat(t *testing.T) {
	h1, _ := Hash(3.14)
	h2, _ := Hash(2.71)
	if h1 == h2 {
		t.Error("different floats should have different hashes")
	}
	_, _ = Hash(float32(1.5))
}

func TestHashComplex(t *testing.T) {
	h1, _ := Hash(complex(1.0, 2.0))
	h2, _ := Hash(complex(2.0, 1.0))
	if h1 == h2 {
		t.Error("different complex should have different hashes")
	}
	_, _ = Hash(complex64(complex(1.0, 2.0)))
}

func TestHashString(t *testing.T) {
	h1, _ := Hash("hello")
	h2, _ := Hash("world")
	if h1 == h2 {
		t.Error("different strings should have different hashes")
	}
	h3, _ := Hash("hello")
	if h1 != h3 {
		t.Error("same string should produce same hash")
	}
	// Empty string
	_, err := Hash("")
	if err != nil {
		t.Fatal(err)
	}
}

// --- Pointer tests ---

func TestHashPointerSingle(t *testing.T) {
	x := 42
	h1, _ := Hash(&x)
	h2, _ := Hash(42)
	// Pointer to 42 and 42 should hash same (pointer is dereferenced)
	if h1 != h2 {
		t.Error("pointer to value should hash same as value")
	}
}

func TestHashPointerDouble(t *testing.T) {
	x := 42
	p := &x
	h1, _ := Hash(&p)
	h2, _ := Hash(42)
	if h1 != h2 {
		t.Error("double pointer should dereference to same hash")
	}
}

func TestHashPointerNil(t *testing.T) {
	var p *int
	h1, _ := Hash(p)
	_ = h1 // should not panic
}

// --- Struct tests ---

func TestHashStruct(t *testing.T) {
	s1 := testStruct{Name: "alice", Age: 30, Email: "a@b.com"}
	s2 := testStruct{Name: "bob", Age: 25, Email: "b@b.com"}
	h1, _ := Hash(s1)
	h2, _ := Hash(s2)
	if h1 == h2 {
		t.Error("different structs should have different hashes")
	}
	h3, _ := Hash(s1)
	if h1 != h3 {
		t.Error("same struct should produce same hash")
	}
}

func TestHashStructUnexported(t *testing.T) {
	s1 := unexportedStruct{Public: "hello"}
	s2 := unexportedStruct{Public: "hello"}
	h1, _ := Hash(s1)
	h2, _ := Hash(s2)
	if h1 != h2 {
		t.Error("unexported fields should be skipped, same exported = same hash")
	}
}

func TestHashNestedStruct(t *testing.T) {
	s := nestedStruct{Inner: testStruct{Name: "a", Age: 1}, Value: 10}
	h1, _ := Hash(s)
	s.Inner.Name = "b"
	h2, _ := Hash(s)
	if h1 == h2 {
		t.Error("different nested values should change hash")
	}
}

// --- Map tests ---

func TestHashMap(t *testing.T) {
	m1 := map[string]int{"a": 1, "b": 2}
	m2 := map[string]int{"b": 2, "a": 1}
	h1, _ := Hash(m1)
	h2, _ := Hash(m2)
	if h1 != h2 {
		t.Error("maps with same content in different order should hash the same")
	}
}

func TestHashMapDifferent(t *testing.T) {
	m1 := map[string]int{"a": 1}
	m2 := map[string]int{"a": 2}
	h1, _ := Hash(m1)
	h2, _ := Hash(m2)
	if h1 == h2 {
		t.Error("maps with different values should have different hashes")
	}
}

func TestHashMapNil(t *testing.T) {
	var m map[string]int
	_, err := Hash(m)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHashMapEmpty(t *testing.T) {
	m := map[string]int{}
	_, err := Hash(m)
	if err != nil {
		t.Fatal(err)
	}
}

// --- Slice/Array tests ---

func TestHashSlice(t *testing.T) {
	s1 := []int{1, 2, 3}
	s2 := []int{1, 2, 3}
	h1, _ := Hash(s1)
	h2, _ := Hash(s2)
	if h1 != h2 {
		t.Error("same slices should produce same hash")
	}
}

func TestHashSliceOrderMatters(t *testing.T) {
	s1 := []int{1, 2, 3}
	s2 := []int{3, 2, 1}
	h1, _ := Hash(s1)
	h2, _ := Hash(s2)
	if h1 == h2 {
		t.Error("different order slices should have different hashes by default")
	}
}

func TestHashSliceNil(t *testing.T) {
	var s []int
	_, err := Hash(s)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHashSliceEmpty(t *testing.T) {
	s := []int{}
	_, err := Hash(s)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHashArray(t *testing.T) {
	a1 := [3]int{1, 2, 3}
	a2 := [3]int{1, 2, 3}
	h1, _ := Hash(a1)
	h2, _ := Hash(a2)
	if h1 != h2 {
		t.Error("same arrays should produce same hash")
	}
}

// --- Interface tests ---

func TestHashInterface(t *testing.T) {
	var i1 any = 42
	var i2 any = 42
	h1, _ := HashAny(i1)
	h2, _ := HashAny(i2)
	if h1 != h2 {
		t.Error("same interface values should produce same hash")
	}
}

func TestHashInterfaceNil(t *testing.T) {
	var i any
	_, err := HashAny(i)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHashInterfaceDifferentTypes(t *testing.T) {
	var i1 any = 42
	var i2 any = "42"
	h1, _ := HashAny(i1)
	h2, _ := HashAny(i2)
	if h1 == h2 {
		t.Error("different types in interface should have different hashes")
	}
}

// --- Struct tag tests ---

func TestTagIgnore(t *testing.T) {
	s1 := taggedStruct{Name: "a", Secret: "x"}
	s2 := taggedStruct{Name: "a", Secret: "y"}
	h1, _ := Hash(s1)
	h2, _ := Hash(s2)
	if h1 != h2 {
		t.Error("ignored field should not affect hash")
	}
}

func TestTagDash(t *testing.T) {
	s1 := taggedStruct{Name: "a", Dash: "x"}
	s2 := taggedStruct{Name: "a", Dash: "y"}
	h1, _ := Hash(s1)
	h2, _ := Hash(s2)
	if h1 != h2 {
		t.Error("dash-tagged field should not affect hash")
	}
}

func TestTagSet(t *testing.T) {
	s1 := taggedStruct{Name: "a", Tags: []string{"x", "y"}}
	s2 := taggedStruct{Name: "a", Tags: []string{"y", "x"}}
	h1, _ := Hash(s1)
	h2, _ := Hash(s2)
	if h1 != h2 {
		t.Error("set-tagged slice should be order-independent")
	}
}

func TestTagString(t *testing.T) {
	s1 := taggedStruct{Name: "a", Label: stringerType{"hello"}}
	s2 := taggedStruct{Name: "a", Label: stringerType{"hello"}}
	h1, _ := Hash(s1)
	h2, _ := Hash(s2)
	if h1 != h2 {
		t.Error("string-tagged field should produce consistent hash")
	}
	s3 := taggedStruct{Name: "a", Label: stringerType{"world"}}
	h3, _ := Hash(s3)
	if h1 == h3 {
		t.Error("different stringer values should produce different hashes")
	}
}

func TestTagOmitempty(t *testing.T) {
	s1 := taggedStruct{Name: "a", Temp: 0}
	s2 := taggedStruct{Name: "a", Temp: 42}
	h1, _ := Hash(s1)
	h2, _ := Hash(s2)
	if h1 == h2 {
		t.Error("non-zero omitempty field should affect hash")
	}
	// When zero, field is skipped - verify by checking against struct without the field contribution
	s3 := taggedStruct{Name: "a"}
	h3, _ := Hash(s3)
	if h1 != h3 {
		t.Error("zero omitempty field should be skipped")
	}
}

// --- Hasher interface tests ---

func TestHasherInterface(t *testing.T) {
	c1 := customHasher{val: 100}
	c2 := customHasher{val: 200}
	h1, _ := Hash(c1)
	h2, _ := Hash(c2)
	if h1 == h2 {
		t.Error("different Hasher values should produce different hashes")
	}
}

func TestHasherInterfaceInStruct(t *testing.T) {
	type s struct {
		C customHasher
	}
	v1 := s{C: customHasher{val: 42}}
	v2 := s{C: customHasher{val: 42}}
	h1, _ := Hash(v1)
	h2, _ := Hash(v2)
	if h1 != h2 {
		t.Error("same Hasher in struct should produce same hash")
	}
}

// --- Option tests ---

func TestWithSlicesAsSets(t *testing.T) {
	s1 := []int{1, 2, 3}
	s2 := []int{3, 2, 1}
	h1, _ := Hash(s1, WithSlicesAsSets(true))
	h2, _ := Hash(s2, WithSlicesAsSets(true))
	if h1 != h2 {
		t.Error("slices as sets should be order-independent")
	}
}

func TestWithZeroNil(t *testing.T) {
	var p *int
	zero := 0
	h1, _ := Hash(p, WithZeroNil(true))
	h2, _ := Hash(&zero, WithZeroNil(true))
	if h1 != h2 {
		t.Error("nil and zero should hash the same with ZeroNil")
	}
}

func TestWithIgnoreZeroValue(t *testing.T) {
	s1 := testStruct{Name: "alice", Age: 0}
	s2 := testStruct{Name: "alice", Age: 0}
	h1, _ := Hash(s1, WithIgnoreZeroValue(true))
	h2, _ := Hash(s2, WithIgnoreZeroValue(true))
	if h1 != h2 {
		t.Error("same struct with ignored zero values should produce same hash")
	}
}

func TestWithUseStringer(t *testing.T) {
	s := stringerType{s: "test"}
	h1, _ := Hash(s, WithUseStringer(true))
	h2, _ := Hash(s, WithUseStringer(false))
	if h1 == h2 {
		t.Error("stringer vs non-stringer should produce different hashes")
	}
}

func TestWithTagName(t *testing.T) {
	type custom struct {
		Name   string
		Secret string `custom:"ignore"`
	}
	s1 := custom{Name: "a", Secret: "x"}
	s2 := custom{Name: "a", Secret: "y"}

	// Default tag name: fields not ignored
	h1, _ := Hash(s1)
	h2, _ := Hash(s2)
	if h1 == h2 {
		t.Error("without custom tag, Secret should be hashed")
	}

	// Custom tag name: fields ignored
	h3, _ := Hash(s1, WithTagName("custom"))
	h4, _ := Hash(s2, WithTagName("custom"))
	if h3 != h4 {
		t.Error("with custom tag, Secret should be ignored")
	}
}

func TestWithHasher(t *testing.T) {
	h := fnv.New64()
	v, err := Hash("test", WithHasher(h))
	if err != nil {
		t.Fatal(err)
	}
	if v == 0 {
		t.Error("hash should be non-zero")
	}
}

func TestWithTypeHasher(t *testing.T) {
	type special struct{ X int }
	fn := func(v any, h hash.Hash64) error {
		_, err := h.Write([]byte("custom"))
		return err
	}
	h1, _ := Hash(special{X: 1}, WithTypeHasher(reflect.TypeOf(special{}), fn))
	h2, _ := Hash(special{X: 2}, WithTypeHasher(reflect.TypeOf(special{}), fn))
	if h1 != h2 {
		t.Error("custom type hasher should override default behavior")
	}
}

// --- MustHash tests ---

func TestMustHash(t *testing.T) {
	h := MustHash("hello")
	if h == 0 {
		t.Error("MustHash should return non-zero for non-empty string")
	}
}

func TestMustHashDeterministic(t *testing.T) {
	h1 := MustHash(testStruct{Name: "a", Age: 1})
	h2 := MustHash(testStruct{Name: "a", Age: 1})
	if h1 != h2 {
		t.Error("MustHash should be deterministic")
	}
}

// --- HashAny tests ---

func TestHashAny(t *testing.T) {
	h1, _ := HashAny(42)
	h2, _ := HashAny(42)
	if h1 != h2 {
		t.Error("HashAny should be deterministic")
	}
}

func TestHashAnyNil(t *testing.T) {
	_, err := HashAny(nil)
	if err != nil {
		t.Fatal(err)
	}
}

// --- Compat layer tests ---

func TestFromHashOptionsNil(t *testing.T) {
	opts := FromHashOptions(nil)
	if opts != nil {
		t.Error("nil input should return nil")
	}
}

func TestFromHashOptionsAll(t *testing.T) {
	hopts := &HashOptions{
		Hasher:          fnv.New64(),
		TagName:         "hash",
		ZeroNil:         true,
		IgnoreZeroValue: true,
		SlicesAsSets:    true,
		UseStringer:     true,
	}
	opts := FromHashOptions(hopts)
	if len(opts) != 6 {
		t.Errorf("expected 6 options, got %d", len(opts))
	}
	// Verify they apply correctly
	cfg := applyOptions(opts)
	if cfg.tagName != "hash" {
		t.Error("TagName not applied")
	}
	if !cfg.zeroNil {
		t.Error("ZeroNil not applied")
	}
	if !cfg.ignoreZeroValue {
		t.Error("IgnoreZeroValue not applied")
	}
	if !cfg.slicesAsSets {
		t.Error("SlicesAsSets not applied")
	}
	if !cfg.useStringer {
		t.Error("UseStringer not applied")
	}
}

func TestFromHashOptionsPartial(t *testing.T) {
	hopts := &HashOptions{SlicesAsSets: true}
	opts := FromHashOptions(hopts)
	if len(opts) != 1 {
		t.Errorf("expected 1 option, got %d", len(opts))
	}
}

func TestFromHashOptionsUsedWithHash(t *testing.T) {
	s1 := []int{1, 2, 3}
	s2 := []int{3, 2, 1}
	opts := FromHashOptions(&HashOptions{SlicesAsSets: true})
	h1, _ := Hash(s1, opts...)
	h2, _ := Hash(s2, opts...)
	if h1 != h2 {
		t.Error("FromHashOptions SlicesAsSets should work with Hash")
	}
}

// --- Error path tests ---

func TestMustHashPanicsOnError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustHash should panic on error")
		}
	}()
	MustHash(make(chan int))
}

func TestHashUnsupportedKind(t *testing.T) {
	_, err := Hash(make(chan int))
	if err == nil {
		t.Error("channel should return error")
	}
}

func TestHasherInterfaceError(t *testing.T) {
	type errHasher struct{}
	// Use HashAny with a type that will fail
	_, err := HashAny(make(chan int))
	if err == nil {
		t.Error("expected error for unsupported type")
	}
}

// pointer-receiver Hasher
type ptrHasher struct{ val uint64 }

func (p *ptrHasher) ImprintHash() (uint64, error) { return p.val, nil }

func TestHasherPointerReceiver(t *testing.T) {
	type s struct {
		H ptrHasher
	}
	v1 := s{H: ptrHasher{val: 42}}
	v2 := s{H: ptrHasher{val: 42}}
	h1, _ := Hash(v1)
	h2, _ := Hash(v2)
	if h1 != h2 {
		t.Error("pointer-receiver Hasher in struct should work")
	}
}

// pointer-receiver Stringer
type ptrStringer struct{ s string }

func (p *ptrStringer) String() string { return p.s }

func TestStringerPointerReceiver(t *testing.T) {
	type s struct {
		Label ptrStringer `imprint:"string"`
	}
	v1 := s{Label: ptrStringer{"hello"}}
	v2 := s{Label: ptrStringer{"hello"}}
	h1, _ := Hash(v1)
	h2, _ := Hash(v2)
	if h1 != h2 {
		t.Error("pointer-receiver Stringer should work with string tag")
	}
}

func TestUseStringerPointerReceiver(t *testing.T) {
	type s struct {
		Label ptrStringer
	}
	v := s{Label: ptrStringer{"test"}}
	h1, _ := Hash(v, WithUseStringer(true))
	h2, _ := Hash(v, WithUseStringer(true))
	if h1 != h2 {
		t.Error("UseStringer with pointer receiver should be deterministic")
	}
}

func TestHashInterfaceInStruct(t *testing.T) {
	type s struct {
		I any
	}
	v1 := s{I: 42}
	v2 := s{I: nil}
	h1, _ := Hash(v1)
	h2, _ := Hash(v2)
	if h1 == h2 {
		t.Error("interface field with value vs nil should differ")
	}
}

func TestHashNilSliceZeroNil(t *testing.T) {
	var s []int
	_, err := Hash(s, WithZeroNil(true))
	if err != nil {
		t.Fatal(err)
	}
}

func TestHashNilPointerZeroNilFalse(t *testing.T) {
	var p *int
	_, err := Hash(p, WithZeroNil(false))
	if err != nil {
		t.Fatal(err)
	}
}

func TestHasherErrorPropagation(t *testing.T) {
	type errHasher struct{}
	// Not implementing Hasher, just testing error return from type hasher
	fn := func(v any, h hash.Hash64) error {
		return fmt.Errorf("custom error")
	}
	type s struct{ X int }
	_, err := Hash(s{X: 1}, WithTypeHasher(reflect.TypeOf(s{}), fn))
	if err == nil || err.Error() != "custom error" {
		t.Error("type hasher error should propagate")
	}
}

// --- Edge cases ---

func TestHashEmptyStruct(t *testing.T) {
	type empty struct{}
	_, err := Hash(empty{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestHashEmptySlice(t *testing.T) {
	h1, _ := Hash([]int{})
	h2, _ := Hash([]string{})
	_ = h1
	_ = h2
}

func TestHashLargeMap(t *testing.T) {
	m := make(map[int]int)
	for i := range 1000 {
		m[i] = i * 2
	}
	h1, _ := Hash(m)
	h2, _ := Hash(m)
	if h1 != h2 {
		t.Error("large map should hash deterministically")
	}
}

func TestHashStructWithMapField(t *testing.T) {
	type s struct {
		M map[string]int
	}
	v1 := s{M: map[string]int{"a": 1, "b": 2}}
	v2 := s{M: map[string]int{"b": 2, "a": 1}}
	h1, _ := Hash(v1)
	h2, _ := Hash(v2)
	if h1 != h2 {
		t.Error("struct with map field should be order-independent on map keys")
	}
}

func TestHashSliceOfStructs(t *testing.T) {
	s1 := []testStruct{{Name: "a"}, {Name: "b"}}
	s2 := []testStruct{{Name: "a"}, {Name: "b"}}
	h1, _ := Hash(s1)
	h2, _ := Hash(s2)
	if h1 != h2 {
		t.Error("same slice of structs should produce same hash")
	}
}

func TestParseTag(t *testing.T) {
	tests := []struct {
		tag  string
		want tagOpts
	}{
		{"ignore", tagOpts{ignore: true}},
		{"-", tagOpts{ignore: true}},
		{"set", tagOpts{set: true}},
		{"string", tagOpts{str: true}},
		{"omitempty", tagOpts{omitempty: true}},
		{"set,omitempty", tagOpts{set: true, omitempty: true}},
		{"", tagOpts{}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("tag=%q", tt.tag), func(t *testing.T) {
			got := parseTag(tt.tag)
			if got != tt.want {
				t.Errorf("parseTag(%q) = %+v, want %+v", tt.tag, got, tt.want)
			}
		})
	}
}

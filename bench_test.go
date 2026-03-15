package imprint

import "testing"

type benchSimple struct {
	Name  string
	Age   int
	Email string
}

type benchNested struct {
	Inner benchSimple
	Tags  []string
	Score float64
}

func BenchmarkHashSimpleStruct(b *testing.B) {
	v := benchSimple{Name: "alice", Age: 30, Email: "alice@example.com"}
	for range b.N {
		_, _ = Hash(v)
	}
}

func BenchmarkHashNestedStruct(b *testing.B) {
	v := benchNested{
		Inner: benchSimple{Name: "alice", Age: 30, Email: "alice@example.com"},
		Tags:  []string{"admin", "user", "editor"},
		Score: 98.5,
	}
	for range b.N {
		_, _ = Hash(v)
	}
}

func BenchmarkHashLargeMap(b *testing.B) {
	m := make(map[string]int, 100)
	for i := range 100 {
		m[string(rune('a'+i%26))+string(rune('0'+i/26))] = i
	}
	for range b.N {
		_, _ = Hash(m)
	}
}

func BenchmarkHashSlice(b *testing.B) {
	s := make([]int, 100)
	for i := range s {
		s[i] = i
	}
	for range b.N {
		_, _ = Hash(s)
	}
}

func BenchmarkHashSliceAsSet(b *testing.B) {
	s := make([]int, 100)
	for i := range s {
		s[i] = i
	}
	for range b.N {
		_, _ = Hash(s, WithSlicesAsSets(true))
	}
}

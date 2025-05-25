package helper

import (
	"cmp"
	"slices"
)

type KeyValueSorter[K cmp.Ordered, V any] struct {
	keys   []K
	values []V
}

func NewKeyValueSorter[K cmp.Ordered, V any](cap int) *KeyValueSorter[K, V] {
	var keys []K
	var values []V

	if cap > 0 {
		keys = make([]K, 0, cap)
		values = make([]V, 0, cap)
	}

	return &KeyValueSorter[K, V]{keys: keys, values: values}
}

func (s *KeyValueSorter[K, V]) Len() int {
	return len(s.keys)
}

func (s *KeyValueSorter[K, V]) Cap() int {
	return cap(s.keys)
}

// TODO add a comment
func (s *KeyValueSorter[K, V]) InsertAndShift(key K, value V) {
	if len(s.keys) < cap(s.keys) {
		s.keys = append(s.keys, key)
		s.values = append(s.values, value)
	}
	if len(s.keys) == cap(s.keys) && key > s.keys[len(s.keys)-1] {
		return
	}

	n, found := slices.BinarySearch(s.keys, key)
	if found || n-1 == cap(s.keys) {
		s.keys[n], s.values[n] = key, value
		return
	}

	if n >= cap(s.keys) {
		return
	}

	copy(s.keys[n+1:], s.keys[n:])
	copy(s.values[n+1:], s.values[n:])
	s.keys[n], s.values[n] = key, value
}

func (s *KeyValueSorter[K, V]) Values() []V {
	return s.values
}

func (s *KeyValueSorter[K, V]) LastKey() K {
	if len(s.keys) == 0 {
		panic("empty sorter")
	}

	return s.keys[len(s.keys)-1]
}

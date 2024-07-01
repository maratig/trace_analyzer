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

func (s *KeyValueSorter[K, V]) Swap(i, j int) {
	if i >= len(s.keys) || j >= len(s.keys) {
		panic("keys out of range")
	}

	s.keys[i], s.keys[j] = s.keys[j], s.keys[i]
	s.values[i], s.values[j] = s.values[j], s.values[i]
}

func (s *KeyValueSorter[K, V]) Less(i, j int) bool {
	if i >= len(s.keys) || j >= len(s.keys) {
		panic("keys out of range")
	}
	return s.keys[i] < s.keys[j]
}

func (s *KeyValueSorter[K, V]) Add(key K, value V) {
	s.keys = append(s.keys, key)
	s.values = append(s.values, value)
}

// InsertAndShift assumes that values are sorted already. It inserts k, v to the appropriate position and removes
// the last item (ie item with the biggest key)
func (s *KeyValueSorter[K, V]) InsertAndShift(key K, value V) {
	if len(s.keys) == 0 {
		s.Add(key, value)
		return
	}

	n, found := slices.BinarySearch(s.keys, key)
	if found {
		s.keys[n], s.values[n] = key, value
		return
	}

	if n >= len(s.keys) {
		return
	}

	if n == len(s.keys)-1 {
		s.keys[n], s.values[n] = key, value
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

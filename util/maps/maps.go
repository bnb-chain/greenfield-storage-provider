package maps

import (
	"sort"

	"golang.org/x/exp/constraints"
)

// SortKeys sort keys of a map
func SortKeys[M ~map[K]V, K constraints.Ordered, V any](m M) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sortSlice(keys)
	return keys
}

// sortSlice return the sort items slice by key
func sortSlice[T constraints.Ordered](s []T) {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}

// ValueToSlice convert values of a map to a slice in order
func ValueToSlice[M ~map[K]V, K constraints.Ordered, V any](m M) []V {
	keys := SortKeys(m)
	s := make([]V, len(m))
	for i, k := range keys {
		s[i] = m[k]
	}
	return s
}

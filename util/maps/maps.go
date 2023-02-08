package maps

import (
	"sort"

	"github.com/bnb-chain/greenfield-storage-provider/util/log"
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

func sortSlice[T constraints.Ordered](s []T) {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}

// ValueToSlice convert values of a map to a slice
func ValueToSlice[M ~map[K]V, K constraints.Ordered, V any](m M) []V {
	keys := SortKeys(m)
	valueSlice := make([]V, len(m))
	log.Infow("jjj", "m length", len(m))
	for index, key := range keys {
		value := m[key]
		valueSlice[index] = value
		//valueSlice = append(valueSlice, value)
	}
	log.Infow("iii", "valueSlice length", len(valueSlice))
	return valueSlice
}

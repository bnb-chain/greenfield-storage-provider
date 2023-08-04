package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortKeys(t *testing.T) {
	slice := SortKeys(map[string]struct{}{
		"c": {},
		"a": {},
		"b": {},
	})
	assert.Equal(t, []string{"a", "b", "c"}, slice)
}

func Test_sortSlice(t *testing.T) {
	slice := []int{1, 3, 2, 9, 6}
	sortSlice(slice)
	assert.Equal(t, []int{1, 2, 3, 6, 9}, slice)
}

func TestValueToSlice(t *testing.T) {
	slice := ValueToSlice(map[string]int{
		"c": 10,
		"a": 7,
		"b": 4,
	})
	assert.Equal(t, []int{7, 4, 10}, slice)
}

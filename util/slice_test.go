package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainOnlyOneDifferentElement(t *testing.T) {
	cases := []struct {
		name        string
		slice1      []uint32
		slice2      []uint32
		wantedIndex int
	}{
		{
			"There are only one different element",
			[]uint32{1, 2, 3, 4, 5, 6},
			[]uint32{1, 2, 3, 7, 5, 6},
			3,
		},
		{
			"There are two different elements",
			[]uint32{1, 2, 3, 4, 5, 6},
			[]uint32{1, 2, 3, 7, 8, 6},
			-1,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			index := ContainOnlyOneDifferentElement(tt.slice1, tt.slice2)
			assert.Equal(t, tt.wantedIndex, index)
		})
	}
}

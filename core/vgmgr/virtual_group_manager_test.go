package vgmgr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPickVGFFilter_Check(t *testing.T) {
	filter := NewPickVGFFilter([]uint32{1, 2})
	result := filter.Check(1)
	assert.Equal(t, true, result)
}

func TestNewExcludeIDFilter(t *testing.T) {
	filter := NewExcludeIDFilter(IDSet{})
	result := filter.Apply(1)
	assert.Equal(t, false, result)
}

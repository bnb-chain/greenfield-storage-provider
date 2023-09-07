package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalVirtualGroup_TableName(t *testing.T) {
	localVirtualGroup := LocalVirtualGroup{ID: 1}
	name := localVirtualGroup.TableName()
	assert.Equal(t, LocalVirtualGroupTableName, name)
}

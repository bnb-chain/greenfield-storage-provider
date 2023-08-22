package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalVirtualGroup_TableName(t *testing.T) {
	globalVirtualGroup := GlobalVirtualGroup{ID: 1}
	name := globalVirtualGroup.TableName()
	assert.Equal(t, GlobalVirtualGroupTableName, name)
}

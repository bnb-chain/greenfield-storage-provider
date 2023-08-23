package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalVirtualGroupFamily_TableName(t *testing.T) {
	globalVirtualGroupFamily := GlobalVirtualGroupFamily{ID: 1}
	name := globalVirtualGroupFamily.TableName()
	assert.Equal(t, GlobalVirtualGroupFamilyTableName, name)
}

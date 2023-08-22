package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermission_TableName(t *testing.T) {
	permission := Permission{ID: 1}
	name := permission.TableName()
	assert.Equal(t, PermissionTableName, name)
}

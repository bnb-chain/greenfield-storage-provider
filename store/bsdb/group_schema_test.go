package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroup_TableName(t *testing.T) {
	group := Group{ID: 1}
	name := group.TableName()
	assert.Equal(t, GroupTableName, name)
}

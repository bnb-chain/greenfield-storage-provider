package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlashPrefixTreeNode_TableName(t *testing.T) {
	slashPrefixTreeNode := SlashPrefixTreeNode{ID: 1}
	name := slashPrefixTreeNode.TableName()
	assert.Equal(t, PrefixTreeTableName, name)
}

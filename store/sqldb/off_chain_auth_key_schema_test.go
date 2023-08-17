package sqldb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOffChainAuthKeyTable_TableName(t *testing.T) {
	table := OffChainAuthKeyTable{UserAddress: mockUser}
	result := table.TableName()
	assert.Equal(t, OffChainAuthKeyTableName, result)
}

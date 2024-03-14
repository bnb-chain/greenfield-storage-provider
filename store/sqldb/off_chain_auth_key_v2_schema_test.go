package sqldb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOffChainAuthKeyV2Table_TableName(t *testing.T) {
	table := OffChainAuthKeyV2Table{UserAddress: mockUser}
	result := table.TableName()
	assert.Equal(t, OffChainAuthKeyV2TableName, result)
}

package sqldb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpInfoTable_TableName(t *testing.T) {
	table := SpInfoTable{OperatorAddress: "mockOperatorAddress"}
	result := table.TableName()
	assert.Equal(t, SpInfoTableName, result)
}

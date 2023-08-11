package sqldb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUploadObjectProgressTable_TableName(t *testing.T) {
	table := UploadObjectProgressTable{ObjectID: 1}
	result := table.TableName()
	assert.Equal(t, UploadObjectProgressTableName, result)
}

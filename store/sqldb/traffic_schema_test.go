package sqldb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBucketTrafficTable_TableName(t *testing.T) {
	table := BucketTrafficTable{BucketID: 1}
	result := table.TableName()
	assert.Equal(t, BucketTrafficTableName, result)
}

func TestReadRecordTable_TableName(t *testing.T) {
	table := ReadRecordTable{ReadRecordID: 1}
	result := table.TableName()
	assert.Equal(t, ReadRecordTableName, result)
}

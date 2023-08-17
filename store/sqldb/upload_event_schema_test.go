package sqldb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPutObjectSuccessTable_TableName(t *testing.T) {
	table := PutObjectSuccessTable{ID: 1}
	result := table.TableName()
	assert.Equal(t, PutObjectSuccessTableName, result)
}

func TestPutObjectEventTable_TableName(t *testing.T) {
	table := PutObjectEventTable{ID: 1}
	result := table.TableName()
	assert.Equal(t, PutObjectEventTableName, result)
}

func TestUploadTimeoutTable_TableName(t *testing.T) {
	table := UploadTimeoutTable{ID: 1}
	result := table.TableName()
	assert.Equal(t, UploadTimeoutTableName, result)
}

func TestReplicateTimeoutTable_TableName(t *testing.T) {
	table := ReplicateTimeoutTable{ID: 1}
	result := table.TableName()
	assert.Equal(t, ReplicateTimeoutTableName, result)
}

func TestSealTimeoutTable_TableName(t *testing.T) {
	table := SealTimeoutTable{ID: 1}
	result := table.TableName()
	assert.Equal(t, SealTimeoutTableName, result)
}

func TestUploadFailedTable_TableName(t *testing.T) {
	table := UploadFailedTable{ID: 1}
	result := table.TableName()
	assert.Equal(t, UploadFailedTableName, result)
}

func TestReplicateFailedTable_TableName(t *testing.T) {
	table := ReplicateFailedTable{ID: 1}
	result := table.TableName()
	assert.Equal(t, ReplicateFailedTableName, result)
}

func TestSealFailedTable_TableName(t *testing.T) {
	table := SealFailedTable{ID: 1}
	result := table.TableName()
	assert.Equal(t, SealFailedTableName, result)
}

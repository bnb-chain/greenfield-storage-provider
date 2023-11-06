package bsdb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataMigrationRecord_TableName(t *testing.T) {
	dataMigrationRecord := DataMigrationRecord{ProcessKey: ProcessKeyUpdateBucketSize, IsCompleted: true}
	name := dataMigrationRecord.TableName()
	assert.Equal(t, DataMigrationRecordTableName, name)
}

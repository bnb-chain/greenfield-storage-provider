package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataMigrationRecord_TableName(t *testing.T) {
	dataMigrationRecord := DataMigrationRecord{ProcessKey: ProcessKeyUpdateBucketSize, IsCompleted: true}
	name := dataMigrationRecord.TableName()
	assert.Equal(t, DataMigrationRecordTableName, name)
}

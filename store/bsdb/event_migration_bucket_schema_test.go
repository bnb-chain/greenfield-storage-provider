package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventMigrationBucket_TableName(t *testing.T) {
	eventMigrationBucket := EventMigrationBucket{ID: 1}
	name := eventMigrationBucket.TableName()
	assert.Equal(t, EventMigrationTableName, name)
}

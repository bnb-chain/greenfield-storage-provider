package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventCancelMigrationBucket_TableName(t *testing.T) {
	eventCancelMigrationBucket := EventCancelMigrationBucket{ID: 1}
	name := eventCancelMigrationBucket.TableName()
	assert.Equal(t, EventCancelMigrationTableName, name)
}

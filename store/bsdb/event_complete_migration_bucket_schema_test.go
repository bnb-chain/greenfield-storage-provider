package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventCompleteMigrationBucket_TableName(t *testing.T) {
	eventCompleteMigrationBucket := EventCompleteMigrationBucket{ID: 1}
	name := eventCompleteMigrationBucket.TableName()
	assert.Equal(t, EventCompleteMigrationTableName, name)
}

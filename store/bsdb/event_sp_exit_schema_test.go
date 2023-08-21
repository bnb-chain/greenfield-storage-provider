package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventStorageProviderExit_TableName(t *testing.T) {
	eventStorageProviderExit := EventStorageProviderExit{ID: 1}
	name := eventStorageProviderExit.TableName()
	assert.Equal(t, EventStorageProviderExitTableName, name)
}

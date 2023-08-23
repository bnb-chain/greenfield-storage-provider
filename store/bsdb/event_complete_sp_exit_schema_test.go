package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventCompleteStorageProviderExit_TableName(t *testing.T) {
	eventCompleteStorageProviderExit := EventCompleteStorageProviderExit{ID: 1}
	name := eventCompleteStorageProviderExit.TableName()
	assert.Equal(t, EventCompleteStorageProviderExitTableName, name)
}

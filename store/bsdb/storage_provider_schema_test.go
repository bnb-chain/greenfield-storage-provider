package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorageProvider_TableName(t *testing.T) {
	storageProvider := StorageProvider{ID: 1}
	name := storageProvider.TableName()
	assert.Equal(t, StorageProviderTableName, name)
}

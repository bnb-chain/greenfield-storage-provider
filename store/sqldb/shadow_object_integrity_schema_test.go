package sqldb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShadowIntegrityMetaTable_TableName(t *testing.T) {
	i := ShadowIntegrityMetaTable{ObjectID: 1}
	result := i.TableName()
	assert.Equal(t, ShadowIntegrityMetaTableName, result)
}

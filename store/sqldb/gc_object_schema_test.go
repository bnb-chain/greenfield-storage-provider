package sqldb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGCObjectProgressTable_TableName(t *testing.T) {
	gc := GCObjectProgressTable{TaskKey: "mock"}
	name := gc.TableName()
	assert.Equal(t, GCObjectProgressTableName, name)
}

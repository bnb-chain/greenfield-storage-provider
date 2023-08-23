package bsdb

import (
	"testing"

	"github.com/forbole/juno/v4/common"
	"github.com/stretchr/testify/assert"
)

func TestObjectIDMap_TableName(t *testing.T) {
	objectIDMap := ObjectIDMap{ObjectID: common.HexToHash("0")}
	name := objectIDMap.TableName()
	assert.Equal(t, ObjectIDMapTableName, name)
}

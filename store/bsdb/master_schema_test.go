package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMasterDB_TableName(t *testing.T) {
	masterDB := MasterDB{IsMaster: true}
	name := masterDB.TableName()
	assert.Equal(t, MasterDBTableName, name)
}

package sqldb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrateSubscribeProgressTable_TableName(t *testing.T) {
	m := MigrateSubscribeProgressTable{EventName: "sp_exit"}
	result := m.TableName()
	assert.Equal(t, MigrateSubscribeProgressTableName, result)
}

func TestSwapOutTable_TableName(t *testing.T) {
	s := SwapOutTable{
		SwapOutKey: "swap",
		IsDestSP:   true,
	}
	result := s.TableName()
	assert.Equal(t, SwapOutTableName, result)
}

func TestMigrateGVGTable_TableName(t *testing.T) {
	m := MigrateGVGTable{MigrateKey: "migrate"}
	result := m.TableName()
	assert.Equal(t, MigrateGVGTableName, result)
}

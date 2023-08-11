package sqldb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPieceHashTable_TableName(t *testing.T) {
	p := PieceHashTable{ObjectID: 1}
	result := p.TableName()
	assert.Equal(t, PieceHashTableName, result)
}

func TestIntegrityMetaTable_TableName(t *testing.T) {
	i := IntegrityMetaTable{ObjectID: 1}
	result := i.TableName()
	assert.Equal(t, IntegrityMetaTableName, result)
}

func TestGetIntegrityMetasTableName(t *testing.T) {
	result := GetIntegrityMetasTableName(10)
	assert.Equal(t, "integrity_meta_00", result)
}

func TestGetIntegrityMetasShardNumberByBucketName(t *testing.T) {
	result := GetIntegrityMetasShardNumberByBucketName(27)
	assert.Equal(t, uint64(0), result)
}

func TestGetIntegrityMetasTableNameByShardNumber(t *testing.T) {
	result := GetIntegrityMetasTableNameByShardNumber(100)
	assert.Equal(t, "integrity_meta_100", result)
}

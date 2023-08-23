package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStreamRecord_TableName(t *testing.T) {
	streamRecord := StreamRecord{ID: 1}
	name := streamRecord.TableName()
	assert.Equal(t, StreamRecordTableName, name)
}

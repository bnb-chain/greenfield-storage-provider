package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObject_TableName(t *testing.T) {
	object := Object{ID: 1}
	name := object.TableName()
	assert.Equal(t, ObjectTableName, name)
}

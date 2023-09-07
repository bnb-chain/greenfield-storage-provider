package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatement_TableName(t *testing.T) {
	statement := Statement{ID: 1}
	name := statement.TableName()
	assert.Equal(t, StatementTableName, name)
}

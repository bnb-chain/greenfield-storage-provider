package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventSwapOut_TableName(t *testing.T) {
	eventSwapOut := EventSwapOut{ID: 1}
	name := eventSwapOut.TableName()
	assert.Equal(t, EventSwapOutTableName, name)
}

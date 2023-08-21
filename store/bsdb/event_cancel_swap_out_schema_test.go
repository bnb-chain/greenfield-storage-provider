package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventCancelSwapOut_TableName(t *testing.T) {
	eventCancelSwapOut := EventCancelSwapOut{ID: 1}
	name := eventCancelSwapOut.TableName()
	assert.Equal(t, EventCancelSwapOutTableName, name)
}

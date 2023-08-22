package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventCompleteSwapOut_TableName(t *testing.T) {
	eventCompleteSwapOut := EventCompleteSwapOut{ID: 1}
	name := eventCompleteSwapOut.TableName()
	assert.Equal(t, EventCompleteSwapOutTableName, name)
}

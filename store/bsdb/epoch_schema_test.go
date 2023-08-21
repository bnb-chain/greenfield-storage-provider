package bsdb

import (
	"testing"

	"github.com/forbole/juno/v4/common"
	"github.com/stretchr/testify/assert"
)

func TestEpoch_TableName(t *testing.T) {
	epoch := Epoch{
		OneRowID:    true,
		BlockHeight: 1000,
		BlockHash:   common.HexToHash("0"),
		UpdateTime:  1000,
	}
	name := epoch.TableName()
	assert.Equal(t, EpochTableName, name)
}

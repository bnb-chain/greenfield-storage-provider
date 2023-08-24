package receiver

import (
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfsptqueue"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/stretchr/testify/assert"
)

func TestNewReceiveModular(t *testing.T) {
	app := &gfspapp.GfSpBaseApp{}
	cfg := &gfspconfig.GfSpConfig{
		Customize: &gfspconfig.Customize{
			NewStrategyTQueueFunc: mockQueueOnStrategy,
		},
	}
	result, err := NewReceiveModular(app, cfg)
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func mockQueueOnStrategy(name string, cap int) taskqueue.TQueueOnStrategy {
	return gfsptqueue.NewGfSpTQueue(name, cap)
}

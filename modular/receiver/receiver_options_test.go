package receiver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zkMeLabs/mechain-storage-provider/base/gfspapp"
	"github.com/zkMeLabs/mechain-storage-provider/base/gfspconfig"
	"github.com/zkMeLabs/mechain-storage-provider/base/gfsptqueue"
	"github.com/zkMeLabs/mechain-storage-provider/core/taskqueue"
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

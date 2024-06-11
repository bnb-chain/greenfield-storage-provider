package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zkMeLabs/mechain-storage-provider/base/gfspapp"
	"github.com/zkMeLabs/mechain-storage-provider/base/gfspconfig"
)

func TestNewExecuteModular(t *testing.T) {
	app := &gfspapp.GfSpBaseApp{}
	cfg := &gfspconfig.GfSpConfig{}
	result, err := NewExecuteModular(app, cfg)
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

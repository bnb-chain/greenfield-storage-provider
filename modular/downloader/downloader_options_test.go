package downloader

import (
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/stretchr/testify/assert"
)

func TestNewDownloadModular(t *testing.T) {
	app := &gfspapp.GfSpBaseApp{}
	cfg := &gfspconfig.GfSpConfig{}
	result, err := NewDownloadModular(app, cfg)
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

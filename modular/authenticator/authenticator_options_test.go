package authenticator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
)

func TestNewAuthenticationModular(t *testing.T) {
	result, err := NewAuthenticationModular(&gfspapp.GfSpBaseApp{}, &gfspconfig.GfSpConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

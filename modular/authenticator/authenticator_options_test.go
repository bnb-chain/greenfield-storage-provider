package authenticator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zkMeLabs/mechain-storage-provider/base/gfspapp"
	"github.com/zkMeLabs/mechain-storage-provider/base/gfspconfig"
)

func TestNewAuthenticationModular(t *testing.T) {
	result, err := NewAuthenticationModular(&gfspapp.GfSpBaseApp{}, &gfspconfig.GfSpConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

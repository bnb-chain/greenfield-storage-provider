package gfspapp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
)

func TestGfSpBaseApp_GfSpSetResourceLimit(t *testing.T) {
	g := setup(t)
	result, err := g.GfSpSetResourceLimit(context.TODO(), &gfspserver.GfSpSetResourceLimitRequest{})
	assert.Nil(t, err)
	assert.Equal(t, ErrFutureSupport, result.GetErr())
}

func TestGfSpBaseApp_GfSpQueryResourceLimit(t *testing.T) {
	g := setup(t)
	result, err := g.GfSpQueryResourceLimit(context.TODO(), &gfspserver.GfSpQueryResourceLimitRequest{})
	assert.Nil(t, err)
	assert.Equal(t, ErrFutureSupport, result.GetErr())
}

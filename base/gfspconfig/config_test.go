package gfspconfig

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var mockErr = errors.New("mock error")

func TestGfSpConfig_ApplySuccess(t *testing.T) {
	cfg := &GfSpConfig{Env: "mainnet"}
	opt := func(cfg *GfSpConfig) error { return nil }
	err := cfg.Apply(opt)
	assert.Equal(t, nil, err)
}

func TestGfSpConfig_ApplyFailure(t *testing.T) {
	cfg := &GfSpConfig{Env: "mainnet"}
	opt := func(cfg *GfSpConfig) error { return mockErr }
	err := cfg.Apply(opt)
	assert.Equal(t, mockErr, err)
}

func TestGfSpConfig_StringSuccess(t *testing.T) {
	cfg := &GfSpConfig{Env: "mainnet"}
	result := cfg.String()
	assert.NotNil(t, result)
}

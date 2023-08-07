package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDefaultServerInterceptor(t *testing.T) {
	options := GetDefaultServerInterceptor()
	assert.Equal(t, 2, len(options))
}

func TestGetDefaultClientInterceptor(t *testing.T) {
	options := GetDefaultClientInterceptor()
	assert.Equal(t, 2, len(options))
}

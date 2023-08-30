package gfspp2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGfSpPing_GetSignBytes(t *testing.T) {
	m := &GfSpPing{
		SpOperatorAddress: "mockSpOperatorAddress",
		Signature:         []byte("mockSig"),
	}
	result := m.GetSignBytes()
	assert.NotNil(t, result)
}

func TestGetSignBytes(t *testing.T) {
	m := &GfSpPong{
		SpOperatorAddress: "mockSpOperatorAddress",
		Signature:         []byte("mockSig"),
	}
	result := m.GetSignBytes()
	assert.NotNil(t, result)
}

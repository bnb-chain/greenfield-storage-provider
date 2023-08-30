package gfsptask

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
)

func TestGfSpTask_Key(t *testing.T) {
	m := &GfSpTask{}
	m.Key()
}

func TestGfSpTask_Type(t *testing.T) {
	m := &GfSpTask{}
	result := m.Type()
	assert.Equal(t, coretask.TypeTaskUnknown, result)
}

func TestGfSpTask_Expired(t *testing.T) {
	m := &GfSpTask{
		Retry:    1,
		MaxRetry: 2,
	}
	result := m.Expired()
	assert.Equal(t, false, result)
}

func TestGfSpTask_EstimateLimit(t *testing.T) {
	m := &GfSpTask{}
	m.EstimateLimit()
}

func TestGfSpTask_Error1(t *testing.T) {
	m := &GfSpTask{Err: &gfsperrors.GfSpError{
		CodeSpace: gfsperrors.DefaultCodeSpace,
		InnerCode: 0,
	}}
	_ = m.Error()
}

func TestGfSpTask_Error2(t *testing.T) {
	m := &GfSpTask{}
	m.SetError(&gfsperrors.GfSpError{InnerCode: 1})
	_ = m.Error()
}

func TestLimitEstimateByPriority(t *testing.T) {
	LimitEstimateByPriority(200)
	LimitEstimateByPriority(100)
}

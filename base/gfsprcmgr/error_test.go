package gfsprcmgr

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
)

var mockErr = errors.New("mock error")

func TestErrMemoryLimitExceeded_Error(t *testing.T) {
	e := &ErrMemoryLimitExceeded{err: mockErr}
	_ = e.Error()
	_ = e.Unwrap()
}

func Test_logValuesMemoryLimit(t *testing.T) {
	err := &ErrMemoryLimitExceeded{
		current:   1,
		attempted: 2,
		limit:     3,
		priority:  4,
		err:       mockErr,
	}
	stat := corercmgr.ScopeStat{Memory: 1}
	result := logValuesMemoryLimit("scope", "edge", stat, err)
	assert.NotNil(t, result)
}

func TestErrConnLimitExceeded_Error(t *testing.T) {
	e := &ErrConnLimitExceeded{err: mockErr}
	_ = e.Error()
	_ = e.Unwrap()
}

func Test_logValuesConnLimit(t *testing.T) {
	err := &ErrConnLimitExceeded{
		current:   1,
		attempted: 2,
		limit:     3,
		err:       mockErr,
	}
	stat := corercmgr.ScopeStat{Memory: 1}
	result := logValuesConnLimit("scope", "edge", corercmgr.DirOutbound, stat, err)
	assert.NotNil(t, result)
}

func TestErrTaskLimitExceeded_Error(t *testing.T) {
	e := &ErrTaskLimitExceeded{err: mockErr}
	_ = e.Error()
	_ = e.Unwrap()
}

func Test_logValuesTaskLimit(t *testing.T) {
	err := &ErrTaskLimitExceeded{
		priority:  corercmgr.MaxLimitInt,
		current:   1,
		attempted: 2,
		limit:     3,
		err:       mockErr,
	}
	stat := corercmgr.ScopeStat{Memory: 1}
	result := logValuesTaskLimit("scope", "edge", stat, err)
	assert.NotNil(t, result)
}

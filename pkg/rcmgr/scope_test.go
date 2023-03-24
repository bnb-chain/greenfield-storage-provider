package rcmgr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceScopeSpan(t *testing.T) {
	limit := &BaseLimit{
		Conns:         1000,
		ConnsInbound:  1000,
		ConnsOutbound: 1000,
		FD:            1000,
		Memory:        1000,
	}
	parentScope := newResourceScope(limit, nil, "parent")
	child1Scope := newResourceScopeSpan(parentScope, parentScope.NextSpanID(), "child1")
	err := child1Scope.ReserveMemory(100, ReservationPriorityAlways)
	assert.NoError(t, err)
	assert.Equal(t, int64(100), child1Scope.rc.memory)
	assert.Equal(t, int64(100), parentScope.rc.memory)

	child2Scope := newResourceScopeSpan(parentScope, parentScope.NextSpanID(), "child2")
	err = child2Scope.ReserveMemory(300, ReservationPriorityHigh)
	assert.NoError(t, err)
	assert.Equal(t, int64(100), child1Scope.rc.memory)
	assert.Equal(t, int64(400), parentScope.rc.memory)

	err = parentScope.ReserveMemory(200, ReservationPriorityLow)
	assert.Equal(t, parentScope.wrapError(ErrResourceLimitExceeded).Error(), err.Error())

	child3Scope := newResourceScopeSpan(child2Scope, child2Scope.NextSpanID(), "child2")
	err = child3Scope.ReserveMemory(300, ReservationPriorityHigh)
	assert.NoError(t, err)
	assert.Equal(t, int64(300), child3Scope.rc.memory)
	assert.Equal(t, int64(600), child2Scope.rc.memory)
	assert.Equal(t, int64(700), parentScope.rc.memory)

	child2Scope.ReleaseMemory(100)
	assert.Equal(t, int64(500), child2Scope.rc.memory)
	assert.Equal(t, int64(600), parentScope.rc.memory)
	assert.Equal(t, int64(300), child3Scope.rc.memory)

	child2Scope.ReleaseMemory(200)
	assert.Equal(t, int64(300), child2Scope.rc.memory)
	assert.Equal(t, int64(400), parentScope.rc.memory)
	assert.Equal(t, int64(300), child3Scope.rc.memory)

	child3Scope.ReleaseMemory(300)
	assert.Equal(t, int64(0), child2Scope.rc.memory)
	assert.Equal(t, int64(100), parentScope.rc.memory)
	assert.Equal(t, int64(0), child3Scope.rc.memory)
}

func TestResourceScopeEdges(t *testing.T) {
	limit := &BaseLimit{
		Conns:         1000,
		ConnsInbound:  1000,
		ConnsOutbound: 1000,
		FD:            1000,
		Memory:        1000,
	}
	parentScope := newResourceScope(limit, nil, "parent")
	neighborScope := newResourceScope(limit, []*resourceScope{parentScope}, "neighbor")

	err := neighborScope.ReserveMemory(500, ReservationPriorityHigh)
	assert.NoError(t, err)
	assert.Equal(t, int64(500), parentScope.rc.memory)
	assert.Equal(t, int64(500), neighborScope.rc.memory)

	neighborChildScope := newResourceScopeSpan(neighborScope, neighborScope.NextSpanID(), "neighbor-child")
	err = neighborChildScope.ReserveMemory(300, ReservationPriorityAlways)
	assert.NoError(t, err)
	assert.Equal(t, int64(800), neighborScope.rc.memory)
	assert.Equal(t, int64(800), parentScope.rc.memory)

	childScope := newResourceScopeSpan(parentScope, parentScope.NextSpanID(), "child")
	err = childScope.ReserveMemory(100, ReservationPriorityAlways)
	assert.NoError(t, err)
	assert.Equal(t, int64(800), neighborScope.rc.memory)
	assert.Equal(t, int64(900), parentScope.rc.memory)

	neighborChildScope.ReleaseMemory(300)
	assert.Equal(t, int64(500), neighborScope.rc.memory)
	assert.Equal(t, int64(600), parentScope.rc.memory)

	neighborScope.ReleaseMemory(300)
	assert.Equal(t, int64(200), neighborScope.rc.memory)
	assert.Equal(t, int64(300), parentScope.rc.memory)

	childScope.ReleaseMemory(200)
	assert.Equal(t, int64(0), childScope.rc.memory)
	assert.Equal(t, int64(100), parentScope.rc.memory)

	neighborScope.ReleaseMemory(200)
	assert.Equal(t, int64(0), neighborScope.rc.memory)
	assert.Equal(t, int64(0), parentScope.rc.memory)
}

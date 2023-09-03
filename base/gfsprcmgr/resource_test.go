package gfsprcmgr

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
)

func setupResources(t *testing.T) *resources {
	return &resources{
		nconnsIn:     1,
		nconnsOut:    2,
		nfd:          3,
		ntasksHigh:   4,
		ntasksMedium: 5,
		ntasksLow:    6,
		memory:       7,
	}
}

func TestResources_remaining1(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().GetMemoryLimit().Return(int64(8)).Times(2)
	m.EXPECT().GetTaskTotalLimit().Return(16).Times(2)
	m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(10).Times(2)
	m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(10).Times(2)
	m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(10).Times(2)

	rc := setupResources(t)
	rc.limit = m
	result := rc.remaining()
	assert.Equal(t, int64(1), result.GetMemoryLimit())
}

func TestResources_remaining2(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().GetMemoryLimit().Return(int64(6)).Times(1)
	m.EXPECT().GetTaskTotalLimit().Return(10).Times(1)
	m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(3).Times(1)
	m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(3).Times(1)
	m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(3).Times(1)

	rc := setupResources(t)
	rc.limit = m
	result := rc.remaining()
	assert.Equal(t, int64(0), result.GetMemoryLimit())
}

func Test_addInt64WithOverflow(t *testing.T) {
	result, ok := addInt64WithOverflow(1, 2)
	assert.Equal(t, int64(3), result)
	assert.Equal(t, true, ok)
}

func Test_mulInt64WithOverflow(t *testing.T) {
	cases := []struct {
		name          string
		a, b          int64
		wantedResult1 int64
		wantedResult2 bool
	}{
		{
			name:          "1",
			a:             0,
			b:             0,
			wantedResult1: 0,
			wantedResult2: true,
		},
		{
			name:          "1",
			a:             math.MaxInt64,
			b:             2,
			wantedResult1: -2,
			wantedResult2: false,
		},
		{
			name:          "1",
			a:             2,
			b:             2,
			wantedResult1: 4,
			wantedResult2: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := mulInt64WithOverflow(tt.a, tt.b)
			assert.Equal(t, tt.wantedResult1, result)
			assert.Equal(t, tt.wantedResult2, ok)
		})
	}
}

func TestResources_checkMemory(t *testing.T) {
	cases := []struct {
		name      string
		rsvp      int64
		prio      uint8
		fn        func() *resources
		wantedErr error
	}{
		{
			name:      "1",
			rsvp:      -1,
			fn:        func() *resources { return setupResources(t) },
			wantedErr: errors.New("cannot reserve negative memory, rsvp=-1"),
		},
		{
			name: "memory limit is math.MaxInt64",
			rsvp: 1,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64)).Times(1)
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
			wantedErr: nil,
		},
		{
			name: "mul int64 is overflow",
			rsvp: 1,
			prio: 2,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64 - 1)).Times(1)
				rc := setupResources(t)
				rc.memory = math.MaxInt64
				rc.limit = m
				return rc
			},
			wantedErr: &ErrMemoryLimitExceeded{
				current:   9223372036854775807,
				attempted: 1,
				limit:     9223372036854775806,
				priority:  0x2,
				err:       ErrResourceLimitExceeded,
			},
		},
		{
			name: "mul int64 is not overflow",
			rsvp: 1,
			prio: 2,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64 - 1)).Times(1)
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
			wantedErr: nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn().checkMemory(tt.rsvp, tt.prio)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestResources_reserveMemory(t *testing.T) {
	cases := []struct {
		name      string
		size      int64
		prio      uint8
		fn        func() *resources
		wantedErr error
	}{
		{
			name: "mul int64 is overflow",
			size: 1,
			prio: 2,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64 - 1)).Times(1)
				rc := setupResources(t)
				rc.memory = math.MaxInt64
				rc.limit = m
				return rc
			},
			wantedErr: &ErrMemoryLimitExceeded{
				current:   9223372036854775807,
				attempted: 1,
				limit:     9223372036854775806,
				priority:  0x2,
				err:       ErrResourceLimitExceeded,
			},
		},
		{
			name: "mul int64 is not overflow",
			size: 1,
			prio: 2,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64 - 1)).Times(1)
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
			wantedErr: nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn().reserveMemory(tt.size, tt.prio)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestResources_releaseMemory(t *testing.T) {
	rc := setupResources(t)
	rc.releaseMemory(8)
}

func TestResources_addTask(t *testing.T) {
	cases := []struct {
		name string
		num  int
		prio corercmgr.ReserveTaskPriority
		fn   func() *resources
	}{
		{
			name: "high priority task",
			num:  2,
			prio: corercmgr.ReserveTaskPriorityHigh,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetTaskTotalLimit().Return(1).AnyTimes()
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
		},
		{
			name: "medium priority task",
			num:  2,
			prio: corercmgr.ReserveTaskPriorityMedium,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetTaskTotalLimit().Return(1).AnyTimes()
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
		},
		{
			name: "low priority task",
			num:  2,
			prio: corercmgr.ReserveTaskPriorityLow,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetTaskTotalLimit().Return(1).AnyTimes()
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn().addTask(tt.num, tt.prio)
			assert.NotNil(t, err)
		})
	}
}

func TestResources_addTasks(t *testing.T) {
	cases := []struct {
		name              string
		high, medium, low int
		fn                func() *resources
		wantedErr         error
	}{
		{
			name: "no error",
			high: 3, medium: 2, low: 1,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetTaskTotalLimit().Return(100).Times(1)
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(20).Times(1)
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(20).Times(1)
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(20).Times(1)
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
			wantedErr: nil,
		},
		{
			name: "total task limit exceeded",
			high: 3, medium: 2, low: 1,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetTaskTotalLimit().Return(1).Times(2)
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
			wantedErr: &ErrTaskLimitExceeded{
				current: 15, attempted: 6, limit: 1,
				err: fmt.Errorf("total task limit exceeded: %w", ErrResourceLimitExceeded),
			},
		},
		{
			name: "high priority task limit exceeded",
			high: 3, medium: 2, low: 1,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetTaskTotalLimit().Return(100).Times(1)
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(1).Times(2)
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
			wantedErr: &ErrTaskLimitExceeded{
				current: 4, attempted: 3, limit: 1,
				err: fmt.Errorf("high priority task limit exceeded: %w", ErrResourceLimitExceeded),
			},
		},
		{
			name: "medium priority task limit exceeded",
			high: 3, medium: 2, low: 1,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetTaskTotalLimit().Return(100).Times(1)
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(20).Times(1)
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(1).Times(2)
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
			wantedErr: &ErrTaskLimitExceeded{
				current: 5, attempted: 2, limit: 1,
				err: fmt.Errorf("medium priority task limit exceeded: %w", ErrResourceLimitExceeded),
			},
		},
		{
			name: "low priority task limit exceeded",
			high: 3, medium: 2, low: 1,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetTaskTotalLimit().Return(100).Times(1)
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(20).Times(1)
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(20).Times(1)
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(1).Times(2)
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
			wantedErr: &ErrTaskLimitExceeded{
				current: 6, attempted: 1, limit: 1,
				err: fmt.Errorf("low priority task limit exceeded: %w", ErrResourceLimitExceeded),
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn().addTasks(tt.high, tt.medium, tt.low)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestResources_removeTask(t *testing.T) {
	cases := []struct {
		name string
		prio corercmgr.ReserveTaskPriority
	}{
		{
			name: "1",
			prio: corercmgr.ReserveTaskPriorityHigh,
		},
		{
			name: "2",
			prio: corercmgr.ReserveTaskPriorityMedium,
		},
		{
			name: "3",
			prio: corercmgr.ReserveTaskPriorityLow,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rc := setupResources(t)
			rc.removeTask(1, tt.prio)
		})
	}
}

func TestResources_removeTasks(t *testing.T) {
	rc := setupResources(t)
	rc.removeTasks(7, 8, 9)
}

func TestResources_addConn(t *testing.T) {
	cases := []struct {
		name      string
		dir       corercmgr.Direction
		fn        func() *resources
		wantedErr error
	}{
		{
			name: "1",
			dir:  corercmgr.DirInbound,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetConnLimit(corercmgr.DirInbound).Return(0).Times(1)
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
			wantedErr: &ErrConnLimitExceeded{
				current:   1,
				attempted: 1,
				limit:     0,
				err:       fmt.Errorf("cannot reserve inbound connection: %w", ErrResourceLimitExceeded),
			},
		},
		{
			name: "2",
			dir:  corercmgr.DirOutbound,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetConnLimit(corercmgr.DirOutbound).Return(0).Times(1)
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
			wantedErr: &ErrConnLimitExceeded{
				current:   2,
				attempted: 1,
				limit:     0,
				err:       fmt.Errorf("cannot reserve outbound connection: %w", ErrResourceLimitExceeded),
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn().addConn(tt.dir)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestResources_addConns(t *testing.T) {
	cases := []struct {
		name                       string
		inCount, outCount, fdCount int
		fn                         func() *resources
		wantedErr                  error
	}{
		{
			name:    "1",
			inCount: 1,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetConnLimit(corercmgr.DirInbound).Return(0).Times(1)
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
			wantedErr: &ErrConnLimitExceeded{
				current:   1,
				attempted: 1,
				limit:     0,
				err:       fmt.Errorf("cannot reserve inbound connection: %w", ErrResourceLimitExceeded),
			},
		},
		{
			name:     "2",
			outCount: 1,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetConnLimit(corercmgr.DirOutbound).Return(0).Times(1)
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
			wantedErr: &ErrConnLimitExceeded{
				current:   2,
				attempted: 1,
				limit:     0,
				err:       fmt.Errorf("cannot reserve outbound connection: %w", ErrResourceLimitExceeded),
			},
		},
		{
			name: "3",
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetConnTotalLimit().Return(0).Times(1)
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
			wantedErr: &ErrConnLimitExceeded{
				current:   3,
				attempted: 0,
				limit:     0,
				err:       fmt.Errorf("cannot reserve connection: %w", ErrResourceLimitExceeded),
			},
		},
		{
			name:    "4",
			fdCount: 1,
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetConnTotalLimit().Return(100).Times(1)
				m.EXPECT().GetFDLimit().Return(0).Times(1)
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
			wantedErr: &ErrConnLimitExceeded{
				current:   3,
				attempted: 1,
				limit:     0,
				err:       fmt.Errorf("cannot reserve file descriptor: %w", ErrResourceLimitExceeded),
			},
		},
		{
			name: "5",
			fn: func() *resources {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetConnTotalLimit().Return(100).Times(1)
				rc := setupResources(t)
				rc.limit = m
				return rc
			},
			wantedErr: nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn().addConns(tt.inCount, tt.outCount, tt.fdCount)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestResources_removeConn(t *testing.T) {
	cases := []struct {
		name string
		dir  corercmgr.Direction
	}{
		{
			name: "1",
			dir:  corercmgr.DirInbound,
		},
		{
			name: "2",
			dir:  corercmgr.DirOutbound,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rc := setupResources(t)
			rc.removeConn(tt.dir)
		})
	}
}

func TestResources_removeConns(t *testing.T) {
	rc := setupResources(t)
	rc.removeConns(7, 8, 9)
}

func TestResources_stat(t *testing.T) {
	rc := setupResources(t)
	result := rc.stat()
	assert.NotNil(t, result)
}

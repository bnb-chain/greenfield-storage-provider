package gfsprcmgr

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
)

func setupResourceScope(t *testing.T) *resourceScope {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimit(ctrl)
	s := &resourceScope{done: false, rc: resources{limit: m}}
	sList := make([]*resourceScope, 0)
	sList = append(sList, s)
	rs := newResourceScope(m, sList, "test")
	return newResourceScopeSpan(rs, 1, "mock")
}

func TestResourceScope_BeginSpan(t *testing.T) {
	cases := []struct {
		name        string
		fn          func() *resourceScope
		wantedIsErr bool
	}{
		{
			name:        "done is true and returns error",
			fn:          func() *resourceScope { return &resourceScope{done: true} },
			wantedIsErr: true,
		},
		{
			name: "no error",
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				return &resourceScope{owner: newResourceScope(m, nil, "")}
			},
			wantedIsErr: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn().BeginSpan()
			if tt.wantedIsErr {
				assert.NotNil(t, err)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestResourceScope_Done(t *testing.T) {
	cases := []struct {
		name string
		fn   func() *resourceScope
	}{
		{
			name: "done is true",
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				rs.done = true
				return rs
			},
		},
		{
			name: "owner is not nil",
			fn:   func() *resourceScope { return setupResourceScope(t) },
		},
		{
			name: "owner is nil and edge is not nil",
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(0)).AnyTimes()
				rs2 := newResourceScope(m, nil, "mock1")
				edge := make([]*resourceScope, 0)
				edge = append(edge, rs2)
				rs := setupResourceScope(t)
				rs.owner = nil
				rs.edges = edge
				return rs
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn().Done()
		})
	}
}

func TestResourceScope_ReserveMemory(t *testing.T) {
	cases := []struct {
		name        string
		size        int64
		prio        uint8
		fn          func() *resourceScope
		wantedIsErr bool
	}{
		{
			name: "done is true and returns error",
			size: 1,
			prio: 1,
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				rs.done = true
				return rs
			},
			wantedIsErr: true,
		},
		{
			name:        "reserveMemory returns error",
			size:        -1,
			prio:        1,
			fn:          func() *resourceScope { return setupResourceScope(t) },
			wantedIsErr: true,
		},
		{
			name: "reserveMemoryForEdges returns error",
			size: 100,
			prio: 1,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64)).AnyTimes()

				m1 := corercmgr.NewMockLimit(ctrl)
				m1.EXPECT().GetMemoryLimit().Return(int64(0)).AnyTimes()
				rs2 := newResourceScope(m1, nil, "mock1")
				edge := make([]*resourceScope, 0)
				edge = append(edge, rs2)

				rs := setupResourceScope(t)
				rs.edges = edge
				rs.rc.limit = m
				rs.owner = nil
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "no error",
			size: 0,
			prio: 1,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(100)).AnyTimes()

				rs := setupResourceScope(t)
				rs.rc.limit = m
				rs.owner = nil
				return rs
			},
			wantedIsErr: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn().ReserveMemory(tt.size, tt.prio)
			if tt.wantedIsErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestResourceScope_reserveMemoryForEdges(t *testing.T) {
	cases := []struct {
		name        string
		size        int64
		prio        uint8
		fn          func() *resourceScope
		wantedIsErr bool
	}{
		{
			name: "owner is not nil and returns error",
			size: -1,
			prio: 1,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(0)).AnyTimes()
				rs := setupResourceScope(t)
				rs.rc.limit = m
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "owner is nil and returns error",
			size: 100,
			prio: 1,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				rs := newResourceScope(m, nil, "test")
				rs.rc.limit = m

				m1 := corercmgr.NewMockLimit(ctrl)
				m1.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64)).AnyTimes()
				rs2 := newResourceScope(m1, nil, "mock1")
				edge := make([]*resourceScope, 0)
				edge = append(edge, rs2)

				m2 := corercmgr.NewMockLimit(ctrl)
				m1.EXPECT().GetMemoryLimit().Return(int64(-1)).AnyTimes()
				rs3 := newResourceScope(m2, nil, "mock2")
				rs3.done = true
				edge = append(edge, rs3)
				rs.edges = edge

				rs.rc.ntasksHigh = 10
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "owner is nil and no error",
			size: 100,
			prio: 1,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(0)).AnyTimes()
				rs := setupResourceScope(t)
				rs.rc.limit = m
				rs.owner = nil
				return rs
			},
			wantedIsErr: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn().reserveMemoryForEdges(tt.size, tt.prio)
			if tt.wantedIsErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestResourceScope_ReserveMemoryForChild(t *testing.T) {
	cases := []struct {
		name        string
		size        int64
		prio        uint8
		fn          func() *resourceScope
		wantedIsErr bool
	}{
		{
			name: "done is true and returns error",
			size: 0,
			prio: 0,
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				rs.done = true
				return rs
			},
			wantedIsErr: true,
		},
		{
			name:        "reserveMemory returns error",
			size:        -1,
			prio:        0,
			fn:          func() *resourceScope { return setupResourceScope(t) },
			wantedIsErr: true,
		},
		{
			name: "no error",
			size: 1,
			prio: 0,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64 - 1)).AnyTimes()
				rs := setupResourceScope(t)
				rs.rc.limit = m
				return rs
			},
			wantedIsErr: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn().ReserveMemoryForChild(tt.size, tt.prio)
			if tt.wantedIsErr {
				assert.NotNil(t, err)
				assert.NotNil(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestResourceScope_ReleaseMemory(t *testing.T) {
	cases := []struct {
		name string
		size int64
		fn   func() *resourceScope
	}{
		{
			name: "done is true",
			size: 1,
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				rs.done = true
				return rs
			},
		},
		{
			name: "done is false",
			size: 1,
			fn:   func() *resourceScope { return setupResourceScope(t) },
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn().ReleaseMemory(tt.size)
		})
	}
}

func TestResourceScope_releaseMemoryForEdges(t *testing.T) {
	rs := setupResourceScope(t)
	rs.releaseMemoryForEdges(1)
}

func TestResourceScope_ReleaseMemoryForChild(t *testing.T) {
	rs := setupResourceScope(t)
	rs.done = true
	rs.ReleaseMemoryForChild(1)
}

func TestResourceScope_AddTask(t *testing.T) {
	cases := []struct {
		name        string
		num         int
		prio        corercmgr.ReserveTaskPriority
		fn          func() *resourceScope
		wantedIsErr bool
	}{
		{
			name: "done is true and returns error",
			num:  1,
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				rs.done = true
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "addTask returns error",
			num:  1,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetTaskTotalLimit().Return(-1).AnyTimes()
				rs := setupResourceScope(t)
				rs.rc.limit = m
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "addTaskForEdges returns error",
			num:  1,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				rs := newResourceScope(m, nil, "test")
				rs.rc.limit = m
				m.EXPECT().GetTaskTotalLimit().Return(100).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(100).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(100).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(100).AnyTimes()

				m1 := corercmgr.NewMockLimit(ctrl)
				m1.EXPECT().GetTaskTotalLimit().Return(-1).AnyTimes()
				rs2 := newResourceScope(m1, nil, "mock")
				rs.owner = rs2

				rs.rc.ntasksHigh = 10
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "no error",
			num:  1,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				rs := newResourceScope(m, nil, "test")
				rs.rc.limit = m

				m.EXPECT().GetTaskTotalLimit().Return(100).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(100).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(100).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(100).AnyTimes()

				rs.rc.ntasksHigh = 10
				return rs
			},
			wantedIsErr: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn().AddTask(tt.num, tt.prio)
			if tt.wantedIsErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestResourceScope_addTaskForEdges(t *testing.T) {
	cases := []struct {
		name        string
		num         int
		prio        corercmgr.ReserveTaskPriority
		fn          func() *resourceScope
		wantedIsErr bool
	}{
		{
			name: "owner is nil and returns error",
			num:  1,
			prio: corercmgr.ReserveTaskPriorityHigh,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				rs := newResourceScope(m, nil, "test")
				rs.rc.limit = m

				m1 := corercmgr.NewMockLimit(ctrl)
				m1.EXPECT().GetTaskTotalLimit().Return(10).AnyTimes()
				m1.EXPECT().GetTaskTotalLimit().Return(100).AnyTimes()
				m1.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(100).AnyTimes()
				m1.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(100).AnyTimes()
				m1.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(100).AnyTimes()

				rs2 := newResourceScope(m1, nil, "mock")
				edge := make([]*resourceScope, 0)
				edge = append(edge, rs2)

				m2 := corercmgr.NewMockLimit(ctrl)
				m2.EXPECT().GetTaskTotalLimit().Return(-1).AnyTimes()
				rs3 := newResourceScope(m2, nil, "mock")
				edge = append(edge, rs3)
				rs.edges = edge

				rs.rc.ntasksHigh = 10
				return rs
			},
			wantedIsErr: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn().addTaskForEdges(tt.num, tt.prio)
			if tt.wantedIsErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestResourceScope_AddTaskForChild(t *testing.T) {
	t.Log("Case description: done is true")
	rs := setupResourceScope(t)
	rs.done = true
	result, err := rs.AddTaskForChild(1, corercmgr.ReserveTaskPriorityHigh)
	assert.NotNil(t, err)
	assert.NotNil(t, result)
}

func TestResourceScope_RemoveTask(t *testing.T) {
	cases := []struct {
		name string
		num  int
		prio corercmgr.ReserveTaskPriority
		fn   func() *resourceScope
	}{
		{
			name: "done is true",
			num:  0,
			prio: corercmgr.ReserveTaskPriorityHigh,
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				rs.done = true
				return rs
			},
		},
		{
			name: "done is false",
			num:  0,
			prio: corercmgr.ReserveTaskPriorityHigh,
			fn:   func() *resourceScope { return setupResourceScope(t) },
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn().RemoveTask(tt.num, tt.prio)
		})
	}
}

func TestResourceScope_removeTaskForEdges(t *testing.T) {
	cases := []struct {
		name string
		num  int
		prio corercmgr.ReserveTaskPriority
		fn   func() *resourceScope
	}{
		{
			name: "owner is not nil and edges is not nil",
			num:  0,
			prio: corercmgr.ReserveTaskPriorityHigh,
			fn:   func() *resourceScope { return setupResourceScope(t) },
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn().removeTaskForEdges(tt.num, tt.prio)
		})
	}
}

func TestResourceScope_RemoveTaskForChild(t *testing.T) {
	t.Log("Case description: done is true")
	rs := setupResourceScope(t)
	rs.done = true
	rs.RemoveTaskForChild(0, corercmgr.ReserveTaskPriorityHigh)
}

func TestResourceScope_AddConn(t *testing.T) {
	cases := []struct {
		name       string
		dir        corercmgr.Direction
		fn         func() *resourceScope
		wanedIsErr bool
	}{
		{
			name: "done is true and returns error",
			dir:  corercmgr.DirInbound,
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				rs.done = true
				return rs
			},
			wanedIsErr: true,
		},
		{
			name: "addConn returns error",
			dir:  corercmgr.DirInbound,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetConnLimit(corercmgr.DirInbound).Return(-1).AnyTimes()
				rs := setupResourceScope(t)
				rs.rc.limit = m
				return rs
			},
			wanedIsErr: true,
		},
		{
			name: "addConnForEdges returns error",
			dir:  corercmgr.DirInbound,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				rs := newResourceScope(m, nil, "test")
				rs.rc.limit = m
				m.EXPECT().GetConnLimit(corercmgr.DirInbound).Return(10).AnyTimes()
				m.EXPECT().GetConnTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetFDLimit().Return(10).AnyTimes()

				m1 := corercmgr.NewMockLimit(ctrl)
				m1.EXPECT().GetConnLimit(corercmgr.DirInbound).Return(-1).AnyTimes()
				rs2 := newResourceScope(m1, nil, "mock")
				rs.owner = rs2

				rs.rc.ntasksHigh = 10
				return rs
			},
			wanedIsErr: true,
		},
		{
			name: "no error",
			dir:  corercmgr.DirInbound,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				rs := newResourceScope(m, nil, "test")
				rs.rc.limit = m

				m.EXPECT().GetConnLimit(corercmgr.DirInbound).Return(10).AnyTimes()
				m.EXPECT().GetConnTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetFDLimit().Return(10).AnyTimes()

				rs.rc.ntasksHigh = 10
				return rs
			},
			wanedIsErr: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn().AddConn(tt.dir)
			if tt.wanedIsErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestResourceScope_addConnForEdges(t *testing.T) {
	cases := []struct {
		name       string
		dir        corercmgr.Direction
		fn         func() *resourceScope
		wanedIsErr bool
	}{
		{
			name: "owner is not nil and no error",
			dir:  corercmgr.DirInbound,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				rs := newResourceScope(m, nil, "test")
				rs.rc.limit = m

				m1 := corercmgr.NewMockLimit(ctrl)
				m1.EXPECT().GetConnLimit(corercmgr.DirInbound).Return(10).AnyTimes()
				m1.EXPECT().GetConnTotalLimit().Return(10).AnyTimes()
				m1.EXPECT().GetFDLimit().Return(10).AnyTimes()
				rs2 := newResourceScope(m1, nil, "mock")
				rs.owner = rs2
				return rs
			},
			wanedIsErr: false,
		},
		{
			name: "owner is nil and returns error",
			dir:  corercmgr.DirInbound,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				rs := newResourceScope(m, nil, "test")
				rs.rc.limit = m

				m1 := corercmgr.NewMockLimit(ctrl)
				m1.EXPECT().GetConnLimit(corercmgr.DirInbound).Return(10).AnyTimes()
				m1.EXPECT().GetConnTotalLimit().Return(10).AnyTimes()
				m1.EXPECT().GetFDLimit().Return(10).AnyTimes()

				rs2 := newResourceScope(m1, nil, "mock")
				edge := make([]*resourceScope, 0)
				edge = append(edge, rs2)

				m2 := corercmgr.NewMockLimit(ctrl)
				m2.EXPECT().GetConnLimit(corercmgr.DirInbound).Return(-1).AnyTimes()
				rs3 := newResourceScope(m2, nil, "mock")
				edge = append(edge, rs3)
				rs.edges = edge

				rs.rc.ntasksHigh = 10
				return rs
			},
			wanedIsErr: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn().addConnForEdges(tt.dir)
			if tt.wanedIsErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestResourceScope_AddConnForChild(t *testing.T) {
	cases := []struct {
		name       string
		dir        corercmgr.Direction
		fn         func() *resourceScope
		wanedIsErr bool
	}{
		{
			name: "done is true",
			dir:  corercmgr.DirInbound,
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				rs.done = true
				return rs
			},
			wanedIsErr: true,
		},
		{
			name: "addConn returns error",
			dir:  corercmgr.DirInbound,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetConnLimit(corercmgr.DirInbound).Return(-1).AnyTimes()
				rs := newResourceScope(m, nil, "test")
				return rs
			},
			wanedIsErr: true,
		},
		{
			name: "no error",
			dir:  corercmgr.DirInbound,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetConnLimit(corercmgr.DirInbound).Return(10).AnyTimes()
				m.EXPECT().GetConnTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetFDLimit().Return(10).AnyTimes()
				rs := newResourceScope(m, nil, "test")
				return rs
			},
			wanedIsErr: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn().AddConnForChild(tt.dir)
			if tt.wanedIsErr {
				assert.NotNil(t, err)
				assert.NotNil(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestResourceScope_RemoveConn(t *testing.T) {
	cases := []struct {
		name string
		dir  corercmgr.Direction
		fn   func() *resourceScope
	}{
		{
			name: "done is true",
			dir:  corercmgr.DirInbound,
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				rs.done = true
				return rs
			},
		},
		{
			name: "done is false",
			dir:  corercmgr.DirInbound,
			fn:   func() *resourceScope { return setupResourceScope(t) },
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn().RemoveConn(tt.dir)
		})
	}
}

func TestResourceScope_removeConnForEdges(t *testing.T) {
	cases := []struct {
		name string
		dir  corercmgr.Direction
		fn   func() *resourceScope
	}{
		{
			name: "owner is not nil",
			dir:  corercmgr.DirInbound,
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				rs2 := newResourceScope(m, nil, "mock")
				rs := setupResourceScope(t)
				rs.owner = rs2
				return rs
			},
		},
		{
			name: "edges is not nil",
			dir:  corercmgr.DirInbound,
			fn:   func() *resourceScope { return setupResourceScope(t) },
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn().removeConnForEdges(tt.dir)
		})
	}
}

func TestResourceScope_RemoveConnForChild(t *testing.T) {
	t.Log("Case description: done is true")
	rs := setupResourceScope(t)
	rs.done = true
	rs.RemoveConnForChild(corercmgr.DirInbound)
}

func TestResourceScope_ReserveForChild(t *testing.T) {
	cases := []struct {
		name        string
		st          corercmgr.ScopeStat
		fn          func() *resourceScope
		wantedIsErr bool
	}{
		{
			name: "done is true",
			st:   corercmgr.ScopeStat{},
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				rs.done = true
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "reserveMemory returns error",
			st:   corercmgr.ScopeStat{Memory: -1},
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				rs := newResourceScope(m, nil, "test")
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "addConns returns error",
			st:   corercmgr.ScopeStat{},
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64)).AnyTimes()
				m.EXPECT().GetConnTotalLimit().Return(-1).AnyTimes()
				rs := newResourceScope(m, nil, "test")
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "add high priority task returns error",
			st:   corercmgr.ScopeStat{},
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64)).AnyTimes()
				m.EXPECT().GetConnTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetTaskTotalLimit().Return(100).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(-1).AnyTimes()
				rs := newResourceScope(m, nil, "test")
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "add medium priority task returns error",
			st:   corercmgr.ScopeStat{NumTasksMedium: 20},
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64)).AnyTimes()
				m.EXPECT().GetConnTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetTaskTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(10).AnyTimes()
				rs := newResourceScope(m, nil, "test")
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "add low priority task returns error",
			st:   corercmgr.ScopeStat{NumTasksLow: 20},
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64)).AnyTimes()
				m.EXPECT().GetConnTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetTaskTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(10).AnyTimes()
				rs := newResourceScope(m, nil, "test")
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "no error",
			st:   corercmgr.ScopeStat{},
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64)).AnyTimes()
				m.EXPECT().GetConnTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetTaskTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(10).AnyTimes()
				rs := newResourceScope(m, nil, "test")
				return rs
			},
			wantedIsErr: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn().ReserveForChild(tt.st)
			if tt.wantedIsErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestResourceScope_ReleaseResources(t *testing.T) {
	cases := []struct {
		name string
		st   corercmgr.ScopeStat
		fn   func() *resourceScope
	}{
		{
			name: "done is true ",
			st:   corercmgr.ScopeStat{},
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				rs.done = true
				return rs
			},
		},
		{
			name: "owner is not nil",
			st:   corercmgr.ScopeStat{},
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				rs1 := newResourceScope(m, nil, "test")
				rs.owner = rs1
				return rs
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn().ReleaseResources(tt.st)
		})
	}
}

func TestResourceScope_Release(t *testing.T) {
	cases := []struct {
		name string
		fn   func() *resourceScope
	}{
		{
			name: "done is true ",
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				rs.done = true
				return rs
			},
		},
		{
			name: "owner is not nil",
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				rs1 := newResourceScope(m, nil, "test")
				rs.owner = rs1
				return rs
			},
		},
		{
			name: "range edges",
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				rs1 := newResourceScope(m, nil, "test")
				edges := make([]*resourceScope, 0)
				edges = append(edges, rs1)
				rs.owner = nil
				rs.edges = edges
				return rs
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn().Release()
		})
	}
}

func TestResourceScope_ReleaseForChild(t *testing.T) {
	t.Log("Case description: done is true")
	rs := setupResourceScope(t)
	rs.done = true
	rs.ReleaseForChild(corercmgr.ScopeStat{})
}

func TestResourceScope_ReserveResources(t *testing.T) {
	cases := []struct {
		name        string
		st          *corercmgr.ScopeStat
		fn          func() *resourceScope
		wantedIsErr bool
	}{
		{
			name: "done is true and returns error",
			st:   &corercmgr.ScopeStat{},
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				rs.done = true
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "reserveMemory returns error",
			st:   &corercmgr.ScopeStat{},
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(-1)).AnyTimes()
				rs := newResourceScope(m, nil, "test")
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "addConns returns error",
			st:   &corercmgr.ScopeStat{},
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64)).AnyTimes()
				m.EXPECT().GetConnTotalLimit().Return(-1).AnyTimes()
				rs := newResourceScope(m, nil, "test")
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "add high priority task returns error",
			st:   &corercmgr.ScopeStat{},
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64)).AnyTimes()
				m.EXPECT().GetConnTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetTaskTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(-1).AnyTimes()
				rs := newResourceScope(m, nil, "test")
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "add medium priority task returns error",
			st:   &corercmgr.ScopeStat{NumTasksMedium: 20},
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64)).AnyTimes()
				m.EXPECT().GetConnTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetTaskTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(10).AnyTimes()
				rs := newResourceScope(m, nil, "test")
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "add low priority task returns error",
			st:   &corercmgr.ScopeStat{NumTasksLow: 20},
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64)).AnyTimes()
				m.EXPECT().GetConnTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetTaskTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(10).AnyTimes()
				rs := newResourceScope(m, nil, "test")
				return rs
			},
			wantedIsErr: true,
		},
		{
			name: "owner is not nil and no error",
			st:   &corercmgr.ScopeStat{},
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64)).AnyTimes()
				m.EXPECT().GetConnTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetTaskTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(10).AnyTimes()
				rs := newResourceScope(m, nil, "test")
				rs1 := newResourceScope(m, nil, "mock")
				rs.owner = rs1
				return rs
			},
			wantedIsErr: false,
		},
		{
			name: "edges is not nil and no error",
			st:   &corercmgr.ScopeStat{},
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64)).AnyTimes()
				m.EXPECT().GetConnTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetTaskTotalLimit().Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(10).AnyTimes()
				m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(10).AnyTimes()

				edges := make([]*resourceScope, 0)
				rs1 := newResourceScope(m, nil, "mock")
				edges = append(edges, rs1)
				rs := newResourceScope(m, nil, "test")
				rs.edges = edges
				return rs
			},
			wantedIsErr: false,
		},
		{
			name: "no error",
			st:   &corercmgr.ScopeStat{},
			fn: func() *resourceScope {
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockLimit(ctrl)
				m.EXPECT().GetMemoryLimit().Return(int64(math.MaxInt64)).AnyTimes()
				m.EXPECT().GetConnTotalLimit().Return(-1).AnyTimes()
				rs := newResourceScope(m, nil, "test")
				return rs
			},
			wantedIsErr: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn().ReserveResources(tt.st)
			if tt.wantedIsErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestResourceScope_RemainingResource(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().GetMemoryLimit().Return(int64(10)).AnyTimes()
	m.EXPECT().GetTaskTotalLimit().Return(10).AnyTimes()
	m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(10).AnyTimes()
	m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(10).AnyTimes()
	m.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(10).AnyTimes()
	rs := newResourceScope(m, nil, "test")
	result, err := rs.RemainingResource()
	assert.NotNil(t, result)
	assert.Nil(t, err)
}

func TestResourceScope_Name(t *testing.T) {
	rs := setupResourceScope(t)
	result := rs.Name()
	assert.Equal(t, "test.span-mock-1", result)
}

func TestResourceScope_IsUnused(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *resourceScope
		wantedResult bool
	}{
		{
			name: "done is true",
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				rs.done = true
				return rs
			},
			wantedResult: true,
		},
		{
			name: "refCnt is greater than 0",
			fn: func() *resourceScope {
				rs := setupResourceScope(t)
				rs.refCnt = 5
				return rs
			},
			wantedResult: false,
		},
		{
			name:         "result is true",
			fn:           func() *resourceScope { return setupResourceScope(t) },
			wantedResult: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn().IsUnused()
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

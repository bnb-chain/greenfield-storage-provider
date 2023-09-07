package gfsplimit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
)

func TestGfSpLimit_GetMemoryLimit(t *testing.T) {
	m := &GfSpLimit{Memory: 1}
	result := m.GetMemoryLimit()
	assert.Equal(t, int64(1), result)
}

func TestGfSpLimit_GetFDLimit(t *testing.T) {
	m := &GfSpLimit{Fd: 1}
	result := m.GetFDLimit()
	assert.Equal(t, 1, result)
}

func TestGfSpLimit_GetConnLimit(t *testing.T) {
	cases := []struct {
		name         string
		direction    rcmgr.Direction
		m            *GfSpLimit
		wantedResult int
	}{
		{
			name:         "1",
			direction:    rcmgr.DirInbound,
			m:            &GfSpLimit{ConnsInbound: 1},
			wantedResult: 1,
		},
		{
			name:         "2",
			direction:    rcmgr.DirOutbound,
			m:            &GfSpLimit{ConnsOutbound: 2},
			wantedResult: 2,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.m.GetConnLimit(tt.direction)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpLimit_GetConnTotalLimit(t *testing.T) {
	m := &GfSpLimit{Conns: 1}
	result := m.GetConnTotalLimit()
	assert.Equal(t, 1, result)
}

func TestGfSpLimit_GetTaskLimit(t *testing.T) {
	cases := []struct {
		name         string
		priority     rcmgr.ReserveTaskPriority
		m            *GfSpLimit
		wantedResult int
	}{
		{
			name:         "1",
			priority:     rcmgr.ReserveTaskPriorityHigh,
			m:            &GfSpLimit{TasksHighPriority: 1},
			wantedResult: 1,
		},
		{
			name:         "2",
			priority:     rcmgr.ReserveTaskPriorityMedium,
			m:            &GfSpLimit{TasksMediumPriority: 2},
			wantedResult: 2,
		},
		{
			name:         "3",
			priority:     rcmgr.ReserveTaskPriorityLow,
			m:            &GfSpLimit{TasksLowPriority: 3},
			wantedResult: 3,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.m.GetTaskLimit(tt.priority)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpLimit_GetTaskTotalLimit(t *testing.T) {
	m := &GfSpLimit{Tasks: 1}
	result := m.GetTaskTotalLimit()
	assert.Equal(t, 1, result)
}

func TestGfSpLimit_ScopeStat(t *testing.T) {
	m := &GfSpLimit{Tasks: 1}
	result := m.ScopeStat()
	assert.NotNil(t, result)
}

func TestGfSpLimit_NotLess1(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(2)).Times(1)
	m := &GfSpLimit{Memory: 1}
	result := m.NotLess(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_NotLess2(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(2).Times(1)
	m := &GfSpLimit{Memory: 1, Fd: 1}
	result := m.NotLess(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_NotLess3(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(2).Times(1)
	m := &GfSpLimit{Memory: 1, Tasks: 1, Fd: 1}
	result := m.NotLess(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_NotLess4(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirOutbound).Return(2).Times(1)
	m := &GfSpLimit{Memory: 1, Tasks: 1, Fd: 1, ConnsOutbound: 1}
	result := m.NotLess(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_NotLess5(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirOutbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirInbound).Return(2).Times(1)
	m := &GfSpLimit{Memory: 1, Tasks: 1, Fd: 1, ConnsInbound: 1, ConnsOutbound: 1}
	result := m.NotLess(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_NotLess6(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirOutbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirInbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnTotalLimit().Return(2).Times(1)
	m := &GfSpLimit{Memory: 1, Tasks: 1, Fd: 1, Conns: 1, ConnsInbound: 1, ConnsOutbound: 1}
	result := m.NotLess(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_NotLess7(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirOutbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirInbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityHigh).Return(2).Times(1)
	m := &GfSpLimit{
		Memory:            1,
		Tasks:             1,
		TasksHighPriority: 1,
		Fd:                1,
		Conns:             1,
		ConnsInbound:      1,
		ConnsOutbound:     1,
	}
	result := m.NotLess(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_NotLess8(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirOutbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirInbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityHigh).Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityMedium).Return(2).Times(1)
	m := &GfSpLimit{
		Memory:              1,
		Tasks:               1,
		TasksHighPriority:   1,
		TasksMediumPriority: 1,
		Fd:                  1,
		Conns:               1,
		ConnsInbound:        1,
		ConnsOutbound:       1,
	}
	result := m.NotLess(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_NotLess9(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirOutbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirInbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityHigh).Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityMedium).Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityLow).Return(2).Times(1)
	m := &GfSpLimit{
		Memory:              1,
		Tasks:               1,
		TasksHighPriority:   1,
		TasksMediumPriority: 1,
		TasksLowPriority:    1,
		Fd:                  1,
		Conns:               1,
		ConnsInbound:        1,
		ConnsOutbound:       1,
	}
	result := m.NotLess(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_NotLess10(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirOutbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirInbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityHigh).Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityMedium).Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityLow).Return(1).Times(1)
	m := &GfSpLimit{
		Memory:              1,
		Tasks:               1,
		TasksHighPriority:   1,
		TasksMediumPriority: 1,
		TasksLowPriority:    1,
		Fd:                  1,
		Conns:               1,
		ConnsInbound:        1,
		ConnsOutbound:       1,
	}
	result := m.NotLess(mockLimit)
	assert.Equal(t, true, result)
}

func TestGfSpLimit_Add(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirOutbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirInbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityHigh).Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityMedium).Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityLow).Return(1).Times(1)
	m := &GfSpLimit{}
	m.Add(mockLimit)
}

func TestGfSpLimit_Sub1(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(2)).Times(1)
	m := &GfSpLimit{Memory: 1}
	result := m.Sub(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_Sub2(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(2)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(2)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(1).Times(2)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirOutbound).Return(1).Times(2)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirInbound).Return(1).Times(2)
	mockLimit.EXPECT().GetConnTotalLimit().Return(1).Times(2)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityHigh).Return(1).Times(2)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityMedium).Return(1).Times(2)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityLow).Return(1).Times(2)
	m := &GfSpLimit{
		Memory:              1,
		Tasks:               1,
		TasksHighPriority:   1,
		TasksMediumPriority: 1,
		TasksLowPriority:    1,
		Fd:                  1,
		Conns:               1,
		ConnsInbound:        1,
		ConnsOutbound:       1,
	}
	result := m.Sub(mockLimit)
	assert.Equal(t, true, result)
}

func TestGfSpLimit_Equal1(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	m := &GfSpLimit{Memory: 2}
	result := m.Equal(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_Equal2(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	m := &GfSpLimit{Memory: 1, Fd: 2}
	result := m.Equal(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_Equal3(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
	m := &GfSpLimit{Memory: 1, Tasks: 2, Fd: 1}
	result := m.Equal(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_Equal4(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirOutbound).Return(1).Times(1)
	m := &GfSpLimit{Memory: 1, Tasks: 1, Fd: 1, ConnsOutbound: 2}
	result := m.Equal(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_Equal5(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirOutbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirInbound).Return(1).Times(1)
	m := &GfSpLimit{Memory: 1, Tasks: 1, Fd: 1, ConnsInbound: 2, ConnsOutbound: 1}
	result := m.Equal(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_Equal6(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirOutbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirInbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnTotalLimit().Return(1).Times(1)
	m := &GfSpLimit{Memory: 1, Tasks: 1, Fd: 1, Conns: 2, ConnsInbound: 1, ConnsOutbound: 1}
	result := m.Equal(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_Equal7(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirOutbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirInbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityHigh).Return(1).Times(1)
	m := &GfSpLimit{
		Memory:            1,
		Tasks:             1,
		TasksHighPriority: 2,
		Fd:                1,
		Conns:             1,
		ConnsInbound:      1,
		ConnsOutbound:     1,
	}
	result := m.Equal(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_Equal8(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirOutbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirInbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityHigh).Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityMedium).Return(1).Times(1)
	m := &GfSpLimit{
		Memory:              1,
		Tasks:               1,
		TasksHighPriority:   1,
		TasksMediumPriority: 2,
		Fd:                  1,
		Conns:               1,
		ConnsInbound:        1,
		ConnsOutbound:       1,
	}
	result := m.Equal(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_Equal9(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirOutbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirInbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityHigh).Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityMedium).Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityLow).Return(1).Times(1)
	m := &GfSpLimit{
		Memory:              1,
		Tasks:               1,
		TasksHighPriority:   1,
		TasksMediumPriority: 1,
		TasksLowPriority:    2,
		Fd:                  1,
		Conns:               1,
		ConnsInbound:        1,
		ConnsOutbound:       1,
	}
	result := m.Equal(mockLimit)
	assert.Equal(t, false, result)
}

func TestGfSpLimit_Equal10(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLimit := rcmgr.NewMockLimit(ctrl)
	mockLimit.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
	mockLimit.EXPECT().GetFDLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirOutbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnLimit(rcmgr.DirInbound).Return(1).Times(1)
	mockLimit.EXPECT().GetConnTotalLimit().Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityHigh).Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityMedium).Return(1).Times(1)
	mockLimit.EXPECT().GetTaskLimit(rcmgr.ReserveTaskPriorityLow).Return(1).Times(1)
	m := &GfSpLimit{
		Memory:              1,
		Tasks:               1,
		TasksHighPriority:   1,
		TasksMediumPriority: 1,
		TasksLowPriority:    1,
		Fd:                  1,
		Conns:               1,
		ConnsInbound:        1,
		ConnsOutbound:       1,
	}
	result := m.Equal(mockLimit)
	assert.Equal(t, true, result)
}

func TestGfSpLimiter_GetSystemLimits(t *testing.T) {
	m := &GfSpLimiter{}
	result := m.GetSystemLimits()
	assert.Nil(t, result)
}

func TestGfSpLimiter_GetTransientLimits(t *testing.T) {
	m := &GfSpLimiter{}
	result := m.GetTransientLimits()
	assert.Nil(t, result)
}

func TestGfSpLimiter_GetServiceLimits1(t *testing.T) {
	m := &GfSpLimiter{ServiceLimit: map[string]*GfSpLimit{"test": &GfSpLimit{Memory: 1}}}
	result := m.GetServiceLimits("test")
	assert.NotNil(t, result)
}

func TestGfSpLimiter_GetServiceLimits2(t *testing.T) {
	m := &GfSpLimiter{}
	result := m.GetServiceLimits("test")
	assert.Nil(t, result)
}

package gfsplimit

import (
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
)

var _ rcmgr.Limit = &GfSpLimit{}
var _ rcmgr.Limiter = &GfSpLimiter{}

func (m *GfSpLimit) GetMemoryLimit() int64 {
	return m.GetMemory()
}

func (m *GfSpLimit) GetFDLimit() int {
	return int(m.GetFd())
}
func (m *GfSpLimit) GetConnLimit(direction rcmgr.Direction) int {
	if direction == rcmgr.DirOutbound {
		return int(m.GetConnsOutbound())
	}
	return int(m.GetConnsInbound())
}

func (m *GfSpLimit) GetConnTotalLimit() int {
	return int(m.GetConns())
}

func (m *GfSpLimit) GetTaskLimit(priority rcmgr.ReserveTaskPriority) int {
	if priority == rcmgr.ReserveTaskPriorityHigh {
		return int(m.GetTasksHighPriority())
	} else if priority == rcmgr.ReserveTaskPriorityMedium {
		return int(m.GetTasksMediumPriority())
	}
	return int(m.GetTasksLowPriority())
}

func (m *GfSpLimit) GetTaskTotalLimit() int {
	return int(m.GetTasks())
}

func (m *GfSpLimit) ScopeStat() *rcmgr.ScopeStat {
	st := &rcmgr.ScopeStat{
		Memory:           m.GetMemory(),
		NumTasksHigh:     int64(m.GetTaskLimit(rcmgr.ReserveTaskPriorityHigh)),
		NumTasksMedium:   int64(m.GetTaskLimit(rcmgr.ReserveTaskPriorityMedium)),
		NumTasksLow:      int64(m.GetTaskLimit(rcmgr.ReserveTaskPriorityLow)),
		NumConnsInbound:  int64(m.GetConnLimit(rcmgr.DirInbound)),
		NumConnsOutbound: int64(m.GetConnLimit(rcmgr.DirOutbound)),
		NumFD:            int64(m.GetFDLimit()),
	}
	return st
}

func (m *GfSpLimit) NotLess(x rcmgr.Limit) bool {
	if m.GetMemoryLimit() < x.GetMemoryLimit() {
		return false
	}
	if m.GetFDLimit() < x.GetFDLimit() {
		return false
	}
	if m.GetTaskTotalLimit() < x.GetTaskTotalLimit() {
		return false
	}
	if m.GetConnLimit(rcmgr.DirOutbound) < x.GetConnLimit(rcmgr.DirOutbound) {
		return false
	}
	if m.GetConnLimit(rcmgr.DirInbound) < x.GetConnLimit(rcmgr.DirInbound) {
		return false
	}
	if m.GetConnTotalLimit() < x.GetConnTotalLimit() {
		return false
	}
	if m.GetTaskLimit(rcmgr.ReserveTaskPriorityHigh) < x.GetTaskLimit(rcmgr.ReserveTaskPriorityHigh) {
		return false
	}
	if m.GetTaskLimit(rcmgr.ReserveTaskPriorityMedium) < x.GetTaskLimit(rcmgr.ReserveTaskPriorityMedium) {
		return false
	}
	if m.GetTaskLimit(rcmgr.ReserveTaskPriorityLow) < x.GetTaskLimit(rcmgr.ReserveTaskPriorityLow) {
		return false
	}
	if m.GetTaskTotalLimit() < x.GetTaskTotalLimit() {
		return false
	}
	return true
}

func (m *GfSpLimit) Add(x rcmgr.Limit) {
	m.Memory += x.GetMemoryLimit()
	m.Tasks += int32(x.GetTaskTotalLimit())
	m.TasksHighPriority += int32(x.GetTaskLimit(rcmgr.ReserveTaskPriorityHigh))
	m.TasksMediumPriority += int32(x.GetTaskLimit(rcmgr.ReserveTaskPriorityMedium))
	m.TasksLowPriority += int32(x.GetTaskLimit(rcmgr.ReserveTaskPriorityLow))
	m.Fd += int32(x.GetFDLimit())
	m.Conns += int32(x.GetConnTotalLimit())
	m.ConnsInbound += int32(x.GetConnLimit(rcmgr.DirInbound))
	m.ConnsOutbound += int32(x.GetConnLimit(rcmgr.DirOutbound))
}

func (m *GfSpLimit) Sub(x rcmgr.Limit) bool {
	if !m.NotLess(x) {
		return false
	}
	m.Memory -= x.GetMemoryLimit()
	m.Tasks -= int32(x.GetTaskTotalLimit())
	m.TasksHighPriority -= int32(x.GetTaskLimit(rcmgr.ReserveTaskPriorityHigh))
	m.TasksMediumPriority -= int32(x.GetTaskLimit(rcmgr.ReserveTaskPriorityMedium))
	m.TasksLowPriority -= int32(x.GetTaskLimit(rcmgr.ReserveTaskPriorityLow))
	m.Fd -= int32(x.GetFDLimit())
	m.Conns -= int32(x.GetConnTotalLimit())
	m.ConnsInbound -= int32(x.GetConnLimit(rcmgr.DirInbound))
	m.ConnsOutbound -= int32(x.GetConnLimit(rcmgr.DirOutbound))
	return true
}

func (m *GfSpLimit) Equal(x rcmgr.Limit) bool {
	if m.GetMemoryLimit() != x.GetMemoryLimit() {
		return false
	}
	if m.GetFDLimit() != x.GetFDLimit() {
		return false
	}
	if m.GetTaskTotalLimit() != x.GetTaskTotalLimit() {
		return false
	}
	if m.GetConnLimit(rcmgr.DirOutbound) != x.GetConnLimit(rcmgr.DirOutbound) {
		return false
	}
	if m.GetConnLimit(rcmgr.DirInbound) != x.GetConnLimit(rcmgr.DirInbound) {
		return false
	}
	if m.GetConnTotalLimit() != x.GetConnTotalLimit() {
		return false
	}
	if m.GetTaskLimit(rcmgr.ReserveTaskPriorityHigh) != x.GetTaskLimit(rcmgr.ReserveTaskPriorityHigh) {
		return false
	}
	if m.GetTaskLimit(rcmgr.ReserveTaskPriorityMedium) != x.GetTaskLimit(rcmgr.ReserveTaskPriorityMedium) {
		return false
	}
	if m.GetTaskLimit(rcmgr.ReserveTaskPriorityLow) != x.GetTaskLimit(rcmgr.ReserveTaskPriorityLow) {
		return false
	}
	if m.GetTaskTotalLimit() != x.GetTaskTotalLimit() {
		return false
	}
	return true
}

func (m *GfSpLimiter) GetSystemLimits() rcmgr.Limit {
	return m.GetSystem()
}

func (m *GfSpLimiter) GetTransientLimits() rcmgr.Limit {
	return m.GetTransient()
}

func (m *GfSpLimiter) GetServiceLimits(svc string) rcmgr.Limit {
	if _, ok := m.GetServiceLimit()[svc]; !ok {
		return nil
	}
	return m.GetServiceLimit()[svc]
}

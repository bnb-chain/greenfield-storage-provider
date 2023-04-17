package types

import "github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"

func NewLimits(rcLimit rcmgr.Limit) *Limit {
	m := &Limit{}
	m.Memory = rcLimit.GetMemoryLimit()
	m.HighPriorityTask = int64(rcLimit.GetTaskLimit(rcmgr.ReserveTaskPriorityHigh))
	m.MediumPriorityTask = int64(rcLimit.GetTaskLimit(rcmgr.ReserveTaskPriorityMedium))
	m.LowPriorityTask = int64(rcLimit.GetTaskLimit(rcmgr.ReserveTaskPriorityLow))
	return m
}

func (m *Limit) TransferRcmgrLimits() rcmgr.Limit {
	if m == nil {
		return rcmgr.InfinitesimalLimit()
	}
	rcLimits := rcmgr.InfinitesimalLimit().(*rcmgr.BaseLimit)
	rcLimits.Memory = m.GetMemory()
	rcLimits.TasksHighPriority = int(m.GetHighPriorityTask())
	rcLimits.TasksMediumPriority = int(m.GetMediumPriorityTask())
	rcLimits.TasksLowPriority = int(m.GetLowPriorityTask())
	return rcLimits
}

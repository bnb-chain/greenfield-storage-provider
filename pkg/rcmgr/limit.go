package rcmgr

import (
	"fmt"
	"math"
)

const (
	MaxLimitInt64 = math.MaxInt64
	MaxLimitInt   = math.MaxInt
)

// Limit is an interface that that specifies basic resource limits.
type Limit interface {
	// GetMemoryLimit returns the (current) memory limit.
	GetMemoryLimit() int64
	// GetFDLimit returns the file descriptor limit.
	GetFDLimit() int
	// GetConnLimit returns the connection limit, for inbound or outbound connections.
	GetConnLimit(Direction) int
	// GetConnTotalLimit returns the total connection limit.
	GetConnTotalLimit() int
	// GetTaskLimit returns the task limit, for high, medium and low priority tasks.
	GetTaskLimit(ReserveTaskPriority) int
	// GetTaskTotalLimit returns the total task limit.
	GetTaskTotalLimit() int
	// Greater returns an indicator whether cover the param limit.
	Greater(Limit) bool
	// Equal returns an indicator whether equal the param limit.
	Equal(Limit) bool
	// String returns the Limit state string.
	String() string
}

// Limiter is an interface for providing limits to the resource manager.
type Limiter interface {
	GetSystemLimits() Limit
	GetTransientLimits() Limit
	GetServiceLimits(svc string) Limit
	String() string
}

var _ Limit = &BaseLimit{}

// BaseLimit is a mixin type for basic resource limits.
type BaseLimit struct {
	Memory              int64 `json:",omitempty"`
	FD                  int   `json:",omitempty"`
	Conns               int   `json:",omitempty"`
	ConnsInbound        int   `json:",omitempty"`
	ConnsOutbound       int   `json:",omitempty"`
	Tasks               int   `json:",omitempty"`
	TasksHighPriority   int   `json:",omitempty"`
	TasksMediumPriority int   `json:",omitempty"`
	TasksLowPriority    int   `json:",omitempty"`
}

// GetMemoryLimit returns the (current) memory limit.
func (limit *BaseLimit) GetMemoryLimit() int64 {
	return limit.Memory
}

// GetFDLimit returns the file descriptor limit.
func (limit *BaseLimit) GetFDLimit() int {
	return limit.FD
}

// GetConnLimit returns the connection limit, for inbound or outbound connections.
func (limit *BaseLimit) GetConnLimit(direction Direction) int {
	if direction == DirInbound {
		return limit.ConnsInbound
	}
	return limit.ConnsOutbound
}

// GetConnTotalLimit returns the total connection limit.
func (limit *BaseLimit) GetConnTotalLimit() int {
	return limit.Conns
}

// GetTaskTotalLimit returns the total task limit.
func (limit *BaseLimit) GetTaskTotalLimit() int {
	return limit.Tasks
}

// GetTaskLimit returns the task limit, for high, medium and low priority tasks.
func (limit *BaseLimit) GetTaskLimit(priority ReserveTaskPriority) int {
	switch priority {
	case ReserveTaskPriorityHigh:
		return limit.TasksHighPriority
	case ReserveTaskPriorityMedium:
		return limit.TasksMediumPriority
	case ReserveTaskPriorityLow:
		return limit.TasksLowPriority
	default:
		return 0
	}
}

// Greater returns an indicator whether cover the param limit.
func (limit *BaseLimit) Greater(x Limit) bool {
	if x == nil {
		return true
	}
	if limit.Memory < x.GetMemoryLimit() {
		return false
	}
	if limit.FD < x.GetFDLimit() {
		return false
	}
	if limit.Conns < x.GetConnTotalLimit() {
		return false
	}
	if limit.ConnsInbound < x.GetConnLimit(DirInbound) {
		return false
	}
	if limit.ConnsOutbound < x.GetConnLimit(DirOutbound) {
		return false
	}
	if limit.Tasks < x.GetTaskTotalLimit() {
		return false
	}
	if limit.TasksHighPriority < x.GetTaskLimit(ReserveTaskPriorityHigh) {
		return false
	}
	if limit.TasksMediumPriority < x.GetTaskLimit(ReserveTaskPriorityMedium) {
		return false
	}
	if limit.TasksLowPriority < x.GetTaskLimit(ReserveTaskPriorityLow) {
		return false
	}
	return true
}

// Equal returns true iff limit is the same with x.
func (limit *BaseLimit) Equal(x Limit) bool {
	if x == nil {
		return false
	}
	if limit.Memory != x.GetMemoryLimit() {
		return false
	}
	if limit.FD != x.GetFDLimit() {
		return false
	}
	if limit.Conns != x.GetConnTotalLimit() {
		return false
	}
	if limit.ConnsInbound != x.GetConnLimit(DirInbound) {
		return false
	}
	if limit.ConnsOutbound != x.GetConnLimit(DirOutbound) {
		return false
	}
	if limit.Tasks != x.GetTaskTotalLimit() {
		return false
	}
	if limit.TasksHighPriority != x.GetTaskLimit(ReserveTaskPriorityHigh) {
		return false
	}
	if limit.TasksMediumPriority != x.GetTaskLimit(ReserveTaskPriorityMedium) {
		return false
	}
	if limit.TasksLowPriority != x.GetTaskLimit(ReserveTaskPriorityLow) {
		return false
	}
	return true
}

// String returns the Limit state string.
// TODO:: supports connection and fd field
func (limit *BaseLimit) String() string {
	return fmt.Sprintf("memory limits %d, task limits [h: %d, m: %d, l: %d]",
		limit.Memory, limit.TasksHighPriority, limit.TasksMediumPriority, limit.TasksLowPriority)
}

// InfiniteLimit returns a limiter that uses unlimited limits, thus effectively not limiting anything.
// Keep in mind that the operating system limits the number of file descriptors that an application can use.
func InfiniteLimit() Limit {
	return &BaseLimit{
		Memory:              MaxLimitInt64,
		FD:                  MaxLimitInt,
		Conns:               MaxLimitInt,
		ConnsInbound:        MaxLimitInt,
		ConnsOutbound:       MaxLimitInt,
		Tasks:               MaxLimitInt,
		TasksHighPriority:   MaxLimitInt,
		TasksMediumPriority: MaxLimitInt,
		TasksLowPriority:    MaxLimitInt,
	}
}

// InfinitesimalLimit returns a limiter that uses zero limits, thus effectively limiting anything.
func InfinitesimalLimit() Limit {
	return &BaseLimit{}
}

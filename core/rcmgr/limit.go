package rcmgr

import (
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
	ScopeStat() *ScopeStat
	// NotLess returns an indicator whether cover the param limit fields.
	NotLess(Limit) bool
	Add(Limit)
	Sub(Limit) bool
	// Equal returns an indicator whether equal the param limit.
	Equal(Limit) bool
	// String returns the Limit state string.
	String() string
}

// Limiter is an interface for providing limits to the resource manager.
type Limiter interface {
	// GetSystemLimits returns the system limits.
	GetSystemLimits() Limit
	// GetTransientLimits returns the transient limits.
	GetTransientLimits() Limit
	// GetServiceLimits returns a service-specific limits.
	GetServiceLimits(svc string) Limit
	// String returns the all kinds of Limit state string
	String() string
}

var _ Limiter = (*NullLimit)(nil)
var _ Limit = (*NullLimit)(nil)

// NullLimit is a stub for tests and initialization of default values
type NullLimit struct{}

func (n *NullLimit) GetSystemLimits() Limit               { return nil }
func (n *NullLimit) GetTransientLimits() Limit            { return nil }
func (n *NullLimit) GetServiceLimits(svc string) Limit    { return nil }
func (n *NullLimit) String() string                       { return "null limit" }
func (n *NullLimit) GetMemoryLimit() int64                { return 0 }
func (n *NullLimit) GetFDLimit() int                      { return 0 }
func (n *NullLimit) GetConnLimit(Direction) int           { return 0 }
func (n *NullLimit) GetConnTotalLimit() int               { return 0 }
func (n *NullLimit) GetTaskLimit(ReserveTaskPriority) int { return 0 }
func (n *NullLimit) GetTaskTotalLimit() int               { return 0 }
func (n *NullLimit) NotLess(Limit) bool                   { return false }
func (n *NullLimit) Add(Limit)                            { return }
func (n *NullLimit) Sub(Limit) bool                       { return false }
func (n *NullLimit) Equal(Limit) bool                     { return false }
func (n *NullLimit) ScopeStat() *ScopeStat                { return nil }

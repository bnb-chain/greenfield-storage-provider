package rcmgr

import (
	"fmt"
	"math"

	"github.com/shirou/gopsutil/mem"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const (
	LimitFactor              = 0.85
	DefaultMemorySize uint64 = 8 * 1024 * 1024
)

// Limit is an interface that that specifies basic resource limits.
type Limit interface {
	// GetMemoryLimit returns the (current) memory limit.
	GetMemoryLimit() int64
	// GetConnLimit returns the connection limit, for inbound or outbound connections.
	GetConnLimit(Direction) int
	// GetConnTotalLimit returns the total connection limit
	GetConnTotalLimit() int
	// GetFDLimit returns the file descriptor limit.
	GetFDLimit() int
	// String returns the Limit state string
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
	Conns         int   `json:",omitempty"`
	ConnsInbound  int   `json:",omitempty"`
	ConnsOutbound int   `json:",omitempty"`
	FD            int   `json:",omitempty"`
	Memory        int64 `json:",omitempty"`
}

// GetMemoryLimit returns the (current) memory limit.
func (limit *BaseLimit) GetMemoryLimit() int64 {
	return limit.Memory
}

// GetConnLimit returns the connection limit, for inbound or outbound connections.
func (limit *BaseLimit) GetConnLimit(direction Direction) int {
	if direction == DirInbound {
		return limit.ConnsInbound
	}
	return limit.ConnsOutbound
}

// GetConnTotalLimit returns the total connection limit
func (limit *BaseLimit) GetConnTotalLimit() int {
	return limit.Conns
}

// GetFDLimit returns the file descriptor limit.
func (limit *BaseLimit) GetFDLimit() int {
	return limit.FD
}

// String returns the Limit state string
// TODO:: supports connection and fd field
func (limit *BaseLimit) String() string {
	return fmt.Sprintf("memory limits %d", limit.Memory)
}

// InfiniteBaseLimit are a limiter configuration that uses unlimited limits, thus effectively not limiting anything.
// Keep in mind that the operating system limits the number of file descriptors that an application can use.
var InfiniteBaseLimit = BaseLimit{
	Conns:         math.MaxInt,
	ConnsInbound:  math.MaxInt,
	ConnsOutbound: math.MaxInt,
	FD:            math.MaxInt,
	Memory:        math.MaxInt64,
}

// DynamicLimits generate limits by os resource
func DynamicLimits() *BaseLimit {
	availableMem := DefaultMemorySize
	virtualMem, err := mem.VirtualMemory()
	if err != nil {
		log.Errorw("failed to get os memory states", "error", err)
	} else {
		availableMem = virtualMem.Available
	}
	limits := &BaseLimit{}
	limits.Memory = int64(float64(availableMem) * LimitFactor)
	// TODO:: get from os and compatible with a variety of os
	limits.FD = math.MaxInt
	limits.Conns = math.MaxInt
	limits.ConnsInbound = math.MaxInt
	limits.ConnsOutbound = math.MaxInt
	return limits
}

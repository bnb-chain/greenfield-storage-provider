package rcmgr

import (
	"fmt"
)

// SP Resource Manager reference go-libp2p, see:
// https://github.com/libp2p/go-libp2p/blob/master/core/network/
// https://github.com/libp2p/go-libp2p/tree/master/p2p/host/resource-manager

// ResourceManager is the interface to the resource management subsystem.
// The ResourceManager tracks and accounts for resource usage in the stack,
// from the internals to the application, and provides a mechanism to limit
// resource usage according to a user configurable policy.
//
// Resource Management through the ResourceManager is based on the concept of
// Resource Management Scopes, whereby resource usage is constrained by a DAG of
// scopes, The following diagram illustrates the structure of the resource
// constraint DAG:
// System
//
//	+------------> Transient........................+................+
//	|                                               .                .
//	+------------> Service------------------------- . ----------+    .
//	|                                               .           |    .
//			   	   +--->  Connection/Memory/Task--- . ----------+    .
//
// The basic resources accounted by the ResourceManager include memory, connections,
// file  descriptors and task. These account for both space and time used by the stack,
// as each resource has a direct effect on the system availability and performance.
//
// The module of the resource manager is to restrict resource usage at the
// time of reservation. When a component of the stack needs to use a resource, it
// reserves it in the appropriate scope. The resource manager gates the reservation
// against the scope applicable limits; if the limit is exceeded, then an error is up
// the component to act accordingly. At the lower levels of the stack, this will normally
// signal a failure of some sorts, like failing to opening a connection, which will
// propagate to the programmer. Some components may be able to handle resource reservation
// failure more gracefully.
// All resources reserved in some scopes are released when the scope is closed. For low
// level scopes, mainly Service and Connection scopes, this happens when the service or
// connection is closed.
//
// Service programmers will typically use the resource manager to reserve memory for their
// subsystem.
// This happens with two avenues: the programmer can attach a connection to a service, whereby
// resources reserved by the connection are automatically accounted in the service budget;
// or the programmer may directly interact with the service scope, by using ViewService through
// the resource manager interface.
//
// Application programmers can also directly reserve memory in some applicable scope. In order
// to facilitate control flow delimited resource accounting, all scopes defined in the system
// allow for the user to create spans. Spans are temporary scopes rooted at some
// other scope and release their resources when the programmer is done with them. Span
// scopes can form trees, with nested spans.
//
// Typical Usage:
//   - Low level components of the system  all have access to the resource manager and create
//     connection scopes through it. These scopes are accessible to the user, albeit with a
//     narrower interface, through Conn objects who have a Scope method.
//   - Services typically center around connections, where the programmer can attach connections
//     to a particular service. They can also directly reserve memory for a service by accessing
//     the service scope using the ResourceManager interface.
//   - Applications that want to account for their resource usage can reserve memory, typically
//     using a span, directly in the System or a Service scope.
type ResourceManager interface {
	ResourceScopeViewer
	// OpenService creates a new Service scope associated with System/Transient Scope
	// The caller owns the returned scope and is responsible for calling Done in order
	// to signify the end of the scope's span.
	OpenService(svc string) (ResourceScope, error)
	// Close closes the resource manager
	Close() error
}

// ResourceScopeViewer is a mixin interface providing view methods for accessing top level scopes
type ResourceScopeViewer interface {
	// ViewSystem views the system-wide resource scope.
	// The system scope is the top level scope that accounts for global
	// resource usage at all levels of the system. This scope constrains all
	// other scopes and institutes global hard limits.
	ViewSystem(func(ResourceScope) error) error
	// ViewTransient views the transient (DMZ) resource scope.
	// The transient scope accounts for resources that are in the process of
	// full establishment.  For instance, a new connection prior to the
	// handshake does not belong to any service, but it still needs to be
	// constrained as this opens an avenue for attacks in transient resource
	// usage.
	ViewTransient(func(ResourceScope) error) error
	// ViewService retrieves a service-specific scope.
	ViewService(string, func(ResourceScope) error) error
	// SystemState output the system resource scope and limit readable
	SystemState() string
	// TransientState output the transient (DMZ)  resource scope and limit readable
	TransientState() string
	// ServiceState output a service-specific resource scope and limit readable
	ServiceState(string) string
}

const (
	// ReservationPriorityLow is a reservation priority that indicates a reservation if the scope
	// memory utilization is at 40% or less.
	ReservationPriorityLow uint8 = 101
	// ReservationPriorityMedium is a reservation priority that indicates a reservation if the scope
	// memory utilization is at 60% or less.
	ReservationPriorityMedium uint8 = 152
	// ReservationPriorityHigh is a reservation priority that indicates a reservation if the scope
	// memory utilization is at 80% or less.
	ReservationPriorityHigh uint8 = 203
	// ReservationPriorityAlways is a reservation priority that indicates a reservation if there is
	// enough memory, regardless of scope utilization.
	ReservationPriorityAlways uint8 = 255
)

// ResourceScope is the interface for all scopes.
type ResourceScope interface {
	// ReserveMemory reserves memory/buffer space in the scope; the unit is bytes.
	//
	// If ReserveMemory returns an error, then no memory was reserved and the caller
	// should handle the failure condition.
	//
	// The priority argument indicates the priority of the memory reservation. A reservation
	// will fail if the available memory is less than (1+prio)/256 of the scope limit, providing
	// a mechanism to gracefully handle optional reservations that might overload the system.
	//
	// There are 4 predefined priority levels, Low, Medium, High and Always, capturing common
	// patterns, but the user is free to use any granularity applicable to his case.
	ReserveMemory(size int64, prio uint8) error
	// ReleaseMemory explicitly releases memory previously reserved with ReserveMemory
	ReleaseMemory(size int64)
	// AddTask reserves task by ReserveTaskPriority in the scope.
	//
	// If ReserveTask returns an error, then no task quota was reserved and the caller
	// should handle the failure condition.
	AddTask(num int, prio ReserveTaskPriority) error
	// RemoveTask explicitly releases task reserved with AddTask.
	RemoveTask(num int, prio ReserveTaskPriority)
	// AddConn reserves connection by Direction in the scope.
	AddConn(dir Direction) error
	// RemoveConn explicitly releases connection reserved with AddConn.
	RemoveConn(dir Direction)
	// ReserveResources reserves the resource by ScopeStat, it will reserve all kinds
	// of resources atomic
	ReserveResources(st *ScopeStat) error
	// RemainingResource returns the remaining resource that have not be reserved
	RemainingResource() (Limit, error)
	// Stat retrieves current resource usage for the scope.
	Stat() ScopeStat
	// Name returns the name of this scope
	Name() string
	// BeginSpan creates a new span scope rooted at this scope
	BeginSpan() (ResourceScopeSpan, error)
	// Release resource at this scope
	Release()
}

// ResourceScopeSpan is a ResourceScope with a delimited span.
// Span scopes are control flow delimited and release all their associated resources
// when the programmer calls Done.
//
// Example:
//
//	s, err := someScope.BeginSpan()
//	if err != nil { ... }
//	defer s.Done()
//
//	if err := s.ReserveMemory(...); err != nil { ... }
//	// ... use memory
type ResourceScopeSpan interface {
	ResourceScope
	// Done ends the span and releases associated resources.
	Done()
}

// Direction represents which peer in a stream initiated a connection.
type Direction int

const (
	// DirUnknown is the default direction.
	DirUnknown Direction = iota
	// DirInbound is for when the remote peer initiated a connection.
	DirInbound
	// DirOutbound is for when the local peer initiated a connection.
	DirOutbound
)

// ReserveTaskPriority represents which priority in a reservation of task.
type ReserveTaskPriority int

const (
	// ReserveTaskPriorityUnknown is the default task priority.
	ReserveTaskPriorityUnknown ReserveTaskPriority = iota
	// ReserveTaskPriorityHigh is a task reservation priority that indicates a
	// reservation consume a high task limit.
	ReserveTaskPriorityHigh
	// ReserveTaskPriorityMedium is a task reservation priority that indicates a
	// reservation consume a medium task limit.
	ReserveTaskPriorityMedium
	// ReserveTaskPriorityLow is a task reservation priority that indicates a
	// reservation consume a low task limit.
	ReserveTaskPriorityLow
)

// ScopeStat is a struct containing resource accounting information.
type ScopeStat struct {
	Memory           int64
	NumTasksHigh     int64
	NumTasksMedium   int64
	NumTasksLow      int64
	NumConnsInbound  int64
	NumConnsOutbound int64
	NumFD            int64
}

// String returns the state string of ScopeStat
// TODO:: supports connections and fd field
func (s ScopeStat) String() string {
	return fmt.Sprintf("memory reserved [%d], task reserved[h: %d, m: %d, l: %d]",
		s.Memory, s.NumTasksHigh, s.NumTasksMedium, s.NumTasksLow)
}

var _ ResourceManager = &NullResourceManager{}

// NullResourceManager is a stub for tests and initialization of default values
type NullResourceManager struct{}

func (n *NullResourceManager) ViewSystem(f func(ResourceScope) error) error {
	return f(&NullScope{})
}
func (n *NullResourceManager) ViewTransient(f func(ResourceScope) error) error {
	return f(&NullScope{})
}
func (n *NullResourceManager) ViewService(svc string, f func(ResourceScope) error) error {
	return f(&NullScope{})
}
func (n *NullResourceManager) SystemLimitString(func(ResourceScope) string) string {
	return ""
}
func (n *NullResourceManager) TransientLimitString(func(ResourceScope) string) string {
	return ""
}
func (n *NullResourceManager) ServiceLimitString(string, func(ResourceScope) string) string {
	return ""
}
func (n *NullResourceManager) SystemState() string        { return "" }
func (n *NullResourceManager) TransientState() string     { return "" }
func (n *NullResourceManager) ServiceState(string) string { return "" }
func (n *NullResourceManager) OpenService(svc string) (ResourceScope, error) {
	return &NullScope{}, nil
}
func (n *NullResourceManager) Close() error {
	return nil
}

var _ ResourceScope = (*NullScope)(nil)
var _ ResourceScopeSpan = (*NullScope)(nil)

// NullScope is a stub for tests and initialization of default values
type NullScope struct{}

func (n *NullScope) RemainingResource() (Limit, error)               { return &Unlimited{}, nil }
func (n *NullScope) ReserveResources(st *ScopeStat) error            { return nil }
func (n *NullScope) ReserveMemory(size int64, prio uint8) error      { return nil }
func (n *NullScope) ReleaseMemory(size int64)                        {}
func (n *NullScope) AddTask(num int, prio ReserveTaskPriority) error { return nil }
func (n *NullScope) RemoveTask(num int, prio ReserveTaskPriority)    {}
func (n *NullScope) AddConn(dir Direction) error                     { return nil }
func (n *NullScope) RemoveConn(dir Direction)                        {}
func (n *NullScope) Stat() ScopeStat                                 { return ScopeStat{} }
func (n *NullScope) BeginSpan() (ResourceScopeSpan, error)           { return &NullScope{}, nil }
func (n *NullScope) Done()                                           {}
func (n *NullScope) Name() string                                    { return "" }
func (n *NullScope) Release()                                        {}

package gfsprcmgr

import (
	"fmt"
	"sync"

	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var _ corercmgr.ResourceScope = &resourceScope{}
var _ corercmgr.ResourceScopeSpan = &resourceScope{}

// A resourceScope can be a DAG, where a downstream node is not allowed to outlive an upstream node
// (ie cannot call Done in the upstream node before the downstream node) and account for resources
// using a linearized parent set.
// A resourceScope can be a span scope, where it has a specific owner; span scopes create a tree rooted
// at the owner (which can be a DAG scope) and can outlive their parents -- this is important because
// span scopes are the main *user* interface for memory management, and the user may call
// Done in a span scope after the system has closed the root of the span tree in some background
// goroutine.
// If we didn't make this distinction we would have a double release problem in that case.
type resourceScope struct {
	sync.Mutex
	done   bool
	refCnt int
	spanID int

	rc    resources
	owner *resourceScope   // set in span scopes, which define trees
	edges []*resourceScope // set in DAG scopes, it's the linearized parent set

	name string // for debugging purposes
}

// newResourceScope returns an instance of resourceScope.
func newResourceScope(
	limit corercmgr.Limit,
	edges []*resourceScope,
	name string) *resourceScope {
	for _, e := range edges {
		e.IncRef()
	}
	r := &resourceScope{
		rc:    resources{limit: limit},
		edges: edges,
		name:  name,
	}
	return r
}

// newResourceScopeSpan returns an instance of span resourceScope.
func newResourceScopeSpan(
	owner *resourceScope,
	id int, name string) *resourceScope {
	r := &resourceScope{
		rc:    resources{limit: owner.rc.limit},
		owner: owner,
		name:  fmt.Sprintf("%s.span-%s-%d", owner.name, name, id),
	}
	return r
}

// BeginSpan creates a new span scope rooted at this scope.
func (s *resourceScope) BeginSpan() (corercmgr.ResourceScopeSpan, error) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return nil, s.wrapError(ErrResourceScopeClosed)
	}
	s.refCnt++
	return newResourceScopeSpan(s, s.nextSpanID(), "temp"), nil
}

// Done ends the span and releases associated resources.
func (s *resourceScope) Done() {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return
	}
	stat := s.rc.stat()
	if s.owner != nil {
		s.owner.ReleaseResources(stat)
		s.owner.DecRef()
	} else {
		for _, e := range s.edges {
			e.ReleaseForChild(stat)
			e.DecRef()
		}
	}
	s.rc.nconnsIn = 0
	s.rc.nconnsOut = 0
	s.rc.nfd = 0
	s.rc.memory = 0

	s.done = true
}

// ReserveMemory reserves memory/buffer space in the scope; the unit is bytes.
func (s *resourceScope) ReserveMemory(size int64, prio uint8) error {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return s.wrapError(ErrResourceScopeClosed)
	}
	if err := s.rc.reserveMemory(size, prio); err != nil {
		log.Debugw("blocked memory reservation", logValuesMemoryLimit(s.name, "", s.rc.stat(), err)...)
		return s.wrapError(err)
	}
	if err := s.reserveMemoryForEdges(size, prio); err != nil {
		s.rc.releaseMemory(size)
		return s.wrapError(err)
	}
	return nil
}

func (s *resourceScope) reserveMemoryForEdges(size int64, prio uint8) error {
	if s.owner != nil {
		return s.owner.ReserveMemory(size, prio)
	}
	var reserved int
	var err error
	for _, e := range s.edges {
		var stat corercmgr.ScopeStat
		stat, err = e.ReserveMemoryForChild(size, prio)
		if err != nil {
			log.Debugw("blocked memory reservation from constraining edge", logValuesMemoryLimit(s.name, e.name, stat, err)...)
			break
		}
		reserved++
	}
	if err != nil {
		// we failed because of a constraint; undo memory reservations
		for _, e := range s.edges[:reserved] {
			e.ReleaseMemoryForChild(size)
		}
	}
	return err
}

// ReserveMemoryForChild reserves memory/buffer space in the child scope.
func (s *resourceScope) ReserveMemoryForChild(
	size int64, prio uint8) (corercmgr.ScopeStat, error) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return s.rc.stat(), s.wrapError(ErrResourceScopeClosed)
	}
	if err := s.rc.reserveMemory(size, prio); err != nil {
		return s.rc.stat(), s.wrapError(err)
	}
	return corercmgr.ScopeStat{}, nil
}

// ReleaseMemory explicitly releases memory previously reserved with ReserveMemory.
func (s *resourceScope) ReleaseMemory(size int64) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return
	}
	s.rc.releaseMemory(size)
	s.releaseMemoryForEdges(size)
}

func (s *resourceScope) releaseMemoryForEdges(size int64) {
	if s.owner != nil {
		s.owner.ReleaseMemory(size)
		return
	}
	for _, e := range s.edges {
		e.ReleaseMemoryForChild(size)
	}
}

// ReleaseMemoryForChild explicitly releases memory in the child scope.
func (s *resourceScope) ReleaseMemoryForChild(size int64) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return
	}
	s.rc.releaseMemory(size)
}

// AddTask reserves task in the scope.
func (s *resourceScope) AddTask(num int, prio corercmgr.ReserveTaskPriority) error {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return s.wrapError(ErrResourceScopeClosed)
	}
	if err := s.rc.addTask(num, prio); err != nil {
		log.Debugw("blocked task", logValuesTaskLimit(s.name, "", s.rc.stat(), err)...)
		return s.wrapError(err)
	}
	if err := s.addTaskForEdges(num, prio); err != nil {
		s.rc.removeTask(num, prio)
		return s.wrapError(err)
	}
	return nil
}

func (s *resourceScope) addTaskForEdges(num int,
	prio corercmgr.ReserveTaskPriority) error {
	if s.owner != nil {
		return s.owner.AddTask(num, prio)
	}
	var err error
	var reserved int
	for _, e := range s.edges {
		var stat corercmgr.ScopeStat
		stat, err = e.AddTaskForChild(num, prio)
		if err != nil {
			log.Debugw("blocked task from constraining edge", logValuesTaskLimit(s.name, e.name, stat, err)...)
			break
		}
		reserved++
	}
	if err != nil {
		for _, e := range s.edges[:reserved] {
			e.RemoveTaskForChild(num, prio)
		}
	}
	return err
}

// AddTaskForChild reserves task in the child scope.
func (s *resourceScope) AddTaskForChild(num int,
	prio corercmgr.ReserveTaskPriority) (corercmgr.ScopeStat, error) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return s.rc.stat(), s.wrapError(ErrResourceScopeClosed)
	}
	if err := s.rc.addTask(num, prio); err != nil {
		return s.rc.stat(), s.wrapError(err)
	}
	return corercmgr.ScopeStat{}, nil
}

// RemoveTask explicitly releases task in the scope.
func (s *resourceScope) RemoveTask(num int,
	prio corercmgr.ReserveTaskPriority) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return
	}
	s.rc.removeTask(num, prio)
	s.removeTaskForEdges(num, prio)
}

func (s *resourceScope) removeTaskForEdges(num int,
	prio corercmgr.ReserveTaskPriority) {
	if s.owner != nil {
		s.owner.RemoveTask(num, prio)
	}
	for _, e := range s.edges {
		e.RemoveTaskForChild(num, prio)
	}
}

// RemoveTaskForChild explicitly releases task in child the scope.
func (s *resourceScope) RemoveTaskForChild(num int,
	prio corercmgr.ReserveTaskPriority) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return
	}
	s.rc.removeTask(num, prio)
}

// AddConn reserves connection in the scope.
func (s *resourceScope) AddConn(dir corercmgr.Direction) error {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return s.wrapError(ErrResourceScopeClosed)
	}
	if err := s.rc.addConn(dir); err != nil {
		log.Debugw("blocked connection", logValuesConnLimit(s.name, "", dir, s.rc.stat(), err)...)
		return s.wrapError(err)
	}
	if err := s.addConnForEdges(dir); err != nil {
		s.rc.removeConn(dir)
		return s.wrapError(err)
	}
	return nil
}

func (s *resourceScope) addConnForEdges(dir corercmgr.Direction) error {
	if s.owner != nil {
		return s.owner.AddConn(dir)
	}
	var err error
	var reserved int
	for _, e := range s.edges {
		var stat corercmgr.ScopeStat
		stat, err = e.AddConnForChild(dir)
		if err != nil {
			log.Debugw("blocked connection from constraining edge", logValuesConnLimit(s.name, e.name, dir, stat, err)...)
			break
		}
		reserved++
	}
	if err != nil {
		for _, e := range s.edges[:reserved] {
			e.RemoveConnForChild(dir)
		}
	}
	return err
}

// AddConnForChild reserves connection in the child scope.
func (s *resourceScope) AddConnForChild(dir corercmgr.Direction) (
	corercmgr.ScopeStat, error) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return s.rc.stat(), s.wrapError(ErrResourceScopeClosed)
	}
	if err := s.rc.addConn(dir); err != nil {
		return s.rc.stat(), s.wrapError(err)
	}
	return corercmgr.ScopeStat{}, nil
}

// RemoveConn explicitly releases connection in the scope.
func (s *resourceScope) RemoveConn(dir corercmgr.Direction) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return
	}
	s.rc.removeConn(dir)
	s.removeConnForEdges(dir)
}

func (s *resourceScope) removeConnForEdges(dir corercmgr.Direction) {
	if s.owner != nil {
		s.owner.RemoveConn(dir)
	}
	for _, e := range s.edges {
		e.RemoveConnForChild(dir)
	}
}

// RemoveConnForChild explicitly releases connection in child the scope.
func (s *resourceScope) RemoveConnForChild(dir corercmgr.Direction) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return
	}
	s.rc.removeConn(dir)
}

// ReserveForChild explicitly releases connection in child the scope.
func (s *resourceScope) ReserveForChild(st corercmgr.ScopeStat) error {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return s.wrapError(ErrResourceScopeClosed)
	}
	if err := s.rc.reserveMemory(st.Memory, corercmgr.ReservationPriorityAlways); err != nil {
		return s.wrapError(err)
	}
	if err := s.rc.addConns(int(st.NumConnsInbound), int(st.NumConnsOutbound), int(st.NumFD)); err != nil {
		s.rc.releaseMemory(st.Memory)
		return s.wrapError(err)
	}
	if err := s.rc.addTask(int(st.NumTasksHigh), corercmgr.ReserveTaskPriorityHigh); err != nil {
		s.rc.releaseMemory(st.Memory)
		s.rc.removeConns(int(st.NumConnsInbound), int(st.NumConnsOutbound), int(st.NumFD))
		return s.wrapError(err)
	}
	if err := s.rc.addTask(int(st.NumTasksHigh), corercmgr.ReserveTaskPriorityMedium); err != nil {
		s.rc.releaseMemory(st.Memory)
		s.rc.removeConns(int(st.NumConnsInbound), int(st.NumConnsOutbound), int(st.NumFD))
		s.rc.removeTask(int(st.NumTasksHigh), corercmgr.ReserveTaskPriorityHigh)
		return s.wrapError(err)
	}
	if err := s.rc.addTask(int(st.NumTasksHigh), corercmgr.ReserveTaskPriorityLow); err != nil {
		s.rc.releaseMemory(st.Memory)
		s.rc.removeConns(int(st.NumConnsInbound), int(st.NumConnsOutbound), int(st.NumFD))
		s.rc.removeTask(int(st.NumTasksHigh), corercmgr.ReserveTaskPriorityHigh)
		s.rc.removeTask(int(st.NumTasksMedium), corercmgr.ReserveTaskPriorityMedium)
		return s.wrapError(err)
	}
	return nil
}

// ReleaseResources explicitly releases resource by ScopeStat.
func (s *resourceScope) ReleaseResources(st corercmgr.ScopeStat) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return
	}
	s.rc.releaseMemory(st.Memory)
	s.rc.removeConns(int(st.NumConnsInbound), int(st.NumConnsOutbound), int(st.NumFD))
	s.rc.removeTask(int(st.NumTasksHigh), corercmgr.ReserveTaskPriorityHigh)
	s.rc.removeTask(int(st.NumTasksMedium), corercmgr.ReserveTaskPriorityMedium)
	s.rc.removeTask(int(st.NumTasksLow), corercmgr.ReserveTaskPriorityLow)
	if s.owner != nil {
		s.owner.ReleaseResources(st)
	} else {
		for _, e := range s.edges {
			e.ReleaseForChild(st)
		}
	}
}

// Release explicitly releases self resource.
func (s *resourceScope) Release() {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return
	}
	st := s.rc.stat()
	s.rc.releaseMemory(st.Memory)
	s.rc.removeConns(int(st.NumConnsInbound), int(st.NumConnsOutbound), int(st.NumFD))
	s.rc.removeTask(int(st.NumTasksHigh), corercmgr.ReserveTaskPriorityHigh)
	s.rc.removeTask(int(st.NumTasksMedium), corercmgr.ReserveTaskPriorityMedium)
	s.rc.removeTask(int(st.NumTasksLow), corercmgr.ReserveTaskPriorityLow)
	if s.owner != nil {
		s.owner.ReleaseResources(st)
	} else {
		for _, e := range s.edges {
			e.ReleaseForChild(st)
		}
	}
}

// ReleaseForChild explicitly releases resource in the child scope by ScopeStat.
func (s *resourceScope) ReleaseForChild(st corercmgr.ScopeStat) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return
	}
	s.rc.releaseMemory(st.Memory)
	s.rc.removeConns(int(st.NumConnsInbound), int(st.NumConnsOutbound), int(st.NumFD))
	s.rc.removeTask(int(st.NumTasksHigh), corercmgr.ReserveTaskPriorityHigh)
	s.rc.removeTask(int(st.NumTasksMedium), corercmgr.ReserveTaskPriorityMedium)
	s.rc.removeTask(int(st.NumTasksLow), corercmgr.ReserveTaskPriorityLow)
}

func (s *resourceScope) ReserveResources(st *corercmgr.ScopeStat) error {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return s.wrapError(ErrResourceScopeClosed)
	}
	if err := s.rc.reserveMemory(st.Memory, corercmgr.ReservationPriorityAlways); err != nil {
		return s.wrapError(err)
	}
	if err := s.rc.addConns(int(st.NumConnsInbound), int(st.NumConnsOutbound), int(st.NumFD)); err != nil {
		s.rc.releaseMemory(st.Memory)
		return s.wrapError(err)
	}
	if err := s.rc.addTask(int(st.NumTasksHigh), corercmgr.ReserveTaskPriorityHigh); err != nil {
		s.rc.releaseMemory(st.Memory)
		s.rc.removeConns(int(st.NumConnsInbound), int(st.NumConnsOutbound), int(st.NumFD))
		return s.wrapError(err)
	}
	if err := s.rc.addTask(int(st.NumTasksHigh), corercmgr.ReserveTaskPriorityMedium); err != nil {
		s.rc.releaseMemory(st.Memory)
		s.rc.removeConns(int(st.NumConnsInbound), int(st.NumConnsOutbound), int(st.NumFD))
		s.rc.removeTask(int(st.NumTasksHigh), corercmgr.ReserveTaskPriorityHigh)
		return s.wrapError(err)
	}
	if err := s.rc.addTask(int(st.NumTasksHigh), corercmgr.ReserveTaskPriorityLow); err != nil {
		s.rc.releaseMemory(st.Memory)
		s.rc.removeConns(int(st.NumConnsInbound), int(st.NumConnsOutbound), int(st.NumFD))
		s.rc.removeTask(int(st.NumTasksHigh), corercmgr.ReserveTaskPriorityHigh)
		s.rc.removeTask(int(st.NumTasksMedium), corercmgr.ReserveTaskPriorityMedium)
		return s.wrapError(err)
	}

	var err error
	if s.owner != nil {
		err = s.owner.ReserveResources(st)
	} else {
		for _, e := range s.edges {
			err = e.ReserveForChild(*st)
			break
		}
	}
	if err != nil {
		s.ReleaseResources(*st)
		return err
	}
	return nil
}

func (s *resourceScope) RemainingResource() (corercmgr.Limit, error) {
	return s.rc.remaining(), nil
}

// Name returns the name of scope.
func (s *resourceScope) Name() string {
	return s.name
}

// Stat returns the state of scope.
func (s *resourceScope) Stat() corercmgr.ScopeStat {
	s.Lock()
	defer s.Unlock()
	return s.rc.stat()
}

func (s *resourceScope) IncRef() {
	s.Lock()
	defer s.Unlock()
	s.refCnt++
}

func (s *resourceScope) DecRef() {
	s.Lock()
	defer s.Unlock()
	s.refCnt--
}

func (s *resourceScope) IsUnused() bool {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return true
	}
	if s.refCnt > 0 {
		return false
	}
	st := s.rc.stat()
	return st.NumConnsInbound == 0 && st.NumConnsOutbound == 0 && st.NumFD == 0
}

func (s *resourceScope) NextSpanID() int {
	s.Lock()
	defer s.Unlock()
	s.spanID++
	return s.spanID
}

func (s *resourceScope) nextSpanID() int {
	s.spanID++
	return s.spanID
}

func (s *resourceScope) wrapError(err error) error {
	return fmt.Errorf("%s: %w", s.name, err)
}

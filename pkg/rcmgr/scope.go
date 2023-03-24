package rcmgr

import (
	"fmt"
	"sync"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var _ ResourceScope = &resourceScope{}
var _ ResourceScopeSpan = &resourceScope{}

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
func newResourceScope(limit Limit, edges []*resourceScope, name string) *resourceScope {
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
func newResourceScopeSpan(owner *resourceScope, id int, name string) *resourceScope {
	r := &resourceScope{
		rc:    resources{limit: owner.rc.limit},
		owner: owner,
		name:  fmt.Sprintf("%s.span-%s-%d", owner.name, name, id),
	}
	return r
}

// BeginSpan creates a new span scope rooted at this scope.
func (s *resourceScope) BeginSpan() (ResourceScopeSpan, error) {
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
func (s *resourceScope) ReserveMemory(size int, prio uint8) error {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return s.wrapError(ErrResourceScopeClosed)
	}
	if err := s.rc.reserveMemory(int64(size), prio); err != nil {
		log.Debugw("blocked memory reservation", logValuesMemoryLimit(s.name, "", s.rc.stat(), err)...)
		return s.wrapError(err)
	}
	if err := s.reserveMemoryForEdges(size, prio); err != nil {
		s.rc.releaseMemory(int64(size))
		return s.wrapError(err)
	}
	return nil
}

func (s *resourceScope) reserveMemoryForEdges(size int, prio uint8) error {
	if s.owner != nil {
		return s.owner.ReserveMemory(size, prio)
	}
	var reserved int
	var err error
	for _, e := range s.edges {
		var stat ScopeStat
		stat, err = e.ReserveMemoryForChild(int64(size), prio)
		if err != nil {
			log.Debugw("blocked memory reservation from constraining edge", logValuesMemoryLimit(s.name, e.name, stat, err)...)
			break
		}
		reserved++
	}
	if err != nil {
		// we failed because of a constraint; undo memory reservations
		for _, e := range s.edges[:reserved] {
			e.ReleaseMemoryForChild(int64(size))
		}
	}
	return err
}

// ReserveMemoryForChild reserves memory/buffer space in the child scope.
func (s *resourceScope) ReserveMemoryForChild(size int64, prio uint8) (ScopeStat, error) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return s.rc.stat(), s.wrapError(ErrResourceScopeClosed)
	}
	if err := s.rc.reserveMemory(size, prio); err != nil {
		return s.rc.stat(), s.wrapError(err)
	}
	return ScopeStat{}, nil
}

// ReleaseMemory explicitly releases memory previously reserved with ReserveMemory.
func (s *resourceScope) ReleaseMemory(size int) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return
	}
	s.rc.releaseMemory(int64(size))
	s.releaseMemoryForEdges(size)
}

func (s *resourceScope) releaseMemoryForEdges(size int) {
	if s.owner != nil {
		s.owner.ReleaseMemory(size)
		return
	}
	for _, e := range s.edges {
		e.ReleaseMemoryForChild(int64(size))
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

// AddConn reserves connection in the scope.
func (s *resourceScope) AddConn(dir Direction) error {
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

func (s *resourceScope) addConnForEdges(dir Direction) error {
	if s.owner != nil {
		return s.owner.AddConn(dir)
	}
	var err error
	var reserved int
	for _, e := range s.edges {
		var stat ScopeStat
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
func (s *resourceScope) AddConnForChild(dir Direction) (ScopeStat, error) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return s.rc.stat(), s.wrapError(ErrResourceScopeClosed)
	}
	if err := s.rc.addConn(dir); err != nil {
		return s.rc.stat(), s.wrapError(err)
	}
	return ScopeStat{}, nil
}

// RemoveConn explicitly releases connection in the scope.
func (s *resourceScope) RemoveConn(dir Direction) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return
	}
	s.rc.removeConn(dir)
	s.removeConnForEdges(dir)
}

func (s *resourceScope) removeConnForEdges(dir Direction) {
	if s.owner != nil {
		s.owner.RemoveConn(dir)
	}
	for _, e := range s.edges {
		e.RemoveConnForChild(dir)
	}
}

// RemoveConnForChild explicitly releases connection in child the scope.
func (s *resourceScope) RemoveConnForChild(dir Direction) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return
	}
	s.rc.removeConn(dir)
}

// ReserveForChild explicitly releases connection in child the scope.
func (s *resourceScope) ReserveForChild(st ScopeStat) error {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return s.wrapError(ErrResourceScopeClosed)
	}
	if err := s.rc.reserveMemory(st.Memory, ReservationPriorityAlways); err != nil {
		return s.wrapError(err)
	}
	if err := s.rc.addConns(st.NumConnsInbound, st.NumConnsOutbound, st.NumFD); err != nil {
		s.rc.releaseMemory(st.Memory)
		return s.wrapError(err)
	}
	return nil
}

// ReleaseResources explicitly releases resource by ScopeStat.
func (s *resourceScope) ReleaseResources(st ScopeStat) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return
	}
	s.rc.releaseMemory(st.Memory)
	s.rc.removeConns(st.NumConnsInbound, st.NumConnsOutbound, st.NumFD)
	if s.owner != nil {
		s.owner.ReleaseResources(st)
	} else {
		for _, e := range s.edges {
			e.ReleaseForChild(st)
		}
	}
}

// ReleaseForChild explicitly releases resource in the child scope by ScopeStat.
func (s *resourceScope) ReleaseForChild(st ScopeStat) {
	s.Lock()
	defer s.Unlock()
	if s.done {
		return
	}
	s.rc.releaseMemory(st.Memory)
	s.rc.removeConns(st.NumConnsInbound, st.NumConnsOutbound, st.NumFD)
}

// Name returns the name of scope.
func (s *resourceScope) Name() string {
	return s.name
}

// Stat returns the state of scope.
func (s *resourceScope) Stat() ScopeStat {
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

package rcmgr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScopeStat_String(t *testing.T) {
	s := ScopeStat{}
	result := s.String()
	assert.NotNil(t, result)
}

func TestNullResourceManager(t *testing.T) {
	n := &NullResourceManager{}
	f := func(ResourceScope) error { return nil }
	_ = n.ViewSystem(f)
	_ = n.ViewTransient(f)
	_ = n.ViewService("", f)
	f1 := func(ResourceScope) string { return "" }
	n.SystemLimitString(f1)
	n.TransientLimitString(f1)
	n.ServiceLimitString("", f1)
	n.SystemState()
	n.TransientState()
	n.ServiceState("")
	_, _ = n.OpenService("")
	_ = n.Close()
}

func TestNullScope(t *testing.T) {
	n := &NullScope{}
	_, _ = n.RemainingResource()
	_ = n.ReserveResources(nil)
	_ = n.ReserveMemory(0, 0)
	n.ReleaseMemory(0)
	_ = n.AddTask(0, 0)
	n.RemoveTask(0, 0)
	_ = n.AddConn(0)
	n.RemoveConn(0)
	n.Stat()
	_, _ = n.BeginSpan()
	n.Done()
	n.Name()
	n.Release()
}

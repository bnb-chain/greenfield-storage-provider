package rcmgr

import "testing"

func TestNullLimit(t *testing.T) {
	n := &NullLimit{}
	n.GetSystemLimits()
	n.GetTransientLimits()
	n.GetServiceLimits("")
	_ = n.String()
	n.GetMemoryLimit()
	n.GetFDLimit()
	n.GetConnLimit(0)
	n.GetConnTotalLimit()
	n.GetTaskLimit(0)
	n.GetTaskTotalLimit()
	n.NotLess(n)
	n.Add(n)
	n.Sub(n)
	n.Equal(n)
	n.ScopeStat()
}

func TestUnlimited(t *testing.T) {
	u := &Unlimited{}
	u.GetSystemLimits()
	u.GetTransientLimits()
	u.GetServiceLimits("")
	_ = u.String()
	u.GetMemoryLimit()
	u.GetFDLimit()
	u.GetConnLimit(0)
	u.GetConnTotalLimit()
	u.GetTaskLimit(0)
	u.GetTaskTotalLimit()
	u.NotLess(u)
	u.Add(u)
	u.Sub(u)
	u.Equal(u)
	u.ScopeStat()
}

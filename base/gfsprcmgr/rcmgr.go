package gfsprcmgr

import (
	"sync"

	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
)

// resourceManager manager resource scopes, include the top level system scope
// and sub service scope
type resourceManager struct {
	limits    corercmgr.Limiter
	system    *resourceScope
	transient *resourceScope

	svc map[string]*resourceScope
	mux sync.Mutex
}

var _ corercmgr.ResourceManager = &resourceManager{}

// OpenService creates a new service resource scope associated with system resource scope
func (r *resourceManager) OpenService(name string) (corercmgr.ResourceScope, error) {
	r.mux.Lock()
	defer r.mux.Unlock()
	limit := r.limits.GetServiceLimits(name)
	var scope *resourceScope
	if limit == nil {
		scope = newResourceScopeSpan(r.system, r.system.NextSpanID(), name)
	} else {
		scope = newResourceScope(limit, []*resourceScope{r.system}, name)
	}
	r.svc[name] = scope
	return scope, nil
}

// Close closes the resource manager
func (r *resourceManager) Close() error {
	return nil
}

// ViewSystem views the system resource scope.
// The system scope is the top level scope that accounts for global
// resource usage at all levels of the system. This scope constrains all
// other scopes and institutes global hard limits.
func (r *resourceManager) ViewSystem(f func(corercmgr.ResourceScope) error) error {
	return f(r.system)
}

// ViewTransient views the transient (DMZ) resource scope.
func (r *resourceManager) ViewTransient(f func(corercmgr.ResourceScope) error) error {
	return f(r.system)
}

// ViewService retrieves a service-specific scope.
func (r *resourceManager) ViewService(name string, f func(corercmgr.ResourceScope) error) error {
	scop, ok := r.svc[name]
	if !ok {
		return nil
	}
	return f(scop)
}

// SystemState output the system resource scope and limit readable
func (r *resourceManager) SystemState() string {
	state := r.system.Stat().String()
	limit := r.limits.GetSystemLimits().String()
	return "use: " + state + "limit: " + limit
}

// TransientState output the transient (DMZ)  resource scope and limit readable
func (r *resourceManager) TransientState() string {
	state := r.transient.Stat().String()
	limit := r.limits.GetTransientLimits().String()
	return "use: " + state + "limit: " + limit
}

// ServiceState output a service-specific resource scope and limit readable
func (r *resourceManager) ServiceState(name string) string {
	scop, ok := r.svc[name]
	if !ok {
		return ""
	}
	state := scop.Stat().String()
	limit := r.limits.GetServiceLimits(state)
	var limitState string
	if limit == nil {
		limitState = r.limits.GetSystemLimits().String()
	} else {
		limitState = limit.String()
	}
	return "use: " + state + "limit: " + limitState
}

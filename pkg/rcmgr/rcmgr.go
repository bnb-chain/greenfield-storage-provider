package rcmgr

import (
	"sync"
)

// resourceManager manager resource scopes, include the top level system scope
// and sub service scope
type resourceManager struct {
	limits    Limiter
	system    *resourceScope
	transient *resourceScope

	svc map[string]*resourceScope
	mux sync.Mutex
}

var _ ResourceManager = &resourceManager{}

var (
	resrcMgr ResourceManager
	once     sync.Once
)

func init() {
	resrcMgr = &NullResourceManager{}
}

// NewResourceManager inits and returns singleton instance of ResourceManager
func NewResourceManager(limits Limiter) (ResourceManager, error) {
	return initResourceManager(limits)
}

// ResrcManager return the global singleton instance of ResourceManager
func ResrcManager() ResourceManager {
	manager, _ := initResourceManager(nil)
	return manager
}

func initResourceManager(limits Limiter) (ResourceManager, error) {
	var err error
	once.Do(func() {
		if limits == nil {
			resrcMgr = &NullResourceManager{}
			return
		}
		r := &resourceManager{
			limits: limits,
			svc:    make(map[string]*resourceScope),
		}
		r.system = newResourceScope(limits.GetSystemLimits(), nil, "system")
		// TODO:: support transient resource scope
		r.transient = r.system
		resrcMgr = r
	})
	return resrcMgr, err
}

// OpenService creates a new service resource scope associated with system resource scope
func (r *resourceManager) OpenService(name string) (ResourceScope, error) {
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
func (r *resourceManager) ViewSystem(f func(ResourceScope) error) error {
	return f(r.system)
}

// ViewTransient views the transient (DMZ) resource scope.
func (r *resourceManager) ViewTransient(f func(ResourceScope) error) error {
	return f(r.system)
}

// ViewService retrieves a service-specific scope.
func (r *resourceManager) ViewService(name string, f func(ResourceScope) error) error {
	scope, ok := r.svc[name]
	if !ok {
		return nil
	}
	return f(scope)
}

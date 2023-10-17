package probe

import (
	"io"
	"net/http"
	"sync/atomic"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

type check func() bool

// HTTPProbe represents health and readiness status of given component, and provides HTTP integration.
type HTTPProbe struct {
	ready   atomic.Uint32
	healthy atomic.Uint32
}

// NewHTTPProbe returns HTTPProbe representing readiness and liveness of given component.
func NewHTTPProbe() *HTTPProbe {
	return &HTTPProbe{}
}

// HealthyHandler returns an HTTP Handler which responds health checks.
func (p *HTTPProbe) HealthyHandler() http.HandlerFunc {
	return p.handler(p.isHealthy)
}

// ReadyHandler returns an HTTP handler which responds readiness checks.
func (p *HTTPProbe) ReadyHandler() http.HandlerFunc {
	return p.handler(p.IsReady)
}

func (p *HTTPProbe) handler(c check) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if !c() {
			http.Error(w, "NOT OK", http.StatusServiceUnavailable)
			return
		}
		if _, err := io.WriteString(w, "OK"); err != nil {
			log.Errorw("failed to write probe response", "error", err)
		}
	}
}

// IsReady returns true if component is ready.
func (p *HTTPProbe) IsReady() bool {
	value := p.ready.Load()
	return value > 0
}

// isHealthy returns true if component is healthy.
func (p *HTTPProbe) isHealthy() bool {
	value := p.healthy.Load()
	return value > 0
}

// Ready sets components status to ready.
func (p *HTTPProbe) Ready() {
	p.ready.Swap(1)
}

// Unready sets components status to unready with given error as a cause.
func (p *HTTPProbe) Unready(err error) {
	p.ready.Swap(0)
}

// Healthy sets components status to healthy.
func (p *HTTPProbe) Healthy() {
	p.healthy.Swap(1)
}

// Unhealthy sets components status to unhealthy with given error as a cause.
func (p *HTTPProbe) Unhealthy(err error) {
	p.healthy.Swap(0)
}

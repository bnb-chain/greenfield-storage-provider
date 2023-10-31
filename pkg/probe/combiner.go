package probe

import (
	"sync"

	coreprober "github.com/bnb-chain/greenfield-storage-provider/core/prober"
)

type combined struct {
	mu     sync.Mutex
	probes []coreprober.Prober
}

// Combine folds given probes into one, reflects their statuses in a thread-safe way.
func Combine(probes ...coreprober.Prober) coreprober.Prober {
	return &combined{probes: probes}
}

// Ready sets components status to ready.
func (p *combined) Ready() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, probe := range p.probes {
		probe.Ready()
	}
}

// Unready sets components status to unready with given error as a cause.
func (p *combined) Unready(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, probe := range p.probes {
		probe.Unready(err)
	}
}

// Healthy sets components status to healthy.
func (p *combined) Healthy() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, probe := range p.probes {
		probe.Healthy()
	}
}

// Unhealthy sets components status to unhealthy with given error as a cause.
func (p *combined) Unhealthy(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, probe := range p.probes {
		probe.Unhealthy(err)
	}
}

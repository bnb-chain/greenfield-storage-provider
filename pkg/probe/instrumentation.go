package probe

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

const (
	ready     = "ready"
	unready   = "unready"
	healthy   = "healthy"
	unhealthy = "unhealthy"
)

// InstrumentationProbe stores instrumentation state of Probe.
// This is created with an intention to combine with other Probe's using prober.Combine.
type InstrumentationProbe struct {
	statusMetric *prometheus.GaugeVec
	mu           sync.Mutex
	statusString string
}

// NewInstrumentation returns InstrumentationProbe records readiness and healthiness for given module.
func NewInstrumentation() *InstrumentationProbe {
	status := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "status_check",
		Help: "Represents status (0 indicates failure, 1 indicates success) of the module.",
	}, []string{"status_check"})
	metrics.AddMetrics(status)

	return &InstrumentationProbe{statusMetric: status}
}

// Ready records the module status when Ready is called, if combined with other Probes.
func (p *InstrumentationProbe) Ready() {
	p.statusMetric.WithLabelValues(ready).Set(1)
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.statusString != ready {
		log.Infof("change probe status to %s", ready)
		p.statusString = ready
	}
}

// Unready records the module status when Unready is called, if combined with other Probes.
func (p *InstrumentationProbe) Unready(err error) {
	p.statusMetric.WithLabelValues(ready).Set(0)
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.statusString != unready {
		log.Infof("change probe status to %s, reason: %v", unready, err)
		p.statusString = unready
	}
}

// Healthy records the module status when Healthy is called, if combined with other Probes.
func (p *InstrumentationProbe) Healthy() {
	p.statusMetric.WithLabelValues(healthy).Set(1)
	log.Infof("change probe status to %s", healthy)
}

// Unhealthy records the module status when UnHealthy is called, if combined with other Probes.
func (p *InstrumentationProbe) Unhealthy(err error) {
	p.statusMetric.WithLabelValues(healthy).Set(0)
	log.Infof("change probe status to %s, reason: %v", unhealthy, err)
}

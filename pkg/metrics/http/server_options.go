package http

type serverMetricsConfig struct {
	counterOpts   counterOptions
	gaugeOpts     gaugeOptions
	histogramOpts histogramOptions
	summaryOpts   summaryOptions
}

type ServerMetricsOption func(*serverMetricsConfig)

func (c *serverMetricsConfig) apply(opts []ServerMetricsOption) {
	for _, o := range opts {
		o(c)
	}
}

// WithServerCounterOptions adds options to counter
func WithServerCounterOptions(opts ...CounterOption) ServerMetricsOption {
	return func(o *serverMetricsConfig) {
		o.counterOpts = opts
	}
}

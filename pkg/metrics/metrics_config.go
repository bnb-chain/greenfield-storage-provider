package metrics

// MetricsMonitorConfig contains the configuration for the metric collection.
type MetricsMonitorConfig struct {
	Enabled     bool   `toml:",omitempty"`
	HTTPAddress string `toml:",omitempty"`
}

// DefaultMetricsMonitorConfig is the default config for metrics used in storage provider
var DefaultMetricsMonitorConfig = &MetricsMonitorConfig{
	Enabled:     true,
	HTTPAddress: "127.0.0.1:9300",
}

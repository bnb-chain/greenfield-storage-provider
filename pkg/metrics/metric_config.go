package metrics

// Config contains the configuration for the metric collection.
type MetricsConfig struct {
	Enabled     bool   `toml:",omitempty"`
	HTTPAddress string `toml:",omitempty"`
}

// DefaultMetricsConfig is the default config for metrics used in storage provider
var DefaultMetricsConfig = &MetricsConfig{
	Enabled:     false,
	HTTPAddress: "127.0.0.1:6060",
}

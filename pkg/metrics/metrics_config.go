package metrics

// MetricsMonitorConfig contains the configuration for the metric collection.
type MetricsMonitorConfig struct {
	Enabled     bool   `toml:",omitempty"`
	HTTPAddress string `toml:",omitempty"`
}

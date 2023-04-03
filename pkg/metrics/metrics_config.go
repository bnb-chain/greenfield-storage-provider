package metrics

// MetricsConfig contains the configuration for the metric collection.
type MetricsConfig struct {
	Enabled     bool   `toml:",omitempty"`
	HTTPAddress string `toml:",omitempty"`
}

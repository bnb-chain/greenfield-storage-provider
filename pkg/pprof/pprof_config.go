package pprof

// PprofConfig contains the configuration for pprof.
type PprofConfig struct {
	Enabled     bool   `toml:",omitempty"`
	HTTPAddress string `toml:",omitempty"`
}

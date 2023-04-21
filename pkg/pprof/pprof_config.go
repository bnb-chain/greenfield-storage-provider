package pprof

// PProfConfig contains the configuration for pprof.
type PProfConfig struct {
	Enabled     bool   `toml:",omitempty"`
	HTTPAddress string `toml:",omitempty"`
}

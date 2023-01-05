package gateway

type GatewayConfig struct {
	Address   string
	Domain    string
	LogConfig struct {
		FilePath string
		Level    string
	}
	UploaderConfig   uploaderClientConfig
	ChainConfig      chainClientConfig
	DownloaderConfig downloaderClientConfig
}

var DefaultGatewayConfig = &GatewayConfig{
	Address: "127.0.0.1:5310",
	Domain:  "bfs.nodereal.com",
	LogConfig: struct {
		FilePath string
		Level    string
	}{FilePath: "./debug/gateway.log", Level: "info"},
	UploaderConfig: struct {
		Mode     string
		DebugDir string
		Address  string
	}{Mode: "DebugMode", DebugDir: "./debug", Address: ""},
	ChainConfig: struct {
		Mode     string
		DebugDir string
	}{Mode: "DebugMode", DebugDir: "./debug"},
	DownloaderConfig: struct {
		Mode     string
		DebugDir string
	}{Mode: "DebugMode", DebugDir: "./debug"},
}

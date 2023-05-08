package config

type GfSpConfig struct {
	appID          string
	grpcAddress    string
	httpAddress    string
	domain         string // external domain name is used for virtual-hosted style url
	operateAddress string
}

package auth

import (
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
)

// AuthConfig is the auth service config
type AuthConfig struct {
	GRPCAddress       string
	SpDBConfig        *config.SQLDBConfig
	SpOperatorAddress string
}

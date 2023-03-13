package blocksyncer

import (
	"fmt"
	"os"
	"time"

	tomlconfig "github.com/forbole/juno/v4/cmd/migrate/toml"
	databaseconfig "github.com/forbole/juno/v4/database/config"
	"github.com/forbole/juno/v4/log/config"
	"github.com/forbole/juno/v4/node/remote"
	parserconfig "github.com/forbole/juno/v4/parser/config"
)

var avgBlockTime = time.Second

const DSN_BLOCK_SYNCER = "BLOCK_SYNCER_DSN"

var DefaultBlockSyncerConfig = &tomlconfig.TomlConfig{
	Node: tomlconfig.NodeConfig{
		Type: "mysql",
		RPC: &remote.RPCConfig{
			ClientName:     "juno",
			Address:        "http://localhost:26750",
			MaxConnections: 20,
		},
		GRPC: &remote.GRPCConfig{
			Address:  "http://localhost:9090",
			Insecure: true,
		},
	},
	Parser: parserconfig.Config{
		AvgBlockTime: &avgBlockTime,
	},
	Database: databaseconfig.Config{
		Secrets: &databaseconfig.Params{
			SecretId: "1",
			Region:   "1",
		},
	},
	Logging: config.DefaultLogConfig(),
}

func getDBConfigFromEnv(dsn string) (string, error) {
	dsnVal, ok := os.LookupEnv(dsn)
	if !ok {
		return "", fmt.Errorf("dsn %s config is not set in environment", dsnVal)
	}
	return dsnVal, nil
}

package utils

import (
	"errors"
	"os"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/pelletier/go-toml/v2"
	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/base/gnfd"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

// MakeConfig loads the configuration from local file and replace the fields by flags.
func MakeConfig(ctx *cli.Context) (*gfspconfig.GfSpConfig, error) {
	cfg := &gfspconfig.GfSpConfig{}
	if ctx.IsSet(ConfigFileFlag.Name) {
		err := LoadConfig(ctx.String(ConfigFileFlag.Name), cfg)
		if err != nil {
			log.Errorw("failed to load config file", "error", err)
			return nil, err
		}
	}
	if ctx.IsSet(ServerFlag.Name) {
		cfg.Server = util.SplitByComma(ctx.String(ServerFlag.Name))
	}
	if ctx.IsSet(MetricsDisableFlag.Name) {
		cfg.Monitor.DisableMetrics = ctx.Bool(MetricsDisableFlag.Name)
	}
	if ctx.IsSet(utils.MetricsHTTPFlag.Name) {
		cfg.Monitor.MetricsHTTPAddress = ctx.String(utils.MetricsHTTPFlag.Name)
	}
	if ctx.IsSet(PProfDisableFlag.Name) {
		cfg.Monitor.DisablePProf = ctx.Bool(PProfDisableFlag.Name)
	}
	if ctx.IsSet(PProfHTTPFlag.Name) {
		cfg.Monitor.PProfHTTPAddress = ctx.String(PProfHTTPFlag.Name)
	}
	if ctx.IsSet(DisableResourceManagerFlag.Name) {
		cfg.Rcmgr.DisableRcmgr = ctx.Bool(DisableResourceManagerFlag.Name)
	}
	return cfg, nil
}

// LoadConfig loads the configuration from file.
func LoadConfig(file string, cfg *gfspconfig.GfSpConfig) error {
	if cfg == nil {
		return errors.New("failed to load config file, the config param invalid")
	}
	bz, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return toml.Unmarshal(bz, cfg)
}

// MakeEnv inits storage provider runtime environment.
func MakeEnv(ctx *cli.Context, cfg *gfspconfig.GfSpConfig) error {
	if err := initLog(ctx, cfg); err != nil {
		return err
	}
	return nil
}

// initLog inits the log configuration from config file and command flags.
func initLog(ctx *cli.Context, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Log.Level == "" {
		cfg.Log.Level = "debug"
	}
	if cfg.Log.Path == "" {
		cfg.Log.Path = "./gnfd-sp.log"
	}
	if ctx.IsSet(LogLevelFlag.Name) {
		cfg.Log.Level = ctx.String(LogLevelFlag.Name)
	}
	if ctx.IsSet(LogPathFlag.Name) {
		cfg.Log.Path = ctx.String(LogPathFlag.Name)
	}
	if ctx.IsSet(LogStdOutputFlag.Name) {
		cfg.Log.Path = ""
	}
	level, err := log.ParseLevel(cfg.Log.Level)
	if err != nil {
		return err
	}
	log.Init(level, cfg.Log.Path)
	return nil
}

func MakeGfSpClient(cfg *gfspconfig.GfSpConfig) *gfspclient.GfSpClient {
	if len(cfg.GRPCAddress) == 0 {
		cfg.GRPCAddress = gfspapp.DefaultGRPCAddress
	}
	if len(cfg.Endpoint.ApproverEndpoint) == 0 {
		cfg.Endpoint.ApproverEndpoint = cfg.GRPCAddress
	}
	if len(cfg.Endpoint.ManagerEndpoint) == 0 {
		cfg.Endpoint.ManagerEndpoint = cfg.GRPCAddress
	}
	if len(cfg.Endpoint.DownloaderEndpoint) == 0 {
		cfg.Endpoint.DownloaderEndpoint = cfg.GRPCAddress
	}
	if len(cfg.Endpoint.ReceiverEndpoint) == 0 {
		cfg.Endpoint.ReceiverEndpoint = cfg.GRPCAddress
	}
	if len(cfg.Endpoint.MetadataEndpoint) == 0 {
		cfg.Endpoint.MetadataEndpoint = cfg.GRPCAddress
	}
	if len(cfg.Endpoint.UploaderEndpoint) == 0 {
		cfg.Endpoint.UploaderEndpoint = cfg.GRPCAddress
	}
	if len(cfg.Endpoint.P2PEndpoint) == 0 {
		cfg.Endpoint.P2PEndpoint = cfg.GRPCAddress
	}
	if len(cfg.Endpoint.SignerEndpoint) == 0 {
		cfg.Endpoint.SignerEndpoint = cfg.GRPCAddress
	}
	if len(cfg.Endpoint.AuthenticatorEndpoint) == 0 {
		cfg.Endpoint.AuthenticatorEndpoint = cfg.GRPCAddress
	}
	client := gfspclient.NewGfSpClient(
		cfg.Endpoint.ApproverEndpoint,
		cfg.Endpoint.ManagerEndpoint,
		cfg.Endpoint.DownloaderEndpoint,
		cfg.Endpoint.ReceiverEndpoint,
		cfg.Endpoint.MetadataEndpoint,
		cfg.Endpoint.UploaderEndpoint,
		cfg.Endpoint.P2PEndpoint,
		cfg.Endpoint.SignerEndpoint,
		cfg.Endpoint.AuthenticatorEndpoint,
		false)
	return client
}

func MakeGnfd(cfg *gfspconfig.GfSpConfig) (*gnfd.Gnfd, error) {
	if len(cfg.Chain.ChainID) == 0 {
		cfg.Chain.ChainID = gfspapp.DefaultChainID
	}
	if len(cfg.Chain.ChainAddress) == 0 {
		cfg.Chain.ChainAddress = []string{gfspapp.DefaultChainAddress}
	}
	gnfdCfg := &gnfd.GnfdChainConfig{
		ChainID:      cfg.Chain.ChainID,
		ChainAddress: cfg.Chain.ChainAddress,
	}
	return gnfd.NewGnfd(gnfdCfg)
}

func MakeSPDB(cfg *gfspconfig.GfSpConfig) (spdb.SPDB, error) {
	if val, ok := os.LookupEnv(sqldb.SpDBUser); ok {
		cfg.SpDB.User = val
	}
	if val, ok := os.LookupEnv(sqldb.SpDBPasswd); ok {
		cfg.SpDB.Passwd = val
	}
	if val, ok := os.LookupEnv(sqldb.SpDBAddress); ok {
		cfg.SpDB.Address = val
	}
	if val, ok := os.LookupEnv(sqldb.SpDBDatabase); ok {
		cfg.SpDB.Database = val
	}
	if cfg.SpDB.User == "" {
		cfg.SpDB.User = "root"
	}
	if cfg.SpDB.Passwd == "" {
		cfg.SpDB.Passwd = "test"
	}
	if cfg.SpDB.Address == "" {
		cfg.SpDB.Address = "127.0.0.1:3306"
	}
	if cfg.SpDB.Database == "" {
		cfg.SpDB.Database = "storage_provider_db"
	}
	dbCfg := &cfg.SpDB
	db, err := sqldb.NewSpDB(dbCfg)
	if err != nil {
		return nil, err
	}
	return db, nil
}

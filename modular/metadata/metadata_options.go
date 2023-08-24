package metadata

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
)

const (
	// DefaultQuerySPParallelPerNode defines the max parallel for retrieving request
	DefaultQuerySPParallelPerNode int64 = 10240
	// DefaultBsDBSwitchCheckIntervalSec defines the default db switch check interval in seconds
	DefaultBsDBSwitchCheckIntervalSec = 30
)

var (
	BsModules []string
	BsWorkers uint

	ChainID      string
	ChainAddress []string

	SpOperatorAddress string
	GatewayDomainName string
)

func NewMetadataModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	metadata := &MetadataModular{baseApp: app}
	if err := DefaultMetadataOptions(metadata, cfg); err != nil {
		return nil, err
	}
	// register metadata service to gfsp base app's grpc server
	types.RegisterGfSpMetadataServiceServer(metadata.baseApp.ServerForRegister(), metadata)
	return metadata, nil
}

func DefaultMetadataOptions(metadata *MetadataModular, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Parallel.QuerySPParallelPerNode == 0 {
		cfg.Parallel.QuerySPParallelPerNode = DefaultQuerySPParallelPerNode
	}

	metadata.maxMetadataRequest = cfg.Parallel.QuerySPParallelPerNode

	metadata.baseApp.SetGfBsDB(metadata.baseApp.GfBsDBMaster())

	BsModules = cfg.BlockSyncer.Modules
	BsWorkers = cfg.BlockSyncer.Workers
	ChainID = cfg.Chain.ChainID
	ChainAddress = cfg.Chain.ChainAddress
	SpOperatorAddress = cfg.SpAccount.SpOperatorAddress
	GatewayDomainName = cfg.Gateway.DomainName

	return nil
}

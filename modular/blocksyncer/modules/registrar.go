package registrar

import (
	"github.com/forbole/juno/v4/modules"
	"github.com/forbole/juno/v4/modules/block"
	"github.com/forbole/juno/v4/modules/bucket"
	"github.com/forbole/juno/v4/modules/epoch"
	"github.com/forbole/juno/v4/modules/group"
	"github.com/forbole/juno/v4/modules/messages"
	"github.com/forbole/juno/v4/modules/payment"
	"github.com/forbole/juno/v4/modules/permission"
	"github.com/forbole/juno/v4/modules/registrar"
	sp "github.com/forbole/juno/v4/modules/storage_provider"

	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/database"
	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/modules/object"
	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/modules/prefixtree"
)

var (
	_ registrar.Registrar = &BlockSyncerRegistrar{}
)

// BlockSyncerRegistrar represents the modules.Registrar that allows to register all modules that are supported by blocksyncer
type BlockSyncerRegistrar struct {
	parser messages.MessageAddressesParser
}

// NewBlockSyncerRegistrar allows to build a new Registrar instance
func NewBlockSyncerRegistrar(parser messages.MessageAddressesParser) *BlockSyncerRegistrar {
	return &BlockSyncerRegistrar{
		parser: parser,
	}
}

// BuildModules implements modules.Registrar
func (r *BlockSyncerRegistrar) BuildModules(ctx registrar.Context) modules.Modules {
	db := database.Cast(ctx.Database)

	return modules.Modules{
		block.NewModule(db),
		bucket.NewModule(db),
		object.NewModule(db),
		epoch.NewModule(db),
		payment.NewModule(db),
		permission.NewModule(db),
		group.NewModule(db),
		sp.NewModule(db),
		prefixtree.NewModule(db),
	}
}

package registrar

import (
	"github.com/forbole/juno/v4/modules"
	"github.com/forbole/juno/v4/modules/epoch"
	"github.com/forbole/juno/v4/modules/messages"
	"github.com/forbole/juno/v4/modules/registrar"

	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/database"
	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/modules/bucket"
	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/modules/events"
	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/modules/general"
	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/modules/group"
	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/modules/object"
	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/modules/objectidmap"
	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/modules/payment"
	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/modules/permission"
	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/modules/prefixtree"
	sp "github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/modules/storage_provider"
	virtualgroup "github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/modules/virtual_group"
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
		epoch.NewModule(db),
		bucket.NewModule(db),
		object.NewModule(db),
		payment.NewModule(db),
		permission.NewModule(db),
		group.NewModule(db),
		sp.NewModule(db),
		prefixtree.NewModule(db),

		//vg related module
		virtualgroup.NewModule(db),
		//vg event module
		events.NewModule(db),
		objectidmap.NewModule(db),
		general.NewModule(db),
	}
}

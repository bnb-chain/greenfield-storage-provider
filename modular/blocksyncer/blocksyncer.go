package blocksyncer

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	tomlconfig "github.com/forbole/juno/v4/cmd/migrate/toml"
	"github.com/forbole/juno/v4/database"
	"github.com/forbole/juno/v4/parser"

	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
)

// BlockSyncer synchronizes storage,payment,permission data to db by handling related events
type BlockSyncerModular struct {
	config    *tomlconfig.TomlConfig
	name      string
	parserCtx *parser.Context
	running   atomic.Value
	context   context.Context
	scope     rcmgr.ResourceScope
}

// Read concurrency required global variables
var (
	blockMap *sync.Map
	eventMap *sync.Map
	txMap    *sync.Map

	MainService   *BlockSyncerModular
	BackupService *BlockSyncerModular

	FlagDB database.Database

	NeedBackup bool

	CancelMain func()
	CtxMain    context.Context
)

func (b *BlockSyncerModular) Name() string {
	return module.BlockSyncerModularName
}

func (b *BlockSyncerModular) Start(ctx context.Context) error {
	if b.running.Swap(true) == true {
		return errors.New("block syncer hub has already started")
	}

	determineMainService()

	CtxMain, CancelMain = context.WithCancel(context.Background())

	go MainService.serve(CtxMain)

	//create backup blocksyncer
	if NeedBackup {
		ctxBackup := context.Background()
		BackupService.context = ctxBackup

		go BackupService.serve(ctxBackup)
		go CheckProgress()
	}

	return nil
}

func (b *BlockSyncerModular) Stop(ctx context.Context) error {
	b.scope.Release()
	return nil
}

func (b *BlockSyncerModular) ReserveResource(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
	span, err := b.scope.BeginSpan()
	if err != nil {
		return nil, err
	}
	err = span.ReserveResources(state)
	if err != nil {
		return nil, err
	}
	return span, nil
}

func (b *BlockSyncerModular) ReleaseResource(ctx context.Context, span rcmgr.ResourceScopeSpan) {
	span.Done()
}

func determineMainService() error {
	masterFlag, err := FlagDB.GetMasterDB(context.TODO())
	if err != nil {
		return err
	}
	if masterFlag.IsMaster {
		return nil
	} else {
		//switch role
		temp := MainService
		MainService = BackupService
		BackupService = temp
	}
	return nil
}

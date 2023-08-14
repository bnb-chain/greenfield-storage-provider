package blocksyncer

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/forbole/juno/v4/parser"
	"github.com/forbole/juno/v4/types/config"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	db "github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/database"
)

var (
	BlockSyncerModularBackupName = strings.ToLower("BlockSyncerBackup")
	// DsnBlockSyncer defines env variable name for block syncer dsn
	DsnBlockSyncer = "BLOCK_SYNCER_DSN"
	// DsnBlockSyncerSwitched defines env variable name for block syncer backup dsn
	DsnBlockSyncerSwitched = "BLOCK_SYNCER_DSN_SWITCHED"
	ErrDSNNotSet           = errors.New("dsn config is not set in environment")
	ErrBlockNotFound       = errors.New("failed to get block from map need retry")
	ErrHandleEvent         = errors.New("failed to handle event")
)

const (
	// MaxRetryCount defines getting the latest height from the RPC client max retry count
	MaxRetryCount = 50
	// DefaultBlockHeightDiff defines default block height diff of main and backup service
	DefaultBlockHeightDiff = 100
	// DefaultCheckDiffPeriod defines check interval of block height diff
	DefaultCheckDiffPeriod = 1
	// MaxHeightGapFactor defines the gap coefficient between the block height in the Map and the processed block height
	MaxHeightGapFactor = 4
)

type MigrateDBKey struct{}

// BlockSyncerModular synchronizes storage,payment,permission data to db by handling related events
type BlockSyncerModular struct {
	config    *config.TomlConfig
	name      string
	parserCtx *parser.Context
	running   atomic.Value
	context   context.Context
	scope     rcmgr.ResourceScope
	baseApp   *gfspapp.GfSpBaseApp
}

// Read concurrency required global variables
var (
	blockMap *sync.Map
	eventMap *sync.Map
	txMap    *sync.Map

	RealTimeStart *atomic.Bool
	CatchEndBlock *atomic.Int64

	MainService   *BlockSyncerModular
	BackupService *BlockSyncerModular

	FlagDB *db.DB

	NeedBackup bool

	CancelMain func()
	CtxMain    context.Context
)

func (b *BlockSyncerModular) Name() string {
	return coremodule.BlockSyncerModularName
}

func (b *BlockSyncerModular) Start(ctx context.Context) error {
	if b.running.Swap(true) == true {
		return errors.New("block syncer hub has already started")
	}

	scope, err := b.baseApp.ResourceManager().OpenService(b.Name())
	if err != nil {
		return err
	}
	b.scope = scope

	determineMainService()

	CtxMain, CancelMain = context.WithCancel(context.Background())
	if !NeedBackup {
		CtxMain = context.WithValue(CtxMain, MigrateDBKey{}, true)
	}

	go MainService.serve(CtxMain)

	// create backup blocksyncer
	if NeedBackup {
		ctxBackup := context.WithValue(context.Background(), MigrateDBKey{}, true)
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
	if !masterFlag.IsMaster {
		// switch role
		temp := MainService
		MainService = BackupService
		BackupService = temp
	}

	return nil
}

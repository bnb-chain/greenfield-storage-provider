package gfspapp

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsplimit"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
)

func TestDefaultStaticOption(t *testing.T) {
	g := setup(t)
	g.server = grpc.NewServer()
	cfg := &gfspconfig.GfSpConfig{}
	err := DefaultStaticOption(g, cfg)
	assert.Nil(t, err)
}

func TestDefaultGfSpClientOption(t *testing.T) {
	g := setup(t)
	cfg := &gfspconfig.GfSpConfig{}
	err := DefaultGfSpClientOption(g, cfg)
	assert.Nil(t, err)
}

func TestDefaultGfSpDBOptionSuccess1(t *testing.T) {
	t.Log("Success case description: customize spdb is not nil")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	cfg := &gfspconfig.GfSpConfig{Customize: &gfspconfig.Customize{GfSpDB: m}}
	err := DefaultGfSpDBOption(g, cfg)
	assert.Nil(t, err)
}

func TestDefaultGfSpDBOptionFailure1(t *testing.T) {
	t.Log("Failure case description: set os env")
	g := setup(t)
	cfg := &gfspconfig.GfSpConfig{
		Server:    []string{module.BlockSyncerModularName, module.ApprovalModularName},
		Customize: &gfspconfig.Customize{GfSpDB: nil},
	}
	_ = os.Setenv(sqldb.SpDBUser, "mockUser")
	_ = os.Setenv(sqldb.SpDBPasswd, "mockPassword")
	_ = os.Setenv(sqldb.SpDBAddress, "mockAddress")
	_ = os.Setenv(sqldb.SpDBDatabase, "mockDatabase")
	defer os.Unsetenv(sqldb.SpDBUser)
	defer os.Unsetenv(sqldb.SpDBPasswd)
	defer os.Unsetenv(sqldb.SpDBAddress)
	defer os.Unsetenv(sqldb.SpDBDatabase)
	assert.Panics(t, func() {
		_ = DefaultGfSpDBOption(g, cfg)
	})
}

func Test_defaultGfSpDB(t *testing.T) {
	cfg := &config.SQLDBConfig{
		User:            "",
		Passwd:          "",
		Address:         "",
		Database:        "",
		ConnMaxLifetime: 0,
		ConnMaxIdleTime: 0,
		MaxIdleConns:    0,
		MaxOpenConns:    0,
	}
	defaultGfSpDB(cfg)
	assert.Equal(t, "storage_provider_db", cfg.Database)
}

func TestDefaultGfBsDBOptionFailure(t *testing.T) {
	t.Log("Failure case description: customize bsdb is nil, cannot link to db")
	g := setup(t)
	cfg := &gfspconfig.GfSpConfig{
		Server: []string{module.ApprovalModularName, module.MetadataModularName},
	}
	_ = os.Setenv(bsdb.BsDBUser, mockUser)
	_ = os.Setenv(bsdb.BsDBPasswd, mockPassword)
	_ = os.Setenv(bsdb.BsDBAddress, mockAddress)
	_ = os.Setenv(bsdb.BsDBDatabase, mockDatabase)
	_ = os.Setenv(bsdb.BsDBSwitchedUser, mockUser)
	_ = os.Setenv(bsdb.BsDBSwitchedPasswd, mockUser)
	_ = os.Setenv(bsdb.BsDBSwitchedAddress, mockUser)
	_ = os.Setenv(bsdb.BsDBSwitchedDatabase, mockUser)
	defer os.Unsetenv(bsdb.BsDBUser)
	defer os.Unsetenv(bsdb.BsDBPasswd)
	defer os.Unsetenv(bsdb.BsDBAddress)
	defer os.Unsetenv(bsdb.BsDBDatabase)
	defer os.Unsetenv(bsdb.BsDBSwitchedUser)
	defer os.Unsetenv(bsdb.BsDBSwitchedPasswd)
	defer os.Unsetenv(bsdb.BsDBSwitchedAddress)
	defer os.Unsetenv(bsdb.BsDBSwitchedDatabase)
	assert.Panics(t, func() {
		_ = DefaultGfBsDBOption(g, cfg)
	})
}

func Test_defaultGfBsDB(t *testing.T) {
	cfg := &config.SQLDBConfig{
		User:            "",
		Passwd:          "",
		Address:         "",
		Database:        "",
		ConnMaxLifetime: 0,
		ConnMaxIdleTime: 0,
		MaxIdleConns:    0,
		MaxOpenConns:    0,
	}
	defaultGfBsDB(cfg)
	assert.Equal(t, "block_syncer_db", cfg.Database)
}

func TestDefaultGfSpPieceStoreOptionSuccess1(t *testing.T) {
	t.Log("Success case description: customize piece store is not nil")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := piecestore.NewMockPieceStore(ctrl)
	cfg := &gfspconfig.GfSpConfig{
		Customize: &gfspconfig.Customize{PieceStore: m},
	}
	err := DefaultGfSpPieceStoreOption(g, cfg)
	assert.Nil(t, err)
}

func TestDefaultGfSpPieceStoreOptionSuccess2(t *testing.T) {
	t.Log("Success case description: use default config to new piece store")
	g := setup(t)
	cfg := &gfspconfig.GfSpConfig{
		Server:    []string{module.ApprovalModularName, module.DownloadModularName},
		Customize: &gfspconfig.Customize{PieceStore: nil},
	}
	err := DefaultGfSpPieceStoreOption(g, cfg)
	assert.Nil(t, err)
}

// func TestDefaultGfSpPieceStoreOptionFailure(t *testing.T) {
// 	t.Log("Failure case description: invalid piece store config would panic")
// 	var testOnce sync.Once
// 	testOnce.Do(func() {
// 		g := setup(t)
// 		cfg := &gfspconfig.GfSpConfig{
// 			Server:     []string{module.DownloadModularName},
// 			Customize:  &gfspconfig.Customize{PieceStore: nil},
// 			PieceStore: storage.PieceStoreConfig{Shards: 257},
// 		}
// 		assert.Panics(t, func() {
// 			_ = DefaultGfSpPieceStoreOption(g, cfg)
// 		})
// 	})
// }

func TestDefaultGfSpPieceOpOption(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := piecestore.NewMockPieceOp(ctrl)
	cases := []struct {
		name string
		cfg  *gfspconfig.GfSpConfig
	}{
		{
			name: "1",
			cfg:  &gfspconfig.GfSpConfig{Customize: &gfspconfig.Customize{PieceOp: m}},
		},
		{
			name: "2",
			cfg:  &gfspconfig.GfSpConfig{Customize: &gfspconfig.Customize{PieceOp: nil}},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := DefaultGfSpPieceOpOption(g, tt.cfg)
			assert.Nil(t, err)
		})
	}
}

func TestDefaultGfSpTQueueOption(t *testing.T) {
	g := setup(t)
	cfg := &gfspconfig.GfSpConfig{Customize: &gfspconfig.Customize{
		NewStrategyTQueueFunc:          nil,
		NewStrategyTQueueWithLimitFunc: nil,
		NewVirtualGroupManagerFunc:     nil,
	}}
	err := DefaultGfSpTQueueOption(g, cfg)
	assert.Nil(t, err)
}

func TestDefaultGfSpResourceManagerOption(t *testing.T) {
	g := setup(t)
	cases := []struct {
		name string
		cfg  *gfspconfig.GfSpConfig
	}{
		{
			name: "1",
			cfg: &gfspconfig.GfSpConfig{
				Customize: &gfspconfig.Customize{
					RcLimiter: nil,
				},
				Rcmgr: gfspconfig.RcmgrConfig{
					DisableRcmgr: false,
					GfSpLimiter:  nil,
				},
			},
		},
		{
			name: "2",
			cfg: &gfspconfig.GfSpConfig{
				Customize: &gfspconfig.Customize{
					RcLimiter: nil,
				},
				Rcmgr: gfspconfig.RcmgrConfig{
					DisableRcmgr: true,
					GfSpLimiter: &gfsplimit.GfSpLimiter{
						System: &gfsplimit.GfSpLimit{
							Memory: 1,
						},
					},
				},
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := DefaultGfSpResourceManagerOption(g, tt.cfg)
			assert.Nil(t, err)
		})
	}
}

func TestDefaultGfSpConsensusOptionSuccess1(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	mockConsensus := consensus.NewMockConsensus(ctrl)
	cfg := &gfspconfig.GfSpConfig{
		Customize: &gfspconfig.Customize{Consensus: mockConsensus},
	}
	err := DefaultGfSpConsensusOption(g, cfg)
	assert.Nil(t, err)
}

func TestDefaultGfSpConsensusOptionSuccess2(t *testing.T) {
	g := setup(t)
	cfg := &gfspconfig.GfSpConfig{Customize: &gfspconfig.Customize{Consensus: nil}}
	err := DefaultGfSpConsensusOption(g, cfg)
	assert.Nil(t, err)
}

func TestDefaultGfSpModuleOptionSuccess(t *testing.T) {
	g := setup(t)
	mockRegisterModular(t)
	defer ClearRegisterModules()
	cfg := &gfspconfig.GfSpConfig{
		Server: []string{module.ApprovalModularName, module.AuthenticationModularName, module.DownloadModularName,
			module.GateModularName, module.ManageModularName, module.P2PModularName, module.ReceiveModularName,
			module.SignModularName, module.UploadModularName, module.ExecuteModularName, module.GateModularName}}
	err := DefaultGfSpModuleOption(g, cfg)
	assert.Nil(t, err)
}

func TestDefaultGfSpModuleOptionFailure(t *testing.T) {
	g := setup(t)
	mock := mockApproverFailure{t: t}
	RegisterModular(module.ApprovalModularName, module.ApprovalModularDescription, mock.new)
	defer ClearRegisterModules()
	cfg := &gfspconfig.GfSpConfig{Server: []string{module.ApprovalModularName}}
	err := DefaultGfSpModuleOption(g, cfg)
	assert.Equal(t, errors.New("mock error"), err)
}

func TestDefaultGfSpMetricOption(t *testing.T) {
	g := setup(t)
	cfg := &gfspconfig.GfSpConfig{
		Monitor: gfspconfig.MonitorConfig{
			DisableMetrics: true,
		},
	}
	err := DefaultGfSpMetricOption(g, cfg)
	assert.Nil(t, err)
}

func TestDefaultGfSpPProfOption(t *testing.T) {
	g := setup(t)
	cfg := &gfspconfig.GfSpConfig{
		Monitor: gfspconfig.MonitorConfig{
			DisablePProf: true,
		},
	}
	err := DefaultGfSpPProfOption(g, cfg)
	assert.Nil(t, err)
}

func TestNewGfSpBaseAppFailure1(t *testing.T) {
	t.Log("Failure case description: init would panic")
	cfg := &gfspconfig.GfSpConfig{Customize: nil}
	assert.Panics(t, func() {
		NewGfSpBaseApp(cfg)
	})
}

func TestNewGfSpBaseAppFailure2(t *testing.T) {
	t.Log("Failure case description: repeated set piece store")
	ctrl := gomock.NewController(t)
	m := piecestore.NewMockPieceStore(ctrl)
	cfg := &gfspconfig.GfSpConfig{
		Customize: &gfspconfig.Customize{
			PieceStore: m,
		},
	}
	result, err := NewGfSpBaseApp(cfg, gfspconfig.CustomizePieceStore(m))
	assert.Equal(t, errors.New("repeated set piece store"), err)
	assert.Nil(t, result)
}

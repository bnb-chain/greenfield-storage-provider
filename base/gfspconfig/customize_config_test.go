package gfspconfig

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretaskqueue "github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
)

func TestCustomizeGfSpDBSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corespdb.NewMockSPDB(ctrl)
	opt := CustomizeGfSpDB(m)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{})
	assert.Nil(t, err)
}

func TestCustomizeGfSpDBFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corespdb.NewMockSPDB(ctrl)
	opt := CustomizeGfSpDB(m)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{Customize: &Customize{GfSpDB: m}})
	assert.Equal(t, errors.New("repeated set sp db"), err)
}

func TestCustomizePieceStoreSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := piecestore.NewMockPieceStore(ctrl)
	opt := CustomizePieceStore(m)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{})
	assert.Nil(t, err)
}

func TestCustomizePieceStoreFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := piecestore.NewMockPieceStore(ctrl)
	opt := CustomizePieceStore(m)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{Customize: &Customize{PieceStore: m}})
	assert.Equal(t, errors.New("repeated set piece store"), err)
}

func TestCustomizePieceOpSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := piecestore.NewMockPieceOp(ctrl)
	opt := CustomizePieceOp(m)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{})
	assert.Nil(t, err)
}

func TestCustomizePieceOpFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := piecestore.NewMockPieceOp(ctrl)
	opt := CustomizePieceOp(m)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{Customize: &Customize{PieceOp: m}})
	assert.Equal(t, errors.New("repeated set piece op"), err)
}

func TestCustomizeRcmgrSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceManager(ctrl)
	opt := CustomizeRcmgr(m)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{})
	assert.Nil(t, err)
}

func TestCustomizeRcmgrFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceManager(ctrl)
	opt := CustomizeRcmgr(m)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{Customize: &Customize{Rcmgr: m}})
	assert.Equal(t, errors.New("repeated set rcmgr"), err)
}

func TestCustomizeRcLimiterSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimiter(ctrl)
	opt := CustomizeRcLimiter(m)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{})
	assert.Nil(t, err)
}

func TestCustomizeRcLimiterFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimiter(ctrl)
	opt := CustomizeRcLimiter(m)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{Customize: &Customize{RcLimiter: m}})
	assert.Equal(t, errors.New("repeated set rc limiter"), err)
}

func TestCustomizeConsensusSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := consensus.NewMockConsensus(ctrl)
	opt := CustomizeConsensus(m)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{})
	assert.Nil(t, err)
}

func TestCustomizeConsensusFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := consensus.NewMockConsensus(ctrl)
	opt := CustomizeConsensus(m)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{Customize: &Customize{Consensus: m}})
	assert.Equal(t, errors.New("repeated set consensus"), err)
}

func TestCustomizeTQueueSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := coretaskqueue.NewMockTQueue(ctrl)
	fn := func(name string, cap int) coretaskqueue.TQueue { return m }
	opt := CustomizeTQueue(fn)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{})
	assert.Nil(t, err)
}

func TestCustomizeTQueueFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := coretaskqueue.NewMockTQueue(ctrl)
	fn := func(name string, cap int) coretaskqueue.TQueue { return m }
	opt := CustomizeTQueue(fn)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{Customize: &Customize{NewTQueueFunc: fn}})
	assert.Equal(t, errors.New("repeated set task queue"), err)
}

func TestCustomizeTQueueWithLimitSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := coretaskqueue.NewMockTQueueWithLimit(ctrl)
	fn := func(name string, cap int) coretaskqueue.TQueueWithLimit { return m }
	opt := CustomizeTQueueWithLimit(fn)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{})
	assert.Nil(t, err)
}

func TestCustomizeTQueueWithLimitFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := coretaskqueue.NewMockTQueueWithLimit(ctrl)
	fn := func(name string, cap int) coretaskqueue.TQueueWithLimit { return m }
	opt := CustomizeTQueueWithLimit(fn)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{Customize: &Customize{NewTQueueWithLimit: fn}})
	assert.Equal(t, errors.New("repeated set strategy task queue with limit"), err)
}

func TestCustomizeStrategyTQueueSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := coretaskqueue.NewMockTQueueOnStrategy(ctrl)
	fn := func(name string, cap int) coretaskqueue.TQueueOnStrategy { return m }
	opt := CustomizeStrategyTQueue(fn)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{})
	assert.Nil(t, err)
}

func TestCustomizeStrategyTQueueFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := coretaskqueue.NewMockTQueueOnStrategy(ctrl)
	fn := func(name string, cap int) coretaskqueue.TQueueOnStrategy { return m }
	opt := CustomizeStrategyTQueue(fn)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{Customize: &Customize{NewStrategyTQueueFunc: fn}})
	assert.Equal(t, errors.New("repeated set strategy task queue"), err)
}

func TestCustomizeStrategyTQueueWithLimitSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := coretaskqueue.NewMockTQueueOnStrategyWithLimit(ctrl)
	fn := func(name string, cap int) coretaskqueue.TQueueOnStrategyWithLimit { return m }
	opt := CustomizeStrategyTQueueWithLimit(fn)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{})
	assert.Nil(t, err)
}

func TestCustomizeStrategyTQueueWithLimitFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := coretaskqueue.NewMockTQueueOnStrategyWithLimit(ctrl)
	fn := func(name string, cap int) coretaskqueue.TQueueOnStrategyWithLimit { return m }
	opt := CustomizeStrategyTQueueWithLimit(fn)
	assert.NotNil(t, opt)
	err := opt(&GfSpConfig{Customize: &Customize{NewStrategyTQueueWithLimitFunc: fn}})
	assert.Equal(t, errors.New("repeated set strategy task queue with limit"), err)
}

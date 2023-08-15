package gfspapp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

func setup(t *testing.T) *GfSpBaseApp {
	client := gfspclient.NewGfSpClient(DefaultGRPCAddress, DefaultGRPCAddress, DefaultGRPCAddress, DefaultGRPCAddress, DefaultGRPCAddress,
		DefaultGRPCAddress, DefaultGRPCAddress, DefaultGRPCAddress, DefaultGRPCAddress, true)
	ctrl := gomock.NewController(t)
	mockSpDB := spdb.NewMockSPDB(ctrl)
	mockBsDB := bsdb.NewMockBSDB(ctrl)
	mockPieceStore := piecestore.NewMockPieceStore(ctrl)
	mockPieceOp := piecestore.NewMockPieceOp(ctrl)
	mockConsensus := consensus.NewMockConsensus(ctrl)
	mockRcmgr := rcmgr.NewMockResourceManager(ctrl)
	return &GfSpBaseApp{
		appID:           "mockAppID",
		grpcAddress:     "mockGRPCAddress",
		operatorAddress: "mockOperatorAddress",
		chainID:         "mockChainID",
		server:          nil,
		client:          client,
		gfSpDB:          mockSpDB,
		gfBsDB:          mockBsDB,
		gfBsDBMaster:    mockBsDB,
		gfBsDBBackup:    mockBsDB,
		pieceStore:      mockPieceStore,
		pieceOp:         mockPieceOp,
		rcmgr:           mockRcmgr,
		chain:           mockConsensus,
		approver:        nil,
		authenticator:   nil,
		downloader:      nil,
		executor:        nil,
		gater:           nil,
		manager:         nil,
		p2p:             nil,
		receiver:        nil,
		signer:          nil,
		uploader:        nil,
		metrics:         nil,
		pprof:           nil,
		appCtx:          nil,
		appCancel:       nil,
		services:        nil,
	}
}

func TestGfSpBaseApp_AppID(t *testing.T) {
	g := setup(t)
	result := g.AppID()
	assert.Equal(t, "mockAppID", result)
}

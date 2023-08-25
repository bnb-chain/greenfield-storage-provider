package gfspapp

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
)

func setup(t *testing.T) *GfSpBaseApp {
	return &GfSpBaseApp{
		appID:           "mockAppID",
		grpcAddress:     "mockGRPCAddress",
		operatorAddress: "mockOperatorAddress",
		chainID:         "mockChainID",
	}
}

func TestGfSpBaseApp_AppID(t *testing.T) {
	g := setup(t)
	result := g.AppID()
	assert.Equal(t, "mockAppID", result)
}

func TestGfSpBaseApp_GfSpClient(t *testing.T) {
	g := setup(t)
	result := g.GfSpClient()
	assert.Nil(t, result)
}

func TestGfSpBaseApp_PieceStore(t *testing.T) {
	g := setup(t)
	result := g.PieceStore()
	assert.Nil(t, result)
}

func TestGfSpBaseApp_PieceOp(t *testing.T) {
	g := setup(t)
	result := g.PieceOp()
	assert.Nil(t, result)
}

func TestGfSpBaseApp_Consensus(t *testing.T) {
	g := setup(t)
	result := g.Consensus()
	assert.Nil(t, result)
}

func TestGfSpBaseApp_OperatorAddress(t *testing.T) {
	g := setup(t)
	result := g.OperatorAddress()
	assert.Equal(t, "mockOperatorAddress", result)
}

func TestGfSpBaseApp_ChainID(t *testing.T) {
	g := setup(t)
	result := g.ChainID()
	assert.Equal(t, "mockChainID", result)
}

func TestGfSpBaseApp_GfSpDB(t *testing.T) {
	g := setup(t)
	result := g.GfSpDB()
	assert.Nil(t, result)
}

func TestGfSpBaseApp_GfBsDB(t *testing.T) {
	g := setup(t)
	result := g.GfBsDB()
	assert.Nil(t, result)
}

func TestGfSpBaseApp_GfBsDBMaster(t *testing.T) {
	g := setup(t)
	result := g.GfBsDBMaster()
	assert.Nil(t, result)
}

func TestGfSpBaseApp_SetGfBsDB(t *testing.T) {
	g := setup(t)
	result := g.SetGfBsDB(nil)
	assert.Nil(t, result)
}

func TestGfSpBaseApp_ServerForRegister(t *testing.T) {
	g := setup(t)
	result := g.ServerForRegister()
	assert.Nil(t, result)
}

func TestGfSpBaseApp_ResourceManager(t *testing.T) {
	g := setup(t)
	result := g.ResourceManager()
	assert.Nil(t, result)
}

func TestGfSpBaseApp_StartAndCloseSuccess(t *testing.T) {
	g := &GfSpBaseApp{grpcAddress: "localhost:0"}
	ctrl := gomock.NewController(t)
	client := gfspclient.NewMockGfSpClientAPI(ctrl)
	g.client = client
	client.EXPECT().Close().DoAndReturn(func() error { return nil }).AnyTimes()
	rc := rcmgr.NewMockResourceManager(ctrl)
	rc.EXPECT().Close().DoAndReturn(func() error { return nil }).AnyTimes()
	g.rcmgr = rc
	chain := consensus.NewMockConsensus(ctrl)
	chain.EXPECT().Close().DoAndReturn(func() error { return nil }).AnyTimes()
	g.chain = chain

	g.server = grpc.NewServer()
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	err := g.Start(ctx)
	cancel()
	fmt.Println(err)
}

func TestGfSpBaseApp_StartFailure(t *testing.T) {
	t.Log("Failure case description: missing port in address")
	g := setup(t)
	err := g.Start(context.TODO())
	assert.Equal(t, err.Error(), "listen tcp: address mockGRPCAddress: missing port in address")
}

func TestGfSpBaseApp_EnableMetrics(t *testing.T) {
	g := setup(t)
	result := g.EnableMetrics()
	assert.Equal(t, false, result)
}

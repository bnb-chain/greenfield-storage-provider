package gater

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

func TestGateModular_Name(t *testing.T) {
	g := setup(t)
	result := g.Name()
	assert.Equal(t, coremodule.GateModularName, result)
}

// func TestGateModular_StartSuccess(t *testing.T) {
// 	g := setup(t)
// 	g.httpAddress = "localhost:0"
// 	g.httpServer.Handler = new(http.ServeMux)
// 	ctrl := gomock.NewController(t)
// 	m := corercmgr.NewMockResourceManager(ctrl)
// 	g.baseApp.SetResourceManager(m)
// 	m1 := corercmgr.NewMockResourceScope(ctrl)
// 	m.EXPECT().OpenService(gomock.Any()).Return(m1, nil).AnyTimes()
// 	err := g.Start(context.TODO())
// 	assert.Nil(t, err)
// }

func TestGateModular_StartFailure(t *testing.T) {
	t.Log("Failure case description: mock OpenService returns error")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceManager(ctrl)
	g.baseApp.SetResourceManager(m)
	m.EXPECT().OpenService(gomock.Any()).Return(nil, mockErr).AnyTimes()
	err := g.Start(context.TODO())
	assert.Equal(t, mockErr, err)
}

// func TestGateModular_serverFailure(t *testing.T) {
// 	t.Log("Failure case description: invalid port")
// 	g := setup(t)
// 	g.httpAddress = "localhost:-1"
// 	g.server(context.TODO())
// }

func TestGateModular_Stop(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceScope(ctrl)
	g.scope = m
	m.EXPECT().Release().AnyTimes()

	g.httpServer = &http.Server{
		Addr: "localhost:0",
	}
	go func() {
		if err := g.httpServer.ListenAndServe(); err != nil {
			log.Errorw("failed to listen", "error", err)
			return
		}
	}()

	err := g.Stop(context.TODO())
	assert.Nil(t, err)
}

func TestGateModular_ReserveResourceSuccess(t *testing.T) {
	g := &GateModular{
		env:     gfspapp.EnvLocal,
		domain:  testDomain,
		baseApp: &gfspapp.GfSpBaseApp{},
	}
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceScope(ctrl)
	g.scope = m
	m1 := corercmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().BeginSpan().DoAndReturn(func() (corercmgr.ResourceScopeSpan, error) {
		return m1, nil
	}).Times(1)
	m1.EXPECT().ReserveResources(gomock.Any()).DoAndReturn(func(st *corercmgr.ScopeStat) error { return nil }).AnyTimes()
	result, err := g.ReserveResource(context.TODO(), &corercmgr.ScopeStat{Memory: 1})
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestGateModular_ReserveResourceFailure1(t *testing.T) {
	t.Log("Failure case description: mock BeginSpan returns error")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceScope(ctrl)
	g.scope = m
	m.EXPECT().BeginSpan().DoAndReturn(func() (corercmgr.ResourceScopeSpan, error) {
		return nil, mockErr
	}).Times(1)
	result, err := g.ReserveResource(context.TODO(), &corercmgr.ScopeStat{Memory: 1})
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestGateModular_ReserveResourceFailure2(t *testing.T) {
	t.Log("Failure case description: mock ReserveResources returns error")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceScope(ctrl)
	g.scope = m
	m1 := corercmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().BeginSpan().DoAndReturn(func() (corercmgr.ResourceScopeSpan, error) {
		return m1, nil
	}).Times(1)
	m1.EXPECT().ReserveResources(gomock.Any()).DoAndReturn(func(st *corercmgr.ScopeStat) error { return mockErr }).AnyTimes()
	result, err := g.ReserveResource(context.TODO(), &corercmgr.ScopeStat{Memory: 1})
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestGateModular_ReleaseResource(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().Done().AnyTimes()
	g.ReleaseResource(context.TODO(), m)
}

func TestGateModular_getSPIDSuccess1(t *testing.T) {
	g := setup(t)
	g.spID = 1
	result, err := g.getSPID()
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), result)
}

func TestGateModular_getSPIDSuccess2(t *testing.T) {
	t.Log("Success case description: query sp id by chain")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := consensus.NewMockConsensus(ctrl)
	g.baseApp.SetConsensus(m)
	m.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 2}, nil).Times(1)
	result, err := g.getSPID()
	assert.Nil(t, err)
	assert.Equal(t, uint32(2), result)
}

func TestGateModular_getSPIDFailure(t *testing.T) {
	t.Log("Failure case description: mock query sp returns error")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := consensus.NewMockConsensus(ctrl)
	g.baseApp.SetConsensus(m)
	m.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	result, err := g.getSPID()
	assert.Equal(t, mockErr, err)
	assert.Equal(t, uint32(0), result)
}

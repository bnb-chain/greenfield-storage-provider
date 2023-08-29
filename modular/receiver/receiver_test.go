package receiver

import (
	"context"
	"errors"
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var mockErr = errors.New("mock error")

func setup(t *testing.T) *ReceiveModular {
	return &ReceiveModular{baseApp: &gfspapp.GfSpBaseApp{}}
}

func TestReceiveModular_Name(t *testing.T) {
	r := setup(t)
	result := r.Name()
	assert.Equal(t, module.ReceiveModularName, result)
}

func TestReceiveModular_StartSuccess(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	m1 := rcmgr.NewMockResourceManager(ctrl)
	r.baseApp.SetResourceManager(m1)
	m2 := rcmgr.NewMockResourceScope(ctrl)
	m1.EXPECT().OpenService(gomock.Any()).DoAndReturn(func(svc string) (rcmgr.ResourceScope, error) {
		return m2, nil
	})
	err := r.Start(context.TODO())
	assert.Nil(t, err)
}

func TestReceiveModular_StartFailure(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	m1 := rcmgr.NewMockResourceManager(ctrl)
	r.baseApp.SetResourceManager(m1)
	m1.EXPECT().OpenService(gomock.Any()).DoAndReturn(func(svc string) (rcmgr.ResourceScope, error) {
		return nil, mockErr
	})
	err := r.Start(context.TODO())
	assert.Equal(t, mockErr, err)
}

func TestReceiveModular_Stop(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	r.scope = m
	m.EXPECT().Release().AnyTimes()
	err := r.Stop(context.TODO())
	assert.Nil(t, err)
}

func TestReceiveModular_ReserveResourceSuccess(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	r.scope = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().BeginSpan().DoAndReturn(func() (rcmgr.ResourceScopeSpan, error) {
		return m1, nil
	})
	m1.EXPECT().ReserveResources(gomock.Any()).DoAndReturn(func(st *rcmgr.ScopeStat) error { return nil }).AnyTimes()
	result, err := r.ReserveResource(context.TODO(), &rcmgr.ScopeStat{Memory: 1})
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestReceiveModular_ReserveResourceFailure1(t *testing.T) {
	t.Log("Failure case description: mock BeginSpan returns error")
	r := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	r.scope = m
	m.EXPECT().BeginSpan().DoAndReturn(func() (rcmgr.ResourceScopeSpan, error) {
		return nil, mockErr
	})
	result, err := r.ReserveResource(context.TODO(), &rcmgr.ScopeStat{Memory: 1})
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestReceiveModular_ReserveResourceFailure2(t *testing.T) {
	t.Log("Failure case description: mock ReserveResources returns error")
	r := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	r.scope = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().BeginSpan().DoAndReturn(func() (rcmgr.ResourceScopeSpan, error) {
		return m1, nil
	})
	m1.EXPECT().ReserveResources(gomock.Any()).DoAndReturn(func(st *rcmgr.ScopeStat) error { return mockErr }).AnyTimes()
	result, err := r.ReserveResource(context.TODO(), &rcmgr.ScopeStat{Memory: 1})
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestReceiveModular_ReleaseResource(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().Done().AnyTimes()
	r.ReleaseResource(context.TODO(), m)
}

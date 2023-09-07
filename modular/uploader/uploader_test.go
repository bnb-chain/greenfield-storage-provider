package uploader

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
)

var mockErr = errors.New("mock error")

func setup(t *testing.T) *UploadModular {
	return &UploadModular{baseApp: &gfspapp.GfSpBaseApp{}}
}

func TestUploadModular_Name(t *testing.T) {
	u := setup(t)
	result := u.Name()
	assert.Equal(t, module.UploadModularName, result)
}

func TestUploadModular_StartSuccess(t *testing.T) {
	u := setup(t)
	ctrl := gomock.NewController(t)
	m1 := rcmgr.NewMockResourceManager(ctrl)
	u.baseApp.SetResourceManager(m1)
	m2 := rcmgr.NewMockResourceScope(ctrl)
	m1.EXPECT().OpenService(gomock.Any()).DoAndReturn(func(svc string) (rcmgr.ResourceScope, error) {
		return m2, nil
	})
	err := u.Start(context.TODO())
	assert.Nil(t, err)
}

func TestUploadModular_StartFailure(t *testing.T) {
	u := setup(t)
	ctrl := gomock.NewController(t)
	m1 := rcmgr.NewMockResourceManager(ctrl)
	u.baseApp.SetResourceManager(m1)
	m1.EXPECT().OpenService(gomock.Any()).DoAndReturn(func(svc string) (rcmgr.ResourceScope, error) {
		return nil, mockErr
	})
	err := u.Start(context.TODO())
	assert.Equal(t, mockErr, err)
}

func TestUploadModular_Stop(t *testing.T) {
	u := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	u.scope = m
	m.EXPECT().Release().AnyTimes()
	err := u.Stop(context.TODO())
	assert.Nil(t, err)
}

func TestUploadModular_ReserveResourceSuccess(t *testing.T) {
	u := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	u.scope = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().BeginSpan().DoAndReturn(func() (rcmgr.ResourceScopeSpan, error) {
		return m1, nil
	})
	m1.EXPECT().ReserveResources(gomock.Any()).DoAndReturn(func(st *rcmgr.ScopeStat) error { return nil }).AnyTimes()
	result, err := u.ReserveResource(context.TODO(), &rcmgr.ScopeStat{Memory: 1})
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestUploadModular_ReserveResourceFailure1(t *testing.T) {
	t.Log("Failure case description: mock BeginSpan returns error")
	u := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	u.scope = m
	m.EXPECT().BeginSpan().DoAndReturn(func() (rcmgr.ResourceScopeSpan, error) {
		return nil, mockErr
	})
	result, err := u.ReserveResource(context.TODO(), &rcmgr.ScopeStat{Memory: 1})
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestUploadModular_ReserveResourceFailure2(t *testing.T) {
	t.Log("Failure case description: mock ReserveResources returns error")
	u := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	u.scope = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().BeginSpan().DoAndReturn(func() (rcmgr.ResourceScopeSpan, error) {
		return m1, nil
	})
	m1.EXPECT().ReserveResources(gomock.Any()).DoAndReturn(func(st *rcmgr.ScopeStat) error { return mockErr }).AnyTimes()
	result, err := u.ReserveResource(context.TODO(), &rcmgr.ScopeStat{Memory: 1})
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestUploadModular_ReleaseResource(t *testing.T) {
	u := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().Done().AnyTimes()
	u.ReleaseResource(context.TODO(), m)
}

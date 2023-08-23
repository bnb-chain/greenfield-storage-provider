package downloader

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

func setup(t *testing.T) *DownloadModular {
	return &DownloadModular{baseApp: &gfspapp.GfSpBaseApp{}}
}

func TestDownloadModular_Name(t *testing.T) {
	d := setup(t)
	result := d.Name()
	assert.Equal(t, module.DownloadModularName, result)
}

func TestDownloadModular_StartSuccess(t *testing.T) {
	d := setup(t)
	ctrl := gomock.NewController(t)
	m1 := rcmgr.NewMockResourceManager(ctrl)
	d.baseApp.SetResourceManager(m1)
	m2 := rcmgr.NewMockResourceScope(ctrl)
	m1.EXPECT().OpenService(gomock.Any()).DoAndReturn(func(svc string) (rcmgr.ResourceScope, error) {
		return m2, nil
	})
	err := d.Start(context.TODO())
	assert.Nil(t, err)
}

func TestDownloadModular_StartFailure(t *testing.T) {
	d := setup(t)
	ctrl := gomock.NewController(t)
	m1 := rcmgr.NewMockResourceManager(ctrl)
	d.baseApp.SetResourceManager(m1)
	m1.EXPECT().OpenService(gomock.Any()).DoAndReturn(func(svc string) (rcmgr.ResourceScope, error) {
		return nil, mockErr
	})
	err := d.Start(context.TODO())
	assert.Equal(t, mockErr, err)
}

func TestDownloadModular_Stop(t *testing.T) {
	d := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	d.scope = m
	m.EXPECT().Release().AnyTimes()
	err := d.Stop(context.TODO())
	assert.Nil(t, err)
}

func TestDownloadModular_ReserveResourceSuccess(t *testing.T) {
	d := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	d.scope = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().BeginSpan().DoAndReturn(func() (rcmgr.ResourceScopeSpan, error) {
		return m1, nil
	})
	m1.EXPECT().ReserveResources(gomock.Any()).DoAndReturn(func(st *rcmgr.ScopeStat) error { return nil }).AnyTimes()
	result, err := d.ReserveResource(context.TODO(), &rcmgr.ScopeStat{Memory: 1})
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestDownloadModular_ReserveResourceFailure1(t *testing.T) {
	t.Log("Failure case description: mock BeginSpan returns error")
	d := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	d.scope = m
	m.EXPECT().BeginSpan().DoAndReturn(func() (rcmgr.ResourceScopeSpan, error) {
		return nil, mockErr
	})
	result, err := d.ReserveResource(context.TODO(), &rcmgr.ScopeStat{Memory: 1})
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestDownloadModular_ReserveResourceFailure2(t *testing.T) {
	t.Log("Failure case description: mock ReserveResources returns error")
	d := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	d.scope = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().BeginSpan().DoAndReturn(func() (rcmgr.ResourceScopeSpan, error) {
		return m1, nil
	})
	m1.EXPECT().ReserveResources(gomock.Any()).DoAndReturn(func(st *rcmgr.ScopeStat) error { return mockErr }).AnyTimes()
	result, err := d.ReserveResource(context.TODO(), &rcmgr.ScopeStat{Memory: 1})
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestDownloadModular_ReleaseResource(t *testing.T) {
	d := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().Done().AnyTimes()
	d.ReleaseResource(context.TODO(), m)
}

func TestDownloadModular_CacheKey(t *testing.T) {
	key := cacheKey("a", 1, 1)
	assert.Equal(t, "piece:a-offset:1-length:1", key)
}

package metadata

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

func setup(t *testing.T) *MetadataModular {
	return &MetadataModular{baseApp: &gfspapp.GfSpBaseApp{}}
}

func TestMetadataModular_Name(t *testing.T) {
	r := setup(t)
	result := r.Name()
	assert.Equal(t, module.MetadataModularName, result)
}

func TestMetadataModular_StartSuccess(t *testing.T) {
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

func TestMetadataModular_StartFailure(t *testing.T) {
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

func TestMetadataModular_Stop(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	r.scope = m
	m.EXPECT().Release().AnyTimes()
	err := r.Stop(context.TODO())
	assert.Nil(t, err)
}

func TestMetadataModular_ReserveResourceSuccess(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	r.scope = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().BeginSpan().DoAndReturn(func() (rcmgr.ResourceScopeSpan, error) {
		return m1, nil
	}).AnyTimes()
	m1.EXPECT().ReserveResources(gomock.Any()).DoAndReturn(func(st *rcmgr.ScopeStat) error { return nil }).AnyTimes()
	result, err := r.ReserveResource(context.TODO(), &rcmgr.ScopeStat{Memory: 1})
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestMetadataModular_ReleaseResource(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().Done().AnyTimes()
	r.ReleaseResource(context.TODO(), m)
}

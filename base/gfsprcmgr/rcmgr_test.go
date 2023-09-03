package gfsprcmgr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
)

func TestResourceManager_OpenService1(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimiter(ctrl)
	m1 := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().GetSystemLimits().Return(m1).AnyTimes()
	m.EXPECT().GetServiceLimits(gomock.Any()).Return(m1).AnyTimes()

	r := NewResourceManager(m)
	result, err := r.OpenService("mockSvc")
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestResourceManager_OpenService2(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimiter(ctrl)
	m1 := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().GetSystemLimits().Return(m1).AnyTimes()
	m.EXPECT().GetServiceLimits(gomock.Any()).Return(nil).AnyTimes()

	r := NewResourceManager(m)
	result, err := r.OpenService("mockSvc")
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestResourceManager_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimiter(ctrl)
	m1 := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().GetSystemLimits().Return(m1).AnyTimes()

	r := NewResourceManager(m)
	err := r.Close()
	assert.Nil(t, err)
}

func TestResourceManager_ViewSystem(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimiter(ctrl)
	m1 := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().GetSystemLimits().Return(m1).AnyTimes()

	f := func(rs corercmgr.ResourceScope) error { return nil }
	r := NewResourceManager(m)
	err := r.ViewSystem(f)
	assert.Nil(t, err)
}

func TestResourceManager_ViewTransient(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimiter(ctrl)
	m1 := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().GetSystemLimits().Return(m1).AnyTimes()

	f := func(rs corercmgr.ResourceScope) error { return nil }
	r := NewResourceManager(m)
	err := r.ViewTransient(f)
	assert.Nil(t, err)
}

func TestResourceManager_ViewService1(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimiter(ctrl)
	m1 := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().GetSystemLimits().Return(m1).AnyTimes()
	m.EXPECT().GetServiceLimits(gomock.Any()).Return(m1).AnyTimes()

	f := func(rs corercmgr.ResourceScope) error { return nil }
	r := NewResourceManager(m)
	result, err := r.OpenService("mockSvc")
	assert.Nil(t, err)
	assert.NotNil(t, result)
	err = r.ViewService("mockSvc", f)
	assert.Nil(t, err)
}

func TestResourceManager_ViewService2(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimiter(ctrl)
	m1 := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().GetSystemLimits().Return(m1).AnyTimes()

	f := func(rs corercmgr.ResourceScope) error { return nil }
	r := NewResourceManager(m)
	err := r.ViewService("mockSvc", f)
	assert.Nil(t, err)
}

func TestResourceManager_SystemState(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimiter(ctrl)
	m1 := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().GetSystemLimits().Return(m1).AnyTimes()
	m1.EXPECT().String().Return("test").AnyTimes()

	r := NewResourceManager(m)
	result := r.SystemState()
	assert.Equal(t, "use: memory reserved [0], task reserved[h: 0, m: 0, l: 0]limit: test", result)
}

func TestResourceManager_TransientState(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimiter(ctrl)
	m1 := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().GetSystemLimits().Return(m1).AnyTimes()
	m.EXPECT().GetTransientLimits().Return(m1).AnyTimes()
	m1.EXPECT().String().Return("test").AnyTimes()

	r := NewResourceManager(m)
	result := r.TransientState()
	assert.Equal(t, "use: memory reserved [0], task reserved[h: 0, m: 0, l: 0]limit: test", result)
}

func TestResourceManager_ServiceState1(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimiter(ctrl)
	m1 := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().GetSystemLimits().Return(m1).AnyTimes()

	r := NewResourceManager(m)
	result := r.ServiceState("")
	assert.Equal(t, "", result)
}

func TestResourceManager_ServiceState2(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimiter(ctrl)
	m1 := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().GetSystemLimits().Return(m1).AnyTimes()
	m.EXPECT().GetServiceLimits(gomock.Any()).Return(nil).AnyTimes()
	m1.EXPECT().String().Return("test").AnyTimes()

	r := NewResourceManager(m)
	result, err := r.OpenService("mockSvc")
	assert.Nil(t, err)
	assert.NotNil(t, result)
	result1 := r.ServiceState("mockSvc")
	assert.Equal(t, "use: memory reserved [0], task reserved[h: 0, m: 0, l: 0]limit: test", result1)
}

func TestResourceManager_ServiceState3(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockLimiter(ctrl)
	m1 := corercmgr.NewMockLimit(ctrl)
	m.EXPECT().GetSystemLimits().Return(m1).AnyTimes()
	m.EXPECT().GetServiceLimits(gomock.Any()).Return(m1).AnyTimes()
	m1.EXPECT().String().Return("test").AnyTimes()

	r := NewResourceManager(m)
	result, err := r.OpenService("mockSvc")
	assert.Nil(t, err)
	assert.NotNil(t, result)
	result1 := r.ServiceState("mockSvc")
	assert.Equal(t, "use: memory reserved [0], task reserved[h: 0, m: 0, l: 0]limit: test", result1)
}

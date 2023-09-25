// Code generated by MockGen. DO NOT EDIT.
// Source: ./virtual_group_manager.go
//
// Generated by this command:
//
//	mockgen -source=./virtual_group_manager.go -destination=./virtual_group_manager_mock.go -package=vgmgr
//
// Package vgmgr is a generated GoMock package.
package vgmgr

import (
	reflect "reflect"

	types "github.com/bnb-chain/greenfield/x/sp/types"
	types0 "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	gomock "go.uber.org/mock/gomock"
)

// MockSPPickFilter is a mock of SPPickFilter interface.
type MockSPPickFilter struct {
	ctrl     *gomock.Controller
	recorder *MockSPPickFilterMockRecorder
}

// MockSPPickFilterMockRecorder is the mock recorder for MockSPPickFilter.
type MockSPPickFilterMockRecorder struct {
	mock *MockSPPickFilter
}

// NewMockSPPickFilter creates a new mock instance.
func NewMockSPPickFilter(ctrl *gomock.Controller) *MockSPPickFilter {
	mock := &MockSPPickFilter{ctrl: ctrl}
	mock.recorder = &MockSPPickFilterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSPPickFilter) EXPECT() *MockSPPickFilterMockRecorder {
	return m.recorder
}

// Check mocks base method.
func (m *MockSPPickFilter) Check(spID uint32) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Check", spID)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Check indicates an expected call of Check.
func (mr *MockSPPickFilterMockRecorder) Check(spID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Check", reflect.TypeOf((*MockSPPickFilter)(nil).Check), spID)
}

// MockVGFPickFilter is a mock of VGFPickFilter interface.
type MockVGFPickFilter struct {
	ctrl     *gomock.Controller
	recorder *MockVGFPickFilterMockRecorder
}

// MockVGFPickFilterMockRecorder is the mock recorder for MockVGFPickFilter.
type MockVGFPickFilterMockRecorder struct {
	mock *MockVGFPickFilter
}

// NewMockVGFPickFilter creates a new mock instance.
func NewMockVGFPickFilter(ctrl *gomock.Controller) *MockVGFPickFilter {
	mock := &MockVGFPickFilter{ctrl: ctrl}
	mock.recorder = &MockVGFPickFilterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockVGFPickFilter) EXPECT() *MockVGFPickFilterMockRecorder {
	return m.recorder
}

// Check mocks base method.
func (m *MockVGFPickFilter) Check(vgfID uint32) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Check", vgfID)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Check indicates an expected call of Check.
func (mr *MockVGFPickFilterMockRecorder) Check(vgfID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Check", reflect.TypeOf((*MockVGFPickFilter)(nil).Check), vgfID)
}

// MockGVGPickFilter is a mock of GVGPickFilter interface.
type MockGVGPickFilter struct {
	ctrl     *gomock.Controller
	recorder *MockGVGPickFilterMockRecorder
}

// MockGVGPickFilterMockRecorder is the mock recorder for MockGVGPickFilter.
type MockGVGPickFilterMockRecorder struct {
	mock *MockGVGPickFilter
}

// NewMockGVGPickFilter creates a new mock instance.
func NewMockGVGPickFilter(ctrl *gomock.Controller) *MockGVGPickFilter {
	mock := &MockGVGPickFilter{ctrl: ctrl}
	mock.recorder = &MockGVGPickFilterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGVGPickFilter) EXPECT() *MockGVGPickFilterMockRecorder {
	return m.recorder
}

// CheckFamily mocks base method.
func (m *MockGVGPickFilter) CheckFamily(familyID uint32) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckFamily", familyID)
	ret0, _ := ret[0].(bool)
	return ret0
}

// CheckFamily indicates an expected call of CheckFamily.
func (mr *MockGVGPickFilterMockRecorder) CheckFamily(familyID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckFamily", reflect.TypeOf((*MockGVGPickFilter)(nil).CheckFamily), familyID)
}

// CheckGVG mocks base method.
func (m *MockGVGPickFilter) CheckGVG(gvgMeta *GlobalVirtualGroupMeta) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckGVG", gvgMeta)
	ret0, _ := ret[0].(bool)
	return ret0
}

// CheckGVG indicates an expected call of CheckGVG.
func (mr *MockGVGPickFilterMockRecorder) CheckGVG(gvgMeta any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckGVG", reflect.TypeOf((*MockGVGPickFilter)(nil).CheckGVG), gvgMeta)
}

// MockGenerateGVGSecondarySPsPolicy is a mock of GenerateGVGSecondarySPsPolicy interface.
type MockGenerateGVGSecondarySPsPolicy struct {
	ctrl     *gomock.Controller
	recorder *MockGenerateGVGSecondarySPsPolicyMockRecorder
}

// MockGenerateGVGSecondarySPsPolicyMockRecorder is the mock recorder for MockGenerateGVGSecondarySPsPolicy.
type MockGenerateGVGSecondarySPsPolicyMockRecorder struct {
	mock *MockGenerateGVGSecondarySPsPolicy
}

// NewMockGenerateGVGSecondarySPsPolicy creates a new mock instance.
func NewMockGenerateGVGSecondarySPsPolicy(ctrl *gomock.Controller) *MockGenerateGVGSecondarySPsPolicy {
	mock := &MockGenerateGVGSecondarySPsPolicy{ctrl: ctrl}
	mock.recorder = &MockGenerateGVGSecondarySPsPolicyMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGenerateGVGSecondarySPsPolicy) EXPECT() *MockGenerateGVGSecondarySPsPolicyMockRecorder {
	return m.recorder
}

// AddCandidateSP mocks base method.
func (m *MockGenerateGVGSecondarySPsPolicy) AddCandidateSP(spID uint32) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddCandidateSP", spID)
}

// AddCandidateSP indicates an expected call of AddCandidateSP.
func (mr *MockGenerateGVGSecondarySPsPolicyMockRecorder) AddCandidateSP(spID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddCandidateSP", reflect.TypeOf((*MockGenerateGVGSecondarySPsPolicy)(nil).AddCandidateSP), spID)
}

// GenerateGVGSecondarySPs mocks base method.
func (m *MockGenerateGVGSecondarySPsPolicy) GenerateGVGSecondarySPs() ([]uint32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateGVGSecondarySPs")
	ret0, _ := ret[0].([]uint32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateGVGSecondarySPs indicates an expected call of GenerateGVGSecondarySPs.
func (mr *MockGenerateGVGSecondarySPsPolicyMockRecorder) GenerateGVGSecondarySPs() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateGVGSecondarySPs", reflect.TypeOf((*MockGenerateGVGSecondarySPsPolicy)(nil).GenerateGVGSecondarySPs))
}

// MockExcludeFilter is a mock of ExcludeFilter interface.
type MockExcludeFilter struct {
	ctrl     *gomock.Controller
	recorder *MockExcludeFilterMockRecorder
}

// MockExcludeFilterMockRecorder is the mock recorder for MockExcludeFilter.
type MockExcludeFilterMockRecorder struct {
	mock *MockExcludeFilter
}

// NewMockExcludeFilter creates a new mock instance.
func NewMockExcludeFilter(ctrl *gomock.Controller) *MockExcludeFilter {
	mock := &MockExcludeFilter{ctrl: ctrl}
	mock.recorder = &MockExcludeFilterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExcludeFilter) EXPECT() *MockExcludeFilterMockRecorder {
	return m.recorder
}

// Apply mocks base method.
func (m *MockExcludeFilter) Apply(id uint32) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Apply", id)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Apply indicates an expected call of Apply.
func (mr *MockExcludeFilterMockRecorder) Apply(id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Apply", reflect.TypeOf((*MockExcludeFilter)(nil).Apply), id)
}

// MockVirtualGroupManager is a mock of VirtualGroupManager interface.
type MockVirtualGroupManager struct {
	ctrl     *gomock.Controller
	recorder *MockVirtualGroupManagerMockRecorder
}

// MockVirtualGroupManagerMockRecorder is the mock recorder for MockVirtualGroupManager.
type MockVirtualGroupManagerMockRecorder struct {
	mock *MockVirtualGroupManager
}

// NewMockVirtualGroupManager creates a new mock instance.
func NewMockVirtualGroupManager(ctrl *gomock.Controller) *MockVirtualGroupManager {
	mock := &MockVirtualGroupManager{ctrl: ctrl}
	mock.recorder = &MockVirtualGroupManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockVirtualGroupManager) EXPECT() *MockVirtualGroupManagerMockRecorder {
	return m.recorder
}

// ForceRefreshMeta mocks base method.
func (m *MockVirtualGroupManager) ForceRefreshMeta() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ForceRefreshMeta")
	ret0, _ := ret[0].(error)
	return ret0
}

// ForceRefreshMeta indicates an expected call of ForceRefreshMeta.
func (mr *MockVirtualGroupManagerMockRecorder) ForceRefreshMeta() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ForceRefreshMeta", reflect.TypeOf((*MockVirtualGroupManager)(nil).ForceRefreshMeta))
}

// FreezeSPAndGVGs mocks base method.
func (m *MockVirtualGroupManager) FreezeSPAndGVGs(spID uint32, gvgs []*types0.GlobalVirtualGroup) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "FreezeSPAndGVGs", spID, gvgs)
}

// FreezeSPAndGVGs indicates an expected call of FreezeSPAndGVGs.
func (mr *MockVirtualGroupManagerMockRecorder) FreezeSPAndGVGs(spID, gvgs any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FreezeSPAndGVGs", reflect.TypeOf((*MockVirtualGroupManager)(nil).FreezeSPAndGVGs), spID, gvgs)
}

// GenerateGlobalVirtualGroupMeta mocks base method.
func (m *MockVirtualGroupManager) GenerateGlobalVirtualGroupMeta(genPolicy GenerateGVGSecondarySPsPolicy) (*GlobalVirtualGroupMeta, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateGlobalVirtualGroupMeta", genPolicy)
	ret0, _ := ret[0].(*GlobalVirtualGroupMeta)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateGlobalVirtualGroupMeta indicates an expected call of GenerateGlobalVirtualGroupMeta.
func (mr *MockVirtualGroupManagerMockRecorder) GenerateGlobalVirtualGroupMeta(genPolicy any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateGlobalVirtualGroupMeta", reflect.TypeOf((*MockVirtualGroupManager)(nil).GenerateGlobalVirtualGroupMeta), genPolicy)
}

// PickGlobalVirtualGroup mocks base method.
func (m *MockVirtualGroupManager) PickGlobalVirtualGroup(vgfID uint32) (*GlobalVirtualGroupMeta, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PickGlobalVirtualGroup", vgfID)
	ret0, _ := ret[0].(*GlobalVirtualGroupMeta)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PickGlobalVirtualGroup indicates an expected call of PickGlobalVirtualGroup.
func (mr *MockVirtualGroupManagerMockRecorder) PickGlobalVirtualGroup(vgfID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PickGlobalVirtualGroup", reflect.TypeOf((*MockVirtualGroupManager)(nil).PickGlobalVirtualGroup), vgfID)
}

// PickGlobalVirtualGroupForBucketMigrate mocks base method.
func (m *MockVirtualGroupManager) PickGlobalVirtualGroupForBucketMigrate(filter GVGPickFilter) (*GlobalVirtualGroupMeta, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PickGlobalVirtualGroupForBucketMigrate", filter)
	ret0, _ := ret[0].(*GlobalVirtualGroupMeta)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PickGlobalVirtualGroupForBucketMigrate indicates an expected call of PickGlobalVirtualGroupForBucketMigrate.
func (mr *MockVirtualGroupManagerMockRecorder) PickGlobalVirtualGroupForBucketMigrate(filter any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PickGlobalVirtualGroupForBucketMigrate", reflect.TypeOf((*MockVirtualGroupManager)(nil).PickGlobalVirtualGroupForBucketMigrate), filter)
}

// PickSPByFilter mocks base method.
func (m *MockVirtualGroupManager) PickSPByFilter(filter SPPickFilter) (*types.StorageProvider, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PickSPByFilter", filter)
	ret0, _ := ret[0].(*types.StorageProvider)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PickSPByFilter indicates an expected call of PickSPByFilter.
func (mr *MockVirtualGroupManagerMockRecorder) PickSPByFilter(filter any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PickSPByFilter", reflect.TypeOf((*MockVirtualGroupManager)(nil).PickSPByFilter), filter)
}

// PickVirtualGroupFamily mocks base method.
func (m *MockVirtualGroupManager) PickVirtualGroupFamily() (*VirtualGroupFamilyMeta, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PickVirtualGroupFamily")
	ret0, _ := ret[0].(*VirtualGroupFamilyMeta)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PickVirtualGroupFamily indicates an expected call of PickVirtualGroupFamily.
func (mr *MockVirtualGroupManagerMockRecorder) PickVirtualGroupFamily() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PickVirtualGroupFamily", reflect.TypeOf((*MockVirtualGroupManager)(nil).PickVirtualGroupFamily))
}

// QuerySPByID mocks base method.
func (m *MockVirtualGroupManager) QuerySPByID(spID uint32) (*types.StorageProvider, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QuerySPByID", spID)
	ret0, _ := ret[0].(*types.StorageProvider)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QuerySPByID indicates an expected call of QuerySPByID.
func (mr *MockVirtualGroupManagerMockRecorder) QuerySPByID(spID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QuerySPByID", reflect.TypeOf((*MockVirtualGroupManager)(nil).QuerySPByID), spID)
}

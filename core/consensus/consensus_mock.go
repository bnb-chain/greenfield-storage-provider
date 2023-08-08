// Code generated by MockGen. DO NOT EDIT.
// Source: ./consensus.go

// Package consensus is a generated GoMock package.
package consensus

import (
	context "context"
	reflect "reflect"

	types "github.com/bnb-chain/greenfield/x/payment/types"
	types0 "github.com/bnb-chain/greenfield/x/sp/types"
	types1 "github.com/bnb-chain/greenfield/x/storage/types"
	types2 "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	types3 "github.com/cosmos/cosmos-sdk/types"
	types4 "github.com/cosmos/cosmos-sdk/x/staking/types"
	gomock "go.uber.org/mock/gomock"
)

// MockConsensus is a mock of Consensus interface.
type MockConsensus struct {
	ctrl     *gomock.Controller
	recorder *MockConsensusMockRecorder
}

// MockConsensusMockRecorder is the mock recorder for MockConsensus.
type MockConsensusMockRecorder struct {
	mock *MockConsensus
}

// NewMockConsensus creates a new mock instance.
func NewMockConsensus(ctrl *gomock.Controller) *MockConsensus {
	mock := &MockConsensus{ctrl: ctrl}
	mock.recorder = &MockConsensusMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConsensus) EXPECT() *MockConsensusMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockConsensus) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockConsensusMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockConsensus)(nil).Close))
}

// ConfirmTransaction mocks base method.
func (m *MockConsensus) ConfirmTransaction(ctx context.Context, txHash string) (*types3.TxResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConfirmTransaction", ctx, txHash)
	ret0, _ := ret[0].(*types3.TxResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ConfirmTransaction indicates an expected call of ConfirmTransaction.
func (mr *MockConsensusMockRecorder) ConfirmTransaction(ctx, txHash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConfirmTransaction", reflect.TypeOf((*MockConsensus)(nil).ConfirmTransaction), ctx, txHash)
}

// CurrentHeight mocks base method.
func (m *MockConsensus) CurrentHeight(ctx context.Context) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CurrentHeight", ctx)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CurrentHeight indicates an expected call of CurrentHeight.
func (mr *MockConsensusMockRecorder) CurrentHeight(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CurrentHeight", reflect.TypeOf((*MockConsensus)(nil).CurrentHeight), ctx)
}

// HasAccount mocks base method.
func (m *MockConsensus) HasAccount(ctx context.Context, account string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasAccount", ctx, account)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasAccount indicates an expected call of HasAccount.
func (mr *MockConsensusMockRecorder) HasAccount(ctx, account interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasAccount", reflect.TypeOf((*MockConsensus)(nil).HasAccount), ctx, account)
}

// ListBondedValidators mocks base method.
func (m *MockConsensus) ListBondedValidators(ctx context.Context) ([]types4.Validator, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListBondedValidators", ctx)
	ret0, _ := ret[0].([]types4.Validator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListBondedValidators indicates an expected call of ListBondedValidators.
func (mr *MockConsensusMockRecorder) ListBondedValidators(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListBondedValidators", reflect.TypeOf((*MockConsensus)(nil).ListBondedValidators), ctx)
}

// ListGlobalVirtualGroupsByFamilyID mocks base method.
func (m *MockConsensus) ListGlobalVirtualGroupsByFamilyID(ctx context.Context, vgfID uint32) ([]*types2.GlobalVirtualGroup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListGlobalVirtualGroupsByFamilyID", ctx, vgfID)
	ret0, _ := ret[0].([]*types2.GlobalVirtualGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListGlobalVirtualGroupsByFamilyID indicates an expected call of ListGlobalVirtualGroupsByFamilyID.
func (mr *MockConsensusMockRecorder) ListGlobalVirtualGroupsByFamilyID(ctx, vgfID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListGlobalVirtualGroupsByFamilyID", reflect.TypeOf((*MockConsensus)(nil).ListGlobalVirtualGroupsByFamilyID), ctx, vgfID)
}

// ListSPs mocks base method.
func (m *MockConsensus) ListSPs(ctx context.Context) ([]*types0.StorageProvider, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListSPs", ctx)
	ret0, _ := ret[0].([]*types0.StorageProvider)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListSPs indicates an expected call of ListSPs.
func (mr *MockConsensusMockRecorder) ListSPs(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListSPs", reflect.TypeOf((*MockConsensus)(nil).ListSPs), ctx)
}

// ListVirtualGroupFamilies mocks base method.
func (m *MockConsensus) ListVirtualGroupFamilies(ctx context.Context, spID uint32) ([]*types2.GlobalVirtualGroupFamily, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListVirtualGroupFamilies", ctx, spID)
	ret0, _ := ret[0].([]*types2.GlobalVirtualGroupFamily)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListVirtualGroupFamilies indicates an expected call of ListVirtualGroupFamilies.
func (mr *MockConsensusMockRecorder) ListVirtualGroupFamilies(ctx, spID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListVirtualGroupFamilies", reflect.TypeOf((*MockConsensus)(nil).ListVirtualGroupFamilies), ctx, spID)
}

// ListenObjectSeal mocks base method.
func (m *MockConsensus) ListenObjectSeal(ctx context.Context, objectID uint64, timeOutHeight int) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListenObjectSeal", ctx, objectID, timeOutHeight)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListenObjectSeal indicates an expected call of ListenObjectSeal.
func (mr *MockConsensusMockRecorder) ListenObjectSeal(ctx, objectID, timeOutHeight interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListenObjectSeal", reflect.TypeOf((*MockConsensus)(nil).ListenObjectSeal), ctx, objectID, timeOutHeight)
}

// ListenRejectUnSealObject mocks base method.
func (m *MockConsensus) ListenRejectUnSealObject(ctx context.Context, objectID uint64, timeoutHeight int) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListenRejectUnSealObject", ctx, objectID, timeoutHeight)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListenRejectUnSealObject indicates an expected call of ListenRejectUnSealObject.
func (mr *MockConsensusMockRecorder) ListenRejectUnSealObject(ctx, objectID, timeoutHeight interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListenRejectUnSealObject", reflect.TypeOf((*MockConsensus)(nil).ListenRejectUnSealObject), ctx, objectID, timeoutHeight)
}

// QueryBucketInfo mocks base method.
func (m *MockConsensus) QueryBucketInfo(ctx context.Context, bucket string) (*types1.BucketInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryBucketInfo", ctx, bucket)
	ret0, _ := ret[0].(*types1.BucketInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryBucketInfo indicates an expected call of QueryBucketInfo.
func (mr *MockConsensusMockRecorder) QueryBucketInfo(ctx, bucket interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryBucketInfo", reflect.TypeOf((*MockConsensus)(nil).QueryBucketInfo), ctx, bucket)
}

// QueryBucketInfoAndObjectInfo mocks base method.
func (m *MockConsensus) QueryBucketInfoAndObjectInfo(ctx context.Context, bucket, object string) (*types1.BucketInfo, *types1.ObjectInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryBucketInfoAndObjectInfo", ctx, bucket, object)
	ret0, _ := ret[0].(*types1.BucketInfo)
	ret1, _ := ret[1].(*types1.ObjectInfo)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// QueryBucketInfoAndObjectInfo indicates an expected call of QueryBucketInfoAndObjectInfo.
func (mr *MockConsensusMockRecorder) QueryBucketInfoAndObjectInfo(ctx, bucket, object interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryBucketInfoAndObjectInfo", reflect.TypeOf((*MockConsensus)(nil).QueryBucketInfoAndObjectInfo), ctx, bucket, object)
}

// QueryBucketInfoById mocks base method.
func (m *MockConsensus) QueryBucketInfoById(ctx context.Context, bucketId uint64) (*types1.BucketInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryBucketInfoById", ctx, bucketId)
	ret0, _ := ret[0].(*types1.BucketInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryBucketInfoById indicates an expected call of QueryBucketInfoById.
func (mr *MockConsensusMockRecorder) QueryBucketInfoById(ctx, bucketId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryBucketInfoById", reflect.TypeOf((*MockConsensus)(nil).QueryBucketInfoById), ctx, bucketId)
}

// QueryGlobalVirtualGroup mocks base method.
func (m *MockConsensus) QueryGlobalVirtualGroup(ctx context.Context, gvgID uint32) (*types2.GlobalVirtualGroup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryGlobalVirtualGroup", ctx, gvgID)
	ret0, _ := ret[0].(*types2.GlobalVirtualGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryGlobalVirtualGroup indicates an expected call of QueryGlobalVirtualGroup.
func (mr *MockConsensusMockRecorder) QueryGlobalVirtualGroup(ctx, gvgID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryGlobalVirtualGroup", reflect.TypeOf((*MockConsensus)(nil).QueryGlobalVirtualGroup), ctx, gvgID)
}

// QueryObjectInfo mocks base method.
func (m *MockConsensus) QueryObjectInfo(ctx context.Context, bucket, object string) (*types1.ObjectInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryObjectInfo", ctx, bucket, object)
	ret0, _ := ret[0].(*types1.ObjectInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryObjectInfo indicates an expected call of QueryObjectInfo.
func (mr *MockConsensusMockRecorder) QueryObjectInfo(ctx, bucket, object interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryObjectInfo", reflect.TypeOf((*MockConsensus)(nil).QueryObjectInfo), ctx, bucket, object)
}

// QueryObjectInfoByID mocks base method.
func (m *MockConsensus) QueryObjectInfoByID(ctx context.Context, objectID string) (*types1.ObjectInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryObjectInfoByID", ctx, objectID)
	ret0, _ := ret[0].(*types1.ObjectInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryObjectInfoByID indicates an expected call of QueryObjectInfoByID.
func (mr *MockConsensusMockRecorder) QueryObjectInfoByID(ctx, objectID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryObjectInfoByID", reflect.TypeOf((*MockConsensus)(nil).QueryObjectInfoByID), ctx, objectID)
}

// QueryPaymentStreamRecord mocks base method.
func (m *MockConsensus) QueryPaymentStreamRecord(ctx context.Context, account string) (*types.StreamRecord, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryPaymentStreamRecord", ctx, account)
	ret0, _ := ret[0].(*types.StreamRecord)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryPaymentStreamRecord indicates an expected call of QueryPaymentStreamRecord.
func (mr *MockConsensusMockRecorder) QueryPaymentStreamRecord(ctx, account interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryPaymentStreamRecord", reflect.TypeOf((*MockConsensus)(nil).QueryPaymentStreamRecord), ctx, account)
}

// QuerySP mocks base method.
func (m *MockConsensus) QuerySP(arg0 context.Context, arg1 string) (*types0.StorageProvider, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QuerySP", arg0, arg1)
	ret0, _ := ret[0].(*types0.StorageProvider)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QuerySP indicates an expected call of QuerySP.
func (mr *MockConsensusMockRecorder) QuerySP(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QuerySP", reflect.TypeOf((*MockConsensus)(nil).QuerySP), arg0, arg1)
}

// QuerySPByID mocks base method.
func (m *MockConsensus) QuerySPByID(arg0 context.Context, arg1 uint32) (*types0.StorageProvider, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QuerySPByID", arg0, arg1)
	ret0, _ := ret[0].(*types0.StorageProvider)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QuerySPByID indicates an expected call of QuerySPByID.
func (mr *MockConsensusMockRecorder) QuerySPByID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QuerySPByID", reflect.TypeOf((*MockConsensus)(nil).QuerySPByID), arg0, arg1)
}

// QuerySPFreeQuota mocks base method.
func (m *MockConsensus) QuerySPFreeQuota(arg0 context.Context, arg1 string) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QuerySPFreeQuota", arg0, arg1)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QuerySPFreeQuota indicates an expected call of QuerySPFreeQuota.
func (mr *MockConsensusMockRecorder) QuerySPFreeQuota(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QuerySPFreeQuota", reflect.TypeOf((*MockConsensus)(nil).QuerySPFreeQuota), arg0, arg1)
}

// QuerySPPrice mocks base method.
func (m *MockConsensus) QuerySPPrice(ctx context.Context, operatorAddress string) (types0.SpStoragePrice, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QuerySPPrice", ctx, operatorAddress)
	ret0, _ := ret[0].(types0.SpStoragePrice)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QuerySPPrice indicates an expected call of QuerySPPrice.
func (mr *MockConsensusMockRecorder) QuerySPPrice(ctx, operatorAddress interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QuerySPPrice", reflect.TypeOf((*MockConsensus)(nil).QuerySPPrice), ctx, operatorAddress)
}

// QueryStorageParams mocks base method.
func (m *MockConsensus) QueryStorageParams(ctx context.Context) (*types1.Params, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryStorageParams", ctx)
	ret0, _ := ret[0].(*types1.Params)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryStorageParams indicates an expected call of QueryStorageParams.
func (mr *MockConsensusMockRecorder) QueryStorageParams(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryStorageParams", reflect.TypeOf((*MockConsensus)(nil).QueryStorageParams), ctx)
}

// QueryStorageParamsByTimestamp mocks base method.
func (m *MockConsensus) QueryStorageParamsByTimestamp(ctx context.Context, timestamp int64) (*types1.Params, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryStorageParamsByTimestamp", ctx, timestamp)
	ret0, _ := ret[0].(*types1.Params)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryStorageParamsByTimestamp indicates an expected call of QueryStorageParamsByTimestamp.
func (mr *MockConsensusMockRecorder) QueryStorageParamsByTimestamp(ctx, timestamp interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryStorageParamsByTimestamp", reflect.TypeOf((*MockConsensus)(nil).QueryStorageParamsByTimestamp), ctx, timestamp)
}

// QueryVirtualGroupFamily mocks base method.
func (m *MockConsensus) QueryVirtualGroupFamily(ctx context.Context, vgfID uint32) (*types2.GlobalVirtualGroupFamily, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryVirtualGroupFamily", ctx, vgfID)
	ret0, _ := ret[0].(*types2.GlobalVirtualGroupFamily)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryVirtualGroupFamily indicates an expected call of QueryVirtualGroupFamily.
func (mr *MockConsensusMockRecorder) QueryVirtualGroupFamily(ctx, vgfID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryVirtualGroupFamily", reflect.TypeOf((*MockConsensus)(nil).QueryVirtualGroupFamily), ctx, vgfID)
}

// QueryVirtualGroupParams mocks base method.
func (m *MockConsensus) QueryVirtualGroupParams(ctx context.Context) (*types2.Params, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryVirtualGroupParams", ctx)
	ret0, _ := ret[0].(*types2.Params)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryVirtualGroupParams indicates an expected call of QueryVirtualGroupParams.
func (mr *MockConsensusMockRecorder) QueryVirtualGroupParams(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryVirtualGroupParams", reflect.TypeOf((*MockConsensus)(nil).QueryVirtualGroupParams), ctx)
}

// VerifyGetObjectPermission mocks base method.
func (m *MockConsensus) VerifyGetObjectPermission(ctx context.Context, account, bucket, object string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyGetObjectPermission", ctx, account, bucket, object)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// VerifyGetObjectPermission indicates an expected call of VerifyGetObjectPermission.
func (mr *MockConsensusMockRecorder) VerifyGetObjectPermission(ctx, account, bucket, object interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyGetObjectPermission", reflect.TypeOf((*MockConsensus)(nil).VerifyGetObjectPermission), ctx, account, bucket, object)
}

// VerifyPutObjectPermission mocks base method.
func (m *MockConsensus) VerifyPutObjectPermission(ctx context.Context, account, bucket, object string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyPutObjectPermission", ctx, account, bucket, object)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// VerifyPutObjectPermission indicates an expected call of VerifyPutObjectPermission.
func (mr *MockConsensusMockRecorder) VerifyPutObjectPermission(ctx, account, bucket, object interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyPutObjectPermission", reflect.TypeOf((*MockConsensus)(nil).VerifyPutObjectPermission), ctx, account, bucket, object)
}

// WaitForNextBlock mocks base method.
func (m *MockConsensus) WaitForNextBlock(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitForNextBlock", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitForNextBlock indicates an expected call of WaitForNextBlock.
func (mr *MockConsensusMockRecorder) WaitForNextBlock(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitForNextBlock", reflect.TypeOf((*MockConsensus)(nil).WaitForNextBlock), ctx)
}

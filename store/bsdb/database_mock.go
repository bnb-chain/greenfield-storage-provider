// Code generated by MockGen. DO NOT EDIT.
// Source: store/bsdb/database.go

// Package bsdb is a generated GoMock package.
package bsdb

import (
	reflect "reflect"

	common "github.com/forbole/juno/v4/common"
	gomock "github.com/golang/mock/gomock"
)

// MockMetadata is a mock of Metadata interface.
type MockMetadata struct {
	ctrl     *gomock.Controller
	recorder *MockMetadataMockRecorder
}

// MockMetadataMockRecorder is the mock recorder for MockMetadata.
type MockMetadataMockRecorder struct {
	mock *MockMetadata
}

// NewMockMetadata creates a new mock instance.
func NewMockMetadata(ctrl *gomock.Controller) *MockMetadata {
	mock := &MockMetadata{ctrl: ctrl}
	mock.recorder = &MockMetadataMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetadata) EXPECT() *MockMetadataMockRecorder {
	return m.recorder
}

// GetBucketByID mocks base method.
func (m *MockMetadata) GetBucketByID(bucketID int64, includePrivate bool) (*Bucket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBucketByID", bucketID, includePrivate)
	ret0, _ := ret[0].(*Bucket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBucketByID indicates an expected call of GetBucketByID.
func (mr *MockMetadataMockRecorder) GetBucketByID(bucketID, includePrivate interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBucketByID", reflect.TypeOf((*MockMetadata)(nil).GetBucketByID), bucketID, includePrivate)
}

// GetBucketByName mocks base method.
func (m *MockMetadata) GetBucketByName(bucketName string, includePrivate bool) (*Bucket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBucketByName", bucketName, includePrivate)
	ret0, _ := ret[0].(*Bucket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBucketByName indicates an expected call of GetBucketByName.
func (mr *MockMetadataMockRecorder) GetBucketByName(bucketName, includePrivate interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBucketByName", reflect.TypeOf((*MockMetadata)(nil).GetBucketByName), bucketName, includePrivate)
}

// GetBucketMetaByName mocks base method.
func (m *MockMetadata) GetBucketMetaByName(bucketName string, includePrivate bool) (*BucketFullMeta, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBucketMetaByName", bucketName, includePrivate)
	ret0, _ := ret[0].(*BucketFullMeta)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBucketMetaByName indicates an expected call of GetBucketMetaByName.
func (mr *MockMetadataMockRecorder) GetBucketMetaByName(bucketName, includePrivate interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBucketMetaByName", reflect.TypeOf((*MockMetadata)(nil).GetBucketMetaByName), bucketName, includePrivate)
}

// GetGroupsByGroupIDAndAccount mocks base method.
func (m *MockMetadata) GetGroupsByGroupIDAndAccount(groupIDList []common.Hash, account common.Address, includeRemoved bool) ([]*Group, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGroupsByGroupIDAndAccount", groupIDList, account, includeRemoved)
	ret0, _ := ret[0].([]*Group)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGroupsByGroupIDAndAccount indicates an expected call of GetGroupsByGroupIDAndAccount.
func (mr *MockMetadataMockRecorder) GetGroupsByGroupIDAndAccount(groupIDList, account, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGroupsByGroupIDAndAccount", reflect.TypeOf((*MockMetadata)(nil).GetGroupsByGroupIDAndAccount), groupIDList, account, includeRemoved)
}

// GetLatestBlockNumber mocks base method.
func (m *MockMetadata) GetLatestBlockNumber() (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLatestBlockNumber")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLatestBlockNumber indicates an expected call of GetLatestBlockNumber.
func (mr *MockMetadataMockRecorder) GetLatestBlockNumber() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLatestBlockNumber", reflect.TypeOf((*MockMetadata)(nil).GetLatestBlockNumber))
}

// GetObjectByName mocks base method.
func (m *MockMetadata) GetObjectByName(objectName, bucketName string, includePrivate bool) (*Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetObjectByName", objectName, bucketName, includePrivate)
	ret0, _ := ret[0].(*Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetObjectByName indicates an expected call of GetObjectByName.
func (mr *MockMetadataMockRecorder) GetObjectByName(objectName, bucketName, includePrivate interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetObjectByName", reflect.TypeOf((*MockMetadata)(nil).GetObjectByName), objectName, bucketName, includePrivate)
}

// GetPaymentByBucketID mocks base method.
func (m *MockMetadata) GetPaymentByBucketID(bucketID int64, includePrivate bool) (*StreamRecord, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPaymentByBucketID", bucketID, includePrivate)
	ret0, _ := ret[0].(*StreamRecord)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPaymentByBucketID indicates an expected call of GetPaymentByBucketID.
func (mr *MockMetadataMockRecorder) GetPaymentByBucketID(bucketID, includePrivate interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPaymentByBucketID", reflect.TypeOf((*MockMetadata)(nil).GetPaymentByBucketID), bucketID, includePrivate)
}

// GetPaymentByBucketName mocks base method.
func (m *MockMetadata) GetPaymentByBucketName(bucketName string, includePrivate bool) (*StreamRecord, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPaymentByBucketName", bucketName, includePrivate)
	ret0, _ := ret[0].(*StreamRecord)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPaymentByBucketName indicates an expected call of GetPaymentByBucketName.
func (mr *MockMetadataMockRecorder) GetPaymentByBucketName(bucketName, includePrivate interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPaymentByBucketName", reflect.TypeOf((*MockMetadata)(nil).GetPaymentByBucketName), bucketName, includePrivate)
}

// GetPaymentByPaymentAddress mocks base method.
func (m *MockMetadata) GetPaymentByPaymentAddress(address common.Address) (*StreamRecord, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPaymentByPaymentAddress", address)
	ret0, _ := ret[0].(*StreamRecord)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPaymentByPaymentAddress indicates an expected call of GetPaymentByPaymentAddress.
func (mr *MockMetadataMockRecorder) GetPaymentByPaymentAddress(address interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPaymentByPaymentAddress", reflect.TypeOf((*MockMetadata)(nil).GetPaymentByPaymentAddress), address)
}

// GetPermissionByResourceAndPrincipal mocks base method.
func (m *MockMetadata) GetPermissionByResourceAndPrincipal(resourceType, principalType, principalValue string, resourceID common.Hash) (*Permission, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPermissionByResourceAndPrincipal", resourceType, principalType, principalValue, resourceID)
	ret0, _ := ret[0].(*Permission)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPermissionByResourceAndPrincipal indicates an expected call of GetPermissionByResourceAndPrincipal.
func (mr *MockMetadataMockRecorder) GetPermissionByResourceAndPrincipal(resourceType, principalType, principalValue, resourceID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPermissionByResourceAndPrincipal", reflect.TypeOf((*MockMetadata)(nil).GetPermissionByResourceAndPrincipal), resourceType, principalType, principalValue, resourceID)
}

// GetPermissionsByResourceAndPrincipleType mocks base method.
func (m *MockMetadata) GetPermissionsByResourceAndPrincipleType(resourceType, principalType string, resourceID common.Hash, includeRemoved bool) ([]*Permission, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPermissionsByResourceAndPrincipleType", resourceType, principalType, resourceID, includeRemoved)
	ret0, _ := ret[0].([]*Permission)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPermissionsByResourceAndPrincipleType indicates an expected call of GetPermissionsByResourceAndPrincipleType.
func (mr *MockMetadataMockRecorder) GetPermissionsByResourceAndPrincipleType(resourceType, principalType, resourceID, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPermissionsByResourceAndPrincipleType", reflect.TypeOf((*MockMetadata)(nil).GetPermissionsByResourceAndPrincipleType), resourceType, principalType, resourceID, includeRemoved)
}

// GetStatementsByPolicyID mocks base method.
func (m *MockMetadata) GetStatementsByPolicyID(policyIDList []common.Hash, includeRemoved bool) ([]*Statement, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStatementsByPolicyID", policyIDList, includeRemoved)
	ret0, _ := ret[0].([]*Statement)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStatementsByPolicyID indicates an expected call of GetStatementsByPolicyID.
func (mr *MockMetadataMockRecorder) GetStatementsByPolicyID(policyIDList, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStatementsByPolicyID", reflect.TypeOf((*MockMetadata)(nil).GetStatementsByPolicyID), policyIDList, includeRemoved)
}

// GetSwitchDBSignal mocks base method.
func (m *MockMetadata) GetSwitchDBSignal() (*MasterDB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSwitchDBSignal")
	ret0, _ := ret[0].(*MasterDB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSwitchDBSignal indicates an expected call of GetSwitchDBSignal.
func (mr *MockMetadataMockRecorder) GetSwitchDBSignal() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSwitchDBSignal", reflect.TypeOf((*MockMetadata)(nil).GetSwitchDBSignal))
}

// GetUserBuckets mocks base method.
func (m *MockMetadata) GetUserBuckets(accountID common.Address, includeRemoved bool) ([]*Bucket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserBuckets", accountID, includeRemoved)
	ret0, _ := ret[0].([]*Bucket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserBuckets indicates an expected call of GetUserBuckets.
func (mr *MockMetadataMockRecorder) GetUserBuckets(accountID, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserBuckets", reflect.TypeOf((*MockMetadata)(nil).GetUserBuckets), accountID, includeRemoved)
}

// GetUserBucketsCount mocks base method.
func (m *MockMetadata) GetUserBucketsCount(accountID common.Address, includeRemoved bool) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserBucketsCount", accountID, includeRemoved)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserBucketsCount indicates an expected call of GetUserBucketsCount.
func (mr *MockMetadataMockRecorder) GetUserBucketsCount(accountID, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserBucketsCount", reflect.TypeOf((*MockMetadata)(nil).GetUserBucketsCount), accountID, includeRemoved)
}

// ListBucketsByBucketID mocks base method.
func (m *MockMetadata) ListBucketsByBucketID(ids []common.Hash, includeRemoved bool) ([]*Bucket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListBucketsByBucketID", ids, includeRemoved)
	ret0, _ := ret[0].([]*Bucket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListBucketsByBucketID indicates an expected call of ListBucketsByBucketID.
func (mr *MockMetadataMockRecorder) ListBucketsByBucketID(ids, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListBucketsByBucketID", reflect.TypeOf((*MockMetadata)(nil).ListBucketsByBucketID), ids, includeRemoved)
}

// ListDeletedObjectsByBlockNumberRange mocks base method.
func (m *MockMetadata) ListDeletedObjectsByBlockNumberRange(startBlockNumber, endBlockNumber int64, includePrivate bool) ([]*Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListDeletedObjectsByBlockNumberRange", startBlockNumber, endBlockNumber, includePrivate)
	ret0, _ := ret[0].([]*Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListDeletedObjectsByBlockNumberRange indicates an expected call of ListDeletedObjectsByBlockNumberRange.
func (mr *MockMetadataMockRecorder) ListDeletedObjectsByBlockNumberRange(startBlockNumber, endBlockNumber, includePrivate interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListDeletedObjectsByBlockNumberRange", reflect.TypeOf((*MockMetadata)(nil).ListDeletedObjectsByBlockNumberRange), startBlockNumber, endBlockNumber, includePrivate)
}

// ListExpiredBucketsBySp mocks base method.
func (m *MockMetadata) ListExpiredBucketsBySp(createAt int64, primarySpID uint32, limit int64) ([]*Bucket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListExpiredBucketsBySp", createAt, primarySpID, limit)
	ret0, _ := ret[0].([]*Bucket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListExpiredBucketsBySp indicates an expected call of ListExpiredBucketsBySp.
func (mr *MockMetadataMockRecorder) ListExpiredBucketsBySp(createAt, primarySpID, limit interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListExpiredBucketsBySp", reflect.TypeOf((*MockMetadata)(nil).ListExpiredBucketsBySp), createAt, primarySpID, limit)
}

// ListGroupsByNameAndSourceType mocks base method.
func (m *MockMetadata) ListGroupsByNameAndSourceType(name, prefix, sourceType string, limit, offset int, includeRemoved bool) ([]*Group, int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListGroupsByNameAndSourceType", name, prefix, sourceType, limit, offset, includeRemoved)
	ret0, _ := ret[0].([]*Group)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ListGroupsByNameAndSourceType indicates an expected call of ListGroupsByNameAndSourceType.
func (mr *MockMetadataMockRecorder) ListGroupsByNameAndSourceType(name, prefix, sourceType, limit, offset, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListGroupsByNameAndSourceType", reflect.TypeOf((*MockMetadata)(nil).ListGroupsByNameAndSourceType), name, prefix, sourceType, limit, offset, includeRemoved)
}

// ListObjectsByBucketName mocks base method.
func (m *MockMetadata) ListObjectsByBucketName(bucketName, continuationToken, prefix, delimiter string, maxKeys int, includeRemoved bool) ([]*ListObjectsResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListObjectsByBucketName", bucketName, continuationToken, prefix, delimiter, maxKeys, includeRemoved)
	ret0, _ := ret[0].([]*ListObjectsResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListObjectsByBucketName indicates an expected call of ListObjectsByBucketName.
func (mr *MockMetadataMockRecorder) ListObjectsByBucketName(bucketName, continuationToken, prefix, delimiter, maxKeys, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListObjectsByBucketName", reflect.TypeOf((*MockMetadata)(nil).ListObjectsByBucketName), bucketName, continuationToken, prefix, delimiter, maxKeys, includeRemoved)
}

// ListObjectsByObjectID mocks base method.
func (m *MockMetadata) ListObjectsByObjectID(ids []common.Hash, includeRemoved bool) ([]*Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListObjectsByObjectID", ids, includeRemoved)
	ret0, _ := ret[0].([]*Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListObjectsByObjectID indicates an expected call of ListObjectsByObjectID.
func (mr *MockMetadataMockRecorder) ListObjectsByObjectID(ids, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListObjectsByObjectID", reflect.TypeOf((*MockMetadata)(nil).ListObjectsByObjectID), ids, includeRemoved)
}

// MockBSDB is a mock of BSDB interface.
type MockBSDB struct {
	ctrl     *gomock.Controller
	recorder *MockBSDBMockRecorder
}

// MockBSDBMockRecorder is the mock recorder for MockBSDB.
type MockBSDBMockRecorder struct {
	mock *MockBSDB
}

// NewMockBSDB creates a new mock instance.
func NewMockBSDB(ctrl *gomock.Controller) *MockBSDB {
	mock := &MockBSDB{ctrl: ctrl}
	mock.recorder = &MockBSDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBSDB) EXPECT() *MockBSDBMockRecorder {
	return m.recorder
}

// GetBucketByID mocks base method.
func (m *MockBSDB) GetBucketByID(bucketID int64, includePrivate bool) (*Bucket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBucketByID", bucketID, includePrivate)
	ret0, _ := ret[0].(*Bucket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBucketByID indicates an expected call of GetBucketByID.
func (mr *MockBSDBMockRecorder) GetBucketByID(bucketID, includePrivate interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBucketByID", reflect.TypeOf((*MockBSDB)(nil).GetBucketByID), bucketID, includePrivate)
}

// GetBucketByName mocks base method.
func (m *MockBSDB) GetBucketByName(bucketName string, includePrivate bool) (*Bucket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBucketByName", bucketName, includePrivate)
	ret0, _ := ret[0].(*Bucket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBucketByName indicates an expected call of GetBucketByName.
func (mr *MockBSDBMockRecorder) GetBucketByName(bucketName, includePrivate interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBucketByName", reflect.TypeOf((*MockBSDB)(nil).GetBucketByName), bucketName, includePrivate)
}

// GetBucketMetaByName mocks base method.
func (m *MockBSDB) GetBucketMetaByName(bucketName string, includePrivate bool) (*BucketFullMeta, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBucketMetaByName", bucketName, includePrivate)
	ret0, _ := ret[0].(*BucketFullMeta)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBucketMetaByName indicates an expected call of GetBucketMetaByName.
func (mr *MockBSDBMockRecorder) GetBucketMetaByName(bucketName, includePrivate interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBucketMetaByName", reflect.TypeOf((*MockBSDB)(nil).GetBucketMetaByName), bucketName, includePrivate)
}

// GetGroupsByGroupIDAndAccount mocks base method.
func (m *MockBSDB) GetGroupsByGroupIDAndAccount(groupIDList []common.Hash, account common.Address, includeRemoved bool) ([]*Group, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGroupsByGroupIDAndAccount", groupIDList, account, includeRemoved)
	ret0, _ := ret[0].([]*Group)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGroupsByGroupIDAndAccount indicates an expected call of GetGroupsByGroupIDAndAccount.
func (mr *MockBSDBMockRecorder) GetGroupsByGroupIDAndAccount(groupIDList, account, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGroupsByGroupIDAndAccount", reflect.TypeOf((*MockBSDB)(nil).GetGroupsByGroupIDAndAccount), groupIDList, account, includeRemoved)
}

// GetLatestBlockNumber mocks base method.
func (m *MockBSDB) GetLatestBlockNumber() (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLatestBlockNumber")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLatestBlockNumber indicates an expected call of GetLatestBlockNumber.
func (mr *MockBSDBMockRecorder) GetLatestBlockNumber() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLatestBlockNumber", reflect.TypeOf((*MockBSDB)(nil).GetLatestBlockNumber))
}

// GetObjectByName mocks base method.
func (m *MockBSDB) GetObjectByName(objectName, bucketName string, includePrivate bool) (*Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetObjectByName", objectName, bucketName, includePrivate)
	ret0, _ := ret[0].(*Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetObjectByName indicates an expected call of GetObjectByName.
func (mr *MockBSDBMockRecorder) GetObjectByName(objectName, bucketName, includePrivate interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetObjectByName", reflect.TypeOf((*MockBSDB)(nil).GetObjectByName), objectName, bucketName, includePrivate)
}

// GetPaymentByBucketID mocks base method.
func (m *MockBSDB) GetPaymentByBucketID(bucketID int64, includePrivate bool) (*StreamRecord, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPaymentByBucketID", bucketID, includePrivate)
	ret0, _ := ret[0].(*StreamRecord)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPaymentByBucketID indicates an expected call of GetPaymentByBucketID.
func (mr *MockBSDBMockRecorder) GetPaymentByBucketID(bucketID, includePrivate interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPaymentByBucketID", reflect.TypeOf((*MockBSDB)(nil).GetPaymentByBucketID), bucketID, includePrivate)
}

// GetPaymentByBucketName mocks base method.
func (m *MockBSDB) GetPaymentByBucketName(bucketName string, includePrivate bool) (*StreamRecord, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPaymentByBucketName", bucketName, includePrivate)
	ret0, _ := ret[0].(*StreamRecord)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPaymentByBucketName indicates an expected call of GetPaymentByBucketName.
func (mr *MockBSDBMockRecorder) GetPaymentByBucketName(bucketName, includePrivate interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPaymentByBucketName", reflect.TypeOf((*MockBSDB)(nil).GetPaymentByBucketName), bucketName, includePrivate)
}

// GetPaymentByPaymentAddress mocks base method.
func (m *MockBSDB) GetPaymentByPaymentAddress(address common.Address) (*StreamRecord, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPaymentByPaymentAddress", address)
	ret0, _ := ret[0].(*StreamRecord)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPaymentByPaymentAddress indicates an expected call of GetPaymentByPaymentAddress.
func (mr *MockBSDBMockRecorder) GetPaymentByPaymentAddress(address interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPaymentByPaymentAddress", reflect.TypeOf((*MockBSDB)(nil).GetPaymentByPaymentAddress), address)
}

// GetPermissionByResourceAndPrincipal mocks base method.
func (m *MockBSDB) GetPermissionByResourceAndPrincipal(resourceType, principalType, principalValue string, resourceID common.Hash) (*Permission, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPermissionByResourceAndPrincipal", resourceType, principalType, principalValue, resourceID)
	ret0, _ := ret[0].(*Permission)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPermissionByResourceAndPrincipal indicates an expected call of GetPermissionByResourceAndPrincipal.
func (mr *MockBSDBMockRecorder) GetPermissionByResourceAndPrincipal(resourceType, principalType, principalValue, resourceID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPermissionByResourceAndPrincipal", reflect.TypeOf((*MockBSDB)(nil).GetPermissionByResourceAndPrincipal), resourceType, principalType, principalValue, resourceID)
}

// GetPermissionsByResourceAndPrincipleType mocks base method.
func (m *MockBSDB) GetPermissionsByResourceAndPrincipleType(resourceType, principalType string, resourceID common.Hash, includeRemoved bool) ([]*Permission, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPermissionsByResourceAndPrincipleType", resourceType, principalType, resourceID, includeRemoved)
	ret0, _ := ret[0].([]*Permission)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPermissionsByResourceAndPrincipleType indicates an expected call of GetPermissionsByResourceAndPrincipleType.
func (mr *MockBSDBMockRecorder) GetPermissionsByResourceAndPrincipleType(resourceType, principalType, resourceID, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPermissionsByResourceAndPrincipleType", reflect.TypeOf((*MockBSDB)(nil).GetPermissionsByResourceAndPrincipleType), resourceType, principalType, resourceID, includeRemoved)
}

// GetStatementsByPolicyID mocks base method.
func (m *MockBSDB) GetStatementsByPolicyID(policyIDList []common.Hash, includeRemoved bool) ([]*Statement, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStatementsByPolicyID", policyIDList, includeRemoved)
	ret0, _ := ret[0].([]*Statement)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStatementsByPolicyID indicates an expected call of GetStatementsByPolicyID.
func (mr *MockBSDBMockRecorder) GetStatementsByPolicyID(policyIDList, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStatementsByPolicyID", reflect.TypeOf((*MockBSDB)(nil).GetStatementsByPolicyID), policyIDList, includeRemoved)
}

// GetSwitchDBSignal mocks base method.
func (m *MockBSDB) GetSwitchDBSignal() (*MasterDB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSwitchDBSignal")
	ret0, _ := ret[0].(*MasterDB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSwitchDBSignal indicates an expected call of GetSwitchDBSignal.
func (mr *MockBSDBMockRecorder) GetSwitchDBSignal() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSwitchDBSignal", reflect.TypeOf((*MockBSDB)(nil).GetSwitchDBSignal))
}

// GetUserBuckets mocks base method.
func (m *MockBSDB) GetUserBuckets(accountID common.Address, includeRemoved bool) ([]*Bucket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserBuckets", accountID, includeRemoved)
	ret0, _ := ret[0].([]*Bucket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserBuckets indicates an expected call of GetUserBuckets.
func (mr *MockBSDBMockRecorder) GetUserBuckets(accountID, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserBuckets", reflect.TypeOf((*MockBSDB)(nil).GetUserBuckets), accountID, includeRemoved)
}

// GetUserBucketsCount mocks base method.
func (m *MockBSDB) GetUserBucketsCount(accountID common.Address, includeRemoved bool) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserBucketsCount", accountID, includeRemoved)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserBucketsCount indicates an expected call of GetUserBucketsCount.
func (mr *MockBSDBMockRecorder) GetUserBucketsCount(accountID, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserBucketsCount", reflect.TypeOf((*MockBSDB)(nil).GetUserBucketsCount), accountID, includeRemoved)
}

// ListBucketsByBucketID mocks base method.
func (m *MockBSDB) ListBucketsByBucketID(ids []common.Hash, includeRemoved bool) ([]*Bucket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListBucketsByBucketID", ids, includeRemoved)
	ret0, _ := ret[0].([]*Bucket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListBucketsByBucketID indicates an expected call of ListBucketsByBucketID.
func (mr *MockBSDBMockRecorder) ListBucketsByBucketID(ids, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListBucketsByBucketID", reflect.TypeOf((*MockBSDB)(nil).ListBucketsByBucketID), ids, includeRemoved)
}

// ListDeletedObjectsByBlockNumberRange mocks base method.
func (m *MockBSDB) ListDeletedObjectsByBlockNumberRange(startBlockNumber, endBlockNumber int64, includePrivate bool) ([]*Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListDeletedObjectsByBlockNumberRange", startBlockNumber, endBlockNumber, includePrivate)
	ret0, _ := ret[0].([]*Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListDeletedObjectsByBlockNumberRange indicates an expected call of ListDeletedObjectsByBlockNumberRange.
func (mr *MockBSDBMockRecorder) ListDeletedObjectsByBlockNumberRange(startBlockNumber, endBlockNumber, includePrivate interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListDeletedObjectsByBlockNumberRange", reflect.TypeOf((*MockBSDB)(nil).ListDeletedObjectsByBlockNumberRange), startBlockNumber, endBlockNumber, includePrivate)
}

// ListExpiredBucketsBySp mocks base method.
func (m *MockBSDB) ListExpiredBucketsBySp(createAt int64, primarySpID uint32, limit int64) ([]*Bucket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListExpiredBucketsBySp", createAt, primarySpID, limit)
	ret0, _ := ret[0].([]*Bucket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListExpiredBucketsBySp indicates an expected call of ListExpiredBucketsBySp.
func (mr *MockBSDBMockRecorder) ListExpiredBucketsBySp(createAt, primarySpID, limit interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListExpiredBucketsBySp", reflect.TypeOf((*MockBSDB)(nil).ListExpiredBucketsBySp), createAt, primarySpID, limit)
}

// ListGroupsByNameAndSourceType mocks base method.
func (m *MockBSDB) ListGroupsByNameAndSourceType(name, prefix, sourceType string, limit, offset int, includeRemoved bool) ([]*Group, int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListGroupsByNameAndSourceType", name, prefix, sourceType, limit, offset, includeRemoved)
	ret0, _ := ret[0].([]*Group)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ListGroupsByNameAndSourceType indicates an expected call of ListGroupsByNameAndSourceType.
func (mr *MockBSDBMockRecorder) ListGroupsByNameAndSourceType(name, prefix, sourceType, limit, offset, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListGroupsByNameAndSourceType", reflect.TypeOf((*MockBSDB)(nil).ListGroupsByNameAndSourceType), name, prefix, sourceType, limit, offset, includeRemoved)
}

// ListObjectsByBucketName mocks base method.
func (m *MockBSDB) ListObjectsByBucketName(bucketName, continuationToken, prefix, delimiter string, maxKeys int, includeRemoved bool) ([]*ListObjectsResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListObjectsByBucketName", bucketName, continuationToken, prefix, delimiter, maxKeys, includeRemoved)
	ret0, _ := ret[0].([]*ListObjectsResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListObjectsByBucketName indicates an expected call of ListObjectsByBucketName.
func (mr *MockBSDBMockRecorder) ListObjectsByBucketName(bucketName, continuationToken, prefix, delimiter, maxKeys, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListObjectsByBucketName", reflect.TypeOf((*MockBSDB)(nil).ListObjectsByBucketName), bucketName, continuationToken, prefix, delimiter, maxKeys, includeRemoved)
}

// ListObjectsByObjectID mocks base method.
func (m *MockBSDB) ListObjectsByObjectID(ids []common.Hash, includeRemoved bool) ([]*Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListObjectsByObjectID", ids, includeRemoved)
	ret0, _ := ret[0].([]*Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListObjectsByObjectID indicates an expected call of ListObjectsByObjectID.
func (mr *MockBSDBMockRecorder) ListObjectsByObjectID(ids, includeRemoved interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListObjectsByObjectID", reflect.TypeOf((*MockBSDB)(nil).ListObjectsByObjectID), ids, includeRemoved)
}

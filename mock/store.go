package mock

import (
	"context"
	reflect "reflect"

	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
	"github.com/golang/mock/gomock"
)

// MockIStore is a mock of IStore interface.
type MockIStore struct {
	ctrl     *gomock.Controller
	recorder *MockIStoreMockRecorder
}

// MockIStoreMockRecorder is the mock recorder for MockIStore.
type MockIStoreMockRecorder struct {
	mock *MockIStore
}

// NewMockIStore creates a new mock instance.
func NewMockIStore(ctrl *gomock.Controller) *MockIStore {
	mock := &MockIStore{ctrl: ctrl}
	mock.recorder = &MockIStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIStore) EXPECT() *MockIStoreMockRecorder {
	return m.recorder
}

// GetAllApps mocks base method.
func (m *MockIStore) GetUserBuckets(ctx context.Context) ([]*model.Bucket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserBuckets", ctx)
	ret0, _ := ret[0].([]*model.Bucket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserBuckets indicates an expected call of GetUserBuckets.
func (mr *MockIStoreMockRecorder) GetUserBuckets(ctx context.Context) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserBuckets", reflect.TypeOf((*MockIStore)(nil).GetUserBuckets), ctx)
}

// GetAllApps mocks base method.

func (m *MockIStore) ListObjectsByBucketName(ctx context.Context, bucketName string) ([]*model.Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListObjectsByBucket", ctx, bucketName)
	ret0, _ := ret[0].([]*model.Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserBuckets indicates an expected call of GetUserBuckets.
func (mr *MockIStoreMockRecorder) ListObjectsByBucketName(ctx context.Context, bucketName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListObjectsByBucket", reflect.TypeOf((*MockIStore)(nil).ListObjectsByBucketName), ctx, bucketName)
}

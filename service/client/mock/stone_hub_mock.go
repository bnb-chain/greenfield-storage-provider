// Code generated by MockGen. DO NOT EDIT.
// Source: ./stone_hub_client.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	v1 "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
)

// MockStoneHubAPI is a mock of StoneHubAPI interface.
type MockStoneHubAPI struct {
	ctrl     *gomock.Controller
	recorder *MockStoneHubAPIMockRecorder
}

// MockStoneHubAPIMockRecorder is the mock recorder for MockStoneHubAPI.
type MockStoneHubAPIMockRecorder struct {
	mock *MockStoneHubAPI
}

// NewMockStoneHubAPI creates a new mock instance.
func NewMockStoneHubAPI(ctrl *gomock.Controller) *MockStoneHubAPI {
	mock := &MockStoneHubAPI{ctrl: ctrl}
	mock.recorder = &MockStoneHubAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStoneHubAPI) EXPECT() *MockStoneHubAPIMockRecorder {
	return m.recorder
}

// AllocStoneJob mocks base method.
func (m *MockStoneHubAPI) AllocStoneJob(ctx context.Context, opts ...grpc.CallOption) (*v1.StoneHubServiceAllocStoneJobResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "AllocStoneJob", varargs...)
	ret0, _ := ret[0].(*v1.StoneHubServiceAllocStoneJobResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AllocStoneJob indicates an expected call of AllocStoneJob.
func (mr *MockStoneHubAPIMockRecorder) AllocStoneJob(ctx interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllocStoneJob", reflect.TypeOf((*MockStoneHubAPI)(nil).AllocStoneJob), varargs...)
}

// BeginUploadPayload mocks base method.
func (m *MockStoneHubAPI) BeginUploadPayload(ctx context.Context, in *v1.StoneHubServiceBeginUploadPayloadRequest, opts ...grpc.CallOption) (*v1.StoneHubServiceBeginUploadPayloadResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "BeginUploadPayload", varargs...)
	ret0, _ := ret[0].(*v1.StoneHubServiceBeginUploadPayloadResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeginUploadPayload indicates an expected call of BeginUploadPayload.
func (mr *MockStoneHubAPIMockRecorder) BeginUploadPayload(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginUploadPayload", reflect.TypeOf((*MockStoneHubAPI)(nil).BeginUploadPayload), varargs...)
}

// BeginUploadPayloadV2 mocks base method.
func (m *MockStoneHubAPI) BeginUploadPayloadV2(ctx context.Context, in *v1.StoneHubServiceBeginUploadPayloadV2Request, opts ...grpc.CallOption) (*v1.StoneHubServiceBeginUploadPayloadV2Response, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "BeginUploadPayloadV2", varargs...)
	ret0, _ := ret[0].(*v1.StoneHubServiceBeginUploadPayloadV2Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeginUploadPayloadV2 indicates an expected call of BeginUploadPayloadV2.
func (mr *MockStoneHubAPIMockRecorder) BeginUploadPayloadV2(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginUploadPayloadV2", reflect.TypeOf((*MockStoneHubAPI)(nil).BeginUploadPayloadV2), varargs...)
}

// Close mocks base method.
func (m *MockStoneHubAPI) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockStoneHubAPIMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStoneHubAPI)(nil).Close))
}

// CreateObject mocks base method.
func (m *MockStoneHubAPI) CreateObject(ctx context.Context, in *v1.StoneHubServiceCreateObjectRequest, opts ...grpc.CallOption) (*v1.StoneHubServiceCreateObjectResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateObject", varargs...)
	ret0, _ := ret[0].(*v1.StoneHubServiceCreateObjectResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateObject indicates an expected call of CreateObject.
func (mr *MockStoneHubAPIMockRecorder) CreateObject(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateObject", reflect.TypeOf((*MockStoneHubAPI)(nil).CreateObject), varargs...)
}

// DonePrimaryPieceJob mocks base method.
func (m *MockStoneHubAPI) DonePrimaryPieceJob(ctx context.Context, in *v1.StoneHubServiceDonePrimaryPieceJobRequest, opts ...grpc.CallOption) (*v1.StoneHubServiceDonePrimaryPieceJobResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DonePrimaryPieceJob", varargs...)
	ret0, _ := ret[0].(*v1.StoneHubServiceDonePrimaryPieceJobResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DonePrimaryPieceJob indicates an expected call of DonePrimaryPieceJob.
func (mr *MockStoneHubAPIMockRecorder) DonePrimaryPieceJob(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DonePrimaryPieceJob", reflect.TypeOf((*MockStoneHubAPI)(nil).DonePrimaryPieceJob), varargs...)
}

// DoneSecondaryPieceJob mocks base method.
func (m *MockStoneHubAPI) DoneSecondaryPieceJob(ctx context.Context, in *v1.StoneHubServiceDoneSecondaryPieceJobRequest, opts ...grpc.CallOption) (*v1.StoneHubServiceDoneSecondaryPieceJobResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DoneSecondaryPieceJob", varargs...)
	ret0, _ := ret[0].(*v1.StoneHubServiceDoneSecondaryPieceJobResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DoneSecondaryPieceJob indicates an expected call of DoneSecondaryPieceJob.
func (mr *MockStoneHubAPIMockRecorder) DoneSecondaryPieceJob(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DoneSecondaryPieceJob", reflect.TypeOf((*MockStoneHubAPI)(nil).DoneSecondaryPieceJob), varargs...)
}

// SetObjectCreateInfo mocks base method.
func (m *MockStoneHubAPI) SetObjectCreateInfo(ctx context.Context, in *v1.StoneHubServiceSetObjectCreateInfoRequest, opts ...grpc.CallOption) (*v1.StoneHubServiceSetObjectCreateInfoResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SetObjectCreateInfo", varargs...)
	ret0, _ := ret[0].(*v1.StoneHubServiceSetObjectCreateInfoResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SetObjectCreateInfo indicates an expected call of SetObjectCreateInfo.
func (mr *MockStoneHubAPIMockRecorder) SetObjectCreateInfo(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetObjectCreateInfo", reflect.TypeOf((*MockStoneHubAPI)(nil).SetObjectCreateInfo), varargs...)
}

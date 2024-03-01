// Code generated by MockGen. DO NOT EDIT.
// Source: ./piecestore.go

// Package piecestore is a generated GoMock package.
package piecestore

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockPieceOp is a mock of PieceOp interface.
type MockPieceOp struct {
	ctrl     *gomock.Controller
	recorder *MockPieceOpMockRecorder
}

// MockPieceOpMockRecorder is the mock recorder for MockPieceOp.
type MockPieceOpMockRecorder struct {
	mock *MockPieceOp
}

// NewMockPieceOp creates a new mock instance.
func NewMockPieceOp(ctrl *gomock.Controller) *MockPieceOp {
	mock := &MockPieceOp{ctrl: ctrl}
	mock.recorder = &MockPieceOpMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPieceOp) EXPECT() *MockPieceOpMockRecorder {
	return m.recorder
}

// ChallengePieceKey mocks base method.
func (m *MockPieceOp) ChallengePieceKey(objectID uint64, segmentIdx uint32, redundancyIdx int32, version int64) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ChallengePieceKey", objectID, segmentIdx, redundancyIdx, version)
	ret0, _ := ret[0].(string)
	return ret0
}

// ChallengePieceKey indicates an expected call of ChallengePieceKey.
func (mr *MockPieceOpMockRecorder) ChallengePieceKey(objectID, segmentIdx, redundancyIdx, version interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChallengePieceKey", reflect.TypeOf((*MockPieceOp)(nil).ChallengePieceKey), objectID, segmentIdx, redundancyIdx, version)
}

// ECPieceKey mocks base method.
func (m *MockPieceOp) ECPieceKey(objectID uint64, segmentIdx, redundancyIdx uint32, version int64) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ECPieceKey", objectID, segmentIdx, redundancyIdx, version)
	ret0, _ := ret[0].(string)
	return ret0
}

// ECPieceKey indicates an expected call of ECPieceKey.
func (mr *MockPieceOpMockRecorder) ECPieceKey(objectID, segmentIdx, redundancyIdx, version interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ECPieceKey", reflect.TypeOf((*MockPieceOp)(nil).ECPieceKey), objectID, segmentIdx, redundancyIdx, version)
}

// ECPieceSize mocks base method.
func (m *MockPieceOp) ECPieceSize(payloadSize uint64, segmentIdx uint32, maxSegmentSize uint64, chunkNum uint32) int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ECPieceSize", payloadSize, segmentIdx, maxSegmentSize, chunkNum)
	ret0, _ := ret[0].(int64)
	return ret0
}

// ECPieceSize indicates an expected call of ECPieceSize.
func (mr *MockPieceOpMockRecorder) ECPieceSize(payloadSize, segmentIdx, maxSegmentSize, chunkNum interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ECPieceSize", reflect.TypeOf((*MockPieceOp)(nil).ECPieceSize), payloadSize, segmentIdx, maxSegmentSize, chunkNum)
}

// MaxSegmentPieceSize mocks base method.
func (m *MockPieceOp) MaxSegmentPieceSize(payloadSize, maxSegmentSize uint64) int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MaxSegmentPieceSize", payloadSize, maxSegmentSize)
	ret0, _ := ret[0].(int64)
	return ret0
}

// MaxSegmentPieceSize indicates an expected call of MaxSegmentPieceSize.
func (mr *MockPieceOpMockRecorder) MaxSegmentPieceSize(payloadSize, maxSegmentSize interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MaxSegmentPieceSize", reflect.TypeOf((*MockPieceOp)(nil).MaxSegmentPieceSize), payloadSize, maxSegmentSize)
}

// ParseChallengeIdx mocks base method.
func (m *MockPieceOp) ParseChallengeIdx(challengeKey string) (uint32, int32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParseChallengeIdx", challengeKey)
	ret0, _ := ret[0].(uint32)
	ret1, _ := ret[1].(int32)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ParseChallengeIdx indicates an expected call of ParseChallengeIdx.
func (mr *MockPieceOpMockRecorder) ParseChallengeIdx(challengeKey interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseChallengeIdx", reflect.TypeOf((*MockPieceOp)(nil).ParseChallengeIdx), challengeKey)
}

// ParseSegmentIdx mocks base method.
func (m *MockPieceOp) ParseSegmentIdx(segmentKey string) (uint32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParseSegmentIdx", segmentKey)
	ret0, _ := ret[0].(uint32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ParseSegmentIdx indicates an expected call of ParseSegmentIdx.
func (mr *MockPieceOpMockRecorder) ParseSegmentIdx(segmentKey interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseSegmentIdx", reflect.TypeOf((*MockPieceOp)(nil).ParseSegmentIdx), segmentKey)
}

// SegmentPieceCount mocks base method.
func (m *MockPieceOp) SegmentPieceCount(payloadSize, maxSegmentSize uint64) uint32 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SegmentPieceCount", payloadSize, maxSegmentSize)
	ret0, _ := ret[0].(uint32)
	return ret0
}

// SegmentPieceCount indicates an expected call of SegmentPieceCount.
func (mr *MockPieceOpMockRecorder) SegmentPieceCount(payloadSize, maxSegmentSize interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SegmentPieceCount", reflect.TypeOf((*MockPieceOp)(nil).SegmentPieceCount), payloadSize, maxSegmentSize)
}

// SegmentPieceKey mocks base method.
func (m *MockPieceOp) SegmentPieceKey(objectID uint64, segmentIdx uint32, version int64) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SegmentPieceKey", objectID, segmentIdx, version)
	ret0, _ := ret[0].(string)
	return ret0
}

// SegmentPieceKey indicates an expected call of SegmentPieceKey.
func (mr *MockPieceOpMockRecorder) SegmentPieceKey(objectID, segmentIdx, version interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SegmentPieceKey", reflect.TypeOf((*MockPieceOp)(nil).SegmentPieceKey), objectID, segmentIdx, version)
}

// SegmentPieceSize mocks base method.
func (m *MockPieceOp) SegmentPieceSize(payloadSize uint64, segmentIdx uint32, maxSegmentSize uint64) int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SegmentPieceSize", payloadSize, segmentIdx, maxSegmentSize)
	ret0, _ := ret[0].(int64)
	return ret0
}

// SegmentPieceSize indicates an expected call of SegmentPieceSize.
func (mr *MockPieceOpMockRecorder) SegmentPieceSize(payloadSize, segmentIdx, maxSegmentSize interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SegmentPieceSize", reflect.TypeOf((*MockPieceOp)(nil).SegmentPieceSize), payloadSize, segmentIdx, maxSegmentSize)
}

// MockPieceStore is a mock of PieceStore interface.
type MockPieceStore struct {
	ctrl     *gomock.Controller
	recorder *MockPieceStoreMockRecorder
}

// MockPieceStoreMockRecorder is the mock recorder for MockPieceStore.
type MockPieceStoreMockRecorder struct {
	mock *MockPieceStore
}

// NewMockPieceStore creates a new mock instance.
func NewMockPieceStore(ctrl *gomock.Controller) *MockPieceStore {
	mock := &MockPieceStore{ctrl: ctrl}
	mock.recorder = &MockPieceStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPieceStore) EXPECT() *MockPieceStoreMockRecorder {
	return m.recorder
}

// DeletePiece mocks base method.
func (m *MockPieceStore) DeletePiece(ctx context.Context, key string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletePiece", ctx, key)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletePiece indicates an expected call of DeletePiece.
func (mr *MockPieceStoreMockRecorder) DeletePiece(ctx, key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePiece", reflect.TypeOf((*MockPieceStore)(nil).DeletePiece), ctx, key)
}

// GetPiece mocks base method.
func (m *MockPieceStore) GetPiece(ctx context.Context, key string, offset, limit int64) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPiece", ctx, key, offset, limit)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPiece indicates an expected call of GetPiece.
func (mr *MockPieceStoreMockRecorder) GetPiece(ctx, key, offset, limit interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPiece", reflect.TypeOf((*MockPieceStore)(nil).GetPiece), ctx, key, offset, limit)
}

// PutPiece mocks base method.
func (m *MockPieceStore) PutPiece(ctx context.Context, key string, value []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutPiece", ctx, key, value)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutPiece indicates an expected call of PutPiece.
func (mr *MockPieceStoreMockRecorder) PutPiece(ctx, key, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutPiece", reflect.TypeOf((*MockPieceStore)(nil).PutPiece), ctx, key, value)
}

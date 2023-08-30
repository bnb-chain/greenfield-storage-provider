package gfspclient

import (
	"context"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func TestGfSpClient_UploadObject(t *testing.T) {
	cases := []struct {
		name        string
		task        coretask.UploadObjectTask
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name: "success",
			task: &gfsptask.GfSpUploadObjectTask{
				Task:       &gfsptask.GfSpTask{TaskPriority: 1},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName3}},
			wantedIsErr: false,
		},
		{
			name: "mock rpc error",
			task: &gfsptask.GfSpUploadObjectTask{
				Task:       &gfsptask.GfSpTask{TaskPriority: 1},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName1}},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name: "mock response returns error",
			task: &gfsptask.GfSpUploadObjectTask{
				Task:       &gfsptask.GfSpTask{TaskPriority: 1},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName2}},
			wantedIsErr: true,
			wantedErr:   ErrNoSuchObject,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			defer s.Close()
			err := s.UploadObject(context.TODO(), tt.task, strings.NewReader(mockTxHash), grpc.WithContextDialer(bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGfSpClient_UploadObjectSuccess(t *testing.T) {
	t.Log("Success case description: mock io.Reader")
	ctrl := gomock.NewController(t)
	m := NewMockIOReader(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(10, io.EOF).AnyTimes()
	s := mockBufClient()
	ta := &gfsptask.GfSpUploadObjectTask{
		Task:       &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName3}}
	err := s.UploadObject(context.TODO(), ta, m, grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.Nil(t, err)
}

func TestGfSpClient_UploadObjectFailure1(t *testing.T) {
	t.Log("Failure case description: client failed to connect uploader")
	ctx, cancel := context.WithCancel(context.TODO())
	s := mockBufClient()
	defer s.Close()
	cancel()
	err := s.UploadObject(ctx, &gfsptask.GfSpUploadObjectTask{}, nil)
	assert.Contains(t, err.Error(), context.Canceled.Error())
}

func TestGfSpClient_UploadObjectFailure2(t *testing.T) {
	t.Log("Failure case description: mock io.Reader returns error")
	ctrl := gomock.NewController(t)
	m := NewMockIOReader(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(0, mockRPCErr).AnyTimes()
	s := mockBufClient()
	ta := &gfsptask.GfSpUploadObjectTask{
		Task:       &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName3}}
	err := s.UploadObject(context.TODO(), ta, m, grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.Equal(t, ErrExceptionsStream, err)
}

func TestGfSpClient_ResumableUploadObject(t *testing.T) {
	cases := []struct {
		name        string
		task        coretask.ResumableUploadObjectTask
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name: "success",
			task: &gfsptask.GfSpResumableUploadObjectTask{
				Task:       &gfsptask.GfSpTask{TaskPriority: 1},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName3}},
			wantedIsErr: false,
		},
		{
			name: "mock rpc error",
			task: &gfsptask.GfSpResumableUploadObjectTask{
				Task:       &gfsptask.GfSpTask{TaskPriority: 1},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName1}},
			wantedIsErr: true,
			wantedErr:   mockRPCErr,
		},
		{
			name: "mock response returns error",
			task: &gfsptask.GfSpResumableUploadObjectTask{
				Task:       &gfsptask.GfSpTask{TaskPriority: 1},
				ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName2}},
			wantedIsErr: true,
			wantedErr:   ErrNoSuchObject,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			defer s.Close()
			err := s.ResumableUploadObject(context.TODO(), tt.task, strings.NewReader(mockTxHash), grpc.WithContextDialer(bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGfSpClient_ResumableUploadObjectSuccess(t *testing.T) {
	t.Log("Success case description: mock io.Reader")
	ctrl := gomock.NewController(t)
	m := NewMockIOReader(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(10, io.EOF).AnyTimes()
	s := mockBufClient()
	ta := &gfsptask.GfSpResumableUploadObjectTask{
		Task:       &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName3}}
	err := s.ResumableUploadObject(context.TODO(), ta, m, grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.Nil(t, err)
}

func TestGfSpClient_ResumableUploadObjectFailure1(t *testing.T) {
	t.Log("Failure case description: client failed to connect uploader")
	ctx, cancel := context.WithCancel(context.TODO())
	s := mockBufClient()
	defer s.Close()
	cancel()
	err := s.ResumableUploadObject(ctx, &gfsptask.GfSpResumableUploadObjectTask{}, nil)
	assert.Contains(t, err.Error(), context.Canceled.Error())
}

func TestGfSpClient_ResumableUploadObjectFailure2(t *testing.T) {
	t.Log("Failure case description: mock io.Reader returns error")
	ctrl := gomock.NewController(t)
	m := NewMockIOReader(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(0, mockRPCErr).AnyTimes()
	s := mockBufClient()
	ta := &gfsptask.GfSpResumableUploadObjectTask{
		Task:       &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName3}}
	err := s.ResumableUploadObject(context.TODO(), ta, m, grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.Equal(t, ErrExceptionsStream, err)
}

type MockIOReader struct {
	ctrl     *gomock.Controller
	recorder *MockIOReaderMockRecorder
}

// MockIOReaderMockRecorder is the mock recorder for MockIOReader.
type MockIOReaderMockRecorder struct {
	mock *MockIOReader
}

// NewMockGfSpClientAPI creates a new mock instance.
func NewMockIOReader(ctrl *gomock.Controller) *MockIOReader {
	mock := &MockIOReader{ctrl: ctrl}
	mock.recorder = &MockIOReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIOReader) EXPECT() *MockIOReaderMockRecorder {
	return m.recorder
}

// Read mocks base method.
func (m *MockIOReader) Read(p []byte) (n int, err error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", p)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockIOReaderMockRecorder) Read(p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockIOReader)(nil).Read), p)
}

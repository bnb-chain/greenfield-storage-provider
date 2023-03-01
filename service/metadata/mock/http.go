package mock

import (
	"net/http"
	"reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockIHTTPClient is a mock of IHTTPClient interface.
type MockIHTTPClient struct {
	ctrl     *gomock.Controller
	recorder *MockIHTTPClientMockRecorder
}

// MockIHTTPClientMockRecorder is the mock recorder for MockIHTTPClient.
type MockIHTTPClientMockRecorder struct {
	mock *MockIHTTPClient
}

// NewMockIHTTPClient creates a new mock instance.
func NewMockIHTTPClient(ctrl *gomock.Controller) *MockIHTTPClient {
	mock := &MockIHTTPClient{ctrl: ctrl}
	mock.recorder = &MockIHTTPClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIHTTPClient) EXPECT() *MockIHTTPClientMockRecorder {
	return m.recorder
}

// Do mocks base method.
func (m *MockIHTTPClient) Do(arg0 *http.Request) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Do", arg0)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Do indicates an expected call of Do.
func (mr *MockIHTTPClientMockRecorder) Do(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Do", reflect.TypeOf((*MockIHTTPClient)(nil).Do), arg0)
}

// Get mocks base method.
func (m *MockIHTTPClient) Get(arg0 string) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockIHTTPClientMockRecorder) Get(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockIHTTPClient)(nil).Get), arg0)
}

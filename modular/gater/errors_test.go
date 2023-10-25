package gater

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrEncodeResponseWithDetail(t *testing.T) {
	err := ErrEncodeResponseWithDetail("mock")
	assert.NotNil(t, err)
}

func TestErrMigrateApprovalWithDetail(t *testing.T) {
	err := ErrMigrateApprovalWithDetail("mock")
	assert.NotNil(t, err)
}

func TestErrNotifySwapOutWithDetail(t *testing.T) {
	err := ErrNotifySwapOutWithDetail("mock")
	assert.NotNil(t, err)
}

func TestErrConsensusWithDetail(t *testing.T) {
	err := ErrConsensusWithDetail("mock")
	assert.NotNil(t, err)
}

type mockResponseWriter struct{ name string }

func (mockResponseWriter) Header() http.Header { return map[string][]string{} }

func (m mockResponseWriter) Write([]byte) (int, error) {
	if m.name == "1" {
		return 0, mockErr
	}
	return 1, nil
}

func (mockResponseWriter) WriteHeader(statusCode int) {}

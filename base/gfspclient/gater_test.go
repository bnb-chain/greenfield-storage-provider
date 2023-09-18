package gfspclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func TestGfSpClient_ReplicatePieceToSecondary(t *testing.T) {
	cases := []struct {
		name         string
		server       *httptest.Server
		endpoint     string
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name:        "success",
			server:      httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})),
			wantedIsErr: false,
		},
		{
			name:         "failure 1",
			server:       nil,
			endpoint:     "\n\r",
			wantedIsErr:  true,
			wantedErrStr: "net/url: invalid control character in URL",
		},
		{
			name:         "failure 2",
			server:       nil,
			endpoint:     emptyString,
			wantedIsErr:  true,
			wantedErrStr: "unsupported protocol scheme",
		},
		{
			name: "failure 3",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			})),
			wantedIsErr:  true,
			wantedErrStr: "failed to replicate piece, status_code",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			if tt.server != nil {
				defer tt.server.Close()
				err := s.ReplicatePieceToSecondary(context.TODO(), tt.server.URL, &gfsptask.GfSpReceivePieceTask{}, mockSignature)
				if tt.wantedIsErr {
					assert.Contains(t, err.Error(), tt.wantedErrStr)
				} else {
					assert.Nil(t, err)
				}
			} else {
				err := s.ReplicatePieceToSecondary(context.TODO(), tt.endpoint, &gfsptask.GfSpReceivePieceTask{}, mockSignature)
				assert.Contains(t, err.Error(), tt.wantedErrStr)
			}
		})
	}
}

func TestGfSpClient_GetPieceFromECChunks(t *testing.T) {
	cases := []struct {
		name         string
		server       *httptest.Server
		endpoint     string
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name:        "success",
			server:      httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})),
			wantedIsErr: false,
		},
		{
			name:         "failure 1",
			server:       nil,
			endpoint:     "\n\r",
			wantedIsErr:  true,
			wantedErrStr: "net/url: invalid control character in URL",
		},
		{
			name:         "failure 2",
			server:       nil,
			endpoint:     emptyString,
			wantedIsErr:  true,
			wantedErrStr: "unsupported protocol scheme",
		},
		{
			name: "failure 3",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			})),
			wantedIsErr:  true,
			wantedErrStr: "failed to get recovery piece, status_code",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			if tt.server != nil {
				defer tt.server.Close()
				result, err := s.GetPieceFromECChunks(context.TODO(), tt.server.URL, &gfsptask.GfSpRecoverPieceTask{})
				if tt.wantedIsErr {
					assert.Contains(t, err.Error(), tt.wantedErrStr)
					assert.Nil(t, result)
				} else {
					assert.Nil(t, err)
					assert.NotNil(t, result)
				}
			} else {
				result, err := s.GetPieceFromECChunks(context.TODO(), tt.endpoint, &gfsptask.GfSpRecoverPieceTask{})
				assert.Contains(t, err.Error(), tt.wantedErrStr)
				assert.Nil(t, result)
			}
		})
	}
}

func TestGfSpClient_DoneReplicatePieceToSecondary(t *testing.T) {
	cases := []struct {
		name         string
		server       *httptest.Server
		endpoint     string
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name:        "success",
			server:      httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})),
			wantedIsErr: false,
		},
		{
			name:         "failure 1",
			server:       nil,
			endpoint:     "\n\r",
			wantedIsErr:  true,
			wantedErrStr: "net/url: invalid control character in URL",
		},
		{
			name:         "failure 2",
			server:       nil,
			endpoint:     emptyString,
			wantedIsErr:  true,
			wantedErrStr: "unsupported protocol scheme",
		},
		{
			name: "failure 3",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			})),
			wantedIsErr:  true,
			wantedErrStr: "failed to done replicate piece, status_code",
		},
		{
			name: "failure 4",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(GnfdIntegrityHashSignatureHeader, mockTxHash)
			})),
			wantedIsErr:  true,
			wantedErrStr: "encoding/hex: invalid byte",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			if tt.server != nil {
				defer tt.server.Close()
				result, err := s.DoneReplicatePieceToSecondary(context.TODO(), tt.server.URL, &gfsptask.GfSpReceivePieceTask{})
				if tt.wantedIsErr {
					assert.Contains(t, err.Error(), tt.wantedErrStr)
					assert.Nil(t, result)
				} else {
					assert.Nil(t, err)
					assert.NotNil(t, result)
				}
			} else {
				result, err := s.DoneReplicatePieceToSecondary(context.TODO(), tt.endpoint, &gfsptask.GfSpReceivePieceTask{})
				assert.Contains(t, err.Error(), tt.wantedErrStr)
				assert.Nil(t, result)
			}
		})
	}
}

func TestGfSpClient_MigratePiece(t *testing.T) {
	cases := []struct {
		name         string
		server       *httptest.Server
		endpoint     string
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name: "success",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(mockSignature)
			})),
			wantedIsErr: false,
		},
		{
			name:         "failure 1",
			server:       nil,
			endpoint:     "\r\n",
			wantedIsErr:  true,
			wantedErrStr: "net/url: invalid control character in URL",
		},
		{
			name:         "failure 2",
			server:       nil,
			endpoint:     "",
			wantedIsErr:  true,
			wantedErrStr: "unsupported protocol scheme",
		},
		{
			name: "failure 3",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			})),
			wantedIsErr:  true,
			wantedErrStr: "failed to migrate pieces, status_code",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			if tt.server != nil {
				defer tt.server.Close()
				result, err := s.MigratePiece(context.TODO(), &gfsptask.GfSpMigrateGVGTask{}, &gfsptask.GfSpMigratePieceTask{SrcSpEndpoint: tt.server.URL})
				if tt.wantedIsErr {
					assert.Contains(t, err.Error(), tt.wantedErrStr)
					assert.Nil(t, result)
				} else {
					assert.Nil(t, err)
					assert.NotNil(t, result)
				}
			} else {
				result, err := s.MigratePiece(context.TODO(), &gfsptask.GfSpMigrateGVGTask{}, &gfsptask.GfSpMigratePieceTask{SrcSpEndpoint: tt.endpoint})
				assert.Contains(t, err.Error(), tt.wantedErrStr)
				assert.Nil(t, result)
			}
		})
	}
}

func TestGfSpClient_NotifyDestSPMigrateSwapOut(t *testing.T) {
	cases := []struct {
		name         string
		server       *httptest.Server
		endpoint     string
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name:        "success",
			server:      httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})),
			wantedIsErr: false,
		},
		{
			name:         "failure 1",
			server:       nil,
			endpoint:     "\r\n",
			wantedIsErr:  true,
			wantedErrStr: "net/url: invalid control character in URL",
		},
		{
			name:         "failure 2",
			server:       nil,
			endpoint:     "",
			wantedIsErr:  true,
			wantedErrStr: "unsupported protocol scheme",
		},
		{
			name: "failure 3",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			})),
			wantedIsErr:  true,
			wantedErrStr: "failed to notify migrate swap out, status_code",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			if tt.server != nil {
				defer tt.server.Close()
				err := s.NotifyDestSPMigrateSwapOut(context.TODO(), tt.server.URL, &virtualgrouptypes.MsgSwapOut{})
				if tt.wantedIsErr {
					assert.Contains(t, err.Error(), tt.wantedErrStr)
				} else {
					assert.Nil(t, err)
				}
			} else {
				err := s.NotifyDestSPMigrateSwapOut(context.TODO(), tt.endpoint, &virtualgrouptypes.MsgSwapOut{})
				assert.Contains(t, err.Error(), tt.wantedErrStr)
			}
		})
	}
}

func TestGfSpClient_GetSecondarySPMigrationBucketApproval(t *testing.T) {
	cases := []struct {
		name         string
		server       *httptest.Server
		endpoint     string
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name:        "success",
			server:      httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})),
			wantedIsErr: false,
		},
		{
			name:         "failure 1",
			server:       nil,
			endpoint:     "\r\n",
			wantedIsErr:  true,
			wantedErrStr: "net/url: invalid control character in URL",
		},
		{
			name:         "failure 2",
			server:       nil,
			endpoint:     "",
			wantedIsErr:  true,
			wantedErrStr: "unsupported protocol scheme",
		},
		{
			name: "failure 3",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			})),
			wantedIsErr:  true,
			wantedErrStr: "failed to get resp body, status_code",
		},
		{
			name: "failure 4",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(GnfdSecondarySPMigrationBucketApprovalHeader, mockTxHash)
			})),
			wantedIsErr:  true,
			wantedErrStr: "encoding/hex: invalid byte",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			if tt.server != nil {
				defer tt.server.Close()
				result, err := s.GetSecondarySPMigrationBucketApproval(context.TODO(), tt.server.URL, &storagetypes.SecondarySpMigrationBucketSignDoc{})
				if tt.wantedIsErr {
					assert.Contains(t, err.Error(), tt.wantedErrStr)
					assert.Nil(t, result)
				} else {
					assert.Nil(t, err)
				}
			} else {
				result, err := s.GetSecondarySPMigrationBucketApproval(context.TODO(), tt.endpoint, &storagetypes.SecondarySpMigrationBucketSignDoc{})
				assert.Contains(t, err.Error(), tt.wantedErrStr)
				assert.Nil(t, result)
			}
		})
	}
}

func TestGfSpClient_GetSwapOutApproval(t *testing.T) {
	cases := []struct {
		name         string
		server       *httptest.Server
		endpoint     string
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name: "success",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(GnfdSignedApprovalMsgHeader, "7b2273746f726167655f70726f7669646572223a226d6f636b547848617368222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22676c6f62616c5f7669727475616c5f67726f75705f696473223a5b5d2c22737563636573736f725f73705f6964223a302c22737563636573736f725f73705f617070726f76616c223a6e756c6c7d")
			})),
			wantedIsErr: false,
		},
		{
			name:         "failure 1",
			server:       nil,
			endpoint:     "\r\n",
			wantedIsErr:  true,
			wantedErrStr: "net/url: invalid control character in URL",
		},
		{
			name:         "failure 2",
			server:       nil,
			endpoint:     "",
			wantedIsErr:  true,
			wantedErrStr: "unsupported protocol scheme",
		},
		{
			name: "failure 3",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			})),
			wantedIsErr:  true,
			wantedErrStr: "failed to get resp body, statue_code",
		},
		{
			name: "failure 4",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(GnfdSignedApprovalMsgHeader, mockTxHash)
			})),
			wantedIsErr:  true,
			wantedErrStr: "encoding/hex: invalid byte",
		},
		{
			name: "failure 5",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(GnfdSignedApprovalMsgHeader, "48656c6c6f20476f7068657221")
			})),
			wantedIsErr:  true,
			wantedErrStr: "invalid character 'H'",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			if tt.server != nil {
				defer tt.server.Close()
				result, err := s.GetSwapOutApproval(context.TODO(), tt.server.URL, &virtualgrouptypes.MsgSwapOut{})
				if tt.wantedIsErr {
					assert.Contains(t, err.Error(), tt.wantedErrStr)
					assert.Nil(t, result)
				} else {
					assert.Nil(t, err)
					assert.NotNil(t, result)
				}
			} else {
				result, err := s.GetSwapOutApproval(context.TODO(), tt.endpoint, &virtualgrouptypes.MsgSwapOut{})
				assert.Contains(t, err.Error(), tt.wantedErrStr)
				assert.Nil(t, result)
			}
		})
	}
}

package gfspclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

func TestGfSpClient_VerifyAuthentication(t *testing.T) {
	cases := []struct {
		name         string
		auth         coremodule.AuthOpType
		wantedResult bool
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name:         "success",
			auth:         coremodule.AuthOpAskCreateBucketApproval,
			wantedResult: true,
			wantedIsErr:  false,
		},
		{
			name:         "mock rpc error",
			auth:         coremodule.AuthOpAskMigrateBucketApproval,
			wantedResult: false,
			wantedIsErr:  true,
			wantedErr:    mockRPCErr,
		},
		{
			name:         "mock response returns error",
			auth:         coremodule.AuthOpAskCreateObjectApproval,
			wantedResult: false,
			wantedIsErr:  true,
			wantedErr:    ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			ctx := context.Background()
			result, err := s.VerifyAuthentication(ctx, tt.auth, emptyString, emptyString, emptyString, grpc.WithContextDialer(bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Equal(t, tt.wantedResult, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantedResult, result)
			}
		})
	}
}

func TestGfSpClient_VerifyAuthenticationFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect authenticator")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.VerifyAuthentication(ctx, coremodule.AuthOpTypeUnKnown, emptyString, emptyString, emptyString)
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Equal(t, false, result)
}

func TestGfSpClient_GetAuthNonce(t *testing.T) {
	cases := []struct {
		name         string
		account      string
		wantedResult int32
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name:         "success",
			account:      mockObjectName3,
			wantedResult: 1,
			wantedIsErr:  false,
		},
		{
			name:         "mock rpc error",
			account:      mockObjectName1,
			wantedResult: 0,
			wantedIsErr:  true,
			wantedErr:    mockRPCErr,
		},
		{
			name:         "mock response returns error",
			account:      mockObjectName2,
			wantedResult: 0,
			wantedIsErr:  true,
			wantedErr:    ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			ctx := context.Background()
			result, _, _, _, err := s.GetAuthNonce(ctx, tt.account, emptyString, grpc.WithContextDialer(bufDialer),
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Equal(t, tt.wantedResult, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantedResult, result)
			}
		})
	}
}

func TestGfSpClient_GetAuthNonceFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect authenticator")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, _, _, _, err := s.GetAuthNonce(ctx, emptyString, emptyString)
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Equal(t, int32(0), result)
}

func TestGfSpClient_UpdateUserPublicKey(t *testing.T) {
	cases := []struct {
		name         string
		account      string
		wantedResult bool
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name:         "success",
			account:      mockObjectName3,
			wantedResult: true,
			wantedIsErr:  false,
		},
		{
			name:         "mock rpc error",
			account:      mockObjectName1,
			wantedResult: false,
			wantedIsErr:  true,
			wantedErr:    mockRPCErr,
		},
		{
			name:         "mock response returns error",
			account:      mockObjectName2,
			wantedResult: false,
			wantedIsErr:  true,
			wantedErr:    ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			ctx := context.Background()
			result, err := s.UpdateUserPublicKey(ctx, tt.account, emptyString, 0, 0, emptyString, 0,
				grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Equal(t, tt.wantedResult, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantedResult, result)
			}
		})
	}
}

func TestGfSpClient_UpdateUserPublicKeyFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect authenticator")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.UpdateUserPublicKey(ctx, emptyString, emptyString, 0, 0, emptyString, 0)
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Equal(t, false, result)
}

func TestGfSpClient_VerifyGNFD1EddsaSignature(t *testing.T) {
	cases := []struct {
		name         string
		account      string
		wantedResult bool
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name:         "success",
			account:      mockObjectName3,
			wantedResult: true,
			wantedIsErr:  false,
		},
		{
			name:         "mock rpc error",
			account:      mockObjectName1,
			wantedResult: false,
			wantedIsErr:  true,
			wantedErr:    mockRPCErr,
		},
		{
			name:         "mock response returns error",
			account:      mockObjectName2,
			wantedResult: false,
			wantedIsErr:  true,
			wantedErr:    ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			ctx := context.Background()
			result, err := s.VerifyGNFD1EddsaSignature(ctx, tt.account, emptyString, emptyString, nil,
				grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Equal(t, tt.wantedResult, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantedResult, result)
			}
		})
	}
}

func TestGfSpClient_VerifyGNFD1EddsaSignatureFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect authenticator")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.VerifyGNFD1EddsaSignature(ctx, emptyString, emptyString, emptyString, nil)
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Equal(t, false, result)
}

func TestGfSpClient_GetAuthKeyV2(t *testing.T) {
	cases := []struct {
		name             string
		account          string
		wantedPublicKey  string
		wantedExpiryDate int64
		wantedIsErr      bool
		wantedErr        error
	}{
		{
			name:             "success",
			account:          mockObjectName3,
			wantedPublicKey:  "test",
			wantedExpiryDate: 0,
			wantedIsErr:      false,
		},
		{
			name:             "mock rpc error",
			account:          mockObjectName1,
			wantedPublicKey:  "",
			wantedExpiryDate: 0,
			wantedIsErr:      true,
			wantedErr:        mockRPCErr,
		},
		{
			name:             "mock response returns error",
			account:          mockObjectName2,
			wantedPublicKey:  "",
			wantedExpiryDate: 0,
			wantedIsErr:      true,
			wantedErr:        ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			ctx := context.Background()
			publicKey, expiryDate, err := s.GetAuthKeyV2(ctx, tt.account, emptyString, emptyString, grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Equal(t, tt.wantedExpiryDate, expiryDate)
				assert.Equal(t, tt.wantedPublicKey, publicKey)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantedExpiryDate, expiryDate)
				assert.Equal(t, tt.wantedPublicKey, publicKey)

			}
		})
	}
}

func TestGfSpClient_GetAuthKeyV2Failure(t *testing.T) {
	t.Log("Failure case description: client failed to connect authenticator")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	publicKey, expiryDate, err := s.GetAuthKeyV2(ctx, emptyString, emptyString, emptyString)
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Equal(t, int64(0), expiryDate)
	assert.Equal(t, "", publicKey)
}

func TestGfSpClient_UpdateUserPublicKeyV2(t *testing.T) {
	cases := []struct {
		name            string
		account         string
		wantedPublicKey string
		wantedResult    bool
		wantedIsErr     bool
		wantedErr       error
	}{
		{
			name:            "success",
			account:         mockObjectName3,
			wantedPublicKey: "test",
			wantedIsErr:     false,
			wantedResult:    true,
		},
		{
			name:            "mock rpc error",
			account:         mockObjectName1,
			wantedPublicKey: "",
			wantedResult:    false,
			wantedIsErr:     true,
			wantedErr:       mockRPCErr,
		},
		{
			name:            "mock response returns error",
			account:         mockObjectName2,
			wantedPublicKey: "",
			wantedResult:    false,
			wantedIsErr:     true,
			wantedErr:       ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			ctx := context.Background()
			result, err := s.UpdateUserPublicKeyV2(ctx, tt.account, emptyString, emptyString, 0, grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Equal(t, tt.wantedResult, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantedResult, result)
			}
		})
	}
}

func TestGfSpClient_UpdateUserPublicKeyV2Failure(t *testing.T) {
	t.Log("Failure case description: client failed to connect authenticator")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.UpdateUserPublicKeyV2(ctx, emptyString, emptyString, emptyString, 0)
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Equal(t, false, result)
}

func TestGfSpClient_VerifyGNFD2EddsaSignature(t *testing.T) {
	cases := []struct {
		name            string
		account         string
		wantedPublicKey string
		wantedResult    bool
		wantedIsErr     bool
		wantedErr       error
	}{
		{
			name:            "success",
			account:         mockObjectName3,
			wantedPublicKey: "test",
			wantedIsErr:     false,
			wantedResult:    true,
		},
		{
			name:            "mock rpc error",
			account:         mockObjectName1,
			wantedPublicKey: "",
			wantedResult:    false,
			wantedIsErr:     true,
			wantedErr:       mockRPCErr,
		},
		{
			name:            "mock response returns error",
			account:         mockObjectName2,
			wantedPublicKey: "",
			wantedResult:    false,
			wantedIsErr:     true,
			wantedErr:       ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			ctx := context.Background()
			result, err := s.VerifyGNFD2EddsaSignature(ctx, tt.account, emptyString, emptyString, emptyString, nil, grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Equal(t, tt.wantedResult, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantedResult, result)
			}
		})
	}
}

func TestGfSpClient_VerifyGNFD2EddsaSignatureFailure(t *testing.T) {
	t.Log("Failure case description: client failed to connect authenticator")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.VerifyGNFD2EddsaSignature(ctx, emptyString, emptyString, emptyString, emptyString, nil)
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Equal(t, false, result)
}

func TestGfSpClient_DeleteAuthKeysV2(t *testing.T) {
	cases := []struct {
		name            string
		account         string
		wantedPublicKey string
		wantedResult    bool
		wantedIsErr     bool
		wantedErr       error
	}{
		{
			name:            "success",
			account:         mockObjectName3,
			wantedPublicKey: "test",
			wantedIsErr:     false,
			wantedResult:    true,
		},
		{
			name:            "mock rpc error",
			account:         mockObjectName1,
			wantedPublicKey: "",
			wantedResult:    false,
			wantedIsErr:     true,
			wantedErr:       mockRPCErr,
		},
		{
			name:            "mock response returns error",
			account:         mockObjectName2,
			wantedPublicKey: "",
			wantedResult:    false,
			wantedIsErr:     true,
			wantedErr:       ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			ctx := context.Background()
			result, err := s.DeleteAuthKeysV2(ctx, tt.account, emptyString, []string{}, grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Equal(t, tt.wantedResult, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantedResult, result)
			}
		})
	}
}

func TestGfSpClient_DeleteAuthKeysV2Failure(t *testing.T) {
	t.Log("Failure case description: client failed to connect authenticator")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.DeleteAuthKeysV2(ctx, emptyString, emptyString, []string{})
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Equal(t, false, result)
}

func TestGfSpClient_ListAuthKeysV2(t *testing.T) {
	cases := []struct {
		name         string
		account      string
		wantedResult []string
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name:         "success",
			account:      mockObjectName3,
			wantedResult: []string{"key1", "key2"},
			wantedIsErr:  false,
		},
		{
			name:         "mock rpc error",
			account:      mockObjectName1,
			wantedResult: nil,
			wantedIsErr:  true,
			wantedErr:    mockRPCErr,
		},
		{
			name:         "mock response returns error",
			account:      mockObjectName2,
			wantedResult: nil,
			wantedIsErr:  true,
			wantedErr:    ErrExceptionsStream,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := mockBufClient()
			ctx := context.Background()
			keys, err := s.ListAuthKeysV2(ctx, tt.account, emptyString, grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Equal(t, tt.wantedResult, []string(nil))
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantedResult, keys)
			}
		})
	}
}

func TestGfSpClient_ListAuthKeysV2Failure(t *testing.T) {
	t.Log("Failure case description: client failed to connect authenticator")
	ctx, cancel := context.WithCancel(context.Background())
	s := mockBufClient()
	defer s.Close()
	cancel()
	result, err := s.ListAuthKeysV2(ctx, emptyString, emptyString)
	assert.Contains(t, err.Error(), context.Canceled.Error())
	assert.Equal(t, []string(nil), result)
}

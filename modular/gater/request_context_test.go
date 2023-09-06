package gater

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
)

func TestRequestContext_SetHTTPCode(t *testing.T) {
	reqCtx := &RequestContext{g: setup(t)}
	reqCtx.SetHTTPCode(http.StatusOK)
}

func TestRequestContext_VerifySignature(t *testing.T) {
	cases := []struct {
		name        string
		fn          func() *GateModular
		request     func() *http.Request
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name: "verifySignatureForGNFD1Ecdsa returns error",
			fn:   func() *GateModular { return setup(t) },
			request: func() *http.Request {
				req := &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Scheme: scheme,
						Host:   testDomain,
						Path:   AuthRequestNoncePath,
					},
					Header: map[string][]string{},
				}
				req.Header.Add(GnfdAuthorizationHeader, "GNFD1-ECDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedIsErr: true,
			wantedErr:   ErrRequestConsistent,
		},
		{
			name: "verifySignatureForGNFD1Eddsa returns error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, mockErr).Times(1)
				g.baseApp.SetGfSpClient(m)
				return g
			},
			request: func() *http.Request {
				req := &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Scheme: scheme,
						Host:   testDomain,
						Path:   AuthRequestNoncePath,
					},
					Header: map[string][]string{},
				}
				req.Header.Add(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedIsErr: true,
			wantedErr:   mockErr,
		},
		{
			name: "verifySignatureForGNFD1Eddsa no error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				g.baseApp.SetGfSpClient(m)
				return g
			},
			request: func() *http.Request {
				req := &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Scheme: scheme,
						Host:   testDomain,
						Path:   AuthRequestNoncePath,
					},
					Header: map[string][]string{},
				}
				req.Header.Add(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedIsErr: false,
			wantedErr:   nil,
		},
		{
			name: "unsupported sign type",
			fn:   func() *GateModular { return setup(t) },
			request: func() *http.Request {
				req := &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Scheme: scheme,
						Host:   testDomain,
						Path:   AuthRequestNoncePath,
					},
					Header: map[string][]string{},
				}
				req.Header.Add(GnfdAuthorizationHeader, "Unsupported,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedIsErr: true,
			wantedErr:   ErrUnsupportedSignType,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			reqCtx := &RequestContext{g: tt.fn(), request: tt.request()}
			result, err := reqCtx.VerifySignature()
			if tt.wantedIsErr {
				assert.Equal(t, tt.wantedErr, err)
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.Empty(t, result)
			}
		})
	}
}

func TestRequestContext_verifySignatureForGNFD1Ecdsa(t *testing.T) {
	cases := []struct {
		name             string
		requestSignature string
		reqCtx           func() *RequestContext
		wantedErrStr     string
	}{
		{
			name:             "failed to parseSignatureFromRequest",
			requestSignature: "Signature",
			reqCtx:           func() *RequestContext { return &RequestContext{g: setup(t)} },
			wantedErrStr:     ErrAuthorizationHeaderFormat.Error(),
		},
		{
			name:             "failed to hex decode string",
			requestSignature: "Signature=1ba5c6d",
			reqCtx:           func() *RequestContext { return &RequestContext{g: setup(t)} },
			wantedErrStr:     "encoding/hex: odd length hex string",
		},
		{
			name:             "failed to hex decode string",
			requestSignature: "Signature=48656c6c6f20476f7068657221",
			reqCtx: func() *RequestContext {
				return &RequestContext{
					g: setup(t),
					request: &http.Request{
						Method: http.MethodGet,
						URL: &url.URL{
							Scheme: scheme,
							Host:   testDomain,
							Path:   AuthRequestNoncePath,
						},
					},
				}
			},
			wantedErrStr: ErrRequestConsistent.Error(),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.reqCtx().verifySignatureForGNFD1Ecdsa(tt.requestSignature)
			assert.Contains(t, err.Error(), tt.wantedErrStr)
			assert.Nil(t, result)
		})
	}
}

func TestRequestContext_verifyTaskSignature1(t *testing.T) {
	reqCtx := &RequestContext{g: setup(t), request: &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: scheme,
			Host:   testDomain,
			Path:   AuthRequestNoncePath,
		},
	}}
	result, err := reqCtx.verifyTaskSignature([]byte("test"), []byte("mock"))
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestRequestContext_verifySignatureForGNFD1Eddsa(t *testing.T) {
	cases := []struct {
		name             string
		fn               func() *RequestContext
		requestSignature string
		account          string
		domain           string
		wantedIsErr      bool
		wantedErr        error
	}{
		{
			name:             "parseSignatureFromRequest returns error",
			fn:               func() *RequestContext { return &RequestContext{g: setup(t)} },
			requestSignature: "Signature",
			wantedIsErr:      true,
			wantedErr:        ErrAuthorizationHeaderFormat,
		},
		{
			name: "failed to verify signature for GNFD1-Eddsa",
			fn: func() *RequestContext {
				g := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, mockErr).Times(1)
				g.baseApp.SetGfSpClient(m)
				reqCtx := &RequestContext{g: g, request: &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Scheme: scheme,
						Host:   testDomain,
						Path:   AuthRequestNoncePath,
					},
				}}
				return reqCtx
			},
			requestSignature: "Signature=1a8b6fe754d",
			wantedIsErr:      true,
			wantedErr:        mockErr,
		},
		{
			name: "success",
			fn: func() *RequestContext {
				g := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				g.baseApp.SetGfSpClient(m)
				reqCtx := &RequestContext{g: g, request: &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Scheme: scheme,
						Host:   testDomain,
						Path:   AuthRequestNoncePath,
					},
				}}
				return reqCtx
			},
			requestSignature: "Signature=1a8b6fe754d",
			account:          "test",
			wantedIsErr:      false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn().verifySignatureForGNFD1Eddsa(tt.requestSignature)
			if tt.wantedIsErr {
				assert.Equal(t, tt.wantedErr, err)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestRequestContext_verifyGNFD1EddsaSignatureFromPreSignedURL(t *testing.T) {
	cases := []struct {
		name              string
		fn                func() *RequestContext
		authenticationStr string
		account           string
		domain            string
		wantedIsErr       bool
		wantedErr         error
	}{
		{
			name: "parseSignatureFromRequest returns error",
			fn: func() *RequestContext {
				return &RequestContext{g: setup(t)}
			},
			authenticationStr: "Signature",
			wantedIsErr:       true,
			wantedErr:         ErrAuthorizationHeaderFormat,
		},
		{
			name: "failed to verify off chain signature",
			fn: func() *RequestContext {
				g := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, mockErr).Times(1)
				g.baseApp.SetGfSpClient(m)
				reqCtx := &RequestContext{g: g, request: &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Scheme: scheme,
						Host:   testDomain,
						Path:   AuthRequestNoncePath,
					},
				}}
				return reqCtx
			},
			authenticationStr: "Signature=1a8b6fe754d",
			wantedIsErr:       true,
			wantedErr:         mockErr,
		},
		{
			name: "success",
			fn: func() *RequestContext {
				g := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				g.baseApp.SetGfSpClient(m)
				reqCtx := &RequestContext{g: g, request: &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Scheme: scheme,
						Host:   testDomain,
						Path:   AuthRequestNoncePath,
					},
				}}
				return reqCtx
			},
			authenticationStr: "Signature=1a8b6fe754d",
			account:           "test",
			wantedIsErr:       false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn().verifyGNFD1EddsaSignatureFromPreSignedURL(tt.authenticationStr, tt.account, tt.domain)
			if tt.wantedIsErr {
				assert.Equal(t, tt.wantedErr, err)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func Test_parseSignatureFromRequest(t *testing.T) {
	cases := []struct {
		name         string
		sig          string
		wantedResult string
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name:         "1",
			sig:          "Signature=1a8b6fe754d",
			wantedResult: "1a8b6fe754d",
			wantedIsErr:  false,
		},
		{
			name:         "2",
			sig:          "Signature",
			wantedResult: "",
			wantedIsErr:  true,
			wantedErr:    ErrAuthorizationHeaderFormat,
		},
		{
			name:         "3",
			sig:          "Sig=1a8b6fe754d",
			wantedResult: "",
			wantedIsErr:  true,
			wantedErr:    ErrAuthorizationHeaderFormat,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSignatureFromRequest(tt.sig)
			if tt.wantedIsErr {
				assert.Equal(t, tt.wantedErr, err)
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantedResult, result)
			}
		})
	}
}

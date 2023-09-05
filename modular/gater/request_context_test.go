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

func TestRequestContext_VerifySignature1(t *testing.T) {
	t.Log("Failure case description: verifySignatureForGNFD1Ecdsa returns error")
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
	reqCtx := &RequestContext{g: setup(t), request: req}
	result, err := reqCtx.VerifySignature()
	assert.Equal(t, ErrRequestConsistent, err)
	assert.Empty(t, result)
}

func TestRequestContext_VerifySignature2(t *testing.T) {
	t.Log("Failure case description: verifySignatureForGNFD1Eddsa returns error")
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
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := gfspclient.NewMockGfSpClientAPI(ctrl)
	m.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		false, mockErr).Times(1)
	g.baseApp.SetGfSpClient(m)
	reqCtx := &RequestContext{g: g, request: req}
	result, err := reqCtx.VerifySignature()
	assert.Equal(t, mockErr, err)
	assert.Empty(t, result)
}

func TestRequestContext_VerifySignature3(t *testing.T) {
	t.Log("Failure case description: verifySignatureForGNFD1Eddsa no error")
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
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := gfspclient.NewMockGfSpClientAPI(ctrl)
	m.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		false, nil).Times(1)
	g.baseApp.SetGfSpClient(m)
	reqCtx := &RequestContext{g: g, request: req}
	result, err := reqCtx.VerifySignature()
	assert.Nil(t, err)
	assert.Empty(t, result)
}

func TestRequestContext_VerifySignature4(t *testing.T) {
	t.Log("Failure case description: unsupported sign type")
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
	g := setup(t)
	reqCtx := &RequestContext{g: g, request: req}
	result, err := reqCtx.VerifySignature()
	assert.Equal(t, ErrUnsupportedSignType, err)
	assert.Empty(t, result)
}

func TestRequestContext_verifySignatureForGNFD1Ecdsa1(t *testing.T) {
	reqCtx := &RequestContext{g: setup(t)}
	result, err := reqCtx.verifySignatureForGNFD1Ecdsa("Signature")
	assert.Equal(t, ErrAuthorizationHeaderFormat, err)
	assert.Nil(t, result)
}

func TestRequestContext_verifySignatureForGNFD1Ecdsa2(t *testing.T) {
	reqCtx := &RequestContext{g: setup(t)}
	result, err := reqCtx.verifySignatureForGNFD1Ecdsa("Signature=1ba5c6d")
	assert.Contains(t, err.Error(), "encoding/hex: odd length hex string")
	assert.Nil(t, result)
}

func TestRequestContext_verifySignatureForGNFD1Ecdsa3(t *testing.T) {
	reqCtx := &RequestContext{g: setup(t), request: &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: scheme,
			Host:   testDomain,
			Path:   AuthRequestNoncePath,
		},
	}}
	result, err := reqCtx.verifySignatureForGNFD1Ecdsa("Signature=48656c6c6f20476f7068657221")
	assert.Equal(t, ErrRequestConsistent, err)
	assert.Nil(t, result)
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

func TestRequestContext_verifySignatureForGNFD1Eddsa1(t *testing.T) {
	reqCtx := &RequestContext{g: setup(t)}
	result, err := reqCtx.verifySignatureForGNFD1Eddsa("Signature")
	assert.Equal(t, ErrAuthorizationHeaderFormat, err)
	assert.Nil(t, result)
}

func TestRequestContext_verifySignatureForGNFD1Eddsa2(t *testing.T) {
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
	result, err := reqCtx.verifySignatureForGNFD1Eddsa("Signature=1a8b6fe754d")
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestRequestContext_verifySignatureForGNFD1Eddsa3(t *testing.T) {
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
	result, err := reqCtx.verifySignatureForGNFD1Eddsa("Signature=1a8b6fe754d")
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestRequestContext_verifyGNFD1EddsaSignatureFromPreSignedURL1(t *testing.T) {
	reqCtx := &RequestContext{g: setup(t)}
	result, err := reqCtx.verifyGNFD1EddsaSignatureFromPreSignedURL("Signature", "", "")
	assert.Equal(t, ErrAuthorizationHeaderFormat, err)
	assert.Nil(t, result)
}

func TestRequestContext_verifyGNFD1EddsaSignatureFromPreSignedURL2(t *testing.T) {
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
	result, err := reqCtx.verifyGNFD1EddsaSignatureFromPreSignedURL("Signature=1a8b6fe754d", "", "")
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestRequestContext_verifyGNFD1EddsaSignatureFromPreSignedURL3(t *testing.T) {
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
	result, err := reqCtx.verifyGNFD1EddsaSignatureFromPreSignedURL("Signature=1a8b6fe754d", "test", "")
	assert.Nil(t, err)
	assert.NotNil(t, result)
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

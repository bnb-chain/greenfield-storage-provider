package gater

import (
	"encoding/xml"
	"fmt"
	commonhttp "github.com/bnb-chain/greenfield-common/go/http"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

// Now used for test
var Now = time.Now

const (
	TestSpAddress           string = "0x1c62EF97a13654A759C7E706Adf9EB3bAb0F807A"
	UnsignedContentTemplate string = `%s wants you to sign in with your BNB Greenfield account:
%s
Register your identity public key %s

URI: %s
Version: 1
Chain ID: 5600
Issued At: %s
Expiration Time: %s
Resources:
- SP %s (name: SP_001) with nonce: %s`

	SampleDAppDomain  = "https://greenfield.dapp.sample.io"
	SamplePublicKey   = "9f708a5c45db9800d57bbdfae1202f31a7569290670609b0f38cab4ee62a12a8"
	SampleUserAccount = "0xa64FdC3B4866CD2aC664998C7b180813fB9B06E6"
	SampleNonce       = "123456"
	SampleExpiryDate  = "test_expiry_date"
)

func TestAuthHandlerVerifySignedContent(t *testing.T) {
	gateway := &GateModular{
		env:    gfspapp.EnvLocal,
		domain: testDomain,
	}
	gateway.baseApp = &gfspapp.GfSpBaseApp{}
	gateway.baseApp.SetOperatorAddress(TestSpAddress)

	unSignedContent := fmt.Sprintf(UnsignedContentTemplate, SampleDAppDomain, SampleUserAccount, SamplePublicKey, SampleDAppDomain, SampleExpiryDate, SampleExpiryDate, TestSpAddress, SampleNonce)

	// Test case 1: all inputs are correct.
	assert.Nil(t, gateway.verifySignedContent(unSignedContent, SampleDAppDomain, SampleNonce, SamplePublicKey, SampleExpiryDate))

	// Test case 2: when signed content does not match template
	assert.Equal(t, ErrSignedMsgNotMatchTemplate, gateway.verifySignedContent("wrong-signed-content", SampleDAppDomain, SampleNonce, SamplePublicKey, SampleExpiryDate))

	// Test case 3: when nonce does not match.
	assert.NotNil(t, gateway.verifySignedContent(unSignedContent, SampleDAppDomain, "invalid_nonce", SamplePublicKey, SampleExpiryDate))

	// Test case 4: when signed content is not for this SP.
	unSignedContentForOtherSP := fmt.Sprintf(UnsignedContentTemplate, SampleDAppDomain, SampleUserAccount, SamplePublicKey, SampleDAppDomain, SampleExpiryDate, SampleExpiryDate, "Other_SP_Address", SampleNonce)
	assert.NotNil(t, gateway.verifySignedContent(unSignedContentForOtherSP, SampleDAppDomain, SampleNonce, SamplePublicKey, SampleExpiryDate))

	// Test case 5: when dapp domain does not match
	assert.NotNil(t, gateway.verifySignedContent(unSignedContent, "invalid_dapp_domain", SampleNonce, SamplePublicKey, SampleExpiryDate))

	// Test case 6: when public key does not match
	assert.NotNil(t, gateway.verifySignedContent(unSignedContent, SampleDAppDomain, SampleNonce, "invalid_public_key", SampleExpiryDate))

	// Test case 7: when expiry date does not match
	assert.NotNil(t, gateway.verifySignedContent(unSignedContent, SampleDAppDomain, SampleNonce, SamplePublicKey, "invalid_expiry_date"))

}

func TestRequestNonceHandler(t *testing.T) {
	type fields struct {
		mockGfSpClient *gfspclient.MockGfSpClientAPI
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	type Body struct {
		fields         fields
		args           args
		wantRespBody   interface{}
		wantRespStatus interface{}
	}
	tm := time.Now().UTC()
	Now = func() time.Time {
		return tm
	}
	tests := []struct {
		name string
		f    func(*testing.T, *gomock.Controller) *Body
	}{
		{
			name: "case 1/requestNonce success",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				// mock data
				currentNonce := int32(0)
				nextNonce := int32(1)
				currentPublicKey := ""
				expiryDate := time.Now().UnixMilli()
				mockedClient := gfspclient.NewMockGfSpClientAPI(c)
				mockedClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return(currentNonce, nextNonce, currentPublicKey, expiryDate, nil).Times(1)

				req := httptest.NewRequest(http.MethodGet, "/auth/request_nonce", nil)
				w := httptest.NewRecorder()

				expectedBody := msgToString(currentNonce, nextNonce, currentPublicKey, expiryDate)
				return &Body{
					fields: fields{
						mockGfSpClient: mockedClient,
					},
					args:           args{w: w, r: req},
					wantRespBody:   expectedBody,
					wantRespStatus: http.StatusOK,
				}
			},
		},
		{
			name: "case 2/no nonce returned",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				mockedClient := gfspclient.NewMockGfSpClientAPI(c)
				mockedClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return(int32(0), int32(0), "", int64(0), gorm.ErrInvalidDB).Times(1)
				req := httptest.NewRequest(http.MethodGet, "/auth/request_nonce", nil)
				w := httptest.NewRecorder()
				return &Body{
					fields: fields{
						mockGfSpClient: mockedClient,
					},
					args:           args{w: w, r: req},
					wantRespBody:   nil,
					wantRespStatus: http.StatusInternalServerError,
				}
			},
		},
	}
	for _, _tt := range tests {
		t.Run(_tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tt := _tt.f(t, ctrl)
			gateway := &GateModular{
				env:    gfspapp.EnvLocal,
				domain: testDomain,
			}
			gateway.baseApp = &gfspapp.GfSpBaseApp{}
			gateway.baseApp.SetGfSpClient(tt.fields.mockGfSpClient)
			gateway.requestNonceHandler(tt.args.w, tt.args.r)

			res := tt.args.w.Result()
			defer res.Body.Close()

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			assert.Equal(t, tt.wantRespStatus, res.StatusCode)
			if tt.wantRespStatus == http.StatusOK {
				assert.Equal(t, tt.wantRespBody, string(data))
			}

		})
	}
}

func getSampleRequestWithAuthSig(domain string, nonce string, eddsaPublicKey string, spAddr string, validExpiryDate string) *http.Request {
	// Account information.
	privateKey, _ := crypto.GenerateKey()

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	log.Infof("address is: " + address.Hex())

	unSignedContent := UnsignedContentTemplate

	unSignedContent = fmt.Sprintf(unSignedContent, domain, address.Hex(), eddsaPublicKey, domain, validExpiryDate, validExpiryDate, spAddr, nonce)

	log.Infof("unSignedContent is: %s", unSignedContent)
	unSignedContentHash := accounts.TextHash([]byte(unSignedContent))

	// Sign data.
	signature, err := crypto.Sign(unSignedContentHash, privateKey)
	if err != nil {
		log.Error(err)
	}
	authString := commonhttp.Gnfd1EthPersonalSign + `,SignedMsg=%s,Signature=%s`
	authString = fmt.Sprintf(authString, unSignedContent, hexutil.Encode(signature))
	log.Infof("authString is: %s", authString)

	// mock request
	req := httptest.NewRequest(http.MethodPost, "/auth/update_key", nil)
	req.Host = "testBucket.gnfd.nodereal.com"

	req.Header.Set(ContentTypeHeader, "application/x-www-form-urlencoded")
	req.Header.Set(GnfdAuthorizationHeader, authString)
	req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDate)
	req.Header.Set(GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
	req.Header.Set(GnfdUserAddressHeader, address.String())
	return req
}

func TestUpdateUserPublicKeyHandler(t *testing.T) {
	type fields struct {
		mockGfSpClient *gfspclient.MockGfSpClientAPI
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	type Body struct {
		fields         fields
		args           args
		wantRespBody   interface{}
		wantRespStatus interface{}
	}
	tm := time.Now().UTC()
	Now = func() time.Time {
		return tm
	}
	tests := []struct {
		name string
		f    func(*testing.T, *gomock.Controller) *Body
	}{
		{
			name: "case UpdateUserPublicKey success",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				// mock data
				currentNonce := int32(0)
				nextNonce := int32(1)
				currentPublicKey := ""
				expiryDate := time.Now().UnixMilli()
				mockedClient := gfspclient.NewMockGfSpClientAPI(c)
				mockedClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return(currentNonce, nextNonce, currentPublicKey, expiryDate, nil).Times(1)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				// validExpiryDate, _ := time.Parse(ExpiryDateFormat, validExpiryDateStr)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDateStr)

				req.Header.Set(GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDateStr)

				mockedClient.EXPECT().UpdateUserPublicKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil).Times(1)

				w := httptest.NewRecorder()
				expectedBody := msgToStringForUpdateKey(true)
				return &Body{
					fields: fields{
						mockGfSpClient: mockedClient,
					},
					args:           args{w: w, r: req},
					wantRespBody:   expectedBody,
					wantRespStatus: http.StatusOK,
				}
			},
		},

		{
			name: "case wrongAuthString",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				currentNonce := int32(0)
				nextNonce := int32(1)
				currentPublicKey := ""
				expiryDate := time.Now().UnixMilli()
				mockedClient := gfspclient.NewMockGfSpClientAPI(c)
				mockedClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return(currentNonce, nextNonce, currentPublicKey, expiryDate, nil).Times(0)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "wrongAuthString")

				req.Header.Set(GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDateStr)

				mockedClient.EXPECT().UpdateUserPublicKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil).Times(0)

				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						mockGfSpClient: mockedClient,
					},
					args:           args{w: w, r: req},
					wantRespBody:   expectedBody,
					wantRespStatus: http.StatusBadRequest,
				}
			},
		},
		{
			name: "case domain is not set",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				currentNonce := int32(0)
				nextNonce := int32(1)
				currentPublicKey := ""
				expiryDate := time.Now().UnixMilli()
				mockedClient := gfspclient.NewMockGfSpClientAPI(c)
				mockedClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return(currentNonce, nextNonce, currentPublicKey, expiryDate, nil).Times(0)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDateStr)

				req.Header.Set("Origin", domain)
				req.Header.Set(GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDateStr)

				mockedClient.EXPECT().UpdateUserPublicKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil).Times(0)
				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						mockGfSpClient: mockedClient,
					},
					args:           args{w: w, r: req},
					wantRespBody:   expectedBody,
					wantRespStatus: http.StatusBadRequest,
				}
			},
		},
		{
			name: "case domain is matching with origin header",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				currentNonce := int32(0)
				nextNonce := int32(1)
				currentPublicKey := ""
				expiryDate := time.Now().UnixMilli()
				mockedClient := gfspclient.NewMockGfSpClientAPI(c)
				mockedClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return(currentNonce, nextNonce, currentPublicKey, expiryDate, nil).Times(0)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDateStr)

				req.Header.Set(GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", "https://another.site.com")
				req.Header.Set(GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDateStr)
				mockedClient.EXPECT().UpdateUserPublicKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil).Times(0)
				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						mockGfSpClient: mockedClient,
					},
					args:           args{w: w, r: req},
					wantRespBody:   expectedBody,
					wantRespStatus: http.StatusBadRequest,
				}
			},
		},
		{
			name: "case userPublicKey is not set",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				currentNonce := int32(0)
				nextNonce := int32(1)
				currentPublicKey := ""
				expiryDate := time.Now().UnixMilli()
				mockedClient := gfspclient.NewMockGfSpClientAPI(c)
				mockedClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return(currentNonce, nextNonce, currentPublicKey, expiryDate, nil).Times(0)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDateStr)
				req.Header.Set(GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDateStr)
				req.Header.Set(GnfdOffChainAuthAppRegPublicKeyHeader, "")

				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						mockGfSpClient: mockedClient,
					},
					args:           args{w: w, r: req},
					wantRespBody:   expectedBody,
					wantRespStatus: http.StatusBadRequest,
				}
			},
		},
		{
			name: "case GetAuthKey failed",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				mockedClient := gfspclient.NewMockGfSpClientAPI(c)
				mockedClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return(int32(0), int32(0), "", int64(0), gorm.ErrInvalidDB).Times(1)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDateStr)
				req.Header.Set(GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDateStr)

				mockedClient.EXPECT().UpdateUserPublicKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil).Times(0)

				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						mockGfSpClient: mockedClient,
					},
					args:           args{w: w, r: req},
					wantRespBody:   expectedBody,
					wantRespStatus: http.StatusInternalServerError,
				}
			},
		},
		{
			name: "case nonce is not a number",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				currentNonce := int32(0)
				nextNonce := int32(1)
				currentPublicKey := ""
				expiryDate := time.Now().UnixMilli()
				mockedClient := gfspclient.NewMockGfSpClientAPI(c)
				mockedClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return(currentNonce, nextNonce, currentPublicKey, expiryDate, nil).Times(1)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDateStr)

				req.Header.Set(GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(GnfdOffChainAuthAppRegNonceHeader, "a wrong nonce")
				req.Header.Set(GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDateStr)
				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						mockGfSpClient: mockedClient,
					},
					args:           args{w: w, r: req},
					wantRespBody:   expectedBody,
					wantRespStatus: http.StatusBadRequest,
				}
			},
		},
		{
			name: "case nonce is not expected for SP",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				currentNonce := int32(0)
				nextNonce := int32(1)
				currentPublicKey := ""
				expiryDate := time.Now().UnixMilli()
				mockedClient := gfspclient.NewMockGfSpClientAPI(c)
				mockedClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return(currentNonce, nextNonce, currentPublicKey, expiryDate, nil).Times(1)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDateStr)

				req.Header.Set(GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(GnfdOffChainAuthAppRegNonceHeader, "2") // sp expects 1 as nonce but the header is set to 2
				req.Header.Set(GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDateStr)
				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						mockGfSpClient: mockedClient,
					},
					args:           args{w: w, r: req},
					wantRespBody:   expectedBody,
					wantRespStatus: http.StatusBadRequest,
				}
			},
		},
		{
			name: "case the format of expiryDate is wrong",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				currentNonce := int32(0)
				nextNonce := int32(1)
				currentPublicKey := ""
				expiryDate := time.Now().UnixMilli()
				mockedClient := gfspclient.NewMockGfSpClientAPI(c)
				mockedClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return(currentNonce, nextNonce, currentPublicKey, expiryDate, nil).Times(1)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDateStr)

				req.Header.Set(GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(GnfdOffChainAuthAppRegExpiryDateHeader, "a wrong date")
				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						mockGfSpClient: mockedClient,
					},
					args:           args{w: w, r: req},
					wantRespBody:   expectedBody,
					wantRespStatus: http.StatusBadRequest,
				}
			},
		},
		{
			name: "case the expiryDate is too far since now",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				currentNonce := int32(0)
				nextNonce := int32(1)
				currentPublicKey := ""
				expiryDate := time.Now().UnixMilli()
				mockedClient := gfspclient.NewMockGfSpClientAPI(c)
				mockedClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return(currentNonce, nextNonce, currentPublicKey, expiryDate, nil).Times(1)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDateStr := time.Now().Add(time.Hour * 24 * 8).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDateStr)

				req.Header.Set(GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)

				req.Header.Set(GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDateStr)
				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						mockGfSpClient: mockedClient,
					},
					args:           args{w: w, r: req},
					wantRespBody:   expectedBody,
					wantRespStatus: http.StatusBadRequest,
				}
			},
		},
		{
			name: "case the unsigned content doesn't match the info set in headers",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				currentNonce := int32(0)
				nextNonce := int32(1)
				currentPublicKey := ""
				expiryDate := time.Now().UnixMilli()
				mockedClient := gfspclient.NewMockGfSpClientAPI(c)
				mockedClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return(currentNonce, nextNonce, currentPublicKey, expiryDate, nil).Times(1)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDateStr := time.Now().Add(time.Hour * 24 * 7).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDateStr)

				req.Header.Set(GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(GnfdOffChainAuthAppRegPublicKeyHeader, "06a3afc1e4bd4a43aa7801db260f6cd0e7429eb6f0c26790eaf0495b36cd3985") // set another key so that unsigned content becomes inconsistent

				req.Header.Set(GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDateStr)
				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						mockGfSpClient: mockedClient,
					},
					args:           args{w: w, r: req},
					wantRespBody:   expectedBody,
					wantRespStatus: http.StatusBadRequest,
				}
			},
		},
		{
			name: "case error occurs when saving key",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				currentNonce := int32(0)
				nextNonce := int32(1)
				currentPublicKey := ""
				expiryDate := time.Now().UnixMilli()
				mockedClient := gfspclient.NewMockGfSpClientAPI(c)
				mockedClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return(currentNonce, nextNonce, currentPublicKey, expiryDate, nil).Times(1)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDateStr := time.Now().Add(time.Hour * 24 * 7).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDateStr)

				req.Header.Set(GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDateStr)

				mockedClient.EXPECT().UpdateUserPublicKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(false, gorm.ErrInvalidDB).Times(1)
				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						mockGfSpClient: mockedClient,
					},
					args:           args{w: w, r: req},
					wantRespBody:   expectedBody,
					wantRespStatus: http.StatusInternalServerError,
				}
			},
		},
	}
	for _, _tt := range tests {
		t.Run(_tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tt := _tt.f(t, ctrl)

			gateway := &GateModular{
				env:    gfspapp.EnvLocal,
				domain: testDomain,
			}
			gateway.baseApp = &gfspapp.GfSpBaseApp{}
			gateway.baseApp.SetGfSpClient(tt.fields.mockGfSpClient)
			gateway.baseApp.SetOperatorAddress(TestSpAddress)

			gateway.updateUserPublicKeyHandler(tt.args.w, tt.args.r)

			res := tt.args.w.Result()
			defer res.Body.Close()

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			assert.Equal(t, tt.wantRespStatus, res.StatusCode)
			if tt.wantRespStatus == http.StatusOK {
				assert.Equal(t, tt.wantRespBody, string(data))
			}

		})
	}
}

// msgToString an util method to convert msg from getAuthNonce API to string
func msgToString(currentNonce int32, nextNonce int32, currentPublicKey string, expiryDate int64) string {
	var resp = &RequestNonceResp{
		CurrentNonce:     currentNonce,
		NextNonce:        nextNonce,
		CurrentPublicKey: currentPublicKey,
		ExpiryDate:       expiryDate,
	}
	b, _ := xml.Marshal(resp)
	return string(b)
}

// msgToStringForUpdateKey an util method to convert msg from UpdateUserPublicKey API to string
func msgToStringForUpdateKey(updateUserPublicKeyResult bool) string {
	var resp = &UpdateUserPublicKeyResp{
		Result: updateUserPublicKeyResult,
	}
	b, _ := xml.Marshal(resp)
	return string(b)
}

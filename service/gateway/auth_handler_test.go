package gateway

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	authclient "github.com/bnb-chain/greenfield-storage-provider/service/auth/client"
	mock_authtypes "github.com/bnb-chain/greenfield-storage-provider/service/auth/mock_client"
	authtypes "github.com/bnb-chain/greenfield-storage-provider/service/auth/types"
)

// Now used for test
var Now = time.Now

const (
	TestSpAddress string = "0x1c62EF97a13654A759C7E706Adf9EB3bAb0F807A"
	TestUserAcct  string = "0xa64fdc3b4866cd2ac664998c7b180813fb9b06e6"

	UnsignedContentTemplate string = `%s wants you to sign in with your BNB Greenfield account:
%s

Register your identity public key %s

URI: %s
Version: 1
Chain ID: 5600
Issued At: %s
Expiration Time: %s
Resources:
- SP %s (name: SP_001) with nonce: %s
- SP 0x20Bb76D063a6d2B18B6DaBb2aC985234a4B9eDe0 (name: SP_002) with nonce: 1`
)

func TestRequestNonceHandler(t *testing.T) {
	type fields struct {
		authClient *authclient.AuthClient
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
				mockAuthClient := mock_authtypes.NewMockAuthServiceClient(c)
				mockData := &authtypes.GetAuthNonceResponse{
					CurrentNonce:     0,
					CurrentPublicKey: "",
					NextNonce:        1,
					ExpiryDate:       time.Now().UnixMilli(),
				}
				mockAuthClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(1)
				req := httptest.NewRequest(http.MethodGet, "/auth/request_nonce", nil)
				w := httptest.NewRecorder()

				expectedBody := protoMsgToString(mockData)
				return &Body{
					fields: fields{
						authClient: &authclient.AuthClient{
							Auth: mockAuthClient,
						},
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
				mockAuthClient := mock_authtypes.NewMockAuthServiceClient(c)
				mockAuthClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any()).Return(nil, gorm.ErrInvalidDB).Times(1)
				req := httptest.NewRequest(http.MethodGet, "/auth/request_nonce", nil)
				w := httptest.NewRecorder()
				return &Body{
					fields: fields{
						authClient: &authclient.AuthClient{
							Auth: mockAuthClient,
						},
					},
					args:           args{w: w, r: req},
					wantRespBody:   nil,
					wantRespStatus: http.StatusInternalServerError,
				}
			},
		},
		{
			name: "case 3/no auth client is set",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				req := httptest.NewRequest(http.MethodGet, "/auth/request_nonce", nil)
				w := httptest.NewRecorder()
				return &Body{
					fields: fields{
						authClient: nil,
					},
					args:           args{w: w, r: req},
					wantRespBody:   nil,
					wantRespStatus: http.StatusNotImplemented,
				}
			},
		},
	}
	for _, _tt := range tests {
		t.Run(_tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tt := _tt.f(t, ctrl)

			gateway := &Gateway{
				config: &GatewayConfig{
					SpOperatorAddress: TestSpAddress,
				},
				auth: tt.fields.authClient,
			}
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
	unSignedContent := fmt.Sprintf(UnsignedContentTemplate, domain, TestUserAcct, eddsaPublicKey, domain, SampleIssueDate, validExpiryDate, TestSpAddress, nonce)

	log.Infof("unSignedContent is: %s", unSignedContent)
	unSignedContentHash := accounts.TextHash([]byte(unSignedContent))
	// Account information.
	privateKey, _ := crypto.GenerateKey()

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	log.Infof("address is: " + address.Hex())
	// Sign data.
	signature, err := crypto.Sign(unSignedContentHash, privateKey)
	if err != nil {
		log.Error(err)
	}
	authString := `PersonalSign ECDSA-secp256k1,SignedMsg=%s,Signature=%s`
	authString = fmt.Sprintf(authString, unSignedContent, hexutil.Encode(signature))
	log.Infof("authString is: %s", authString)

	// mock request
	req := httptest.NewRequest(http.MethodPost, "/auth/update_key", nil)
	req.Host = "testBucket.gnfd.nodereal.com"

	req.Header.Set(model.ContentTypeHeader, "application/x-www-form-urlencoded")
	req.Header.Set(model.GnfdAuthorizationHeader, authString)
	req.Header.Set(model.GnfdUserAddressHeader, address.String())
	return req
}

func TestUpdateUserPublicKeyHandler(t *testing.T) {
	type fields struct {
		authClient *authclient.AuthClient
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
				mockData := &authtypes.GetAuthNonceResponse{
					CurrentNonce:     0,
					CurrentPublicKey: "",
					NextNonce:        1,
					ExpiryDate:       time.Now().UnixMilli(),
				}
				mockAuthClient := mock_authtypes.NewMockAuthServiceClient(c)
				mockAuthClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(1)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				validExpiryDate, _ := time.Parse(ExpiryDateFormat, validExpiryDateStr)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDateStr)

				req.Header.Set(model.GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(model.GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(model.GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(model.GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDateStr)

				updateUserPublicKeyReq := &authtypes.UpdateUserPublicKeyRequest{
					AccountId:     req.Header.Get(model.GnfdUserAddressHeader),
					Domain:        domain,
					CurrentNonce:  mockData.CurrentNonce,
					Nonce:         mockData.CurrentNonce + 1,
					UserPublicKey: eddsaPublicKey,
					ExpiryDate:    validExpiryDate.UnixMilli(),
				}
				mockAuthClient.EXPECT().UpdateUserPublicKey(gomock.Any(), updateUserPublicKeyReq).Return(&authtypes.UpdateUserPublicKeyResponse{Result: true}, nil).Times(1)

				w := httptest.NewRecorder()
				expectedBody := protoMsgToString(&authtypes.UpdateUserPublicKeyResponse{Result: true})
				return &Body{
					fields: fields{
						authClient: &authclient.AuthClient{
							Auth: mockAuthClient,
						},
					},
					args:           args{w: w, r: req},
					wantRespBody:   expectedBody,
					wantRespStatus: http.StatusOK,
				}
			},
		},
		{
			name: "case no auth client is set",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDate := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDate)

				req.Header.Set(model.GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(model.GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(model.GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(model.GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDate)
				w := httptest.NewRecorder()
				return &Body{
					fields: fields{
						authClient: nil,
					},
					args:           args{w: w, r: req},
					wantRespBody:   nil,
					wantRespStatus: http.StatusNotImplemented,
				}
			},
		},
		{
			name: "case wrongAuthString",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				mockData := &authtypes.GetAuthNonceResponse{
					CurrentNonce:     0,
					CurrentPublicKey: "",
					NextNonce:        1,
					ExpiryDate:       time.Now().UnixMilli(),
				}
				mockAuthClient := mock_authtypes.NewMockAuthServiceClient(c)
				mockAuthClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(0)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDateStr)
				req.Header.Set(model.GnfdAuthorizationHeader, "wrongAuthString")

				req.Header.Set(model.GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(model.GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(model.GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(model.GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDateStr)

				mockAuthClient.EXPECT().UpdateUserPublicKey(gomock.Any(), gomock.Any()).Return(&authtypes.UpdateUserPublicKeyResponse{Result: true}, nil).Times(0)

				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						authClient: &authclient.AuthClient{
							Auth: mockAuthClient,
						},
					},
					args:           args{w: w, r: req},
					wantRespBody:   expectedBody,
					wantRespStatus: http.StatusForbidden,
				}
			},
		},
		{
			name: "case domain is not set",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				mockData := &authtypes.GetAuthNonceResponse{
					CurrentNonce:     0,
					CurrentPublicKey: "",
					NextNonce:        1,
					ExpiryDate:       time.Now().UnixMilli(),
				}
				mockAuthClient := mock_authtypes.NewMockAuthServiceClient(c)
				mockAuthClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(0)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDateStr)

				req.Header.Set("Origin", domain)
				req.Header.Set(model.GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(model.GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(model.GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDateStr)

				mockAuthClient.EXPECT().UpdateUserPublicKey(gomock.Any(), gomock.Any()).Return(&authtypes.UpdateUserPublicKeyResponse{Result: true}, nil).Times(0)

				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						authClient: &authclient.AuthClient{
							Auth: mockAuthClient,
						},
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
				mockData := &authtypes.GetAuthNonceResponse{
					CurrentNonce:     0,
					CurrentPublicKey: "",
					NextNonce:        1,
					ExpiryDate:       time.Now().UnixMilli(),
				}
				mockAuthClient := mock_authtypes.NewMockAuthServiceClient(c)
				mockAuthClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(0)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDateStr)

				req.Header.Set(model.GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", "https://another.site.com")
				req.Header.Set(model.GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(model.GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(model.GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDateStr)
				mockAuthClient.EXPECT().UpdateUserPublicKey(gomock.Any(), gomock.Any()).Return(&authtypes.UpdateUserPublicKeyResponse{Result: true}, nil).Times(0)

				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						authClient: &authclient.AuthClient{
							Auth: mockAuthClient,
						},
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
				mockData := &authtypes.GetAuthNonceResponse{
					CurrentNonce:     0,
					CurrentPublicKey: "",
					NextNonce:        1,
					ExpiryDate:       time.Now().UnixMilli(),
				}
				mockAuthClient := mock_authtypes.NewMockAuthServiceClient(c)
				mockAuthClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(0)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDate := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDate)
				req.Header.Set(model.GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(model.GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(model.GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDate)
				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						authClient: &authclient.AuthClient{
							Auth: mockAuthClient,
						},
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
				mockAuthClient := mock_authtypes.NewMockAuthServiceClient(c)
				mockAuthClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any()).Return(nil, gorm.ErrInvalidDB).Times(1)
				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDate := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDate)
				req.Header.Set(model.GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(model.GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(model.GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(model.GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDate)
				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						authClient: &authclient.AuthClient{
							Auth: mockAuthClient,
						},
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
				mockData := &authtypes.GetAuthNonceResponse{
					CurrentNonce:     0,
					CurrentPublicKey: "",
					NextNonce:        1,
					ExpiryDate:       time.Now().UnixMilli(),
				}
				mockAuthClient := mock_authtypes.NewMockAuthServiceClient(c)
				mockAuthClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(1)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDate := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDate)
				req.Header.Set(model.GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(model.GnfdOffChainAuthAppRegNonceHeader, "a wrong nonce")
				req.Header.Set(model.GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(model.GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDate)
				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						authClient: &authclient.AuthClient{
							Auth: mockAuthClient,
						},
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
				mockData := &authtypes.GetAuthNonceResponse{
					CurrentNonce:     0,
					CurrentPublicKey: "",
					NextNonce:        1,
					ExpiryDate:       time.Now().UnixMilli(),
				}
				mockAuthClient := mock_authtypes.NewMockAuthServiceClient(c)
				mockAuthClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(1)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDate := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDate)
				req.Header.Set(model.GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(model.GnfdOffChainAuthAppRegNonceHeader, "2") // sp expects 1 as nonce but the header is set to 2
				req.Header.Set(model.GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(model.GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDate)
				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						authClient: &authclient.AuthClient{
							Auth: mockAuthClient,
						},
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
				mockData := &authtypes.GetAuthNonceResponse{
					CurrentNonce:     0,
					CurrentPublicKey: "",
					NextNonce:        1,
					ExpiryDate:       time.Now().UnixMilli(),
				}
				mockAuthClient := mock_authtypes.NewMockAuthServiceClient(c)
				mockAuthClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(1)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDate := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDate)

				req.Header.Set(model.GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(model.GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(model.GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(model.GnfdOffChainAuthAppRegExpiryDateHeader, "a wrong date")
				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						authClient: &authclient.AuthClient{
							Auth: mockAuthClient,
						},
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
				mockData := &authtypes.GetAuthNonceResponse{
					CurrentNonce:     0,
					CurrentPublicKey: "",
					NextNonce:        1,
					ExpiryDate:       time.Now().UnixMilli(),
				}
				mockAuthClient := mock_authtypes.NewMockAuthServiceClient(c)
				mockAuthClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(1)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDate := time.Now().Add(time.Hour*24*7 + time.Minute).Format(ExpiryDateFormat) // a too far future date
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDate)

				req.Header.Set(model.GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(model.GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(model.GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)

				req.Header.Set(model.GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDate)
				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						authClient: &authclient.AuthClient{
							Auth: mockAuthClient,
						},
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
				mockData := &authtypes.GetAuthNonceResponse{
					CurrentNonce:     0,
					CurrentPublicKey: "",
					NextNonce:        1,
					ExpiryDate:       time.Now().UnixMilli(),
				}
				mockAuthClient := mock_authtypes.NewMockAuthServiceClient(c)
				mockAuthClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(1)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDate := time.Now().Add(time.Hour * 7).Format(ExpiryDateFormat) // a too far future date
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDate)

				req.Header.Set(model.GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(model.GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(model.GnfdOffChainAuthAppRegPublicKeyHeader, "some other key") // set another key so that unsigned content becomes inconsistent

				req.Header.Set(model.GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDate)
				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						authClient: &authclient.AuthClient{
							Auth: mockAuthClient,
						},
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
				mockData := &authtypes.GetAuthNonceResponse{
					CurrentNonce:     0,
					CurrentPublicKey: "",
					NextNonce:        1,
					ExpiryDate:       time.Now().UnixMilli(),
				}
				mockAuthClient := mock_authtypes.NewMockAuthServiceClient(c)
				mockAuthClient.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(1)

				domain := SampleDAppDomain
				nonce := "1"
				eddsaPublicKey := SamplePublicKey
				spAddr := TestSpAddress
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req := getSampleRequestWithAuthSig(domain, nonce, eddsaPublicKey, spAddr, validExpiryDateStr)

				req.Header.Set(model.GnfdOffChainAuthAppDomainHeader, domain)
				req.Header.Set("Origin", domain)
				req.Header.Set(model.GnfdOffChainAuthAppRegNonceHeader, nonce)
				req.Header.Set(model.GnfdOffChainAuthAppRegPublicKeyHeader, eddsaPublicKey)
				req.Header.Set(model.GnfdOffChainAuthAppRegExpiryDateHeader, validExpiryDateStr)

				mockAuthClient.EXPECT().UpdateUserPublicKey(gomock.Any(), gomock.Any()).Return(nil, gorm.ErrInvalidDB).Times(1)

				w := httptest.NewRecorder()
				expectedBody := ""
				return &Body{
					fields: fields{
						authClient: &authclient.AuthClient{
							Auth: mockAuthClient,
						},
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

			gateway := &Gateway{
				config: &GatewayConfig{
					SpOperatorAddress: TestSpAddress,
				},
				auth: tt.fields.authClient,
			}
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

func TestAuthHandlerVerifySignedContentInEip4361Template(t *testing.T) {
	gateway := &Gateway{
		config: &GatewayConfig{
			SpOperatorAddress: TestSpAddress,
		},
	}
	expectedDomain := SampleDAppDomain
	expectedNonce := "123456"
	expectedPublicKey := SamplePublicKey
	expectedExpiryDate := "2023-04-18T16:25:24Z"
	expectedIssueDate := SampleIssueDate
	unSignedContent := fmt.Sprintf(UnsignedContentTemplate, expectedDomain, TestUserAcct, expectedPublicKey, expectedDomain, expectedIssueDate, expectedExpiryDate, TestSpAddress, expectedNonce)
	log.Infof(unSignedContent)
	// Test case 1: all inputs are correct.
	assert.Nil(t, gateway.verifySignedContent(unSignedContent, expectedDomain, expectedNonce, expectedPublicKey, expectedExpiryDate))

	// Test case 2: when signed content does not match template
	assert.Equal(t, SignedMsgNotMatchTemplate, gateway.verifySignedContent("wrong-signed-content", expectedDomain, expectedNonce, expectedPublicKey, expectedExpiryDate))

	// Test case 3: when nonce does not match.
	assert.NotNil(t, gateway.verifySignedContent(unSignedContent, expectedDomain, "invalid_nonce", expectedPublicKey, expectedExpiryDate))

	// Test case 4: when signed content is not for this SP.
	assert.NotNil(t, gateway.verifySignedContent("Register your identity of dapp test_domain \nwith your identity key test_key .\nIn the following SPs:\n- SP SP_OtherAddress (name:SP_Name) with nonce:123456\n\nThe expiry date is test_expiry_date", expectedDomain, expectedNonce, expectedPublicKey, expectedExpiryDate))

	// Test case 5: when dapp domain does not match
	assert.NotNil(t, gateway.verifySignedContent(unSignedContent, "invalid_dapp_domain", expectedNonce, expectedPublicKey, expectedExpiryDate))

	// Test case 6: when public key does not match
	assert.NotNil(t, gateway.verifySignedContent(unSignedContent, expectedDomain, expectedNonce, "invalid_public_key", expectedExpiryDate))

	// Test case 7: when expiry date does not match
	assert.NotNil(t, gateway.verifySignedContent(unSignedContent, expectedDomain, expectedNonce, expectedPublicKey, "invalid_expiry_date"))

}

// protoMsgToString an util method to convert protoMsg to string
func protoMsgToString(pb proto.Message) string {
	var b bytes.Buffer
	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	m.Marshal(&b, pb)
	return b.String()
}

package auth

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	authclient "github.com/bnb-chain/greenfield-storage-provider/service/auth/client"
	authtypes "github.com/bnb-chain/greenfield-storage-provider/service/auth/types"
	mock_store "github.com/bnb-chain/greenfield-storage-provider/store/mock"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
)

// Now used for test
var Now = time.Now()
var testSeed = []byte{15, 128, 7, 103, 18, 78, 41, 63, 47, 120, 41, 247, 209, 152, 27, 43, 225, 223, 242, 203, 250, 160, 164, 252, 210, 82, 236, 192, 196, 78, 97, 41, 78, 38, 181, 77, 175, 253, 155, 97, 176, 146, 76, 158, 41, 102, 57, 177, 37, 177, 201, 227, 174, 159, 83, 80, 175, 202, 121, 136, 211, 114, 126, 199, 28}

const (
	testSpAddress string = "0x1c62EF97a13654A759C7E706Adf9EB3bAb0F807A"
)

func TestGetAuthNonce(t *testing.T) {
	type fields struct {
		store sqldb.SPDB
		auth  *authclient.AuthClient
	}
	type args struct {
		req *authtypes.GetAuthNonceRequest
	}
	type Body struct {
		fields    fields
		args      args
		wantedRes *authtypes.GetAuthNonceResponse
	}

	tests := []struct {
		name string
		f    func(*testing.T, *gomock.Controller) *Body
	}{
		{
			name: "case 1/getAuthNonce success",
			f: func(t *testing.T, c *gomock.Controller) *Body {

				mockSPDB := mock_store.NewMockSPDB(c)
				mockData := &sqldb.OffChainAuthKeyTable{
					UserAddress:      "0xa64FdC3B4866CD2aC664998C7b180813fB9B06E6",
					Domain:           "https://a_dapp.com",
					CurrentNonce:     0,
					CurrentPublicKey: "",
					NextNonce:        1,
					ExpiryDate:       Now,
					CreatedTime:      Now,
					ModifiedTime:     Now,
				}
				mockSPDB.EXPECT().GetAuthKey(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(1)
				req := &authtypes.GetAuthNonceRequest{
					AccountId: "0xa64FdC3B4866CD2aC664998C7b180813fB9B06E6",
					Domain:    "https://a_dapp.com",
				}
				expectResp := &authtypes.GetAuthNonceResponse{
					CurrentNonce:     0,
					NextNonce:        1,
					CurrentPublicKey: "",
					ExpiryDate:       Now.UnixMilli(),
				}
				return &Body{
					fields: fields{
						store: mockSPDB,
					},
					args:      args{req: req},
					wantedRes: expectResp,
				}
			},
		},
	}
	for _, _tt := range tests {
		t.Run(_tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tt := _tt.f(t, ctrl)

			authServer := &AuthServer{
				config: &AuthConfig{
					SpOperatorAddress: testSpAddress,
				},
				spDB: tt.fields.store,
			}
			actualResp, err := authServer.GetAuthNonce(context.Background(), tt.args.req)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantedRes, actualResp)

		})
	}
}

func TestUpdateUserPublicKey(t *testing.T) {
	type fields struct {
		store sqldb.SPDB
		auth  *authclient.AuthClient
	}
	type args struct {
		req *authtypes.UpdateUserPublicKeyRequest
	}
	type Body struct {
		fields    fields
		args      args
		wantedRes *authtypes.UpdateUserPublicKeyResponse
	}

	tests := []struct {
		name string
		f    func(*testing.T, *gomock.Controller) *Body
	}{
		{
			name: "case 1/UpdateUserPublicKey success",
			f: func(t *testing.T, c *gomock.Controller) *Body {

				mockSPDB := mock_store.NewMockSPDB(c)
				mockData := &sqldb.OffChainAuthKeyTable{
					UserAddress:      "0xa64FdC3B4866CD2aC664998C7b180813fB9B06E6",
					Domain:           "https://a_dapp.com",
					CurrentNonce:     0,
					CurrentPublicKey: "",
					NextNonce:        1,
					ExpiryDate:       Now,
					CreatedTime:      Now,
					ModifiedTime:     Now,
				}
				mockSPDB.EXPECT().GetAuthKey(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(0)

				expectResp := &authtypes.UpdateUserPublicKeyResponse{
					Result: true,
				}
				mockSPDB.EXPECT().UpdateAuthKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				req := &authtypes.UpdateUserPublicKeyRequest{
					AccountId:     "0xa64FdC3B4866CD2aC664998C7b180813fB9B06E6",
					Domain:        "https://a_dapp.com",
					CurrentNonce:  0,
					Nonce:         1,
					UserPublicKey: "a user public key",
					ExpiryDate:    Now.UnixMilli(),
				}
				return &Body{
					fields: fields{
						store: mockSPDB,
					},
					args:      args{req: req},
					wantedRes: expectResp,
				}
			},
		},
	}
	for _, _tt := range tests {
		t.Run(_tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tt := _tt.f(t, ctrl)

			authServer := &AuthServer{
				config: &AuthConfig{
					SpOperatorAddress: testSpAddress,
				},
				spDB: tt.fields.store,
			}
			actualResp, err := authServer.UpdateUserPublicKey(context.Background(), tt.args.req)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantedRes, actualResp)

		})
	}
}

// generateOffChainSigForTesting an util method to help generate signature by giving testSeed and msgToSign
// the msgToSign format is expected to format as `${actionContent}_${timestamp}
func generateOffChainSigForTesting(testSeed []byte, msgToSign string) string {
	sk, _ := GenerateEddsaPrivateKey(string(testSeed))
	hFunc := mimc.NewMiMC()
	sig, _ := sk.Sign([]byte(msgToSign), hFunc)
	log.Infof("sig is %s", hex.EncodeToString(sig))
	return hex.EncodeToString(sig)
}

func TestVerifyOffChainSignature(t *testing.T) {
	type fields struct {
		store sqldb.SPDB
		auth  *authclient.AuthClient
	}
	type args struct {
		req *authtypes.VerifyOffChainSignatureRequest
	}
	type Body struct {
		fields    fields
		args      args
		wantedRes *authtypes.VerifyOffChainSignatureResponse
		wantedErr error
	}

	tests := []struct {
		name string
		f    func(*testing.T, *gomock.Controller) *Body
	}{
		{
			name: "case 1/VerifyOffChainSignature success",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				mockSPDB := mock_store.NewMockSPDB(c)
				mockData := &sqldb.OffChainAuthKeyTable{
					UserAddress:      "0xa64FdC3B4866CD2aC664998C7b180813fB9B06E6",
					Domain:           "https://a_dapp.com",
					CurrentNonce:     0,
					CurrentPublicKey: "4db642fe6bc2ceda2e002feb8d78dfbcb2879d8fe28e84e02b7a940bc0440083",
					NextNonce:        1,
					ExpiryDate:       Now,
					CreatedTime:      Now,
					ModifiedTime:     Now,
				}
				mockSPDB.EXPECT().GetAuthKey(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(1)

				expectResp := &authtypes.VerifyOffChainSignatureResponse{
					Result: true,
				}
				realMsgToSign := fmt.Sprintf("Request_API_Download_%d", time.Now().Add(time.Minute*4).UnixMilli())
				offChainSig := generateOffChainSigForTesting(testSeed, realMsgToSign)
				req := &authtypes.VerifyOffChainSignatureRequest{
					AccountId:     "0xa64FdC3B4866CD2aC664998C7b180813fB9B06E6",
					Domain:        "https://a_dapp.com",
					OffChainSig:   offChainSig,
					RealMsgToSign: realMsgToSign,
				}
				return &Body{
					fields: fields{
						store: mockSPDB,
					},
					args:      args{req: req},
					wantedRes: expectResp,
				}
			},
		},
		{
			name: "case 2/wrong msg format",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				mockSPDB := mock_store.NewMockSPDB(c)
				mockData := &sqldb.OffChainAuthKeyTable{
					UserAddress:      "0xa64FdC3B4866CD2aC664998C7b180813fB9B06E6",
					Domain:           "https://a_dapp.com",
					CurrentNonce:     0,
					CurrentPublicKey: "4db642fe6bc2ceda2e002feb8d78dfbcb2879d8fe28e84e02b7a940bc0440083",
					NextNonce:        1,
					ExpiryDate:       Now,
					CreatedTime:      Now,
					ModifiedTime:     Now,
				}
				mockSPDB.EXPECT().GetAuthKey(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(1)

				realMsgToSign := fmt.Sprintf("RequestAPIDownload%d", time.Now().Add(time.Minute*4).UnixMilli())
				offChainSig := generateOffChainSigForTesting(testSeed, realMsgToSign)
				req := &authtypes.VerifyOffChainSignatureRequest{
					AccountId:     "0xa64FdC3B4866CD2aC664998C7b180813fB9B06E6",
					Domain:        "https://a_dapp.com",
					OffChainSig:   offChainSig,
					RealMsgToSign: realMsgToSign,
				}
				return &Body{
					fields: fields{
						store: mockSPDB,
					},
					args:      args{req: req},
					wantedRes: nil,
					wantedErr: fmt.Errorf("signed msg must be formated as ${actionContent}_${expiredTimestamp}"),
				}
			},
		},

		{
			name: "case 3/expired msg ",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				mockSPDB := mock_store.NewMockSPDB(c)
				mockData := &sqldb.OffChainAuthKeyTable{
					UserAddress:      "0xa64FdC3B4866CD2aC664998C7b180813fB9B06E6",
					Domain:           "https://a_dapp.com",
					CurrentNonce:     0,
					CurrentPublicKey: "4db642fe6bc2ceda2e002feb8d78dfbcb2879d8fe28e84e02b7a940bc0440083",
					NextNonce:        1,
					ExpiryDate:       Now,
					CreatedTime:      Now,
					ModifiedTime:     Now,
				}
				mockSPDB.EXPECT().GetAuthKey(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(1)

				realMsgToSign := fmt.Sprintf("Request_API_Download_%d", time.Now().Add(-time.Minute*4).UnixMilli())
				offChainSig := generateOffChainSigForTesting(testSeed, realMsgToSign)
				req := &authtypes.VerifyOffChainSignatureRequest{
					AccountId:     "0xa64FdC3B4866CD2aC664998C7b180813fB9B06E6",
					Domain:        "https://a_dapp.com",
					OffChainSig:   offChainSig,
					RealMsgToSign: realMsgToSign,
				}
				return &Body{
					fields: fields{
						store: mockSPDB,
					},
					args:      args{req: req},
					wantedRes: nil,
					wantedErr: fmt.Errorf("expiredTimestamp in signed msg must be within %d seconds", OffChainAuthSigExpiryAgeInSec),
				}
			},
		},

		{
			name: "case 4/msg expiredTimestamp is set too far ",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				mockSPDB := mock_store.NewMockSPDB(c)
				mockData := &sqldb.OffChainAuthKeyTable{
					UserAddress:      "0xa64FdC3B4866CD2aC664998C7b180813fB9B06E6",
					Domain:           "https://a_dapp.com",
					CurrentNonce:     0,
					CurrentPublicKey: "4db642fe6bc2ceda2e002feb8d78dfbcb2879d8fe28e84e02b7a940bc0440083",
					NextNonce:        1,
					ExpiryDate:       Now,
					CreatedTime:      Now,
					ModifiedTime:     Now,
				}
				mockSPDB.EXPECT().GetAuthKey(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(1)

				realMsgToSign := fmt.Sprintf("Request_API_Download_%d", time.Now().Add(+time.Minute*8).UnixMilli())
				offChainSig := generateOffChainSigForTesting(testSeed, realMsgToSign)
				req := &authtypes.VerifyOffChainSignatureRequest{
					AccountId:     "0xa64FdC3B4866CD2aC664998C7b180813fB9B06E6",
					Domain:        "https://a_dapp.com",
					OffChainSig:   offChainSig,
					RealMsgToSign: realMsgToSign,
				}
				return &Body{
					fields: fields{
						store: mockSPDB,
					},
					args:      args{req: req},
					wantedRes: nil,
					wantedErr: fmt.Errorf("expiredTimestamp in signed msg must be within %d seconds", OffChainAuthSigExpiryAgeInSec),
				}
			},
		},

		{
			name: "case 5/wrong Timestamp format",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				mockSPDB := mock_store.NewMockSPDB(c)
				mockData := &sqldb.OffChainAuthKeyTable{
					UserAddress:      "0xa64FdC3B4866CD2aC664998C7b180813fB9B06E6",
					Domain:           "https://a_dapp.com",
					CurrentNonce:     0,
					CurrentPublicKey: "4db642fe6bc2ceda2e002feb8d78dfbcb2879d8fe28e84e02b7a940bc0440083",
					NextNonce:        1,
					ExpiryDate:       Now,
					CreatedTime:      Now,
					ModifiedTime:     Now,
				}
				mockSPDB.EXPECT().GetAuthKey(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(1)

				realMsgToSign := fmt.Sprintf("RequestAPIDownload_%d_wrongformat", time.Now().Add(time.Minute*4).UnixMilli())
				offChainSig := generateOffChainSigForTesting(testSeed, realMsgToSign)
				req := &authtypes.VerifyOffChainSignatureRequest{
					AccountId:     "0xa64FdC3B4866CD2aC664998C7b180813fB9B06E6",
					Domain:        "https://a_dapp.com",
					OffChainSig:   offChainSig,
					RealMsgToSign: realMsgToSign,
				}
				return &Body{
					fields: fields{
						store: mockSPDB,
					},
					args:      args{req: req},
					wantedRes: nil,
					wantedErr: fmt.Errorf("expiredTimestamp in signed msg must be a unix epoch time in milliseconds"),
				}
			},
		},
		{
			name: "case 6/VerifyOffChainSignature failed",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				mockSPDB := mock_store.NewMockSPDB(c)
				mockData := &sqldb.OffChainAuthKeyTable{
					UserAddress:      "0xa64FdC3B4866CD2aC664998C7b180813fB9B06E6",
					Domain:           "https://a_dapp.com",
					CurrentNonce:     0,
					CurrentPublicKey: "4db642fe6bc2ceda2e002feb8d78dfbcb2879d8fe28e84e02b7a940bc0440083",
					NextNonce:        1,
					ExpiryDate:       Now,
					CreatedTime:      Now,
					ModifiedTime:     Now,
				}
				mockSPDB.EXPECT().GetAuthKey(gomock.Any(), gomock.Any()).Return(mockData, nil).Times(1)

				realMsgToSign := fmt.Sprintf("Request_API_Download_%d", time.Now().Add(time.Minute*4).UnixMilli())
				offChainSig := generateOffChainSigForTesting(testSeed, realMsgToSign)
				req := &authtypes.VerifyOffChainSignatureRequest{
					AccountId:     "0xa64FdC3B4866CD2aC664998C7b180813fB9B06E6",
					Domain:        "https://a_dapp.com",
					OffChainSig:   offChainSig,
					RealMsgToSign: fmt.Sprintf("Request_API_Update_%d", time.Now().Add(time.Minute*4).UnixMilli()), // msg and sig are not matched
				}
				return &Body{
					fields: fields{
						store: mockSPDB,
					},
					args:      args{req: req},
					wantedRes: nil,
					wantedErr: errors.New("invalid signature"),
				}
			},
		},
	}
	for _, _tt := range tests {
		t.Run(_tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tt := _tt.f(t, ctrl)

			authServer := &AuthServer{
				config: &AuthConfig{
					SpOperatorAddress: testSpAddress,
				},
				spDB: tt.fields.store,
			}
			actualResp, err := authServer.VerifyOffChainSignature(context.Background(), tt.args.req)
			if tt.wantedErr != nil {
				assert.Equal(t, tt.wantedErr, err)
			}
			assert.Equal(t, tt.wantedRes, actualResp)
		})
	}
}

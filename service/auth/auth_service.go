package auth

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	authtypes "github.com/bnb-chain/greenfield-storage-provider/service/auth/types"
)

var _ authtypes.AuthServiceServer = &AuthServer{}

const (
	OffChainAuthSigExpiryAgeInSec int32 = 60 * 5 // in 300 seconds
)

// GetAuthNonce get the auth nonce for which the Dapp or client can generate EDDSA key pairs.
func (auth *AuthServer) GetAuthNonce(ctx context.Context, req *authtypes.GetAuthNonceRequest) (resp *authtypes.GetAuthNonceResponse, err error) {
	domain := req.Domain

	ctx = log.Context(ctx, req)
	authKey, err := auth.spDB.GetAuthKey(req.AccountId, domain)
	if err != nil {
		log.Errorw("failed to GetAuthKey", "error", err)
		return nil, err
	}
	resp = &authtypes.GetAuthNonceResponse{
		CurrentNonce:     authKey.CurrentNonce,
		NextNonce:        authKey.NextNonce,
		CurrentPublicKey: authKey.CurrentPublicKey,
		ExpiryDate:       authKey.ExpiryDate.UnixMilli(),
	}
	log.CtxInfow(ctx, "succeed to GetAuthNonce")
	return resp, nil
}

// UpdateUserPublicKey updates the user public key once the Dapp or client generates the EDDSA key pairs.
func (auth *AuthServer) UpdateUserPublicKey(ctx context.Context, req *authtypes.UpdateUserPublicKeyRequest) (*authtypes.UpdateUserPublicKeyResponse, error) {
	err := auth.spDB.UpdateAuthKey(req.AccountId, req.Domain, req.CurrentNonce, req.Nonce, req.UserPublicKey, time.UnixMilli(req.ExpiryDate))
	if err != nil {
		log.Errorw("failed to updateUserPublicKey when saving key")
		return nil, err
	}
	resp := &authtypes.UpdateUserPublicKeyResponse{
		Result: true,
	}
	log.CtxInfow(ctx, "succeed to UpdateUserPublicKey")
	return resp, nil
}

// VerifyOffChainSignature verifies the signature signed by user's EDDSA private key.
func (auth *AuthServer) VerifyOffChainSignature(ctx context.Context, req *authtypes.VerifyOffChainSignatureRequest) (*authtypes.VerifyOffChainSignatureResponse, error) {

	signedMsg := req.RealMsgToSign
	sigString := req.OffChainSig

	signature, err := hex.DecodeString(sigString)
	if err != nil {
		return nil, err
	}

	getAuthNonceReq := &authtypes.GetAuthNonceRequest{
		AccountId: req.AccountId,
		Domain:    req.Domain,
	}
	getAuthNonceResp, err := auth.GetAuthNonce(ctx, getAuthNonceReq)
	if err != nil {
		return nil, err
	}
	userPublicKey := getAuthNonceResp.CurrentPublicKey

	// signedMsg must be formatted as `${actionContent}_${expiredTimestamp}` and timestamp must be within $OffChainAuthSigExpiryAgeInSec seconds, actionContent could be any string
	signedMsgParts := strings.Split(signedMsg, "_")
	if len(signedMsgParts) < 2 {
		err = fmt.Errorf("signed msg must be formated as ${actionContent}_${expiredTimestamp}")
		return nil, err
	}

	signedMsgExpiredTimestamp, err := strconv.Atoi(signedMsgParts[len(signedMsgParts)-1])
	if err != nil {
		err = fmt.Errorf("expiredTimestamp in signed msg must be a unix epoch time in milliseconds")
		return nil, err
	}
	expiredAge := time.Until(time.UnixMilli(int64(signedMsgExpiredTimestamp))).Seconds()
	if float64(OffChainAuthSigExpiryAgeInSec) < expiredAge || expiredAge < 0 { // nonce must be the same as NextNonce
		err = fmt.Errorf("expiredTimestamp in signed msg must be within %d seconds", OffChainAuthSigExpiryAgeInSec)
		return nil, err
	}

	err = VerifyEddsaSignature(userPublicKey, signature, []byte(signedMsg))
	if err != nil {
		return nil, err
	}
	log.Infof("verifyOffChainSignature: err %s", err)
	resp := &authtypes.VerifyOffChainSignatureResponse{
		Result: true,
	}
	log.CtxInfow(ctx, "succeed to VerifyOffChainSignature")
	return resp, nil
}

package authorizer

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"

	"strconv"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	ErrUnsupportedAuthType = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20001, "unsupported auth op type")
	ErrMismatchSp          = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20002, "mismatched primary sp")
	ErrNotCreatedState     = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20003, "object has not been created state")
	ErrNotSealedState      = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20004, "object has not been sealed state")
	ErrPaymentState        = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20005, "payment account is not active")
	ErrNoSuchAccount       = gfsperrors.Register(module.AuthorizationModularName, http.StatusNotFound, 20006, "no such account")
	ErrNoSuchBucket        = gfsperrors.Register(module.AuthorizationModularName, http.StatusNotFound, 20007, "no such bucket")
	ErrNoSuchObject        = gfsperrors.Register(module.AuthorizationModularName, http.StatusNotFound, 20008, "no such object")
	ErrRepeatedBucket      = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20009, "repeated bucket")
	ErrRepeatedObject      = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20010, "repeated object")
	ErrNoPermission        = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20011, "no permission")

	ErrBadSignature           = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20012, "bad signature")
	ErrSignedMsgFormat        = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20013, "signed msg must be formatted as ${actionContent}_${expiredTimestamp}")
	ErrExpiredTimestampFormat = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20014, "expiredTimestamp in signed msg must be a unix epoch time in milliseconds")

	ErrConsensus = gfsperrors.Register(module.AuthorizationModularName, http.StatusInternalServerError, 25002, "server slipped away, try again later")
)

var _ module.Authorizer = &AuthorizeModular{}

type AuthorizeModular struct {
	baseApp *gfspapp.GfSpBaseApp
	scope   rcmgr.ResourceScope
}

func (a *AuthorizeModular) Name() string {
	return module.AuthorizationModularName
}

func (a *AuthorizeModular) Start(ctx context.Context) error {
	scope, err := a.baseApp.ResourceManager().OpenService(a.Name())
	if err != nil {
		return err
	}
	a.scope = scope
	return nil
}

func (a *AuthorizeModular) Stop(ctx context.Context) error {
	a.scope.Release()
	return nil
}

func (a *AuthorizeModular) ReserveResource(
	ctx context.Context,
	state *rcmgr.ScopeStat) (
	rcmgr.ResourceScopeSpan,
	error) {
	span, err := a.scope.BeginSpan()
	if err != nil {
		return nil, err
	}
	err = span.ReserveResources(state)
	if err != nil {
		return nil, err
	}
	return span, nil
}

func (a *AuthorizeModular) ReleaseResource(
	ctx context.Context,
	span rcmgr.ResourceScopeSpan) {
	span.Done()
}

const (
	OffChainAuthSigExpiryAgeInSec int32 = 60 * 5 // in 300 seconds
)

// GetAuthNonce get the auth nonce for which the Dapp or client can generate EDDSA key pairs.
func (a *AuthorizeModular) GetAuthNonce(ctx context.Context, req *gfspserver.GetAuthNonceRequest) (*spdb.OffChainAuthKey, error) {
	domain := req.Domain

	ctx = log.Context(ctx, req)
	authKey, err := a.baseApp.GfSpDB().GetAuthKey(req.AccountId, domain)
	if err != nil {
		log.CtxErrorw(ctx, "failed to GetAuthKey", "error", err)
		return nil, err
	}
	log.CtxInfow(ctx, "succeed to GetAuthNonce")
	return authKey, nil
}

// UpdateUserPublicKey updates the user public key once the Dapp or client generates the EDDSA key pairs.
func (a *AuthorizeModular) UpdateUserPublicKey(ctx context.Context, req *gfspserver.UpdateUserPublicKeyRequest) (bool, error) {
	err := a.baseApp.GfSpDB().UpdateAuthKey(req.AccountId, req.Domain, req.CurrentNonce, req.Nonce, req.UserPublicKey, time.UnixMilli(req.ExpiryDate))
	if err != nil {
		log.Errorw("failed to updateUserPublicKey when saving key")
		return false, err
	}
	log.CtxInfow(ctx, "succeed to UpdateUserPublicKey")
	return true, nil
}

// VerifyOffChainSignature verifies the signature signed by user's EDDSA private key.
func (a *AuthorizeModular) VerifyOffChainSignature(ctx context.Context, req *gfspserver.VerifyOffChainSignatureRequest) (bool, error) {

	signedMsg := req.RealMsgToSign
	sigString := req.OffChainSig

	signature, err := hex.DecodeString(sigString)
	if err != nil {
		return false, ErrBadSignature
	}

	getAuthNonceReq := &gfspserver.GetAuthNonceRequest{
		AccountId: req.AccountId,
		Domain:    req.Domain,
	}
	getAuthNonceResp, err := a.GetAuthNonce(ctx, getAuthNonceReq)
	if err != nil {
		return false, err
	}
	userPublicKey := getAuthNonceResp.CurrentPublicKey

	// signedMsg must be formatted as `${actionContent}_${expiredTimestamp}` and timestamp must be within $OffChainAuthSigExpiryAgeInSec seconds, actionContent could be any string
	signedMsgParts := strings.Split(signedMsg, "_")
	if len(signedMsgParts) < 2 {
		err = fmt.Errorf("signed msg must be formated as ${actionContent}_${expiredTimestamp}")
		return false, ErrSignedMsgFormat
	}

	signedMsgExpiredTimestamp, err := strconv.Atoi(signedMsgParts[len(signedMsgParts)-1])
	if err != nil {
		err = fmt.Errorf("expiredTimestamp in signed msg must be a unix epoch time in milliseconds")
		return false, ErrExpiredTimestampFormat
	}
	expiredAge := time.Until(time.UnixMilli(int64(signedMsgExpiredTimestamp))).Seconds()

	if float64(OffChainAuthSigExpiryAgeInSec) < expiredAge || expiredAge < 0 { // nonce must be the same as NextNonce
		err = fmt.Errorf("expiredTimestamp in signed msg must be within %d seconds. ExpiredTimestamp in sig is %d, while the current server timestamp is %d ", OffChainAuthSigExpiryAgeInSec, signedMsgExpiredTimestamp, time.Now().UnixMilli())
		return false, gfsperrors.MakeGfSpError(err)
	}

	err = VerifyEddsaSignature(userPublicKey, signature, []byte(signedMsg))
	if err != nil {
		return false, gfsperrors.MakeGfSpError(err)
	}

	log.CtxInfow(ctx, "succeed to VerifyOffChainSignature")
	return true, nil
}

// VerifyAuthorize verifies the account has the operation's permission.
// TODO:: supports permission path verification and query
func (a *AuthorizeModular) VerifyAuthorize(
	ctx context.Context,
	authType coremodule.AuthOpType,
	account, bucket, object string) (
	bool, error) {
	has, err := a.baseApp.Consensus().HasAccount(ctx, account)
	if err != nil {
		log.CtxErrorw(ctx, "failed to check account from consensus", "error", err)
		return false, ErrConsensus
	}
	if !has {
		log.CtxErrorw(ctx, "no such account from consensus")
		return false, ErrNoSuchAccount
	}

	switch authType {
	case coremodule.AuthOpAskCreateBucketApproval:
		bucketInfo, _ := a.baseApp.Consensus().QueryBucketInfo(ctx, bucket)
		if bucketInfo != nil {
			log.CtxErrorw(ctx, "failed to verify authorize of asking create bucket "+
				"approval, bucket repeated", "bucket", bucket)
			return false, ErrRepeatedBucket
		}
		return true, nil
	case coremodule.AuthOpAskCreateObjectApproval:
		bucketInfo, objectInfo, _ := a.baseApp.Consensus().QueryBucketInfoAndObjectInfo(ctx, bucket, object)
		if bucketInfo == nil {
			log.CtxErrorw(ctx, "failed to verify authorize of asking create object "+
				"approval, no such bucket to ask create object approval", "bucket", bucket, "object", object)
			return false, ErrNoSuchBucket
		}
		if objectInfo != nil {
			log.CtxErrorw(ctx, "failed to verify authorize of asking create object "+
				"approval, object has been created", "bucket", bucket, "object", object)
			return false, ErrRepeatedObject
		}
		return true, nil
	case coremodule.AuthOpTypePutObject:
		bucketInfo, objectInfo, err := a.baseApp.Consensus().QueryBucketInfoAndObjectInfo(ctx, bucket, object)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get bucket and object info from consensus", "error", err)
			// refer to https://github.com/bnb-chain/greenfield/blob/master/x/storage/types/errors.go
			if strings.Contains(err.Error(), "No such bucket") {
				return false, ErrNoSuchBucket
			}
			if strings.Contains(err.Error(), "No such object") {
				return false, ErrNoSuchObject
			}
			return false, ErrConsensus
		}
		if bucketInfo.GetPrimarySpAddress() != a.baseApp.OperateAddress() {
			log.CtxErrorw(ctx, "sp operator address mismatch", "current", a.baseApp.OperateAddress(),
				"require", bucketInfo.GetPrimarySpAddress())
			return false, ErrMismatchSp
		}
		if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED {
			log.CtxErrorw(ctx, "object state is not sealed", "state", objectInfo.GetObjectStatus())
			return false, ErrNotCreatedState
		}
		allow, err := a.baseApp.Consensus().VerifyPutObjectPermission(ctx, account, bucket, object)
		if err != nil {
			log.CtxErrorw(ctx, "failed to verify put object permission from consensus", "error", err)
			return false, ErrConsensus
		}
		return allow, nil
	case coremodule.AuthOpTypeGetUploadingState:
		bucketInfo, objectInfo, err := a.baseApp.Consensus().QueryBucketInfoAndObjectInfo(ctx, bucket, object)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get bucket and object info from consensus", "error", err)
			// refer to https://github.com/bnb-chain/greenfield/blob/master/x/storage/types/errors.go
			if strings.Contains(err.Error(), "No such bucket") {
				return false, ErrNoSuchBucket
			}
			if strings.Contains(err.Error(), "No such object") {
				return false, ErrNoSuchObject
			}
			return false, ErrConsensus
		}
		if bucketInfo.GetPrimarySpAddress() != a.baseApp.OperateAddress() {
			log.CtxErrorw(ctx, "sp operator address mismatch", "current", a.baseApp.OperateAddress(),
				"require", bucketInfo.GetPrimarySpAddress())
			return false, ErrMismatchSp
		}
		if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED {
			log.CtxErrorw(ctx, "object state is not created", "state", objectInfo.GetObjectStatus())
			return false, ErrNotCreatedState
		}
		allow, err := a.baseApp.Consensus().VerifyPutObjectPermission(ctx, account, bucket, object)
		if err != nil {
			log.CtxErrorw(ctx, "failed to verify put object permission from consensus", "error", err)
			return false, ErrConsensus
		}
		return allow, nil
	case coremodule.AuthOpTypeGetObject:
		bucketInfo, objectInfo, err := a.baseApp.Consensus().QueryBucketInfoAndObjectInfo(ctx, bucket, object)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get bucket and object info from consensus", "error", err)
			// refer to https://github.com/bnb-chain/greenfield/blob/master/x/storage/types/errors.go
			if strings.Contains(err.Error(), "No such bucket") {
				return false, ErrNoSuchBucket
			}
			if strings.Contains(err.Error(), "No such object") {
				return false, ErrNoSuchObject
			}
			return false, ErrConsensus
		}
		if bucketInfo.GetPrimarySpAddress() != a.baseApp.OperateAddress() {
			log.CtxErrorw(ctx, "sp operator address mismatch", "current", a.baseApp.OperateAddress(),
				"require", bucketInfo.GetPrimarySpAddress())
			return false, ErrMismatchSp
		}
		if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
			log.CtxErrorw(ctx, "object state is not sealed", "state", objectInfo.GetObjectStatus())
			return false, ErrNotSealedState
		}
		streamRecord, err := a.baseApp.Consensus().QueryPaymentStreamRecord(ctx, bucketInfo.GetPaymentAddress())
		if err != nil {
			log.CtxErrorw(ctx, "failed to query payment stream record from consensus", "error", err)
			return false, ErrConsensus
		}
		if streamRecord.Status != paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE {
			log.CtxErrorw(ctx, "failed to check payment due to account status is not active", "status", streamRecord.Status)
			return false, ErrPaymentState
		}
		allow, err := a.baseApp.Consensus().VerifyGetObjectPermission(ctx, account, bucket, object)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get bucket and object info from consensus", "error", err)
			return false, ErrConsensus
		}
		return allow, nil
	case coremodule.AuthOpTypeGetBucketQuota, coremodule.AuthOpTypeListBucketReadRecord:
		bucketInfo, err := a.baseApp.Consensus().QueryBucketInfo(ctx, bucket)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get bucket info from consensus", "error", err)
			// refer to https://github.com/bnb-chain/greenfield/blob/master/x/storage/types/errors.go
			if strings.Contains(err.Error(), "No such bucket") {
				return false, ErrNoSuchBucket
			}
			return false, ErrConsensus
		}
		if bucketInfo.GetPrimarySpAddress() != a.baseApp.OperateAddress() {
			log.CtxErrorw(ctx, "sp operator address mismatch", "current", a.baseApp.OperateAddress(),
				"require", bucketInfo.GetPrimarySpAddress())
			return false, ErrMismatchSp
		}
		if bucketInfo.GetOwner() != account {
			log.CtxErrorw(ctx, "only owner can get bucket quota", "current", account,
				"bucket_owner", bucketInfo.GetOwner())
			return false, ErrNoPermission
		}
		return true, nil
	case coremodule.AuthOpTypeChallengePiece:
		//TODO:: only validator has permission to challenge piece
		bucketInfo, objectInfo, err := a.baseApp.Consensus().QueryBucketInfoAndObjectInfo(ctx, bucket, object)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get object info from consensus", "error", err)
			// refer to https://github.com/bnb-chain/greenfield/blob/master/x/storage/types/errors.go
			if strings.Contains(err.Error(), "No such bucket") {
				return false, ErrNoSuchBucket
			}
			if strings.Contains(err.Error(), "No such object") {
				return false, ErrNoSuchObject
			}
			return false, ErrConsensus
		}
		if bucketInfo.GetPrimarySpAddress() == a.baseApp.OperateAddress() {
			return true, nil
		}
		for _, address := range objectInfo.GetSecondarySpAddresses() {
			if address == a.baseApp.OperateAddress() {
				return true, nil
			}
		}
		log.CtxErrorw(ctx, "sp operator address mismatch", "current", a.baseApp.OperateAddress())
		return false, ErrMismatchSp
	default:
		return false, ErrUnsupportedAuthType
	}
}

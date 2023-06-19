package authenticator

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

var (
	ErrUnsupportedAuthType = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20001, "unsupported auth op type")
	ErrMismatchSp          = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20002, "mismatched primary sp")
	ErrNotCreatedState     = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20003, "object has not been created state")
	ErrNotSealedState      = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20004, "object has not been sealed state")
	ErrPaymentState        = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20005, "payment account is not active")
	ErrNoSuchAccount       = gfsperrors.Register(module.AuthenticationModularName, http.StatusNotFound, 20006, "no such account")
	ErrNoSuchBucket        = gfsperrors.Register(module.AuthenticationModularName, http.StatusNotFound, 20007, "no such bucket")
	ErrNoSuchObject        = gfsperrors.Register(module.AuthenticationModularName, http.StatusNotFound, 20008, "no such object")
	ErrRepeatedBucket      = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20009, "repeated bucket")
	ErrRepeatedObject      = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20010, "repeated object")
	ErrNoPermission        = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20011, "no permission")

	ErrBadSignature           = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20012, "bad signature")
	ErrSignedMsgFormat        = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20013, "signed msg must be formatted as ${actionContent}_${expiredTimestamp}")
	ErrExpiredTimestampFormat = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20014, "expiredTimestamp in signed msg must be a unix epoch time in milliseconds")
	ErrPublicKeyExpired       = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20015, "user public key is expired")

	ErrConsensus = gfsperrors.Register(module.AuthenticationModularName, http.StatusInternalServerError, 25002, "server slipped away, try again later")
)

var _ module.Authenticator = &AuthenticationModular{}

type AuthenticationModular struct {
	baseApp *gfspapp.GfSpBaseApp
	scope   rcmgr.ResourceScope
}

func (a *AuthenticationModular) Name() string {
	return module.AuthenticationModularName
}

func (a *AuthenticationModular) Start(ctx context.Context) error {
	scope, err := a.baseApp.ResourceManager().OpenService(a.Name())
	if err != nil {
		return err
	}
	a.scope = scope
	return nil
}

func (a *AuthenticationModular) Stop(ctx context.Context) error {
	a.scope.Release()
	return nil
}

func (a *AuthenticationModular) ReserveResource(
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

func (a *AuthenticationModular) ReleaseResource(
	ctx context.Context,
	span rcmgr.ResourceScopeSpan) {
	span.Done()
}

const (
	OffChainAuthSigExpiryAgeInSec int32 = 60 * 5 // in 300 seconds
)

// GetAuthNonce get the auth nonce for which the Dapp or client can generate EDDSA key pairs.
func (a *AuthenticationModular) GetAuthNonce(ctx context.Context, account string, domain string) (*spdb.OffChainAuthKey, error) {
	authKey, err := a.baseApp.GfSpDB().GetAuthKey(account, domain)
	if err != nil {
		log.CtxErrorw(ctx, "failed to GetAuthKey", "error", err)
		return nil, err
	}
	log.CtxInfow(ctx, "succeed to GetAuthNonce")
	return authKey, nil
}

// UpdateUserPublicKey updates the user public key once the Dapp or client generates the EDDSA key pairs.
func (a *AuthenticationModular) UpdateUserPublicKey(ctx context.Context, account string, domain string, currentNonce int32, nonce int32, userPublicKey string, expiryDate int64) (bool, error) {
	err := a.baseApp.GfSpDB().UpdateAuthKey(account, domain, currentNonce, nonce, userPublicKey, time.UnixMilli(expiryDate))
	if err != nil {
		log.CtxErrorw(ctx, "failed to updateUserPublicKey when saving key")
		return false, err
	}
	log.CtxInfow(ctx, "succeed to UpdateUserPublicKey")
	return true, nil
}

// VerifyOffChainSignature verifies the signature signed by user's EDDSA private key.
func (a *AuthenticationModular) VerifyOffChainSignature(ctx context.Context, account string, domain string, offChainSig string, realMsgToSign string) (bool, error) {
	signature, err := hex.DecodeString(offChainSig)
	if err != nil {
		return false, ErrBadSignature
	}

	getAuthNonceResp, err := a.GetAuthNonce(ctx, account, domain)
	if err != nil {
		return false, err
	}
	if time.Until(getAuthNonceResp.ExpiryDate).Seconds() < 0 {
		return false, ErrPublicKeyExpired
	}
	userPublicKey := getAuthNonceResp.CurrentPublicKey

	// signedMsg must be formatted as `${actionContent}_${expiredTimestamp}` and timestamp must be within $OffChainAuthSigExpiryAgeInSec seconds, actionContent could be any string
	signedMsgParts := strings.Split(realMsgToSign, "_")
	if len(signedMsgParts) < 2 {
		log.CtxErrorw(ctx, "signed msg must be formated as ${actionContent}_${expiredTimestamp}")
		return false, ErrSignedMsgFormat
	}

	signedMsgExpiredTimestamp, err := strconv.Atoi(signedMsgParts[len(signedMsgParts)-1])
	if err != nil {
		log.CtxErrorw(ctx, "expiredTimestamp in signed msg must be a unix epoch time in milliseconds")
		return false, ErrExpiredTimestampFormat
	}
	expiredAge := time.Until(time.UnixMilli(int64(signedMsgExpiredTimestamp))).Seconds()

	if float64(OffChainAuthSigExpiryAgeInSec) < expiredAge || expiredAge < 0 { // nonce must be the same as NextNonce
		err = fmt.Errorf("expiredTimestamp in signed msg must be within %d seconds. ExpiredTimestamp in sig is %d, while the current server timestamp is %d ", OffChainAuthSigExpiryAgeInSec, signedMsgExpiredTimestamp, time.Now().UnixMilli())
		return false, gfsperrors.MakeGfSpError(err)
	}

	err = VerifyEddsaSignature(userPublicKey, signature, []byte(realMsgToSign))
	if err != nil {
		return false, gfsperrors.MakeGfSpError(err)
	}

	log.CtxInfow(ctx, "succeed to VerifyOffChainSignature")
	return true, nil
}

// VerifyAuthentication verifies the account has the operation's permission.
// TODO:: supports permission path verification and query
func (a *AuthenticationModular) VerifyAuthentication(
	ctx context.Context,
	authType coremodule.AuthOpType,
	account, bucket, object string) (
	bool, error) {
	startTime := time.Now()
	has, err := a.baseApp.Consensus().HasAccount(ctx, account)
	metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_check_has_account_time").Observe(time.Since(startTime).Seconds())
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
		queryTime := time.Now()
		bucketInfo, _ := a.baseApp.Consensus().QueryBucketInfo(ctx, bucket)
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_create_bucket_approval_query_bucket_time").Observe(time.Since(queryTime).Seconds())
		if bucketInfo != nil {
			log.CtxErrorw(ctx, "failed to verify authentication of asking create bucket "+
				"approval, bucket repeated", "bucket", bucket)
			return false, ErrRepeatedBucket
		}
		return true, nil
	case coremodule.AuthOpAskCreateObjectApproval:
		queryTime := time.Now()
		bucketInfo, objectInfo, _ := a.baseApp.Consensus().QueryBucketInfoAndObjectInfo(ctx, bucket, object)
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_create_object_approval_query_bucket_object_time").Observe(time.Since(queryTime).Seconds())
		if bucketInfo == nil {
			log.CtxErrorw(ctx, "failed to verify authentication of asking create object "+
				"approval, no such bucket to ask create object approval", "bucket", bucket, "object", object)
			return false, ErrNoSuchBucket
		}
		if objectInfo != nil {
			log.CtxErrorw(ctx, "failed to verify authentication of asking create object "+
				"approval, object has been created", "bucket", bucket, "object", object)
			return false, ErrRepeatedObject
		}
		return true, nil
	case coremodule.AuthOpTypePutObject:
		queryTime := time.Now()
		bucketInfo, objectInfo, err := a.baseApp.Consensus().QueryBucketInfoAndObjectInfo(ctx, bucket, object)
		// TODO:
		_ = bucketInfo
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_put_object_query_bucket_object_time").Observe(time.Since(queryTime).Seconds())
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
		// TODO:
		//if bucketInfo.GetPrimarySpAddress() != a.baseApp.OperatorAddress() {
		//	log.CtxErrorw(ctx, "sp operator address mismatch", "current", a.baseApp.OperatorAddress(),
		//		"require", bucketInfo.GetPrimarySpAddress())
		//	return false, ErrMismatchSp
		//}
		if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED {
			log.CtxErrorw(ctx, "object state is not sealed", "state", objectInfo.GetObjectStatus())
			return false, ErrNotCreatedState
		}
		permissionTime := time.Now()
		allow, err := a.baseApp.Consensus().VerifyPutObjectPermission(ctx, account, bucket, object)
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_put_object_verify_permission_time").Observe(time.Since(permissionTime).Seconds())
		if err != nil {
			log.CtxErrorw(ctx, "failed to verify put object permission from consensus", "error", err)
			return false, ErrConsensus
		}
		return allow, nil
	case coremodule.AuthOpTypeGetUploadingState:
		queryTime := time.Now()
		bucketInfo, objectInfo, err := a.baseApp.Consensus().QueryBucketInfoAndObjectInfo(ctx, bucket, object)
		// TODO:
		_ = bucketInfo
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_get_object_process_query_bucket_object_time").Observe(time.Since(queryTime).Seconds())
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
		// TODO:
		//if bucketInfo.GetPrimarySpAddress() != a.baseApp.OperatorAddress() {
		//	log.CtxErrorw(ctx, "sp operator address mismatch", "current", a.baseApp.OperatorAddress(),
		//		"require", bucketInfo.GetPrimarySpAddress())
		//	return false, ErrMismatchSp
		//}
		if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED {
			log.CtxErrorw(ctx, "object state is not created", "state", objectInfo.GetObjectStatus())
			return false, ErrNotCreatedState
		}
		permissionTime := time.Now()
		allow, err := a.baseApp.Consensus().VerifyPutObjectPermission(ctx, account, bucket, object)
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_get_object_process_verify_permission_time").Observe(time.Since(permissionTime).Seconds())
		if err != nil {
			log.CtxErrorw(ctx, "failed to verify put object permission from consensus", "error", err)
			return false, ErrConsensus
		}
		return allow, nil
	case coremodule.AuthOpTypeGetObject:
		queryTime := time.Now()
		bucketInfo, objectInfo, err := a.baseApp.Consensus().QueryBucketInfoAndObjectInfo(ctx, bucket, object)
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_get_object_query_bucket_object_time").Observe(time.Since(queryTime).Seconds())
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
		//if bucketInfo.GetPrimarySpAddress() != a.baseApp.OperatorAddress() {
		//	log.CtxErrorw(ctx, "sp operator address mismatch", "current", a.baseApp.OperatorAddress(),
		//		"require", bucketInfo.GetPrimarySpAddress())
		//	return false, ErrMismatchSp
		//}
		if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
			log.CtxErrorw(ctx, "object state is not sealed", "state", objectInfo.GetObjectStatus())
			return false, ErrNotSealedState
		}
		streamTime := time.Now()
		streamRecord, err := a.baseApp.Consensus().QueryPaymentStreamRecord(ctx, bucketInfo.GetPaymentAddress())
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_get_object_query_stream_time").Observe(time.Since(streamTime).Seconds())
		if err != nil {
			log.CtxErrorw(ctx, "failed to query payment stream record from consensus", "error", err)
			return false, ErrConsensus
		}
		if streamRecord.Status != paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE {
			log.CtxErrorw(ctx, "failed to check payment due to account status is not active", "status", streamRecord.Status)
			return false, ErrPaymentState
		}
		permissionTime := time.Now()
		allow, err := a.baseApp.Consensus().VerifyGetObjectPermission(ctx, account, bucket, object)
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_get_object_verify_permission_time").Observe(time.Since(permissionTime).Seconds())
		if err != nil {
			log.CtxErrorw(ctx, "failed to get bucket and object info from consensus", "error", err)
			return false, ErrConsensus
		}
		return allow, nil
	case coremodule.AuthOpTypeGetBucketQuota, coremodule.AuthOpTypeListBucketReadRecord:
		queryTime := time.Now()
		bucketInfo, err := a.baseApp.Consensus().QueryBucketInfo(ctx, bucket)
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_get_bucket_quota_query_bucket_time").Observe(time.Since(queryTime).Seconds())
		if err != nil {
			log.CtxErrorw(ctx, "failed to get bucket info from consensus", "error", err)
			// refer to https://github.com/bnb-chain/greenfield/blob/master/x/storage/types/errors.go
			if strings.Contains(err.Error(), "No such bucket") {
				return false, ErrNoSuchBucket
			}
			return false, ErrConsensus
		}
		//if bucketInfo.GetPrimarySpAddress() != a.baseApp.OperatorAddress() {
		//	log.CtxErrorw(ctx, "sp operator address mismatch", "current", a.baseApp.OperatorAddress(),
		//		"require", bucketInfo.GetPrimarySpAddress())
		//	return false, ErrMismatchSp
		//}
		if bucketInfo.GetOwner() != account {
			log.CtxErrorw(ctx, "only owner can get bucket quota", "current", account,
				"bucket_owner", bucketInfo.GetOwner())
			return false, ErrNoPermission
		}
		return true, nil
	case coremodule.AuthOpTypeGetChallengePieceInfo:
		challengeIsFromValidator := false
		queryTime := time.Now()
		validators, err := a.baseApp.Consensus().ListBondedValidators(ctx)
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_challenge_query_validator_time").Observe(time.Since(queryTime).Seconds())
		if err != nil {
			log.CtxErrorw(ctx, "failed to list validator from consensus", "error", err)
			return false, ErrConsensus
		}
		for _, validator := range validators {
			if strings.EqualFold(validator.ChallengerAddress, account) {
				challengeIsFromValidator = true
				break
			}
		}
		if !challengeIsFromValidator {
			log.CtxErrorw(ctx, "failed to get challenge info due to it is not in bonded validator list",
				"actual_challenge_address", account)
			return false, ErrNoPermission
		}
		queryTime = time.Now()
		bucketInfo, objectInfo, err := a.baseApp.Consensus().QueryBucketInfoAndObjectInfo(ctx, bucket, object)
		// TODO:
		_ = bucketInfo
		_ = objectInfo
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_challenge_query_bucket_object_time").Observe(time.Since(queryTime).Seconds())
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
		// TODO:
		//if strings.EqualFold(bucketInfo.GetPrimarySpAddress(), a.baseApp.OperatorAddress()) {
		//	return true, nil
		//}
		//for _, address := range objectInfo.GetSecondarySpAddresses() {
		//	if strings.EqualFold(address, a.baseApp.OperatorAddress()) {
		//		return true, nil
		//	}
		//}
		log.CtxErrorw(ctx, "sp operator address mismatch", "current", a.baseApp.OperatorAddress())
		return false, ErrMismatchSp
	default:
		return false, ErrUnsupportedAuthType
	}
}

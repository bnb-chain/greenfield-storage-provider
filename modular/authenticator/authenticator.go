package authenticator

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	permissiontypes "github.com/bnb-chain/greenfield/x/permission/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	ErrUnsupportedAuthType    = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20001, "unsupported auth op type")
	ErrMismatchSp             = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20002, "mismatched primary sp")
	ErrUnexpectedObjectStatus = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20003, "")
	ErrNotSealedState         = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20004, "object has not been sealed state")
	ErrPaymentState           = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20005, "payment account is not active")
	ErrInvalidAddress         = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20006, "the user address format is invalid")
	ErrNoSuchBucket           = gfsperrors.Register(module.AuthenticationModularName, http.StatusNotFound, 20007, "no such bucket")
	ErrNoSuchObject           = gfsperrors.Register(module.AuthenticationModularName, http.StatusNotFound, 20008, "no such object")
	ErrRepeatedBucket         = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20009, "repeated bucket")
	ErrRepeatedObject         = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20010, "repeated object")
	ErrNoPermission           = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20011, "no permission")

	ErrBadSignature           = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20012, "bad signature")
	ErrSignedMsgFormat        = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20013, "signed msg must be formatted as ${actionContent}_${expiredTimestamp}")
	ErrExpiredTimestampFormat = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20014, "expiredTimestamp in signed msg must be a unix epoch time in milliseconds")
	ErrPublicKeyExpired       = gfsperrors.Register(module.AuthenticationModularName, http.StatusBadRequest, 20015, "user public key is expired")
)

func ErrUnexpectedObjectStatusWithDetail(objectName string, expectedStatus storagetypes.ObjectStatus, actualStatus storagetypes.ObjectStatus) *gfsperrors.GfSpError {
	return &gfsperrors.GfSpError{
		CodeSpace:      module.AuthenticationModularName,
		HttpStatusCode: int32(http.StatusBadRequest),
		InnerCode:      int32(20003),
		Description:    fmt.Sprintf("object %s is expected to be %s status but actually %s status", objectName, expectedStatus, actualStatus),
	}
}

func ErrConsensusWithDetail(detail string) *gfsperrors.GfSpError {
	return gfsperrors.Register(module.AuthenticationModularName, http.StatusInternalServerError, 25002, detail)
}

var _ module.Authenticator = &AuthenticationModular{}

type AuthenticationModular struct {
	baseApp *gfspapp.GfSpBaseApp
	scope   rcmgr.ResourceScope
	spID    uint32
}

func (a *AuthenticationModular) getSPID() (uint32, error) {
	if a.spID != 0 {
		return a.spID, nil
	}
	spInfo, err := a.baseApp.Consensus().QuerySP(context.Background(), a.baseApp.OperatorAddress())
	if err != nil {
		return 0, err
	}
	a.spID = spInfo.GetId()
	return a.spID, nil
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

// VerifyGNFD1EddsaSignature verifies the signature signed by user's EDDSA private key.
// no need to verify if the sig is expired.  This method only need verify the account address and leave the expiration checking to gateway.
func (a *AuthenticationModular) VerifyGNFD1EddsaSignature(ctx context.Context, account string, domain string, offChainSig string, realMsgToSign []byte) (bool, error) {
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

	err = VerifyEddsaSignature(userPublicKey, signature, realMsgToSign)
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
	// check the account if it is a valid address
	_, err := sdk.AccAddressFromHexUnsafe(account)
	if err != nil {
		return false, ErrInvalidAddress
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
	case coremodule.AuthOpAskMigrateBucketApproval:
		queryTime := time.Now()
		bucketInfo, _ := a.baseApp.Consensus().QueryBucketInfo(ctx, bucket)
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_migrate_bucket_approval_query_bucket_time").Observe(time.Since(queryTime).Seconds())
		if bucketInfo == nil {
			log.CtxErrorw(ctx, "failed to verify authentication of asking migrate bucket "+
				"approval, bucket not existed", "bucket", bucket)
			return false, ErrNoSuchBucket
		}
		return true, nil
	case coremodule.AuthOpTypeQueryBucketMigrationProgress:
		queryTime := time.Now()
		bucketInfo, _ := a.baseApp.Consensus().QueryBucketInfo(ctx, bucket)
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_migrate_bucket_progress_query_bucket_time").Observe(time.Since(queryTime).Seconds())
		if bucketInfo == nil {
			log.CtxErrorw(ctx, "failed to verify authentication of asking migrate bucket "+
				"progress, bucket not existed", "bucket", bucket)
			return false, ErrNoSuchBucket
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
			return false, ErrConsensusWithDetail("failed to get bucket and object info from consensus, error: " + err.Error())
		}

		spID, err := a.getSPID()
		if err != nil {
			return false, ErrConsensusWithDetail("getSPID error: " + err.Error())
		}
		bucketSPID, err := util.GetBucketPrimarySPID(ctx, a.baseApp.Consensus(), bucketInfo)
		if err != nil {
			return false, ErrConsensusWithDetail("GetBucketPrimarySPID error: " + err.Error())
		}
		if bucketSPID != spID {
			log.CtxErrorw(ctx, "sp operator address mismatch", "actual_sp_id", spID,
				"expected_sp_id", bucketSPID)
			return false, ErrMismatchSp
		}
		if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED {
			log.CtxErrorw(ctx, "object state should be OBJECT_STATUS_CREATED", "state", objectInfo.GetObjectStatus())
			return false, ErrUnexpectedObjectStatusWithDetail(objectInfo.ObjectName, storagetypes.OBJECT_STATUS_CREATED, objectInfo.GetObjectStatus())
		}
		permissionTime := time.Now()
		allow, err := a.baseApp.Consensus().VerifyPutObjectPermission(ctx, account, bucket, object)
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_put_object_verify_permission_time").Observe(time.Since(permissionTime).Seconds())
		if err != nil {
			log.CtxErrorw(ctx, "failed to verify put object permission from consensus", "error", err)
			return false, err
		}
		return allow, nil
	case coremodule.AuthOpTypeGetUploadingState:
		queryTime := time.Now()
		bucketInfo, _, err := a.baseApp.Consensus().QueryBucketInfoAndObjectInfo(ctx, bucket, object)
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
			return false, ErrConsensusWithDetail("failed to get bucket and object info from consensus, error: " + err.Error())
		}
		spID, err := a.getSPID()
		if err != nil {
			return false, ErrConsensusWithDetail("getSPID error: " + err.Error())
		}
		bucketSPID, err := util.GetBucketPrimarySPID(ctx, a.baseApp.Consensus(), bucketInfo)
		if err != nil {
			return false, ErrConsensusWithDetail("GetBucketPrimarySPID error: " + err.Error())
		}
		if bucketSPID != spID {
			log.CtxErrorw(ctx, "sp operator address mismatch", "actual_sp_id", spID,
				"expected_sp_id", bucketSPID)
			return false, ErrMismatchSp
		}
		return true, nil
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
			return false, ErrConsensusWithDetail("failed to get bucket and object info from consensus, error: " + err.Error())
		}
		spID, err := a.getSPID()
		if err != nil {
			return false, ErrConsensusWithDetail("getSPID error: " + err.Error())
		}
		bucketSPID, err := util.GetBucketPrimarySPID(ctx, a.baseApp.Consensus(), bucketInfo)
		if err != nil {
			return false, ErrConsensusWithDetail("GetBucketPrimarySPID error: " + err.Error())
		}
		if bucketSPID != spID {
			log.CtxErrorw(ctx, "sp operator address mismatch", "actual_sp_id", spID,
				"expected_sp_id", bucketSPID)
			return false, ErrMismatchSp
		}
		if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED &&
			objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED {
			log.CtxErrorw(ctx, "object state is not sealed or created", "state", objectInfo.GetObjectStatus())
			return false, ErrNoPermission
		}
		streamTime := time.Now()
		streamRecord, err := a.baseApp.Consensus().QueryPaymentStreamRecord(ctx, bucketInfo.GetPaymentAddress())
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_get_object_query_stream_time").Observe(time.Since(streamTime).Seconds())
		if err != nil {
			log.CtxErrorw(ctx, "failed to query payment stream record from consensus", "error", err)
			return false, err
		}
		if streamRecord.Status != paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE {
			log.CtxErrorw(ctx, "failed to check payment due to account status is not active", "status", streamRecord.Status)
			return false, ErrPaymentState
		}
		permissionTime := time.Now()
		var permission *permissiontypes.Effect
		var allow bool
		permission, err = a.baseApp.GfSpClient().VerifyPermission(ctx, account, bucket, object, permissiontypes.ACTION_GET_OBJECT)
		metrics.PerfAuthTimeHistogram.WithLabelValues("auth_server_get_object_verify_permission_time").Observe(time.Since(permissionTime).Seconds())
		if err != nil {
			log.CtxErrorw(ctx, "failed to get bucket and object info from meta", "error", err)
			return false, err
		}
		if *permission == permissiontypes.EFFECT_ALLOW {
			allow = true
		}
		return allow, nil
	case coremodule.AuthOpTypeGetRecoveryPiece:
		bucketInfo, objectInfo, err := a.baseApp.Consensus().QueryBucketInfoAndObjectInfo(ctx, bucket, object)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get bucket and object info from consensus", "error", err)
			if strings.Contains(err.Error(), "No such bucket") {
				return false, ErrNoSuchBucket
			}
			if strings.Contains(err.Error(), "No such object") {
				return false, ErrNoSuchObject
			}
			return false, ErrConsensusWithDetail("failed to get bucket and object info from consensus, error: " + err.Error())
		}
		spID, err := a.getSPID()
		if err != nil {
			return false, ErrConsensusWithDetail("getSPID error: " + err.Error())
		}
		_, isSecondarySp, err := util.ValidateAndGetSPIndexWithinGVGSecondarySPs(ctx, a.baseApp.GfSpClient(), spID, bucketInfo.Id.Uint64(), objectInfo.LocalVirtualGroupId)
		if err != nil {
			log.CtxErrorw(ctx, "failed to global virtual group info from metaData", "error", err)
			return false, ErrConsensusWithDetail("failed to global virtual group info from metaData, error: " + err.Error())
		}
		if !isSecondarySp {
			return false, ErrMismatchSp
		}
		if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
			log.CtxErrorw(ctx, "object state is not sealed", "state", objectInfo.GetObjectStatus())
			return false, ErrNotSealedState
		}

		var permission *permissiontypes.Effect
		var allow bool
		permission, err = a.baseApp.GfSpClient().VerifyPermission(ctx, account, bucket, object, permissiontypes.ACTION_GET_OBJECT)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get bucket and object info from meta", "error", err)
			return false, err
		}
		if *permission == permissiontypes.EFFECT_ALLOW {
			allow = true
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
			return false, ErrConsensusWithDetail("failed to get bucket info from consensus, error: " + err.Error())
		}
		spID, err := a.getSPID()
		if err != nil {
			return false, ErrConsensusWithDetail("getSPID error: " + err.Error())
		}
		bucketSPID, err := util.GetBucketPrimarySPID(ctx, a.baseApp.Consensus(), bucketInfo)
		if err != nil {
			return false, ErrConsensusWithDetail("GetBucketPrimarySPID error: " + err.Error())
		}
		if bucketSPID != spID {
			log.CtxErrorw(ctx, "sp operator address mismatch", "actual_sp_id", spID,
				"expected_sp_id", bucketSPID)
			return false, ErrMismatchSp
		}
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
			return false, ErrConsensusWithDetail("failed to list validator from consensus, error: " + err.Error())
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
			return false, ErrConsensusWithDetail("failed to get object info from consensus, error: " + err.Error())
		}
		spID, err := a.getSPID()
		if err != nil {
			return false, ErrConsensusWithDetail("getSPID error: " + err.Error())
		}
		bucketSPID, err := util.GetBucketPrimarySPID(ctx, a.baseApp.Consensus(), bucketInfo)
		if err != nil {
			return false, ErrConsensusWithDetail("GetBucketPrimarySPID error: " + err.Error())
		}
		if bucketSPID == spID {
			return true, nil
		}
		_, isSecondarySp, err := util.ValidateAndGetSPIndexWithinGVGSecondarySPs(ctx, a.baseApp.GfSpClient(), spID, bucketInfo.Id.Uint64(), objectInfo.LocalVirtualGroupId)
		if err != nil {
			log.CtxErrorw(ctx, "failed to global virtual group info from metaData", "error", err)
			return false, ErrConsensusWithDetail("failed to global virtual group info from metaData, error: " + err.Error())
		}
		if !isSecondarySp {
			return false, ErrMismatchSp
		}
		return true, nil
	default:
		return false, ErrUnsupportedAuthType
	}
}

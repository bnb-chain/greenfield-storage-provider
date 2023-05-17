package authorizer

import (
	"context"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	ErrUnsupportedAuthType = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20001, "unsupported auth op type")
	ErrMismatchSp          = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20002, "mismatched primary sp")
	ErrNotCreatedState     = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20003, "object has not been created state")
	ErrNotSealedState      = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20004, "object has not been sealed")
	ErrPaymentState        = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20005, "payment account is not active")
	ErrNoSuchAccount       = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20006, "no such account")
	ErrNoSuchBucket        = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20007, "no such bucket")
	ErrRepeatedBucket      = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20008, "repeated bucket")
	ErrRepeatedObject      = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20009, "repeated object")
	ErrNoPermission        = gfsperrors.Register(module.AuthorizationModularName, http.StatusBadRequest, 20010, "no permission")
	ErrConsensus           = gfsperrors.Register(module.AuthorizationModularName, http.StatusInternalServerError, 25002, "server slipped away, try again later")
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

func (a *AuthorizeModular) ReserveResource(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
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

func (a *AuthorizeModular) ReleaseResource(ctx context.Context, span rcmgr.ResourceScopeSpan) {
	span.Done()
	return
}

func (a *AuthorizeModular) VerifyAuthorize(ctx context.Context,
	auth coremodule.AuthOpType, account, bucket, object string) (bool, error) {
	has, err := a.baseApp.Consensus().HasAccount(ctx, account)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket and object info from consensus", "error", err)
		return false, ErrConsensus
	}
	if !has {
		log.CtxErrorw(ctx, "no such account from consensus")
		return false, ErrNoSuchAccount
	}
	switch auth {
	case coremodule.AuthOpAskCreateBucketApproval:
		bucketInfo, _ := a.baseApp.Consensus().QueryBucketInfo(ctx, bucket)
		if bucketInfo != nil {
			log.CtxErrorw(ctx, "failed to ask create bucket approval, bucket repeated",
				"bucket", bucket, "object", object)
			return false, ErrRepeatedBucket
		}
		return true, nil
	case coremodule.AuthOpAskCreateObjectApproval:
		bucketInfo, objectInfo, _ := a.baseApp.Consensus().QueryBucketInfoAndObjectInfo(ctx, bucket, object)
		if bucketInfo == nil {
			log.CtxErrorw(ctx, "failed to ask create object approval, no such bucket to ask create object approval",
				"bucket", bucket, "object", object)
			return false, ErrNoSuchBucket
		}
		if objectInfo != nil {
			log.CtxErrorw(ctx, "failed to ask create object approval, object has been created",
				"bucket", bucket, "object", object)
			return false, ErrRepeatedObject
		}
		return true, nil
	case coremodule.AuthOpTypePutObject:
		bucketInfo, objectInfo, err := a.baseApp.Consensus().QueryBucketInfoAndObjectInfo(ctx, bucket, object)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get bucket and object info from consensus", "error", err)
			return false, ErrConsensus
		}
		if bucketInfo.GetPrimarySpAddress() != a.baseApp.OperateAddress() {
			log.CtxErrorw(ctx, "sp operator address mismatch", "current", a.baseApp.OperateAddress(),
				"require", bucketInfo.GetPrimarySpAddress())
			return false, ErrMismatchSp
		}
		if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED {
			log.CtxErrorw(ctx, "object state is noe create", "state", objectInfo.GetObjectStatus())
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
			return false, ErrConsensus
		}
		if bucketInfo.GetPrimarySpAddress() != a.baseApp.OperateAddress() {
			log.CtxErrorw(ctx, "sp operator address mismatch", "current", a.baseApp.OperateAddress(),
				"require", bucketInfo.GetPrimarySpAddress())
			return false, ErrMismatchSp
		}
		if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED {
			log.CtxErrorw(ctx, "object state is not create", "state", objectInfo.GetObjectStatus())
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
		streamRecord, err := a.baseApp.Consensus().QueryPaymentStreamRecord(ctx, bucketInfo.GetPrimarySpAddress())
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
			return false, ErrConsensus
		}
		if bucketInfo.GetPrimarySpAddress() != a.baseApp.OperateAddress() {
			log.CtxErrorw(ctx, "sp operator address mismatch", "current", a.baseApp.OperateAddress(),
				"require", bucketInfo.GetPrimarySpAddress())
			return false, ErrMismatchSp
		}
		if bucketInfo.GetOwner() != account {
			log.CtxErrorw(ctx, "only owner can get bucket qouta", "current", account,
				"bucket_owner", bucketInfo.GetOwner())
			return false, ErrNoPermission
		}
		return true, nil
	case coremodule.AuthOpTypeChallengePiece:
		//TODO:: only validator has permission to challenge piece
		bucketInfo, objectInfo, err := a.baseApp.Consensus().QueryBucketInfoAndObjectInfo(ctx, bucket, object)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get object info from consensus", "error", err)
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
	return false, ErrUnsupportedAuthType
}

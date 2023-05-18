package retriever

import (
	"context"
	"time"

	"cosmossdk.io/math"
	chaintypes "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/resource"
	gnfdresource "github.com/bnb-chain/greenfield/types/resource"
	"github.com/bnb-chain/greenfield/types/s3util"
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
	"github.com/bnb-chain/greenfield/x/storage/keeper"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/juno/v4/common"

	"github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

// GfSpVerifyPermission Verify the input account’s permission to input items
func (r *RetrieveModular) GfSpVerifyPermission(ctx context.Context, req *storagetypes.QueryVerifyPermissionRequest) (resp *storagetypes.QueryVerifyPermissionResponse, err error) {
	var (
		operator   sdk.AccAddress
		bucketInfo *bsdb.Bucket
		objectInfo *bsdb.Object
		effect     permtypes.Effect
	)

	ctx = log.Context(ctx, req)

	if req == nil {
		log.CtxErrorw(ctx, "invalid request", "error", err)
		return nil, errors.ErrInvalidParams
	}

	operator, err = sdk.AccAddressFromHexUnsafe(req.Operator)
	if err != nil {
		log.CtxErrorw(ctx, "failed to creates an AccAddress from a HEX-encoded string", "req.Operator", operator.String(), "error", err)
		return nil, err
	}

	if err = s3util.CheckValidBucketName(req.BucketName); err != nil {
		log.Errorw("failed to check bucket name", "bucket_name", req.BucketName, "error", err)
		return nil, errors.ErrInvalidBucketName
	}

	bucketInfo, err = r.baseApp.GfBsDB().GetBucketByName(req.BucketName, true)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket info", "error", err)
		return nil, err
	}
	if bucketInfo == nil {
		log.CtxError(ctx, "no such bucket")
		return nil, errors.ErrNoSuchBucket
	}

	if req.ObjectName == "" {
		effect, err = r.VerifyBucketPermission(ctx, bucketInfo, operator, req.ActionType, nil)
		if err != nil {
			log.CtxErrorw(ctx, "failed to verify bucket permission", "error", err)
			return nil, err
		}
	} else {
		objectInfo, err = r.baseApp.GfBsDB().GetObjectByName(req.BucketName, req.ObjectName, true)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get object info", "error", err)
			return nil, err
		}
		if objectInfo == nil {
			log.CtxError(ctx, "no such object")
			return nil, errors.ErrNoSuchObject
		}
		effect, err = r.VerifyObjectPermission(ctx, bucketInfo, objectInfo, operator, req.ActionType)
		if err != nil {
			log.CtxErrorw(ctx, "failed to verify object permission", "error", err)
			return nil, err
		}
	}

	resp = &storagetypes.QueryVerifyPermissionResponse{Effect: effect}
	log.CtxInfow(ctx, "succeed to verify permission")
	return resp, nil
}

// VerifyBucketPermission verify bucket permission
func (r *RetrieveModular) VerifyBucketPermission(ctx context.Context, bucketInfo *bsdb.Bucket, operator sdk.AccAddress,
	action permtypes.ActionType, options *permtypes.VerifyOptions) (permtypes.Effect, error) {
	var (
		err    error
		owner  sdk.AccAddress
		effect permtypes.Effect
	)

	// if bucket is public, anyone can read but can not write it.
	if bucketInfo.Visibility == storagetypes.VISIBILITY_TYPE_PUBLIC_READ.String() && keeper.PublicReadBucketAllowedActions[action] {
		return permtypes.EFFECT_ALLOW, nil
	}

	owner, err = sdk.AccAddressFromHexUnsafe(bucketInfo.Owner.String())
	if err != nil {
		log.CtxErrorw(ctx, "failed to creates an AccAddress from a HEX-encoded string", "bucketInfo.Owner.String()", bucketInfo.Owner.String(), "error", err)
		return permtypes.EFFECT_DENY, err
	}

	// the owner has full permissions
	if operator.Equals(owner) {
		return permtypes.EFFECT_ALLOW, nil
	}

	// verify policy
	effect, err = r.VerifyPolicy(ctx, math.NewUintFromBigInt(bucketInfo.BucketID.Big()), gnfdresource.RESOURCE_TYPE_BUCKET, operator, action, options)
	if err != nil {
		log.CtxErrorw(ctx, "failed to verify bucket policy", "error", err)
		return permtypes.EFFECT_DENY, err
	}

	if effect == permtypes.EFFECT_ALLOW {
		return permtypes.EFFECT_ALLOW, nil
	}
	return permtypes.EFFECT_DENY, nil
}

// VerifyObjectPermission verify object permission
func (r *RetrieveModular) VerifyObjectPermission(ctx context.Context, bucketInfo *bsdb.Bucket, objectInfo *bsdb.Object,
	operator sdk.AccAddress, action permtypes.ActionType) (permtypes.Effect, error) {
	var (
		visibility   bool
		err          error
		ownerAcc     sdk.AccAddress
		opts         *permtypes.VerifyOptions
		bucketEffect permtypes.Effect
		objectEffect permtypes.Effect
	)

	// check the permission of object and bucket
	if objectInfo.Visibility == storagetypes.VISIBILITY_TYPE_PUBLIC_READ.String() ||
		(objectInfo.Visibility == storagetypes.VISIBILITY_TYPE_INHERIT.String() && bucketInfo.Visibility == storagetypes.VISIBILITY_TYPE_PUBLIC_READ.String()) {
		visibility = true
	}

	if visibility && keeper.PublicReadObjectAllowedActions[action] {
		return permtypes.EFFECT_ALLOW, nil
	}

	// the owner has full permissions
	ownerAcc, err = sdk.AccAddressFromHexUnsafe(objectInfo.Owner.String())
	if err != nil {
		log.CtxErrorw(ctx, "failed to creates an AccAddress from a HEX-encoded string", "objectInfo.Owner.String()", objectInfo.Owner.String(), "error", err)
		return permtypes.EFFECT_DENY, err
	}

	if ownerAcc.Equals(operator) {
		return permtypes.EFFECT_ALLOW, nil
	}

	// verify policy
	opts = &permtypes.VerifyOptions{
		Resource: chaintypes.NewObjectGRN(objectInfo.BucketName, objectInfo.ObjectName).String(),
	}
	bucketEffect, err = r.VerifyPolicy(ctx, math.NewUintFromBigInt(bucketInfo.BucketID.Big()), gnfdresource.RESOURCE_TYPE_BUCKET, operator, action, opts)
	if err != nil || bucketEffect == permtypes.EFFECT_DENY {
		log.CtxErrorw(ctx, "failed to verify object policy", "error", err)
		return permtypes.EFFECT_DENY, err
	}

	objectEffect, err = r.VerifyPolicy(ctx, math.NewUintFromBigInt(objectInfo.ObjectID.Big()), gnfdresource.RESOURCE_TYPE_OBJECT, operator, action,
		nil)
	if err != nil || objectEffect == permtypes.EFFECT_DENY {
		log.CtxErrorw(ctx, "failed to verify object policy", "error", err)
		return permtypes.EFFECT_DENY, err
	}

	if bucketEffect == permtypes.EFFECT_ALLOW || objectEffect == permtypes.EFFECT_ALLOW {
		return permtypes.EFFECT_ALLOW, nil
	}
	return permtypes.EFFECT_DENY, nil
}

// VerifyPolicy verify policy of permission
func (r *RetrieveModular) VerifyPolicy(ctx context.Context, resourceID math.Uint, resourceType resource.ResourceType,
	operator sdk.AccAddress, action permtypes.ActionType, opts *permtypes.VerifyOptions) (permtypes.Effect, error) {
	var (
		err                    error
		allowed                bool
		effect                 permtypes.Effect
		permission             *bsdb.Permission
		permissions            []*bsdb.Permission
		groups                 []*bsdb.Group
		statements             []*bsdb.Statement
		groupIDMap             = make(map[common.Hash]bool)
		accountPolicyID        = make([]common.Hash, 0)
		policyIDList           = make([]common.Hash, 0)
		filteredPermissionList = make([]*bsdb.Permission, 0)
	)

	// verify policy which grant permission to account
	permission, err = r.baseApp.GfBsDB().GetPermissionByResourceAndPrincipal(resourceType.String(), resourceID.String(), permtypes.PRINCIPAL_TYPE_GNFD_ACCOUNT.String(), operator.String())
	if err != nil || permission == nil {
		log.CtxErrorw(ctx, "failed to get permission by resource and principal", "error", err)
		return permtypes.EFFECT_DENY, err
	}

	accountPolicyID = append(accountPolicyID, permission.PolicyID)
	statements, err = r.baseApp.GfBsDB().GetStatementsByPolicyID(accountPolicyID)
	if err != nil || statements == nil {
		log.CtxErrorw(ctx, "failed to get statements by policy id", "error", err)
		return permtypes.EFFECT_DENY, err
	}

	effect = permission.Eval(action, time.Now(), opts, statements)
	if effect != permtypes.EFFECT_UNSPECIFIED {
		return effect, nil
	}

	// verify policy which grant permission to group
	permissions, err = r.baseApp.GfBsDB().GetPermissionsByResourceAndPrincipleType(resourceType.String(), resourceID.String(), permtypes.PRINCIPAL_TYPE_GNFD_GROUP.String())
	if err != nil || permissions == nil {
		log.CtxErrorw(ctx, "failed to get permission by resource and principle type", "error", err)
		return permtypes.EFFECT_DENY, err
	}

	groupIDList := make([]common.Hash, len(permissions))
	for i, perm := range permissions {
		groupIDList[i] = common.HexToHash(perm.PrincipalValue)
	}

	// filter group id by account
	groups, err = r.baseApp.GfBsDB().GetGroupsByGroupIDAndAccount(groupIDList, common.HexToHash(operator.String()))
	if err != nil || groups == nil {
		log.CtxErrorw(ctx, "failed to get groups by group id and account", "error", err)
		return permtypes.EFFECT_DENY, err
	}

	// store the group id into map
	for _, group := range groups {
		groupIDMap[group.GroupID] = true
	}

	// use group id map to filter the above permission list and get the permissions which related to the specific account
	for _, perm := range permissions {
		_, ok := groupIDMap[common.HexToHash(perm.PrincipalValue)]
		if ok {
			policyIDList = append(policyIDList, perm.PolicyID)
			filteredPermissionList = append(filteredPermissionList, perm)
		}
	}

	statements, err = r.baseApp.GfBsDB().GetStatementsByPolicyID(policyIDList)
	if err != nil || statements == nil {
		log.CtxErrorw(ctx, "failed to get statements by policy id", "error", err)
		return permtypes.EFFECT_DENY, err
	}

	// eval permission and get effect result
	for _, perm := range filteredPermissionList {
		effect = perm.Eval(action, time.Now(), opts, statements)
		if effect != permtypes.EFFECT_UNSPECIFIED {
			if effect == permtypes.EFFECT_ALLOW {
				allowed = true
			} else if effect == permtypes.EFFECT_DENY {
				return permtypes.EFFECT_DENY, nil
			}
		}
	}

	if allowed {
		return permtypes.EFFECT_ALLOW, nil
	}
	return permtypes.EFFECT_UNSPECIFIED, nil
}

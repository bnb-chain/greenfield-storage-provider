package metadata

import (
	"context"
	"errors"
	"strings"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	chaintypes "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/resource"
	gnfdresource "github.com/bnb-chain/greenfield/types/resource"
	"github.com/bnb-chain/greenfield/types/s3util"
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
	"github.com/bnb-chain/greenfield/x/storage/keeper"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// GfSpVerifyPermission Verify the input account’s permission to input items
func (r *MetadataModular) GfSpVerifyPermission(ctx context.Context, req *storagetypes.QueryVerifyPermissionRequest) (resp *storagetypes.QueryVerifyPermissionResponse, err error) {
	var (
		operator   sdk.AccAddress
		bucketInfo *bsdb.Bucket
		objectInfo *bsdb.Object
		effect     permtypes.Effect
	)

	ctx = log.Context(ctx, req)

	if req == nil {
		log.CtxErrorw(ctx, "invalid request", "error", err)
		return nil, ErrInvalidParams
	}

	operator, err = sdk.AccAddressFromHexUnsafe(req.Operator)
	if err != nil && err != sdk.ErrEmptyHexAddress {
		log.CtxErrorw(ctx, "failed to creates an AccAddress from a HEX-encoded string", "req.Operator", operator.String(), "error", err)
		return nil, ErrInvalidParams
	}

	if err = s3util.CheckValidBucketName(req.BucketName); err != nil {
		log.Errorw("failed to check bucket name", "bucket_name", req.BucketName, "error", err)
		return nil, ErrInvalidBucketName
	}

	bucketInfo, err = r.baseApp.GfBsDB().GetBucketByName(req.BucketName, true)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket info", "error", err)
		return nil, err
	}
	if bucketInfo == nil || bucketInfo.Removed {
		log.CtxError(ctx, "no such bucket")
		return nil, ErrNoSuchBucket
	}

	if req.ObjectName == "" {
		effect, err = r.VerifyBucketPermission(ctx, bucketInfo, operator, req.ActionType, nil)
		if err != nil {
			log.CtxErrorw(ctx, "failed to verify bucket permission", "error", err)
			return nil, err
		}
	} else {
		objectInfo, err = r.baseApp.GfBsDB().GetObjectByName(req.ObjectName, req.BucketName, true)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get object info", "error", err)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrNoSuchObject
			}
			return nil, err
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

// GfSpVerifyPermissionByID Verify the input account’s permission to input source type and resource id
func (r *MetadataModular) GfSpVerifyPermissionByID(ctx context.Context, req *types.GfSpVerifyPermissionByIDRequest) (resp *types.GfSpVerifyPermissionByIDResponse, err error) {
	var effect permtypes.Effect

	ctx = log.Context(ctx, req)

	if req == nil {
		log.CtxErrorw(ctx, "invalid request", "error", err)
		return nil, ErrInvalidParams
	}

	switch req.ResourceType.String() {
	case "RESOURCE_TYPE_OBJECT":
		effect, err = r.verifyObject(ctx, req)
	case "RESOURCE_TYPE_BUCKET":
		effect, err = r.verifyBucket(ctx, req)
	case "RESOURCE_TYPE_GROUP":
		effect, err = r.verifyGroup(ctx, req)
	default:
		effect = permtypes.EFFECT_DENY
	}

	if err != nil {
		return nil, err
	}
	resp = &types.GfSpVerifyPermissionByIDResponse{Effect: effect}
	log.CtxInfow(ctx, "succeed to verify permission by id")
	return resp, nil
}

func (r *MetadataModular) verifyObject(ctx context.Context, req *types.GfSpVerifyPermissionByIDRequest) (permtypes.Effect, error) {
	var (
		operator   sdk.AccAddress
		bucketInfo *bsdb.Bucket
		objectInfo *bsdb.Object
		effect     permtypes.Effect
		err        error
	)

	operator, err = sdk.AccAddressFromHexUnsafe(req.Operator)
	if err != nil {
		log.CtxErrorw(ctx, "failed to creates an AccAddress from a HEX-encoded string", "req.Operator", operator.String(), "error", err)
		return permtypes.EFFECT_DENY, err
	}

	objectInfo, err = r.baseApp.GfBsDB().GetObjectByID(int64(req.ResourceId), false)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get object info", "error", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return permtypes.EFFECT_DENY, ErrNoSuchObject
		}
		return permtypes.EFFECT_DENY, err
	}
	bucketInfo, err = r.baseApp.GfBsDB().GetBucketByID(objectInfo.BucketID.Big().Int64(), true)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket info", "error", err)
		return permtypes.EFFECT_DENY, err
	}
	if bucketInfo == nil {
		log.CtxError(ctx, "no such bucket")
		return permtypes.EFFECT_DENY, ErrNoSuchBucket
	}
	effect, err = r.VerifyObjectPermission(ctx, bucketInfo, objectInfo, operator, req.ActionType)
	if err != nil {
		log.CtxErrorw(ctx, "failed to verify object permission", "error", err)
		return effect, err
	}
	return effect, nil
}

func (r *MetadataModular) verifyBucket(ctx context.Context, req *types.GfSpVerifyPermissionByIDRequest) (permtypes.Effect, error) {
	var (
		operator   sdk.AccAddress
		bucketInfo *bsdb.Bucket
		effect     permtypes.Effect
		err        error
	)

	operator, err = sdk.AccAddressFromHexUnsafe(req.Operator)
	if err != nil {
		log.CtxErrorw(ctx, "failed to creates an AccAddress from a HEX-encoded string", "req.Operator", operator.String(), "error", err)
		return permtypes.EFFECT_DENY, err
	}

	bucketInfo, err = r.baseApp.GfBsDB().GetBucketByID(int64(req.ResourceId), true)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket info", "error", err)
		return permtypes.EFFECT_DENY, err
	}
	if bucketInfo == nil {
		log.CtxError(ctx, "no such bucket")
		return permtypes.EFFECT_DENY, ErrNoSuchBucket
	}
	effect, err = r.VerifyBucketPermission(ctx, bucketInfo, operator, req.ActionType, nil)
	if err != nil {
		log.CtxErrorw(ctx, "failed to verify bucket permission", "error", err)
		return effect, err
	}
	return effect, nil
}

func (r *MetadataModular) verifyGroup(ctx context.Context, req *types.GfSpVerifyPermissionByIDRequest) (permtypes.Effect, error) {
	var (
		operator  sdk.AccAddress
		groupInfo *bsdb.Group
		effect    permtypes.Effect
		err       error
	)

	operator, err = sdk.AccAddressFromHexUnsafe(req.Operator)
	if err != nil {
		log.CtxErrorw(ctx, "failed to creates an AccAddress from a HEX-encoded string", "req.Operator", operator.String(), "error", err)
		return permtypes.EFFECT_DENY, err
	}

	groupInfo, err = r.baseApp.GfBsDB().GetGroupByID(int64(req.ResourceId), false)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get group info", "error", err)
		return permtypes.EFFECT_DENY, err
	}
	if groupInfo == nil {
		log.CtxError(ctx, "no such group")
		return permtypes.EFFECT_DENY, ErrNoSuchGroup
	}
	effect, err = r.VerifyGroupPermission(ctx, groupInfo, operator, req.ActionType)
	if err != nil {
		log.CtxErrorw(ctx, "failed to verify group permission", "error", err)
		return effect, err
	}
	return effect, nil
}

// VerifyBucketPermission verify bucket permission
func (r *MetadataModular) VerifyBucketPermission(ctx context.Context, bucketInfo *bsdb.Bucket, operator sdk.AccAddress,
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
		return permtypes.EFFECT_DENY, ErrInvalidParams
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
func (r *MetadataModular) VerifyObjectPermission(ctx context.Context, bucketInfo *bsdb.Bucket, objectInfo *bsdb.Object,
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

// VerifyGroupPermission verify group permission
func (r *MetadataModular) VerifyGroupPermission(ctx context.Context, groupInfo *bsdb.Group, operator sdk.AccAddress,
	action permtypes.ActionType) (permtypes.Effect, error) {
	var (
		err    error
		effect permtypes.Effect
	)

	// The owner has full permissions
	if strings.EqualFold(groupInfo.Owner.String(), operator.String()) {
		return permtypes.EFFECT_ALLOW, nil
	}

	// verify policy
	effect, err = r.VerifyPolicy(ctx, math.NewUintFromBigInt(groupInfo.GroupID.Big()), gnfdresource.RESOURCE_TYPE_GROUP, operator, action, nil)
	if err != nil {
		log.CtxErrorw(ctx, "failed to verify bucket policy", "error", err)
		return permtypes.EFFECT_DENY, err
	}

	if effect == permtypes.EFFECT_ALLOW {
		return permtypes.EFFECT_ALLOW, nil
	}
	return permtypes.EFFECT_DENY, nil
}

// VerifyPolicy verify policy of permission
func (r *MetadataModular) VerifyPolicy(ctx context.Context, resourceID math.Uint, resourceType resource.ResourceType,
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
	permission, err = r.baseApp.GfBsDB().GetPermissionByResourceAndPrincipal(resourceType.String(), permtypes.PRINCIPAL_TYPE_GNFD_ACCOUNT.String(), operator.String(), common.BigToHash(resourceID.BigInt()))
	if err != nil {
		log.CtxErrorw(ctx, "failed to get permission by resource and principal", "error", err)
		return permtypes.EFFECT_DENY, err
	}

	if permission != nil {
		accountPolicyID = append(accountPolicyID, permission.PolicyID)
		statements, err = r.baseApp.GfBsDB().GetStatementsByPolicyID(accountPolicyID, false)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get statements by policy id", "error", err)
			return permtypes.EFFECT_DENY, err
		}
	}
	if statements != nil {
		effect = permission.Eval(action, time.Now(), opts, statements)
		if effect != permtypes.EFFECT_UNSPECIFIED {
			return effect, nil
		}
	}

	// verify policy which grant permission to group
	permissions, err = r.baseApp.GfBsDB().GetPermissionsByResourceAndPrincipleType(resourceType.String(), permtypes.PRINCIPAL_TYPE_GNFD_GROUP.String(), common.BigToHash(resourceID.BigInt()), false)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get permission by resource and principle type", "error", err)
		return permtypes.EFFECT_DENY, err
	}
	if permissions != nil {
		groupIDList := make([]common.Hash, len(permissions))
		for i, perm := range permissions {
			groupIDList[i] = common.BigToHash(math.NewUintFromString(perm.PrincipalValue).BigInt())
		}
		groups, err = r.baseApp.GfBsDB().GetGroupsByGroupIDAndAccount(groupIDList, common.HexToAddress(operator.String()), false)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get groups by group id and account", "error", err)
			return permtypes.EFFECT_DENY, err
		}
		if groups != nil {
			var filteredGroups []*bsdb.Group
			// filter the group member if they are expired
			for _, group := range groups {
				if group.ExpirationTime == 0 || time.Unix(group.ExpirationTime, 0).After(time.Now()) {
					filteredGroups = append(filteredGroups, group)
				}
			}
			// store the group id into map
			for _, group := range filteredGroups {
				groupIDMap[group.GroupID] = true
			}
			// use group id map to filter the above permission list and get the permissions which related to the specific account
			for _, perm := range permissions {
				_, ok := groupIDMap[common.BigToHash(math.NewUintFromString(perm.PrincipalValue).BigInt())]
				if ok {
					policyIDList = append(policyIDList, perm.PolicyID)
					filteredPermissionList = append(filteredPermissionList, perm)
				}
			}
		}
	}

	statements, err = r.baseApp.GfBsDB().GetStatementsByPolicyID(policyIDList, false)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get statements by policy id", "error", err)
		return permtypes.EFFECT_DENY, err
	}
	if statements != nil {
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
	}
	return permtypes.EFFECT_UNSPECIFIED, nil
}

// GfSpListObjectPolicies list policies by object info
func (r *MetadataModular) GfSpListObjectPolicies(ctx context.Context, req *types.GfSpListObjectPoliciesRequest) (resp *types.GfSpListObjectPoliciesResponse, err error) {
	var (
		object      *bsdb.Object
		limit       int
		policies    []*types.Policy
		permissions []*bsdb.Permission
	)

	ctx = log.Context(ctx, req)
	limit = int(req.Limit)
	//if the user doesn't specify a limit, the default value is LisPoliciesDefaultLimit
	if req.Limit == 0 {
		limit = bsdb.LisPoliciesDefaultLimit
	}

	// If the user specifies a value exceeding LisPoliciesLimitSize, the response will only return up to LisPoliciesLimitSize policies
	if req.Limit > bsdb.LisPoliciesLimitSize {
		limit = bsdb.LisPoliciesLimitSize
	}

	object, err = r.baseApp.GfBsDB().GetObjectByName(req.ObjectName, req.BucketName, true)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get object info", "error", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoSuchObject
		}
		return nil, err
	}

	permissions, err = r.baseApp.GfBsDB().ListObjectPolicies(object.ObjectID, req.ActionType, common.BigToHash(math.NewUint(req.StartAfter).BigInt()), limit)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list policies by object info", "error", err)
		return nil, err
	}

	policies = make([]*types.Policy, len(permissions))
	for i, perm := range permissions {
		policies[i] = &types.Policy{
			PrincipalType:   perm.PrincipalType,
			PrincipalValue:  perm.PrincipalValue,
			ResourceType:    gnfdresource.ResourceType(gnfdresource.ResourceType_value[perm.ResourceType]),
			ResourceId:      perm.ResourceID.String(),
			CreateTimestamp: perm.CreateTimestamp,
			UpdateTimestamp: perm.UpdateTimestamp,
			ExpirationTime:  perm.ExpirationTime,
		}
	}

	resp = &types.GfSpListObjectPoliciesResponse{Policies: policies}
	log.CtxInfo(ctx, "succeed to list objects by object ids")
	return resp, nil
}

// GfSpVerifyMigrateGVGPermission verify the destination sp id of bucket migration & swap out
// When bucketID is not 0, it means bucket migration; otherwise, it means SP exit
func (r *MetadataModular) GfSpVerifyMigrateGVGPermission(ctx context.Context, req *types.GfSpVerifyMigrateGVGPermissionRequest) (resp *types.GfSpVerifyMigrateGVGPermissionResponse, err error) {
	var effect permtypes.Effect

	ctx = log.Context(ctx, req)

	if req.BucketId != 0 {
		effect, err = r.verifyBucketMigration(ctx, req.BucketId, req.DstSpId)
		if err != nil {
			log.CtxErrorw(ctx, "failed to verify the destination sp id of bucket migration", "error", err)
			return nil, err
		}
	} else {
		effect, err = r.verifySwapOut(ctx, req.GvgId, req.DstSpId)
		if err != nil {
			log.CtxErrorw(ctx, "failed to verify the destination sp id of swap out", "error", err)
			return nil, err
		}
	}

	resp = &types.GfSpVerifyMigrateGVGPermissionResponse{Effect: effect}
	log.CtxInfow(ctx, "succeed to verify the destination sp id of bucket migration & swap out")
	return resp, nil
}

// verifyBucketMigration verify bucket migration sp id
func (r *MetadataModular) verifyBucketMigration(ctx context.Context, bucketID uint64, dstSpID uint32) (permtypes.Effect, error) {
	var (
		err   error
		event *bsdb.EventMigrationBucket
	)

	event, err = r.baseApp.GfBsDB().GetEventMigrationBucketByBucketID(common.BigToHash(math.NewUint(bucketID).BigInt()))
	if err != nil {
		log.CtxErrorw(ctx, "failed to get migration bucket event by bucket id", "error", err)
		return permtypes.EFFECT_DENY, err
	}

	if event == nil || event.DstPrimarySpId != dstSpID {
		return permtypes.EFFECT_DENY, nil
	}

	return permtypes.EFFECT_ALLOW, nil
}

// verifySwapOut verify swap out sp id
func (r *MetadataModular) verifySwapOut(ctx context.Context, gvgID uint32, dstSpID uint32) (permtypes.Effect, error) {
	var (
		err   error
		event *bsdb.EventSwapOut
	)

	event, err = r.baseApp.GfBsDB().GetEventSwapOutByGvgID(gvgID)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get migration bucket event by bucket id", "error", err)
		return permtypes.EFFECT_DENY, err
	}

	if event.SuccessorSpId != dstSpID {
		return permtypes.EFFECT_DENY, nil
	}

	return permtypes.EFFECT_ALLOW, nil
}

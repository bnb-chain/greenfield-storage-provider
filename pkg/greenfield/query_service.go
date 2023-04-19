package greenfield

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	permissiontypes "github.com/bnb-chain/greenfield/x/permission/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	errorstypes "github.com/bnb-chain/greenfield-storage-provider/pkg/errors/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// GetCurrentHeight the block height sub one as the stable height.
func (greenfield *Greenfield) GetCurrentHeight(ctx context.Context) (uint64, error) {
	resp, err := greenfield.client.chainClient.TmClient.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{})
	if err != nil {
		log.Errorw("get latest block height failed", "node_addr", greenfield.client.Provider, "error", err)
		return 0, errorstypes.Error(merrors.ChainGetLatestBlockErrCode, err.Error())
	}
	return (uint64)(resp.SdkBlock.Header.Height), nil
}

// HasAccount returns an indication of the existence of address.
func (greenfield *Greenfield) HasAccount(ctx context.Context, address string) (bool, error) {
	client := greenfield.getCurrentClient().GnfdClient()
	resp, err := client.Account(ctx, &authtypes.QueryAccountRequest{Address: address})
	if err != nil {
		log.Errorw("failed to query account", "address", address, "error", err)
		return false, errorstypes.Error(merrors.ChainQueryAccountErrCode, err.Error())
	}
	return resp.GetAccount() != nil, nil
}

// QuerySPInfo returns the list of storage provider info.
func (greenfield *Greenfield) QuerySPInfo(ctx context.Context) ([]*sptypes.StorageProvider, error) {
	client := greenfield.getCurrentClient().GnfdClient()
	var spInfos []*sptypes.StorageProvider
	resp, err := client.StorageProviders(ctx, &sptypes.QueryStorageProvidersRequest{
		Pagination: &query.PageRequest{
			Offset: 0,
			Limit:  math.MaxUint64,
		},
	})
	if err != nil {
		log.Errorw("failed to query storage provider list", "error", err)
		return spInfos, errorstypes.Error(merrors.ChainQuerySPListErrCode, err.Error())
	}
	for i := 0; i < len(resp.GetSps()); i++ {
		spInfos = append(spInfos, resp.GetSps()[i])
	}
	return spInfos, nil
}

// QueryStorageParams returns storage params
func (greenfield *Greenfield) QueryStorageParams(ctx context.Context) (params *storagetypes.Params, err error) {
	client := greenfield.getCurrentClient().GnfdClient()
	resp, err := client.StorageQueryClient.Params(ctx, &storagetypes.QueryParamsRequest{})
	if err != nil {
		log.Errorw("failed to query storage params", "error", err)
		return nil, errorstypes.Error(merrors.ChainQueryStorageParamsErrCode, err.Error())
	}
	return &resp.Params, nil
}

// QueryBucketInfo return the bucket info by name.
func (greenfield *Greenfield) QueryBucketInfo(ctx context.Context, bucket string) (*storagetypes.BucketInfo, error) {
	client := greenfield.getCurrentClient().GnfdClient()
	resp, err := client.HeadBucket(ctx, &storagetypes.QueryHeadBucketRequest{BucketName: bucket})
	if errors.Is(err, storagetypes.ErrNoSuchBucket) {
		return nil, errorstypes.Error(merrors.NoSuchBucketErrCode, merrors.ErrNoSuchBucket.Error())
	}
	if err != nil {
		log.Errorw("failed to query bucket", "bucket_name", bucket, "error", err)
		return nil, errorstypes.Error(merrors.ChainHeadBucketErrCode, err.Error())
	}
	return resp.GetBucketInfo(), nil
}

// QueryObjectInfo return the object info by name.
func (greenfield *Greenfield) QueryObjectInfo(ctx context.Context, bucket, object string) (*storagetypes.ObjectInfo, error) {
	client := greenfield.getCurrentClient().GnfdClient()
	resp, err := client.HeadObject(ctx, &storagetypes.QueryHeadObjectRequest{
		BucketName: bucket,
		ObjectName: object,
	})
	if strings.Contains(err.Error(), storagetypes.ErrNoSuchObject.Error()) {
		return nil, errorstypes.Error(merrors.NoSuchObjectErrCode, merrors.ErrNoSuchObject.Error())
	}
	if err != nil {
		log.Errorw("failed to query object", "bucket_name", bucket, "object_name", object, "error", err)
		return nil, errorstypes.Error(merrors.ChainHeadObjectErrCode, err.Error())
	}
	return resp.GetObjectInfo(), nil
}

// QueryObjectInfoByID return the object info by name.
func (greenfield *Greenfield) QueryObjectInfoByID(ctx context.Context, objectID string) (*storagetypes.ObjectInfo, error) {
	client := greenfield.getCurrentClient().GnfdClient()
	resp, err := client.HeadObjectById(ctx, &storagetypes.QueryHeadObjectByIdRequest{
		ObjectId: objectID,
	})
	if errors.Is(err, storagetypes.ErrNoSuchObject) {
		return nil, errorstypes.Error(merrors.NoSuchObjectErrCode, merrors.ErrNoSuchObject.Error())
	}
	if err != nil {
		log.Errorw("failed to query object", "object_id", objectID, "error", err)
		return nil, errorstypes.Error(merrors.ChainHeadObjectByIDErrCode, err.Error())
	}
	return resp.GetObjectInfo(), nil
}

// QueryBucketInfoAndObjectInfo return bucket info and object info, if not found, return the corresponding error code
func (greenfield *Greenfield) QueryBucketInfoAndObjectInfo(ctx context.Context, bucket, object string) (*storagetypes.BucketInfo, *storagetypes.ObjectInfo, error) {
	bucketInfo, err := greenfield.QueryBucketInfo(ctx, bucket)
	if err != nil {
		return nil, nil, err
	}
	objectInfo, err := greenfield.QueryObjectInfo(ctx, bucket, object)
	if err != nil {
		return nil, nil, err
	}
	return bucketInfo, objectInfo, nil
}

// ListenObjectSeal return an indication of the object is sealed.
// TODO:: retrieve service support seal event subscription
func (greenfield *Greenfield) ListenObjectSeal(ctx context.Context, bucket, object string, timeoutHeight int) (err error) {
	var objectInfo *storagetypes.ObjectInfo
	for i := 0; i < timeoutHeight; i++ {
		time.Sleep(ExpectedOutputBlockInternal * time.Second)
		objectInfo, err = greenfield.QueryObjectInfo(ctx, bucket, object)
		if err != nil {
			continue
		}
		if objectInfo.GetObjectStatus() == storagetypes.OBJECT_STATUS_SEALED {
			return nil
		}
	}
	log.Errorw("seal object timeout", "bucket_name", bucket, "object_name", object)
	return errorstypes.Error(merrors.ChainSealObjectTimeoutErrCode, merrors.ErrSealTimeout.Error())
}

// QueryStreamRecord return the steam record info by account.
func (greenfield *Greenfield) QueryStreamRecord(ctx context.Context, account string) (*paymenttypes.StreamRecord, error) {
	client := greenfield.getCurrentClient().GnfdClient()
	resp, err := client.StreamRecord(ctx, &paymenttypes.QueryGetStreamRecordRequest{
		Account: account,
	})
	if err != nil {
		log.Errorw("failed to query stream record", "account", account, "error", err)
		return nil, errorstypes.Error(merrors.ChainQueryStreamRecordErrCode, err.Error())
	}
	return &resp.StreamRecord, nil
}

// VerifyGetObjectPermission verify get object permission.
func (greenfield *Greenfield) VerifyGetObjectPermission(ctx context.Context, account, bucket, object string) (bool, error) {
	client := greenfield.getCurrentClient().GnfdClient()
	resp, err := client.VerifyPermission(ctx, &storagetypes.QueryVerifyPermissionRequest{
		Operator:   account,
		BucketName: bucket,
		ObjectName: object,
		ActionType: permissiontypes.ACTION_GET_OBJECT,
	})
	if err != nil {
		log.Errorw("failed to verify get object permission", "account", account, "error", err)
		return false, errorstypes.Error(merrors.ChainGetObjetVerifyPermissionErrCode, err.Error())
	}
	if resp.GetEffect() == permissiontypes.EFFECT_ALLOW {
		return true, nil
	}
	return false, nil
}

// VerifyPutObjectPermission verify put object permission.
func (greenfield *Greenfield) VerifyPutObjectPermission(ctx context.Context, account, bucket, object string) (bool, error) {
	_ = object
	client := greenfield.getCurrentClient().GnfdClient()
	resp, err := client.VerifyPermission(ctx, &storagetypes.QueryVerifyPermissionRequest{
		Operator:   account,
		BucketName: bucket,
		// TODO: Polish the function interface according to the semantics
		// ObjectName: object,
		ActionType: permissiontypes.ACTION_CREATE_OBJECT,
	})
	if err != nil {
		log.Errorw("failed to verify put object permission", "account", account, "error", err)
		return false, errorstypes.Error(merrors.ChainPutObjetVerifyPermissionErrCode, err.Error())
	}
	if resp.GetEffect() == permissiontypes.EFFECT_ALLOW {
		return true, nil
	}
	return false, nil
}

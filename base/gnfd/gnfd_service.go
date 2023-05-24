package gnfd

import (
	"context"
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	permissiontypes "github.com/bnb-chain/greenfield/x/permission/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// CurrentHeight the block height sub one as the stable height.
func (g *Gnfd) CurrentHeight(
	ctx context.Context) (uint64, error) {
	resp, err := g.getCurrentClient().chainClient.TmClient.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{})
	if err != nil {
		log.CtxErrorw(ctx, "get latest block height failed", "node_addr", g.client.Provider, "error", err)
		return 0, err
	}
	return (uint64)(resp.SdkBlock.Header.Height), nil
}

// HasAccount returns an indication of the existence of address.
func (g *Gnfd) HasAccount(
	ctx context.Context,
	address string) (bool, error) {
	client := g.getCurrentClient().GnfdClient()
	resp, err := client.Account(ctx, &authtypes.QueryAccountRequest{Address: address})
	if err != nil {
		log.CtxErrorw(ctx, "failed to query account", "address", address, "error", err)
		return false, err
	}
	return resp.GetAccount() != nil, nil
}

// QuerySPInfo returns the list of storage provider info.
func (g *Gnfd) QuerySPInfo(
	ctx context.Context) (
	[]*sptypes.StorageProvider, error) {
	client := g.getCurrentClient().GnfdClient()
	var spInfos []*sptypes.StorageProvider
	resp, err := client.StorageProviders(ctx, &sptypes.QueryStorageProvidersRequest{
		Pagination: &query.PageRequest{
			Offset: 0,
			Limit:  math.MaxUint64,
		},
	})
	if err != nil {
		log.Errorw("failed to query storage provider list", "error", err)
		return spInfos, err
	}
	for i := 0; i < len(resp.GetSps()); i++ {
		spInfos = append(spInfos, resp.GetSps()[i])
	}
	return spInfos, nil
}

// QueryStorageParams returns storage params
func (g *Gnfd) QueryStorageParams(
	ctx context.Context) (
	params *storagetypes.Params, err error) {
	client := g.getCurrentClient().GnfdClient()
	resp, err := client.StorageQueryClient.Params(ctx, &storagetypes.QueryParamsRequest{})
	if err != nil {
		log.CtxErrorw(ctx, "failed to query storage params", "error", err)
		return nil, err
	}
	return &resp.Params, nil
}

// QueryBucketInfo return the bucket info by name.
func (g *Gnfd) QueryBucketInfo(
	ctx context.Context,
	bucket string) (
	*storagetypes.BucketInfo, error) {
	client := g.getCurrentClient().GnfdClient()
	resp, err := client.HeadBucket(ctx, &storagetypes.QueryHeadBucketRequest{BucketName: bucket})
	if errors.Is(err, storagetypes.ErrNoSuchBucket) {
		return nil, ErrNoSuchBucket
	}
	if err != nil {
		log.CtxErrorw(ctx, "failed to query bucket", "bucket_name", bucket, "error", err)
		return nil, err
	}
	return resp.GetBucketInfo(), nil
}

// QueryObjectInfo return the object info by name.
func (g *Gnfd) QueryObjectInfo(
	ctx context.Context,
	bucket, object string) (
	*storagetypes.ObjectInfo, error) {
	client := g.getCurrentClient().GnfdClient()
	resp, err := client.HeadObject(ctx, &storagetypes.QueryHeadObjectRequest{
		BucketName: bucket,
		ObjectName: object,
	})
	if err != nil {
		log.CtxErrorw(ctx, "failed to query object", "bucket_name", bucket, "object_name", object, "error", err)
		return nil, err
	}
	return resp.GetObjectInfo(), nil
}

// QueryObjectInfoByID return the object info by name.
func (g *Gnfd) QueryObjectInfoByID(
	ctx context.Context,
	objectID string) (
	*storagetypes.ObjectInfo, error) {
	client := g.getCurrentClient().GnfdClient()
	resp, err := client.HeadObjectById(ctx, &storagetypes.QueryHeadObjectByIdRequest{
		ObjectId: objectID,
	})
	if errors.Is(err, storagetypes.ErrNoSuchObject) {
		return nil, merrors.ErrNoSuchObject
	}
	if err != nil {
		log.CtxErrorw(ctx, "failed to query object", "object_id", objectID, "error", err)
		return nil, err
	}
	return resp.GetObjectInfo(), nil
}

// QueryBucketInfoAndObjectInfo return bucket info and object info, if not found, return the corresponding error code
func (g *Gnfd) QueryBucketInfoAndObjectInfo(
	ctx context.Context,
	bucket, object string) (
	*storagetypes.BucketInfo,
	*storagetypes.ObjectInfo,
	error) {
	bucketInfo, err := g.QueryBucketInfo(ctx, bucket)
	if errors.Is(err, storagetypes.ErrNoSuchBucket) {
		return nil, nil, merrors.ErrNoSuchBucket
	}
	if err != nil {
		return nil, nil, err
	}
	objectInfo, err := g.QueryObjectInfo(ctx, bucket, object)
	if errors.Is(err, storagetypes.ErrNoSuchObject) {
		return bucketInfo, nil, merrors.ErrNoSuchObject
	}
	if err != nil {
		return bucketInfo, nil, err
	}
	return bucketInfo, objectInfo, nil
}

// ListenObjectSeal return an indication of the object is sealed.
// TODO:: retrieve service support seal event subscription
func (g *Gnfd) ListenObjectSeal(
	ctx context.Context,
	objectID uint64,
	timeOutHeight int) (bool, error) {
	var (
		objectInfo *storagetypes.ObjectInfo
		err        error
	)
	for i := 0; i < timeOutHeight; i++ {
		time.Sleep(ExpectedOutputBlockInternal * time.Second)
		objectInfo, err = g.QueryObjectInfoByID(ctx, strconv.FormatUint(objectID, 10))
		if err != nil {
			continue
		}
		if objectInfo.GetObjectStatus() == storagetypes.OBJECT_STATUS_SEALED {
			log.CtxDebugw(ctx, "succeed to listen object stat")
			return true, nil
		}
	}
	if err == nil {
		log.CtxErrorw(ctx, "seal object timeout", "object_id", objectID)
		return false, ErrSealTimeout
	}
	log.CtxErrorw(ctx, "listen seal object failed", "object_id", objectID, "error", err)
	return false, err
}

// QueryPaymentStreamRecord return the steam record info by account.
func (g *Gnfd) QueryPaymentStreamRecord(
	ctx context.Context,
	account string) (
	*paymenttypes.StreamRecord,
	error) {
	client := g.getCurrentClient().GnfdClient()
	resp, err := client.StreamRecord(ctx, &paymenttypes.QueryGetStreamRecordRequest{
		Account: account,
	})
	if err != nil {
		log.CtxErrorw(ctx, "failed to query stream record", "account", account, "error", err)
		return nil, err
	}
	return &resp.StreamRecord, nil
}

// VerifyGetObjectPermission verify get object permission.
func (g *Gnfd) VerifyGetObjectPermission(
	ctx context.Context,
	account, bucket, object string) (
	bool, error) {
	client := g.getCurrentClient().GnfdClient()
	resp, err := client.VerifyPermission(ctx, &storagetypes.QueryVerifyPermissionRequest{
		Operator:   account,
		BucketName: bucket,
		ObjectName: object,
		ActionType: permissiontypes.ACTION_GET_OBJECT,
	})
	if err != nil {
		log.CtxErrorw(ctx, "failed to verify get object permission", "account", account, "error", err)
		return false, err
	}
	if resp.GetEffect() == permissiontypes.EFFECT_ALLOW {
		return true, err
	}
	return false, err
}

// VerifyPutObjectPermission verify put object permission.
func (g *Gnfd) VerifyPutObjectPermission(
	ctx context.Context,
	account, bucket, object string) (
	bool, error) {
	_ = object
	client := g.getCurrentClient().GnfdClient()
	resp, err := client.VerifyPermission(ctx, &storagetypes.QueryVerifyPermissionRequest{
		Operator:   account,
		BucketName: bucket,
		// TODO: Polish the function interface according to the semantics
		// ObjectName: object,
		ActionType: permissiontypes.ACTION_CREATE_OBJECT,
	})
	if err != nil {
		log.CtxErrorw(ctx, "failed to verify put object permission", "account", account, "error", err)
		return false, err
	}
	if resp.GetEffect() == permissiontypes.EFFECT_ALLOW {
		return true, err
	}
	return false, err
}

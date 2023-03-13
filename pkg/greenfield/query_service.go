package greenfield

import (
	"context"
	"errors"
	"math"
	"time"

	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// GetCurrentHeight the block height sub one as the stable height.
func (greenfield *Greenfield) GetCurrentHeight(ctx context.Context) (int64, error) {
	status, err := greenfield.getCurrentClient().GnfdCompositeClient().RpcClient.TmClient.Status(ctx)
	if err != nil {
		log.Errorw("failed to query status", "error", err)
		return 0, err
	}
	height := status.SyncInfo.LatestBlockHeight
	if height > 0 {
		height = height - 1
	}
	return height, nil
}

// HasAccount returns an indication of the existence of address.
func (greenfield *Greenfield) HasAccount(ctx context.Context, address string) (bool, error) {
	client := greenfield.getCurrentClient().GnfdCompositeClient()
	resp, err := client.Account(ctx, &authtypes.QueryAccountRequest{Address: address})
	if err != nil {
		log.Errorw("failed to query account", "address", address, "error", err)
		return false, err
	}
	return resp.GetAccount() != nil, nil
}

// QuerySPInfo returns the list of storage provider info.
func (greenfield *Greenfield) QuerySPInfo(ctx context.Context) ([]*sptypes.StorageProvider, error) {
	client := greenfield.getCurrentClient().GnfdCompositeClient()
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
		spInfos = append(spInfos, &resp.GetSps()[i])
	}
	return spInfos, nil
}

// QueryStorageParams returns storage params
func (greenfield *Greenfield) QueryStorageParams(ctx context.Context) (params *storagetypes.Params, err error) {
	client := greenfield.getCurrentClient().GnfdCompositeClient()
	resp, err := client.StorageQueryClient.Params(ctx, &storagetypes.QueryParamsRequest{})
	if err != nil {
		log.Errorw("failed to query storage params", "error", err)
		return nil, err
	}
	return &resp.Params, nil
}

// QueryBucketInfo return the bucket info by name.
func (greenfield *Greenfield) QueryBucketInfo(ctx context.Context, bucket string) (*storagetypes.BucketInfo, error) {
	client := greenfield.getCurrentClient().GnfdCompositeClient()
	resp, err := client.HeadBucket(ctx, &storagetypes.QueryHeadBucketRequest{BucketName: bucket})
	if err != nil {
		log.Errorw("failed to query bucket", "bucket_name", bucket, "error", err)
		return nil, err
	}
	return resp.GetBucketInfo(), nil
}

// QueryObjectInfo return the object info by name.
func (greenfield *Greenfield) QueryObjectInfo(ctx context.Context, bucket, object string) (*storagetypes.ObjectInfo, error) {
	client := greenfield.getCurrentClient().GnfdCompositeClient()
	resp, err := client.HeadObject(ctx, &storagetypes.QueryHeadObjectRequest{
		BucketName: bucket,
		ObjectName: object,
	})
	if err != nil {
		log.Errorw("failed to query object", "bucket_name", bucket, "object_name", object, "error", err)
		return nil, err
	}
	return resp.GetObjectInfo(), nil
}

// QueryBucketInfoAndObjectInfo return bucket info and object info, if not found, return the corresponding error code
func (greenfield *Greenfield) QueryBucketInfoAndObjectInfo(ctx context.Context, bucket, object string) (*storagetypes.BucketInfo, *storagetypes.ObjectInfo, error) {
	bucketInfo, err := greenfield.QueryBucketInfo(ctx, bucket)
	if errors.Is(err, storagetypes.ErrNoSuchBucket) {
		return nil, nil, merrors.ErrNoSuchBucket
	}
	objectInfo, err := greenfield.QueryObjectInfo(ctx, bucket, object)
	if errors.Is(err, storagetypes.ErrNoSuchObject) {
		return nil, nil, merrors.ErrNoSuchObject
	}
	return bucketInfo, objectInfo, nil
}

// ListenObjectSeal return an indication of the object is sealed.
// TODO:: retrieve service support seal event subscription
func (greenfield *Greenfield) ListenObjectSeal(ctx context.Context, bucket, object string, timeOutHeight int) (seal bool, err error) {
	var objectInfo *storagetypes.ObjectInfo
	for i := 0; i < timeOutHeight; i++ {
		time.Sleep(ListenChainEventInternal * time.Second)
		objectInfo, err = greenfield.QueryObjectInfo(ctx, bucket, object)
		if err != nil {
			continue
		}
		if objectInfo.GetObjectStatus() == storagetypes.OBJECT_STATUS_SEALED {
			seal = true
			err = nil
			return
		}
	}
	log.Errorw("seal object timeout", "bucket_name", bucket, "object_name", object)
	err = merrors.ErrSealTimeout
	return
}

// QueryStreamRecord return the steam record info by account.
func (greenfield *Greenfield) QueryStreamRecord(ctx context.Context, account string) (*paymenttypes.StreamRecord, error) {
	client := greenfield.getCurrentClient().GnfdCompositeClient()
	resp, err := client.StreamRecord(ctx, &paymenttypes.QueryGetStreamRecordRequest{
		Account: account,
	})
	if err != nil {
		log.Errorw("failed to query stream record", "account", account, "error", err)
		return nil, err
	}
	return &resp.StreamRecord, nil
}

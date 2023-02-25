package greenfield

import (
	"context"
	"math"
	"time"

	merror "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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
		log.Errorw("failed to query account", "error", err, "address", address)
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
		log.Errorw("failed to query sp list", "error", err)
		return spInfos, err
	}
	for i := 0; i < len(resp.GetSps()); i++ {
		spInfos = append(spInfos, &resp.GetSps()[i])
	}
	return spInfos, nil
}

// QueryBucketInfo return the bucket info by name.
func (greenfield *Greenfield) QueryBucketInfo(ctx context.Context, bucket string) (*storagetypes.BucketInfo, error) {
	client := greenfield.getCurrentClient().GnfdCompositeClient()
	resp, err := client.HeadBucket(ctx, &storagetypes.QueryHeadBucketRequest{BucketName: bucket})
	if err != nil {
		log.Errorw("failed to query bucket", "error", err, "bucket_name", bucket)
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
		log.Errorw("failed to query object", "error", err, "bucket_name", bucket, "object_name", object)
		return nil, err
	}
	return resp.GetObjectInfo(), nil
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
		if objectInfo.GetObjectStatus() == storagetypes.OBJECT_STATUS_IN_SERVICE {
			seal = true
			err = nil
			return
		}
	}
	log.Errorw("listen seal object timeout", "error", err, "bucket_name", bucket, "object_name", object)
	err = merror.ErrSealTimeout
	return
}

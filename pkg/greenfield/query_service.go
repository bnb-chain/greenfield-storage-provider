package greenfield

import (
	"context"
	"math"
	"time"

	merror "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// GetCurrentHeight the block height sub one as the stable height.
func (greenfield *Greenfield) GetCurrentHeight(ctx context.Context) (int64, error) {
	client := greenfield.GetGreenfieldClient().Tendermint()
	resp, err := client.TmClient.Status(ctx)
	if err != nil {
		return 0, err
	}
	height := resp.SyncInfo.LatestBlockHeight
	if height > 0 {
		height = height - 1
	}
	return height, nil
}

// HasAccount returns an indication of the existence of address.
func (greenfield *Greenfield) HasAccount(ctx context.Context, address string) (bool, error) {
	client := greenfield.GetGreenfieldClient().Greenfield()
	resp, err := client.Account(ctx, &authtypes.QueryAccountRequest{Address: address})
	if err != nil {
		return false, err
	}
	return resp.GetAccount() != nil, nil
}

// QuerySPInfo returns the list of storage provider info.
func (greenfield *Greenfield) QuerySPInfo(ctx context.Context) ([]*sptypes.StorageProvider, error) {
	client := greenfield.GetGreenfieldClient().Greenfield()
	var spInfos []*sptypes.StorageProvider
	resp, err := client.StorageProviders(ctx, &sptypes.QueryStorageProvidersRequest{
		Pagination: &query.PageRequest{
			Offset: 0,
			Limit:  math.MaxUint64,
		},
	})
	if err != nil {
		return spInfos, err
	}
	for i := 0; i < len(resp.GetSps()); i++ {
		spInfos = append(spInfos, &resp.GetSps()[i])
	}
	return spInfos, nil
}

// QueryBucketInfo return the bucket info by name.
func (greenfield *Greenfield) QueryBucketInfo(ctx context.Context, bucket string) (*storagetypes.BucketInfo, error) {
	client := greenfield.GetGreenfieldClient().Greenfield()
	resp, err := client.Bucket(ctx, &storagetypes.QueryBucketRequest{BucketName: bucket})
	if err != nil {
		return nil, err
	}
	return resp.GetBucketInfo(), nil
}

// QueryObjectInfo return the object info by name.
func (greenfield *Greenfield) QueryObjectInfo(ctx context.Context, bucket, object string) (*storagetypes.ObjectInfo, error) {
	client := greenfield.GetGreenfieldClient().Greenfield()
	resp, err := client.Object(ctx, &storagetypes.QueryObjectRequest{
		BucketName: bucket,
		ObjectName: object,
	})
	if err != nil {
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
	err = merror.ErrSealTimeout
	return
}

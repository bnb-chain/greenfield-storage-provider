package gnfd

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	permissiontypes "github.com/bnb-chain/greenfield/x/permission/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// CurrentHeight the block height sub one as the stable height.
func (g *Gnfd) CurrentHeight(ctx context.Context) (uint64, error) {
	startTime := time.Now()
	defer metrics.GnfdChainHistogram.WithLabelValues("query_height").Observe(time.Since(startTime).Seconds())
	resp, err := g.getCurrentWsClient().ABCIInfo(ctx)
	if err != nil {
		log.CtxErrorw(ctx, "get latest block height failed", "node_addr",
			g.getCurrentWsClient().Remote(), "error", err)
		return 0, err
	}
	return (uint64)(resp.Response.LastBlockHeight), nil
}

// HasAccount returns an indication of the existence of address.
func (g *Gnfd) HasAccount(ctx context.Context, address string) (bool, error) {
	startTime := time.Now()
	defer metrics.GnfdChainHistogram.WithLabelValues("query_account").Observe(time.Since(startTime).Seconds())
	client := g.getCurrentClient().GnfdClient()
	resp, err := client.Account(ctx, &authtypes.QueryAccountRequest{Address: address})
	if err != nil {
		log.CtxErrorw(ctx, "failed to query account", "address", address, "error", err)
		return false, err
	}
	return resp.GetAccount() != nil, nil
}

// ListSPs returns the list of storage provider info.
func (g *Gnfd) ListSPs(ctx context.Context) ([]*sptypes.StorageProvider, error) {
	startTime := time.Now()
	defer metrics.GnfdChainHistogram.WithLabelValues("list_sps").Observe(time.Since(startTime).Seconds())
	client := g.getCurrentClient().GnfdClient()
	var spInfos []*sptypes.StorageProvider
	resp, err := client.StorageProviders(ctx, &sptypes.QueryStorageProvidersRequest{
		Pagination: &query.PageRequest{
			Offset: 0,
			Limit:  math.MaxUint64,
		},
	})
	if err != nil {
		log.Errorw("failed to list storage providers", "error", err)
		return spInfos, err
	}
	for i := 0; i < len(resp.GetSps()); i++ {
		spInfos = append(spInfos, resp.GetSps()[i])
	}
	return spInfos, nil
}

// ListBondedValidators returns the list of bonded validators.
func (g *Gnfd) ListBondedValidators(ctx context.Context) ([]stakingtypes.Validator, error) {
	startTime := time.Now()
	defer metrics.GnfdChainHistogram.WithLabelValues("list_bonded_validators").Observe(time.Since(startTime).Seconds())
	client := g.getCurrentClient().GnfdClient()
	var validators []stakingtypes.Validator
	resp, err := client.Validators(ctx, &stakingtypes.QueryValidatorsRequest{Status: "BOND_STATUS_BONDED"})
	if err != nil {
		log.Errorw("failed to list validators", "error", err)
		return validators, err
	}
	for i := 0; i < len(resp.GetValidators()); i++ {
		validators = append(validators, resp.GetValidators()[i])
	}
	return validators, nil
}

// QueryStorageParams returns storage params
func (g *Gnfd) QueryStorageParams(ctx context.Context) (params *storagetypes.Params, err error) {
	startTime := time.Now()
	defer metrics.GnfdChainHistogram.WithLabelValues("query_storage_params").Observe(time.Since(startTime).Seconds())
	client := g.getCurrentClient().GnfdClient()
	resp, err := client.StorageQueryClient.Params(ctx, &storagetypes.QueryParamsRequest{})
	if err != nil {
		log.CtxErrorw(ctx, "failed to query storage params", "error", err)
		return nil, err
	}
	return &resp.Params, nil
}

// QueryStorageParamsByTimestamp returns storage params by block create time.
func (g *Gnfd) QueryStorageParamsByTimestamp(ctx context.Context, timestamp int64) (params *storagetypes.Params, err error) {
	startTime := time.Now()
	defer metrics.GnfdChainHistogram.WithLabelValues("query_storage_params_by_timestamp").Observe(time.Since(startTime).Seconds())
	client := g.getCurrentClient().GnfdClient()
	resp, err := client.StorageQueryClient.QueryParamsByTimestamp(ctx,
		&storagetypes.QueryParamsByTimestampRequest{Timestamp: timestamp})
	if err != nil {
		log.CtxErrorw(ctx, "failed to query storage params", "error", err)
		return nil, err
	}
	return &resp.Params, nil
}

// QueryBucketInfo returns the bucket info by name.
func (g *Gnfd) QueryBucketInfo(ctx context.Context, bucket string) (*storagetypes.BucketInfo, error) {
	startTime := time.Now()
	defer metrics.GnfdChainHistogram.WithLabelValues("query_bucket").Observe(time.Since(startTime).Seconds())
	client := g.getCurrentClient().GnfdClient()
	resp, err := client.HeadBucket(ctx, &storagetypes.QueryHeadBucketRequest{BucketName: bucket})
	if err != nil {
		log.CtxErrorw(ctx, "failed to query bucket", "bucket_name", bucket, "error", err)
		return nil, err
	}
	return resp.GetBucketInfo(), nil
}

// QueryObjectInfo returns the object info by name.
func (g *Gnfd) QueryObjectInfo(ctx context.Context, bucket, object string) (*storagetypes.ObjectInfo, error) {
	startTime := time.Now()
	defer metrics.GnfdChainHistogram.WithLabelValues("query_object").Observe(time.Since(startTime).Seconds())
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

// QueryObjectInfoByID returns the object info by name.
func (g *Gnfd) QueryObjectInfoByID(ctx context.Context, objectID string) (*storagetypes.ObjectInfo, error) {
	startTime := time.Now()
	defer metrics.GnfdChainHistogram.WithLabelValues("query_object_by_id").Observe(time.Since(startTime).Seconds())
	client := g.getCurrentClient().GnfdClient()
	resp, err := client.HeadObjectById(ctx, &storagetypes.QueryHeadObjectByIdRequest{
		ObjectId: objectID,
	})
	if err != nil {
		log.CtxErrorw(ctx, "failed to query object", "object_id", objectID, "error", err)
		return nil, err
	}
	return resp.GetObjectInfo(), nil
}

// QueryBucketInfoAndObjectInfo returns bucket info and object info, if not found, return the corresponding error code
func (g *Gnfd) QueryBucketInfoAndObjectInfo(ctx context.Context, bucket, object string) (*storagetypes.BucketInfo,
	*storagetypes.ObjectInfo, error) {
	bucketInfo, err := g.QueryBucketInfo(ctx, bucket)
	if err != nil {
		return nil, nil, err
	}
	objectInfo, err := g.QueryObjectInfo(ctx, bucket, object)
	if err != nil {
		return bucketInfo, nil, err
	}
	return bucketInfo, objectInfo, nil
}

// ListenObjectSeal returns an indication of the object is sealed.
// TODO:: retrieve service support seal event subscription
func (g *Gnfd) ListenObjectSeal(ctx context.Context, objectID uint64, timeoutHeight int) (bool, error) {
	startTime := time.Now()
	defer metrics.GnfdChainHistogram.WithLabelValues("wait_object_seal").Observe(time.Since(startTime).Seconds())
	var (
		objectInfo *storagetypes.ObjectInfo
		err        error
	)
	for i := 0; i < timeoutHeight; i++ {
		objectInfo, err = g.QueryObjectInfoByID(ctx, strconv.FormatUint(objectID, 10))
		if err != nil {
			time.Sleep(ExpectedOutputBlockInternal * time.Second)
			continue
		}
		if objectInfo.GetObjectStatus() == storagetypes.OBJECT_STATUS_SEALED {
			log.CtxDebugw(ctx, "succeed to listen object stat")
			return true, nil
		}
		time.Sleep(ExpectedOutputBlockInternal * time.Second)
	}
	if err == nil {
		log.CtxErrorw(ctx, "seal object timeout", "object_id", objectID)
		return false, ErrSealTimeout
	}
	log.CtxErrorw(ctx, "failed to listen seal object", "object_id", objectID, "error", err)
	return false, err
}

// QueryPaymentStreamRecord returns the steam record info by account.
func (g *Gnfd) QueryPaymentStreamRecord(ctx context.Context, account string) (*paymenttypes.StreamRecord, error) {
	startTime := time.Now()
	defer metrics.GnfdChainHistogram.WithLabelValues("query_payment_stream_record").Observe(time.Since(startTime).Seconds())
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

// VerifyGetObjectPermission verifies get object permission.
func (g *Gnfd) VerifyGetObjectPermission(ctx context.Context, account, bucket, object string) (bool, error) {
	startTime := time.Now()
	defer metrics.GnfdChainHistogram.WithLabelValues("verify_get_object_permission").Observe(time.Since(startTime).Seconds())
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

// VerifyPutObjectPermission verifies put object permission.
func (g *Gnfd) VerifyPutObjectPermission(ctx context.Context, account, bucket, object string) (bool, error) {
	startTime := time.Now()
	defer metrics.GnfdChainHistogram.WithLabelValues("verify_put_object_permission").Observe(time.Since(startTime).Seconds())
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

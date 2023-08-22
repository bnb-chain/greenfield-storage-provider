package gater

import (
	"context"

	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

const (
	ExpireIntervalSecond      = 30 * 60 * time.Second
	CheckExpireIntervalSecond = 60 * time.Second
)

type SPCachePool struct {
	cachePool   sync.Map
	chainClient consensus.Consensus
}

type SPCacheItem struct {
	SPInfo                *sptypes.StorageProvider
	ExpireTimestampSecond int64
}

func NewSPCachePool(chainClient consensus.Consensus) *SPCachePool {
	spCachePool := &SPCachePool{sync.Map{}, chainClient}
	go spCachePool.loopCheckExpire()
	return spCachePool
}

func (s *SPCachePool) QuerySPByAddress(spAddr string) (*sptypes.StorageProvider, error) {
	item, found := s.cachePool.Load(spAddr)
	if found {
		return item.(*SPCacheItem).SPInfo, nil
	}
	spInfo, err := s.chainClient.QuerySP(context.Background(), spAddr)
	if err != nil {
		return nil, err
	}
	s.cachePool.Store(spInfo.GetOperatorAddress(), &SPCacheItem{SPInfo: spInfo, ExpireTimestampSecond: time.Now().Add(ExpireIntervalSecond).Unix()})
	return spInfo, nil
}

func (s *SPCachePool) loopCheckExpire() {
	ticker := time.NewTicker(CheckExpireIntervalSecond)
	for range ticker.C {
		s.cachePool.Range(func(k interface{}, v interface{}) bool {
			spCacheItem := v.(*SPCacheItem)
			if time.Now().Unix() > spCacheItem.ExpireTimestampSecond {
				s.cachePool.Delete(k)
			}
			return true
		})
	}
}

func (g *GateModular) getObjectChainMeta(ctx context.Context, objectName, bucketName string) (*storagetypes.ObjectInfo,
	*storagetypes.BucketInfo, *storagetypes.Params, error) {
	objectInfo, err := g.baseApp.Consensus().QueryObjectInfo(ctx, bucketName, objectName)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get object info from consensus", "error", err)
		return nil, nil, nil, ErrConsensusWithDetail("failed to get object info from consensus, error: " + err.Error())
	}

	bucketInfo, err := g.baseApp.Consensus().QueryBucketInfo(ctx, objectInfo.GetBucketName())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket info from consensus", "error", err)
		return nil, nil, nil, ErrConsensusWithDetail("failed to get bucket info from consensus, error: " + err.Error())
	}

	params, err := g.baseApp.Consensus().QueryStorageParamsByTimestamp(ctx, objectInfo.GetCreateAt())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get storage params", "error", err)
		return nil, nil, nil, ErrConsensusWithDetail("failed to get storage params, error: " + err.Error())
	}

	return objectInfo, bucketInfo, params, nil
}

// checkSPAndBucketStatus check sp and bucket is in right status
func (g *GateModular) checkSPAndBucketStatus(ctx context.Context, bucketName string, creatorAddr string) error {
	spInfo, err := g.baseApp.Consensus().QuerySP(ctx, g.baseApp.OperatorAddress())
	if err != nil {
		log.Errorw("failed to query sp by operator address", "operator_address", g.baseApp.OperatorAddress(),
			"error", err)
		return ErrConsensusWithDetail("failed to query sp by operator address, operator_address: " + g.baseApp.OperatorAddress() +
			", error: " + err.Error())
	}
	spStatus := spInfo.GetStatus()
	if spStatus != sptypes.STATUS_IN_SERVICE && !fromSpMaintenanceAcct(spStatus, spInfo.MaintenanceAddress, creatorAddr) {
		log.Errorw("sp is not in service status", "operator_address", g.baseApp.OperatorAddress(),
			"sp_status", spStatus, "sp_id", spInfo.GetId(), "endpoint", spInfo.GetEndpoint())
		return ErrSPUnavailable
	}

	bucketInfo, err := g.baseApp.Consensus().QueryBucketInfo(ctx, bucketName)
	if err != nil {
		log.Errorw("failed to query bucket info by bucket name", "bucket_name", bucketName, "error", err)
		return ErrConsensusWithDetail("failed to query bucket info by bucket name, bucket_name: " + bucketName + ", error: " + err.Error())
	}
	bucketStatus := bucketInfo.GetBucketStatus()
	if bucketStatus != storagetypes.BUCKET_STATUS_CREATED {
		log.Errorw("bucket is not in created status", "bucket_name", bucketName, "bucket_status", bucketStatus,
			"bucket_id", bucketInfo.Id.String())
		return ErrBucketUnavailable
	}
	return nil
}

// getBucketTotalSize return the total size of the bucket
func (g *GateModular) getBucketTotalSize(ctx context.Context, bucketID uint64) (uint64, error) {
	bucketSize, err := g.baseApp.GfSpClient().GetBucketSize(
		ctx, bucketID)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket size", "bucket_id", bucketID, "error", err)
		return 0, err
	}
	quotaNeed, err := util.StringToUint64(bucketSize)
	if err != nil {
		log.CtxErrorw(ctx, "failed to convert bucket size to uint64", "bucket_id",
			bucketID, "bucket_size", bucketSize, "error", err)
		return 0, err
	}
	return quotaNeed, nil
}
func fromSpMaintenanceAcct(spStatus sptypes.Status, spMaintenanceAddr, creatorAddr string) bool {
	return spStatus == sptypes.STATUS_IN_MAINTENANCE && spMaintenanceAddr == creatorAddr
}

package metadata

import (
	"context"
	"cosmossdk.io/math"
	"errors"
	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"
	"net/http"
	"sync/atomic"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	model "github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield/types/s3util"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	storage_types "github.com/bnb-chain/greenfield/x/storage/types"
	virtual_types "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var (
	ErrDanglingPointer   = gfsperrors.Register(coremodule.MetadataModularName, http.StatusBadRequest, 90001, "OoooH... request lost, try again later")
	ErrExceedRequest     = gfsperrors.Register(coremodule.MetadataModularName, http.StatusNotAcceptable, 90002, "request exceed")
	ErrNoRecord          = gfsperrors.Register(coremodule.MetadataModularName, http.StatusNotFound, 90003, "no uploading record")
	ErrNoSuchSP          = gfsperrors.Register(coremodule.MetadataModularName, http.StatusNotFound, 90004, "no such sp")
	ErrExceedBlockHeight = gfsperrors.Register(coremodule.MetadataModularName, http.StatusBadRequest, 90005, "request block height exceed latest height")
	// ErrInvalidParams defines invalid params
	ErrInvalidParams = gfsperrors.Register(coremodule.MetadataModularName, http.StatusBadRequest, 90006, "invalid params")
	// ErrInvalidBucketName defines invalid bucket name
	ErrInvalidBucketName = gfsperrors.Register(coremodule.MetadataModularName, http.StatusBadRequest, 90007, "invalid bucket name")
	// ErrNoSuchBucket defines not existed bucket error
	ErrNoSuchBucket = gfsperrors.Register(coremodule.MetadataModularName, http.StatusNotFound, 90008, "the specified bucket does not exist")
	// ErrNoSuchGroup defines not existed group error
	ErrNoSuchGroup = gfsperrors.Register(coremodule.MetadataModularName, http.StatusNotFound, 90009, "the specified group does not exist")
	// ErrNoSuchObject defines not existed object error
	ErrNoSuchObject = gfsperrors.Register(coremodule.MetadataModularName, http.StatusNotFound, 90010, "the specified object does not exist")
)

var _ types.GfSpMetadataServiceServer = &MetadataModular{}

func ErrGfSpDBWithDetail(detail string) *gfsperrors.GfSpError {
	return gfsperrors.Register(coremodule.MetadataModularName, http.StatusInternalServerError, 95202, detail)
}

func (r *MetadataModular) GfSpGetUserBuckets(ctx context.Context, req *types.GfSpGetUserBucketsRequest) (
	resp *types.GfSpGetUserBucketsResponse, err error) {
	ctx = log.Context(ctx, req)
	buckets, err := r.baseApp.GfBsDB().GetUserBuckets(common.HexToAddress(req.AccountId), req.GetIncludeRemoved())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get user buckets", "error", err)
		return
	}

	vgfIDs := make([]uint32, len(buckets))
	for i, bucket := range buckets {
		vgfIDs[i] = bucket.GlobalVirtualGroupFamilyID
	}

	families, err := r.baseApp.GfBsDB().ListVirtualGroupFamiliesByVgfIDs(vgfIDs)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list vgf by vgf ids", "error", err)
		return
	}

	vgfMap := make(map[uint32]*model.GlobalVirtualGroupFamily)
	for _, family := range families {
		vgfMap[family.GlobalVirtualGroupFamilyId] = family
	}

	res := make([]*types.VGFInfoBucket, 0)
	for _, bucket := range buckets {
		vgf := vgfMap[bucket.GlobalVirtualGroupFamilyID]
		b := &types.VGFInfoBucket{
			BucketInfo: &storage_types.BucketInfo{
				Owner:                      bucket.Owner.String(),
				BucketName:                 bucket.BucketName,
				Visibility:                 storage_types.VisibilityType(storage_types.VisibilityType_value[bucket.Visibility]),
				Id:                         math.NewUintFromBigInt(bucket.BucketID.Big()),
				SourceType:                 storage_types.SourceType(storage_types.SourceType_value[bucket.SourceType]),
				CreateAt:                   bucket.CreateTime,
				PaymentAddress:             bucket.PaymentAddress.String(),
				GlobalVirtualGroupFamilyId: bucket.GlobalVirtualGroupFamilyID,
				ChargedReadQuota:           bucket.ChargedReadQuota,
				BucketStatus:               storage_types.BucketStatus(storage_types.BucketStatus_value[bucket.Status]),
			},
			Removed:      bucket.Removed,
			DeleteAt:     bucket.DeleteAt,
			DeleteReason: bucket.DeleteReason,
			Operator:     bucket.Operator.String(),
			CreateTxHash: bucket.CreateTxHash.String(),
			UpdateTxHash: bucket.UpdateTxHash.String(),
			UpdateAt:     bucket.UpdateAt,
			UpdateTime:   bucket.UpdateTime,
			StorageSize:  bucket.StorageSize.String(),
		}
		if vgf != nil {
			b.Vgf = &virtual_types.GlobalVirtualGroupFamily{
				Id:                    vgf.GlobalVirtualGroupFamilyId,
				PrimarySpId:           vgf.PrimarySpId,
				GlobalVirtualGroupIds: vgf.GlobalVirtualGroupIds,
				VirtualPaymentAddress: vgf.VirtualPaymentAddress.String(),
			}
		}
		res = append(res, b)
	}
	resp = &types.GfSpGetUserBucketsResponse{Buckets: res}
	log.CtxInfow(ctx, "succeed to get user buckets")
	return resp, nil
}

// GfSpGetBucketByBucketName get buckets info by a bucket name
func (r *MetadataModular) GfSpGetBucketByBucketName(ctx context.Context, req *types.GfSpGetBucketByBucketNameRequest) (resp *types.GfSpGetBucketByBucketNameResponse, err error) {
	var (
		bucket *model.Bucket
		res    *types.Bucket
	)

	ctx = log.Context(ctx, req)
	if err = s3util.CheckValidBucketName(req.BucketName); err != nil {
		log.Errorw("failed to check bucket name", "bucket_name", req.BucketName, "error", err)
		return nil, ErrInvalidBucketName
	}

	bucket, err = r.baseApp.GfBsDB().GetBucketByName(req.BucketName, req.IncludePrivate)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket by bucket name", "error", err)
		return nil, err
	}

	if bucket != nil {
		res = &types.Bucket{
			BucketInfo: &storage_types.BucketInfo{
				Owner:                      bucket.Owner.String(),
				BucketName:                 bucket.BucketName,
				Visibility:                 storage_types.VisibilityType(storage_types.VisibilityType_value[bucket.Visibility]),
				Id:                         math.NewUintFromBigInt(bucket.BucketID.Big()),
				SourceType:                 storage_types.SourceType(storage_types.SourceType_value[bucket.SourceType]),
				CreateAt:                   bucket.CreateTime,
				PaymentAddress:             bucket.PaymentAddress.String(),
				GlobalVirtualGroupFamilyId: bucket.GlobalVirtualGroupFamilyID,
				ChargedReadQuota:           bucket.ChargedReadQuota,
				BucketStatus:               storage_types.BucketStatus(storage_types.BucketStatus_value[bucket.Status]),
			},
			Removed:      bucket.Removed,
			DeleteAt:     bucket.DeleteAt,
			DeleteReason: bucket.DeleteReason,
			Operator:     bucket.Operator.String(),
			CreateTxHash: bucket.CreateTxHash.String(),
			UpdateTxHash: bucket.UpdateTxHash.String(),
			UpdateAt:     bucket.UpdateAt,
			UpdateTime:   bucket.UpdateTime,
			StorageSize:  bucket.StorageSize.String(),
		}
	}
	resp = &types.GfSpGetBucketByBucketNameResponse{Bucket: res}
	log.CtxInfo(ctx, "succeed to get bucket by bucket name")
	return resp, nil
}

// GfSpGetBucketByBucketID get buckets info by by a bucket id
func (r *MetadataModular) GfSpGetBucketByBucketID(ctx context.Context, req *types.GfSpGetBucketByBucketIDRequest) (
	resp *types.GfSpGetBucketByBucketIDResponse, err error) {
	var (
		bucket *model.Bucket
		res    *types.Bucket
	)

	ctx = log.Context(ctx, req)
	bucket, err = r.baseApp.GfBsDB().GetBucketByID(req.BucketId, req.IncludePrivate)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket by bucket id", "error", err)
		return nil, err
	}

	if bucket != nil {
		res = &types.Bucket{
			BucketInfo: &storage_types.BucketInfo{
				Owner:                      bucket.Owner.String(),
				BucketName:                 bucket.BucketName,
				Visibility:                 storage_types.VisibilityType(storage_types.VisibilityType_value[bucket.Visibility]),
				Id:                         math.NewUintFromBigInt(bucket.BucketID.Big()),
				SourceType:                 storage_types.SourceType(storage_types.SourceType_value[bucket.SourceType]),
				CreateAt:                   bucket.CreateTime,
				PaymentAddress:             bucket.PaymentAddress.String(),
				GlobalVirtualGroupFamilyId: bucket.GlobalVirtualGroupFamilyID,
				ChargedReadQuota:           bucket.ChargedReadQuota,
				BucketStatus:               storage_types.BucketStatus(storage_types.BucketStatus_value[bucket.Status]),
			},
			Removed:      bucket.Removed,
			DeleteAt:     bucket.DeleteAt,
			DeleteReason: bucket.DeleteReason,
			Operator:     bucket.Operator.String(),
			CreateTxHash: bucket.CreateTxHash.String(),
			UpdateTxHash: bucket.UpdateTxHash.String(),
			UpdateAt:     bucket.UpdateAt,
			UpdateTime:   bucket.UpdateTime,
			StorageSize:  bucket.StorageSize.String(),
		}
	}
	resp = &types.GfSpGetBucketByBucketIDResponse{Bucket: res}
	log.CtxInfow(ctx, "succeed to get bucket by bucket id")
	return resp, nil
}

// GfSpGetUserBucketsCount get buckets count by a user address
func (r *MetadataModular) GfSpGetUserBucketsCount(ctx context.Context, req *types.GfSpGetUserBucketsCountRequest) (
	resp *types.GfSpGetUserBucketsCountResponse, err error) {
	ctx = log.Context(ctx, req)

	count, err := r.baseApp.GfBsDB().GetUserBucketsCount(common.HexToAddress(req.AccountId), req.GetIncludeRemoved())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get user buckets count", "error", err)
		return
	}

	resp = &types.GfSpGetUserBucketsCountResponse{Count: count}
	log.CtxInfow(ctx, "succeed to get buckets count by a user address")
	return resp, nil
}

// GfSpListExpiredBucketsBySp list expired bucket by sp
func (r *MetadataModular) GfSpListExpiredBucketsBySp(ctx context.Context, req *types.GfSpListExpiredBucketsBySpRequest) (
	resp *types.GfSpListExpiredBucketsBySpResponse, err error) {
	ctx = log.Context(ctx, req)
	buckets, err := r.baseApp.GfBsDB().ListExpiredBucketsBySp(req.GetCreateAt(), req.GetPrimarySpId(), req.GetLimit())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get user buckets", "error", err)
		return
	}

	res := make([]*types.Bucket, 0)
	for _, bucket := range buckets {
		res = append(res, &types.Bucket{
			BucketInfo: &storage_types.BucketInfo{
				Owner:                      bucket.Owner.String(),
				BucketName:                 bucket.BucketName,
				Visibility:                 storage_types.VisibilityType(storage_types.VisibilityType_value[bucket.Visibility]),
				Id:                         math.NewUintFromBigInt(bucket.BucketID.Big()),
				SourceType:                 storage_types.SourceType(storage_types.SourceType_value[bucket.SourceType]),
				CreateAt:                   bucket.CreateTime,
				PaymentAddress:             bucket.PaymentAddress.String(),
				GlobalVirtualGroupFamilyId: bucket.GlobalVirtualGroupFamilyID,
				ChargedReadQuota:           bucket.ChargedReadQuota,
				BucketStatus:               storage_types.BucketStatus(storage_types.BucketStatus_value[bucket.Status]),
			},
			Removed:      bucket.Removed,
			DeleteAt:     bucket.DeleteAt,
			DeleteReason: bucket.DeleteReason,
		})
	}
	resp = &types.GfSpListExpiredBucketsBySpResponse{Buckets: res}
	log.CtxInfow(ctx, "succeed to get user buckets")
	return resp, nil
}

// GfSpGetBucketMeta get bucket metadata
func (r *MetadataModular) GfSpGetBucketMeta(ctx context.Context, req *types.GfSpGetBucketMetaRequest) (
	resp *types.GfSpGetBucketMetaResponse, err error) {
	var (
		bucket          *model.Bucket
		bucketRes       *types.VGFInfoBucket
		streamRecord    *model.StreamRecord
		streamRecordRes *paymenttypes.StreamRecord
	)

	ctx = log.Context(ctx, req)
	bucketFullMeta, err := r.baseApp.GfBsDB().GetBucketMetaByName(req.GetBucketName(), req.GetIncludePrivate())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket meta by name", "error", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoSuchBucket
		}
		return
	}
	bucket = &bucketFullMeta.Bucket
	streamRecord = &bucketFullMeta.StreamRecord

	family, err := r.baseApp.GfBsDB().ListVirtualGroupFamiliesByVgfIDs([]uint32{bucket.GlobalVirtualGroupFamilyID})
	if err != nil {
		log.CtxErrorw(ctx, "failed to list vgf by vgf ids", "error", err)
		return
	}

	if bucket != nil {
		bucketRes = &types.VGFInfoBucket{
			BucketInfo: &storage_types.BucketInfo{
				Owner:                      bucket.Owner.String(),
				BucketName:                 bucket.BucketName,
				Visibility:                 storage_types.VisibilityType(storage_types.VisibilityType_value[bucket.Visibility]),
				Id:                         math.NewUintFromBigInt(bucket.BucketID.Big()),
				SourceType:                 storage_types.SourceType(storage_types.SourceType_value[bucket.SourceType]),
				CreateAt:                   bucket.CreateTime,
				PaymentAddress:             bucket.PaymentAddress.String(),
				GlobalVirtualGroupFamilyId: bucket.GlobalVirtualGroupFamilyID,
				ChargedReadQuota:           bucket.ChargedReadQuota,
				BucketStatus:               storage_types.BucketStatus(storage_types.BucketStatus_value[bucket.Status]),
			},
			Removed:      bucket.Removed,
			DeleteAt:     bucket.DeleteAt,
			DeleteReason: bucket.DeleteReason,
			Operator:     bucket.Operator.String(),
			CreateTxHash: bucket.CreateTxHash.String(),
			UpdateTxHash: bucket.UpdateTxHash.String(),
			UpdateAt:     bucket.UpdateAt,
			UpdateTime:   bucket.UpdateTime,
			Vgf: &virtual_types.GlobalVirtualGroupFamily{
				Id:                    family[0].GlobalVirtualGroupFamilyId,
				PrimarySpId:           family[0].PrimarySpId,
				GlobalVirtualGroupIds: family[0].GlobalVirtualGroupIds,
				VirtualPaymentAddress: family[0].VirtualPaymentAddress.String(),
			},
			StorageSize: bucket.StorageSize.String(),
		}
	}

	if streamRecord != nil {
		streamRecordRes = &paymenttypes.StreamRecord{
			Account:           streamRecord.Account.String(),
			CrudTimestamp:     streamRecord.CrudTimestamp,
			NetflowRate:       math.NewIntFromBigInt(streamRecord.NetflowRate.Raw()),
			StaticBalance:     math.NewIntFromBigInt(streamRecord.StaticBalance.Raw()),
			BufferBalance:     math.NewIntFromBigInt(streamRecord.BufferBalance.Raw()),
			LockBalance:       math.NewIntFromBigInt(streamRecord.LockBalance.Raw()),
			Status:            paymenttypes.StreamAccountStatus(paymenttypes.StreamAccountStatus_value[streamRecord.Status]),
			SettleTimestamp:   streamRecord.SettleTimestamp,
			OutFlowCount:      streamRecord.OutFlowCount,
			FrozenNetflowRate: math.NewIntFromBigInt(streamRecord.FrozenNetflowRate.Raw()),
		}
	}

	resp = &types.GfSpGetBucketMetaResponse{Bucket: bucketRes, StreamRecord: streamRecordRes}
	log.CtxInfow(ctx, "succeed to get bucket meta by name")
	return resp, nil
}

func (r *MetadataModular) GfSpGetBucketReadQuota(
	ctx context.Context,
	req *types.GfSpGetBucketReadQuotaRequest) (
	*types.GfSpGetBucketReadQuotaResponse, error) {
	if req.GetBucketInfo() == nil {
		return nil, ErrDanglingPointer
	}
	defer atomic.AddInt64(&r.retrievingRequest, -1)
	if atomic.AddInt64(&r.retrievingRequest, 1) >
		atomic.LoadInt64(&r.maxMetadataRequest) {
		return nil, ErrExceedRequest
	}
	bucketTraffic, err := r.baseApp.GfSpDB().GetBucketTraffic(
		req.GetBucketInfo().Id.Uint64(), req.YearMonth)
	if err != nil {
		// if the traffic table has not been created and initialized yet, return the chain info
		if errors.Is(err, gorm.ErrRecordNotFound) || bucketTraffic == nil {
			var freeQuotaSize uint64
			freeQuotaSize, err = r.baseApp.Consensus().QuerySPFreeQuota(ctx, r.baseApp.OperatorAddress())
			if err != nil {
				log.Errorw("failed to get free quota on chain",
					"bucket_name", req.GetBucketInfo().GetBucketName(),
					"bucket_id", req.GetBucketInfo().Id.String(), "error", err)
				freeQuotaSize = 0
			}

			return &types.GfSpGetBucketReadQuotaResponse{
				ChargedQuotaSize:     req.GetBucketInfo().GetChargedReadQuota(),
				SpFreeQuotaSize:      freeQuotaSize,
				ConsumedSize:         0,
				FreeQuotaConsumeSize: 0,
			}, nil
		} else {
			log.Errorw("failed to get bucket traffic",
				"bucket_name", req.GetBucketInfo().GetBucketName(),
				"bucket_id", req.GetBucketInfo().Id.String(), "error", err)
			return &types.GfSpGetBucketReadQuotaResponse{Err: ErrGfSpDBWithDetail("failed to get bucket traffic" +
				", bucket_name: " + req.GetBucketInfo().GetBucketName() +
				", bucket_id: " + req.GetBucketInfo().Id.String() + ", error: " + err.Error())}, nil
		}
	}
	// if the traffic table has been created, return the db info from meta service
	return &types.GfSpGetBucketReadQuotaResponse{
		ChargedQuotaSize:     req.GetBucketInfo().GetChargedReadQuota(),
		SpFreeQuotaSize:      bucketTraffic.FreeQuotaSize,
		ConsumedSize:         bucketTraffic.ReadConsumedSize,
		FreeQuotaConsumeSize: bucketTraffic.FreeQuotaConsumedSize,
	}, nil
}

func (r *MetadataModular) GfSpGetLatestBucketReadQuota(
	ctx context.Context,
	req *types.GfSpGetLatestBucketReadQuotaRequest) (
	*types.GfSpGetLatestBucketReadQuotaResponse, error) {

	defer atomic.AddInt64(&r.retrievingRequest, -1)
	if atomic.AddInt64(&r.retrievingRequest, 1) >
		atomic.LoadInt64(&r.maxMetadataRequest) {
		return nil, ErrExceedRequest
	}
	bucketTraffic, err := r.baseApp.GfSpDB().GetLatestBucketTraffic(
		req.GetBucketId())
	if err != nil {
		// if the traffic table has not been created and initialized yet, return the chain info
		if errors.Is(err, gorm.ErrRecordNotFound) || bucketTraffic == nil {
			var freeQuotaSize uint64
			freeQuotaSize, err = r.baseApp.Consensus().QuerySPFreeQuota(ctx, r.baseApp.OperatorAddress())
			if err != nil {
				log.Errorw("failed to get free quota on chain",
					"bucket_id", req.GetBucketId(), "error", err)
				freeQuotaSize = 0
			}
			quota := &gfsptask.GfSpBucketQuotaInfo{
				BucketName:            bucketTraffic.BucketName,
				BucketId:              bucketTraffic.BucketID,
				Month:                 bucketTraffic.YearMonth,
				ReadConsumedSize:      bucketTraffic.ReadConsumedSize,
				FreeQuotaConsumedSize: bucketTraffic.FreeQuotaConsumedSize,
				FreeQuotaSize:         freeQuotaSize,
				ChargedQuotaSize:      bucketTraffic.ChargedQuotaSize,
			}

			return &types.GfSpGetLatestBucketReadQuotaResponse{
				Quota: quota,
			}, nil
		} else {
			log.Errorw("failed to get bucket traffic",
				"bucket_id", req.GetBucketId(), "error", err)
			return &types.GfSpGetLatestBucketReadQuotaResponse{Err: ErrGfSpDBWithDetail("failed to get bucket traffic" +
				", bucket_id: " + util.Uint64ToString(req.GetBucketId()) + ", error: " + err.Error())}, nil
		}
	}
	// if the traffic table has been created, return the db info from meta service
	quota := &gfsptask.GfSpBucketQuotaInfo{
		BucketName:            bucketTraffic.BucketName,
		BucketId:              bucketTraffic.BucketID,
		Month:                 bucketTraffic.YearMonth,
		ReadConsumedSize:      bucketTraffic.ReadConsumedSize,
		FreeQuotaConsumedSize: bucketTraffic.FreeQuotaConsumedSize,
		FreeQuotaSize:         bucketTraffic.FreeQuotaSize,
		ChargedQuotaSize:      bucketTraffic.ChargedQuotaSize,
	}
	return &types.GfSpGetLatestBucketReadQuotaResponse{
		Quota: quota,
	}, nil
}

func (r *MetadataModular) GfSpListBucketReadRecord(
	ctx context.Context,
	req *types.GfSpListBucketReadRecordRequest) (
	*types.GfSpListBucketReadRecordResponse,
	error) {
	if req.GetBucketInfo() == nil {
		return nil, ErrDanglingPointer
	}
	defer atomic.AddInt64(&r.retrievingRequest, -1)
	if atomic.AddInt64(&r.retrievingRequest, 1) >
		atomic.LoadInt64(&r.maxMetadataRequest) {
		return nil, ErrExceedRequest
	}
	records, err := r.baseApp.GfSpDB().GetBucketReadRecord(req.GetBucketInfo().Id.Uint64(),
		&spdb.TrafficTimeRange{
			StartTimestampUs: req.StartTimestampUs,
			EndTimestampUs:   req.EndTimestampUs,
			LimitNum:         int(req.MaxRecordNum),
		})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &types.GfSpListBucketReadRecordResponse{
			NextStartTimestampUs: 0,
		}, nil
	}
	if err != nil {
		log.Errorw("failed to list bucket read record",
			"bucket_name", req.GetBucketInfo().GetBucketName(),
			"bucket_id", req.GetBucketInfo().Id.String(), "error", err)
		return &types.GfSpListBucketReadRecordResponse{Err: ErrGfSpDBWithDetail("failed to list bucket read record" +
			", bucket_name: " + req.GetBucketInfo().GetBucketName() +
			", bucket_id: " + req.GetBucketInfo().Id.String() + ",error: " + err.Error())}, nil
	}
	var nextStartTimestampUs int64
	readRecords := make([]*types.ReadRecord, 0)
	for _, record := range records {
		readRecords = append(readRecords, &types.ReadRecord{
			ObjectName:     record.ObjectName,
			ObjectId:       record.ObjectID,
			AccountAddress: record.UserAddress,
			TimestampUs:    record.ReadTimestampUs,
			ReadSize:       record.ReadSize,
		})
		if record.ReadTimestampUs >= nextStartTimestampUs {
			nextStartTimestampUs = record.ReadTimestampUs + 1
		}
	}
	resp := &types.GfSpListBucketReadRecordResponse{
		ReadRecords:          readRecords,
		NextStartTimestampUs: nextStartTimestampUs,
	}
	return resp, nil
}

// GfSpListBucketsByIDs list buckets by bucket ids
func (r *MetadataModular) GfSpListBucketsByIDs(ctx context.Context, req *types.GfSpListBucketsByIDsRequest) (
	resp *types.GfSpListBucketsByIDsResponse, err error) {
	var (
		buckets    []*model.Bucket
		ids        []common.Hash
		bucketsMap map[uint64]*types.Bucket
	)

	ids = make([]common.Hash, len(req.BucketIds))
	for i, id := range req.BucketIds {
		ids[i] = common.BigToHash(math.NewUint(id).BigInt())
	}

	ctx = log.Context(ctx, req)
	buckets, err = r.baseApp.GfBsDB().ListBucketsByIDs(ids, req.IncludeRemoved)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list buckets by bucket ids", "error", err)
		return nil, err
	}

	bucketsMap = make(map[uint64]*types.Bucket)
	for _, id := range req.BucketIds {
		bucketsMap[id] = nil
	}

	for _, bucket := range buckets {
		bucketsMap[bucket.BucketID.Big().Uint64()] = &types.Bucket{
			BucketInfo: &storage_types.BucketInfo{
				Owner:                      bucket.Owner.String(),
				BucketName:                 bucket.BucketName,
				Visibility:                 storage_types.VisibilityType(storage_types.VisibilityType_value[bucket.Visibility]),
				Id:                         math.NewUintFromBigInt(bucket.BucketID.Big()),
				SourceType:                 storage_types.SourceType(storage_types.SourceType_value[bucket.SourceType]),
				CreateAt:                   bucket.CreateTime,
				PaymentAddress:             bucket.PaymentAddress.String(),
				GlobalVirtualGroupFamilyId: bucket.GlobalVirtualGroupFamilyID,
				ChargedReadQuota:           bucket.ChargedReadQuota,
				BucketStatus:               storage_types.BucketStatus(storage_types.BucketStatus_value[bucket.Status]),
			},
			Removed:      bucket.Removed,
			DeleteAt:     bucket.DeleteAt,
			DeleteReason: bucket.DeleteReason,
			Operator:     bucket.Operator.String(),
			CreateTxHash: bucket.CreateTxHash.String(),
			UpdateTxHash: bucket.UpdateTxHash.String(),
			UpdateAt:     bucket.UpdateAt,
			UpdateTime:   bucket.UpdateTime,
			StorageSize:  bucket.StorageSize.String(),
		}
	}
	resp = &types.GfSpListBucketsByIDsResponse{Buckets: bucketsMap}
	log.CtxInfow(ctx, "succeed to list buckets by bucket ids")
	return resp, nil
}

// GfSpGetBucketSize get bucket total object size
func (r *MetadataModular) GfSpGetBucketSize(ctx context.Context, req *types.GfSpGetBucketSizeRequest) (resp *types.GfSpGetBucketSizeResponse, err error) {
	ctx = log.Context(ctx, req)

	size, err := r.baseApp.GfBsDB().GetBucketSizeByID(req.BucketId)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket total object size", "error", err)
		return
	}

	resp = &types.GfSpGetBucketSizeResponse{BucketSize: size.String()}
	log.CtxInfow(ctx, "succeed to get bucket total object size", "bucket_size", size)
	return resp, nil
}

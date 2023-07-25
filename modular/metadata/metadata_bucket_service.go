package metadata

import (
	"context"
	systemerrors "errors"
	"net/http"
	"sync/atomic"

	"cosmossdk.io/math"
	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	model "github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	"github.com/bnb-chain/greenfield/types/s3util"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	storage_types "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	ErrDanglingPointer = gfsperrors.Register(MetadataModularName, http.StatusBadRequest, 90001, "OoooH... request lost, try again later")
	ErrExceedRequest   = gfsperrors.Register(MetadataModularName, http.StatusNotAcceptable, 90002, "request exceed")
	ErrNoRecord        = gfsperrors.Register(MetadataModularName, http.StatusNotFound, 90003, "no uploading record")
	ErrGfSpDB          = gfsperrors.Register(MetadataModularName, http.StatusInternalServerError, 95202, "server slipped away, try again later")
	ErrNoSuchSP        = gfsperrors.Register(MetadataModularName, http.StatusNotFound, 90004, "no such sp")
)

var _ types.GfSpMetadataServiceServer = &MetadataModular{}

func (r *MetadataModular) GfSpGetUserBuckets(
	ctx context.Context,
	req *types.GfSpGetUserBucketsRequest) (
	resp *types.GfSpGetUserBucketsResponse, err error) {
	ctx = log.Context(ctx, req)
	buckets, err := r.baseApp.GfBsDB().GetUserBuckets(common.HexToAddress(req.AccountId), req.GetIncludeRemoved())
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
			Operator:     bucket.Operator.String(),
			CreateTxHash: bucket.CreateTxHash.String(),
			UpdateTxHash: bucket.UpdateTxHash.String(),
			UpdateAt:     bucket.UpdateAt,
			UpdateTime:   bucket.UpdateTime,
		})
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
		return nil, err
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
		}
	}
	resp = &types.GfSpGetBucketByBucketNameResponse{Bucket: res}
	log.CtxInfo(ctx, "succeed to get bucket by bucket name")
	return resp, nil
}

// GfSpGetBucketByBucketID get buckets info by by a bucket id
func (r *MetadataModular) GfSpGetBucketByBucketID(ctx context.Context, req *types.GfSpGetBucketByBucketIDRequest) (resp *types.GfSpGetBucketByBucketIDResponse, err error) {
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
		}
	}
	resp = &types.GfSpGetBucketByBucketIDResponse{Bucket: res}
	log.CtxInfow(ctx, "succeed to get bucket by bucket id")
	return resp, nil
}

// GfSpGetUserBucketsCount get buckets count by a user address
func (r *MetadataModular) GfSpGetUserBucketsCount(ctx context.Context, req *types.GfSpGetUserBucketsCountRequest) (resp *types.GfSpGetUserBucketsCountResponse, err error) {
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
func (r *MetadataModular) GfSpListExpiredBucketsBySp(ctx context.Context, req *types.GfSpListExpiredBucketsBySpRequest) (resp *types.GfSpListExpiredBucketsBySpResponse, err error) {
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
func (r *MetadataModular) GfSpGetBucketMeta(
	ctx context.Context,
	req *types.GfSpGetBucketMetaRequest) (
	resp *types.GfSpGetBucketMetaResponse, err error) {
	var (
		bucket          *model.Bucket
		bucketRes       *types.Bucket
		streamRecord    *model.StreamRecord
		streamRecordRes *paymenttypes.StreamRecord
	)

	ctx = log.Context(ctx, req)
	bucketFullMeta, err := r.baseApp.GfBsDB().GetBucketMetaByName(req.GetBucketName(), req.GetIncludePrivate())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket meta by name", "error", err)
		return
	}
	bucket = &bucketFullMeta.Bucket
	streamRecord = &bucketFullMeta.StreamRecord

	if bucket != nil {
		bucketRes = &types.Bucket{
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
		req.GetBucketInfo().Id.Uint64(), req.GetYearMonth())
	if systemerrors.Is(err, gorm.ErrRecordNotFound) {
		return &types.GfSpGetBucketReadQuotaResponse{
			ChargedQuotaSize: req.GetBucketInfo().GetChargedReadQuota(),
			SpFreeQuotaSize:  r.freeQuotaPerBucket,
			ConsumedSize:     0,
		}, nil
	}
	if err != nil {
		log.Errorw("failed to get bucket traffic",
			"bucket_name", req.GetBucketInfo().GetBucketName(),
			"bucket_id", req.GetBucketInfo().Id.String(), "error", err)
		return &types.GfSpGetBucketReadQuotaResponse{Err: ErrGfSpDB}, nil
	}
	return &types.GfSpGetBucketReadQuotaResponse{
		ChargedQuotaSize: req.GetBucketInfo().GetChargedReadQuota(),
		SpFreeQuotaSize:  r.freeQuotaPerBucket,
		ConsumedSize:     bucketTraffic.ReadConsumedSize,
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
	if systemerrors.Is(err, gorm.ErrRecordNotFound) {
		return &types.GfSpListBucketReadRecordResponse{
			NextStartTimestampUs: 0,
		}, nil
	}
	if err != nil {
		log.Errorw("failed to list bucket read record",
			"bucket_name", req.GetBucketInfo().GetBucketName(),
			"bucket_id", req.GetBucketInfo().Id.String(), "error", err)
		return &types.GfSpListBucketReadRecordResponse{Err: ErrGfSpDB}, nil
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

// GfSpListBucketsByBucketID list buckets by bucket ids
func (r *MetadataModular) GfSpListBucketsByBucketID(ctx context.Context, req *types.GfSpListBucketsByBucketIDRequest) (resp *types.GfSpListBucketsByBucketIDResponse, err error) {
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
	buckets, err = r.baseApp.GfBsDB().ListBucketsByBucketID(ids, req.IncludeRemoved)
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
		}
	}
	resp = &types.GfSpListBucketsByBucketIDResponse{Buckets: bucketsMap}
	log.CtxInfow(ctx, "succeed to list buckets by bucket ids")
	return resp, nil
}

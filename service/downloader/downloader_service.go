package downloader

import (
	"context"
	"errors"
	"fmt"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	"gorm.io/gorm"
)

var _ types.DownloaderServiceServer = &Downloader{}

// GetObject downloads the payload of the object.
func (downloader *Downloader) GetObject(req *types.GetObjectRequest,
	stream types.DownloaderService_GetObjectServer) (err error) {
	if req.GetObjectInfo() == nil {
		return merrors.ErrDanglingPointer
	}
	var (
		scope       rcmgr.ResourceScopeSpan
		sendSize    int
		objectInfo  = req.GetObjectInfo()
		bucketInfo  = req.GetBucketInfo()
		resp        = &types.GetObjectResponse{}
		startOffset uint64
		endOffset   uint64
		ctx         = log.WithValue(context.Background(), "object_id", objectInfo.Id.String())
	)
	defer func() {
		if scope != nil {
			scope.Done()
		}
		log.CtxInfow(ctx, "finish to get object", "send_size", sendSize,
			"resource_state", rcmgr.GetServiceState(model.DownloaderService), "error", err)
	}()

	scope, err = downloader.rcScope.BeginSpan()
	if err != nil {
		log.CtxErrorw(ctx, "failed to begin reserve resource", "error", err)
		return
	}
	startOffset = uint64(0)
	endOffset = objectInfo.GetPayloadSize() - 1
	if req.GetIsRange() {
		startOffset = req.GetRangeStart()
		endOffset = req.GetRangeEnd()
	}
	readSize := endOffset - startOffset + 1
	err = scope.ReserveMemory(int(readSize), rcmgr.ReservationPriorityAlways)
	if err != nil {
		log.CtxErrorw(ctx, "failed to reserve memory from resource manager",
			"reserve_size", readSize, "error", err)
		return
	}
	if err = downloader.spDB.CheckQuotaAndAddReadRecord(
		&sqldb.ReadRecord{
			BucketID:        bucketInfo.Id.Uint64(),
			ObjectID:        objectInfo.Id.Uint64(),
			UserAddress:     req.GetUserAddress(),
			BucketName:      bucketInfo.GetBucketName(),
			ObjectName:      objectInfo.GetObjectName(),
			ReadSize:        readSize,
			ReadTimestampUs: sqldb.GetCurrentTimestampUs(),
		},
		&sqldb.BucketQuota{
			ReadQuotaSize: bucketInfo.GetChargedReadQuota() + model.DefaultSpFreeReadQuotaSize,
		},
	); err != nil {
		log.CtxErrorw(ctx, "failed to check billing due to bucket quota", "error", err)
		return merrors.InnerErrorToGRPCError(err)
	}
	pieceInfos, err := downloader.SplitToSegmentPieceInfos(objectInfo.Id.Uint64(), objectInfo.GetPayloadSize(), startOffset, endOffset)
	if err != nil {
		return
	}
	for _, pInfo := range pieceInfos {
		resp.Data, err = downloader.pieceStore.GetPiece(ctx, pInfo.segmentPieceKey, int64(pInfo.offset), int64(pInfo.length))
		if err != nil {
			return
		}
		if err = stream.Send(resp); err != nil {
			return
		}
		sendSize += len(resp.Data)
	}
	return
}

type segmentPieceInfo struct {
	segmentPieceKey string
	offset          uint64
	length          uint64
}

// SplitToSegmentPieceInfos compute the piece store info for get object, object data range [start, end].
func (downloader *Downloader) SplitToSegmentPieceInfos(objectID, objectSize, start, end uint64) (pieceInfos []*segmentPieceInfo, err error) {
	if objectSize == 0 || start >= objectSize || end >= objectSize || end < start {
		return pieceInfos, fmt.Errorf("failed to check param, object size: %d, start: %d, end: %d", objectSize, start, end)
	}
	params, err := downloader.spDB.GetStorageParams()
	if err != nil {
		return pieceInfos, err
	}

	segmentSize := params.GetMaxSegmentSize()
	segmentCount := objectSize / segmentSize
	if objectSize%segmentSize != 0 {
		segmentCount++
	}

	for segmentPieceIndex := uint64(0); segmentPieceIndex < segmentCount; segmentPieceIndex++ {
		currentStart := segmentPieceIndex * segmentSize
		currentEnd := (segmentPieceIndex+1)*segmentSize - 1
		if start > currentEnd {
			continue
		}
		if start > currentStart {
			currentStart = start
		}

		if end <= currentEnd {
			currentEnd = end
			offsetInPiece := currentStart - (segmentPieceIndex * segmentSize)
			lengthInPiece := currentEnd - currentStart + 1
			pieceInfos = append(pieceInfos, &segmentPieceInfo{
				segmentPieceKey: piecestore.EncodeSegmentPieceKey(objectID, uint32(segmentPieceIndex)),
				offset:          offsetInPiece,
				length:          lengthInPiece,
			})
			// break to finish
			break
		} else {
			offsetInPiece := currentStart - (segmentPieceIndex * segmentSize)
			lengthInPiece := currentEnd - currentStart + 1
			pieceInfos = append(pieceInfos, &segmentPieceInfo{
				segmentPieceKey: piecestore.EncodeSegmentPieceKey(objectID, uint32(segmentPieceIndex)),
				offset:          offsetInPiece,
				length:          lengthInPiece,
			})
		}
	}
	return
}

// GetBucketReadQuota get the quota info of the specified month.
func (downloader *Downloader) GetBucketReadQuota(ctx context.Context, req *types.GetBucketReadQuotaRequest) (*types.GetBucketReadQuotaResponse, error) {
	bucketTraffic, err := downloader.spDB.GetBucketTraffic(req.GetBucketInfo().Id.Uint64(), req.GetYearMonth())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &types.GetBucketReadQuotaResponse{
			ChargedQuotaSize: req.GetBucketInfo().GetChargedReadQuota(),
			SpFreeQuotaSize:  model.DefaultSpFreeReadQuotaSize,
			ConsumedSize:     0,
		}, nil
	}
	if err != nil {
		log.Errorw("failed to get bucket traffic", "error", err)
		return nil, err
	}
	return &types.GetBucketReadQuotaResponse{
		ChargedQuotaSize: req.GetBucketInfo().GetChargedReadQuota(),
		SpFreeQuotaSize:  model.DefaultSpFreeReadQuotaSize,
		ConsumedSize:     bucketTraffic.ReadConsumedSize,
	}, nil
}

// ListBucketReadRecord get read record list of the specified time range.
func (downloader *Downloader) ListBucketReadRecord(ctx context.Context, req *types.ListBucketReadRecordRequest) (*types.ListBucketReadRecordResponse, error) {
	records, err := downloader.spDB.GetBucketReadRecord(req.GetBucketInfo().Id.Uint64(), &sqldb.TrafficTimeRange{
		StartTimestampUs: req.StartTimestampUs,
		EndTimestampUs:   req.EndTimestampUs,
		LimitNum:         int(req.MaxRecordNum),
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &types.ListBucketReadRecordResponse{
			NextStartTimestampUs: 0,
		}, nil
	}
	if err != nil {
		log.Errorw("failed to list bucket read record", "error", err)
		return nil, err
	}
	var nextStartTimestampUs int64
	readRecords := make([]*types.ReadRecord, 0)
	for _, r := range records {
		readRecords = append(readRecords, &types.ReadRecord{
			ObjectName:     r.ObjectName,
			ObjectId:       r.ObjectID,
			AccountAddress: r.UserAddress,
			TimestampUs:    r.ReadTimestampUs,
			ReadSize:       r.ReadSize,
		})
		if r.ReadTimestampUs >= nextStartTimestampUs {
			nextStartTimestampUs = r.ReadTimestampUs + 1
		}
	}
	resp := &types.ListBucketReadRecordResponse{
		ReadRecords:          readRecords,
		NextStartTimestampUs: nextStartTimestampUs,
	}
	return resp, nil
}

// GetSpByAddress get sp by sp address
func (downloader *Downloader) GetSpByAddress(ctx context.Context, req *types.GetSpByAddressRequest) (resp *types.GetSpByAddressResponse, err error) {
	ctx = log.Context(ctx, req)

	sp, err := downloader.spDB.GetSpByAddress(req.BucketName, sqldb.OperatorAddressType)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get sp", "error", err)
		return
	}

	resp = &types.GetSpByAddressResponse{Endpoint: sp.Endpoint}
	log.CtxInfow(ctx, "succeed to get sp by a sp address")
	return resp, nil
}

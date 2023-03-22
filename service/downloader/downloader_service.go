package downloader

import (
	"context"
	"errors"
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	"gorm.io/gorm"
)

var _ types.DownloaderServiceServer = &Downloader{}

// GetObject downloads the payload of the object.
func (downloader *Downloader) GetObject(req *types.GetObjectRequest,
	stream types.DownloaderService_GetObjectServer) (err error) {
	var (
		size         int
		offset       uint64
		length       uint64
		isValidRange bool
	)

	ctx := log.Context(context.Background(), req)
	resp := &types.GetObjectResponse{}
	defer func() {
		if err != nil {
			return
		}
		resp.IsValidRange = isValidRange
		log.CtxInfow(ctx, "succeed to get object", "send_size", size)
	}()

	bucketInfo := req.GetBucketInfo()
	objectInfo := req.GetObjectInfo()
	if err = downloader.spDB.CheckQuotaAndAddReadRecord(
		// TODO: support range read
		&sqldb.ReadRecord{
			BucketID:        bucketInfo.Id.Uint64(),
			ObjectID:        objectInfo.Id.Uint64(),
			UserAddress:     req.GetUserAddress(),
			BucketName:      bucketInfo.GetBucketName(),
			ObjectName:      objectInfo.GetObjectName(),
			ReadSize:        objectInfo.GetPayloadSize(),
			ReadTimestampUs: sqldb.GetCurrentTimestampUs(),
		},
		&sqldb.BucketQuota{
			ReadQuotaSize: bucketInfo.GetReadQuota() + model.DefaultSpFreeReadQuotaSize,
		},
	); err != nil {
		log.Errorw("failed to check billing due to bucket quota", "error", err)
		return merrors.InnerErrorToGRPCError(err)
	}

	// TODO: It will be optimized
	// if length == 0, download all object data
	if req.RangeStart >= 0 && req.RangeStart < int64(objectInfo.GetPayloadSize()) &&
		req.RangeEnd >= 0 && req.RangeEnd < int64(objectInfo.GetPayloadSize()) {
		isValidRange = true
		offset = uint64(req.RangeStart)
		length = uint64(req.RangeEnd-req.RangeStart) + 1
	} else if req.RangeStart > 0 && req.RangeStart < int64(objectInfo.GetPayloadSize()) && req.RangeEnd < 0 {
		isValidRange = true
		offset = uint64(req.RangeStart)
		length = objectInfo.GetPayloadSize() - uint64(req.RangeStart)
	} else {
		offset, length = 0, objectInfo.GetPayloadSize()
	}
	var segmentInfo segments
	segmentInfo, err = downloader.DownloadPieceInfo(objectInfo.Id.Uint64(), objectInfo.GetPayloadSize(), offset, offset+length-1)
	if err != nil {
		return
	}
	for _, segment := range segmentInfo {
		resp.Data, err = downloader.pieceStore.GetSegment(ctx, segment.pieceKey, int64(segment.offset), int64(segment.offset)+int64(segment.length))
		if err != nil {
			return
		}
		resp.IsValidRange = isValidRange
		if err = stream.Send(resp); err != nil {
			return
		}
		size = size + len(resp.Data)
	}
	return nil
}

type segment struct {
	pieceKey string
	offset   uint64
	length   uint64
}

type segments []*segment

// DownloadPieceInfo compute the piece store info for download.
// download interval [start, end]
func (downloader *Downloader) DownloadPieceInfo(objectID, objectSize, start, end uint64) (pieceInfo segments, err error) {
	if objectSize == 0 || start > objectSize || end < start {
		return pieceInfo, fmt.Errorf("failed to check param, object size: %d, start: %d, end: %d", objectSize, start, end)
	}
	params, err := downloader.spDB.GetStorageParams()
	if err != nil {
		return pieceInfo, err
	}
	segmentSize := params.GetMaxSegmentSize()
	segmentCount := int(objectSize / segmentSize)
	if objectSize%segmentSize != 0 {
		segmentCount++
	}
	for idx := 0; idx < segmentCount; idx++ {
		finish := false
		currentStart := uint64(idx) * segmentSize
		currentEnd := uint64(idx+1)*segmentSize - 1
		if currentEnd >= end {
			currentEnd = end
			finish = true
		}
		if start >= currentStart && start <= currentEnd {
			pieceInfo = append(pieceInfo, &segment{
				pieceKey: piecestore.EncodeSegmentPieceKey(objectID, uint32(idx)),
				offset:   start - currentStart,
				length:   currentEnd - start + 1,
			})
			if finish {
				break
			}
		}
		if end >= currentStart && end <= currentEnd {
			pieceInfo = append(pieceInfo, &segment{
				pieceKey: piecestore.EncodeSegmentPieceKey(objectID, uint32(idx)),
				offset:   0,
				length:   end - currentStart + 1,
			})
			break
		}
		if start < currentStart && end > currentEnd {
			pieceInfo = append(pieceInfo, &segment{
				pieceKey: piecestore.EncodeSegmentPieceKey(objectID, uint32(idx)),
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
			QuotaSize:       req.GetBucketInfo().GetReadQuota(),
			SpFreeQuotaSize: model.DefaultSpFreeReadQuotaSize,
			ConsumedSize:    0,
		}, nil
	}
	if err != nil {
		log.Errorw("failed to get bucket traffic", "error", err)
		return nil, err
	}
	return &types.GetBucketReadQuotaResponse{
		QuotaSize:       req.GetBucketInfo().GetReadQuota(),
		SpFreeQuotaSize: model.DefaultSpFreeReadQuotaSize,
		ConsumedSize:    bucketTraffic.ReadConsumedSize,
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

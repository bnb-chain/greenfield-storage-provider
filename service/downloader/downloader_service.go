package downloader

import (
	"context"
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var _ types.DownloaderServiceServer = &Downloader{}

// DownloaderObject download the payload of the object.
func (downloader *Downloader) DownloaderObject(req *types.DownloaderObjectRequest,
	stream types.DownloaderService_DownloaderObjectServer) (err error) {
	var (
		objectInfo   *storagetypes.ObjectInfo
		size         int
		offset       uint64
		length       uint64
		isValidRange bool
	)
	ctx := log.Context(context.Background(), req)
	resp := &types.DownloaderObjectResponse{}
	defer func() {
		if err != nil {
			return
		}
		resp.IsValidRange = isValidRange
		log.CtxInfow(ctx, "finish to download object", "error", err, "sendSize", size)
	}()

	chainObjectInfo, err := downloader.chain.QueryObjectInfo(ctx, req.BucketName, req.ObjectName)
	if err != nil {
		log.Errorf("failed to query chain", "err", err)
		return
	}
	objectInfo = &storagetypes.ObjectInfo{
		Id:          chainObjectInfo.Id,
		PayloadSize: chainObjectInfo.PayloadSize,
	}

	// TODO: It will be optimized here after connecting with the chain
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
	//offset, length = req.GetOffset(), req.GetLength()
	//if req.GetLength() == 0 {
	//	offset, length = 0, objectInfo.Size
	//}
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
		return pieceInfo, fmt.Errorf("param error, object size: %d, start: %d, end: %d", objectSize, start, end)
	}
	param, err := downloader.spDB.GetAllParam()
	if err != nil {
		return pieceInfo, err
	}
	segmentSize := param.GetMaxSegmentSize()
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

package downloader

import (
	"context"
	"fmt"

	"github.com/bnb-chain/inscription-storage-provider/model"
	merrors "github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/model/piecestore"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

var _ service.DownloaderServiceServer = &Downloader{}

// DownloaderSegment download the segment data and return to client.
func (downloader *Downloader) DownloaderSegment(ctx context.Context, req *service.DownloaderServiceDownloaderSegmentRequest) (resp *service.DownloaderServiceDownloaderSegmentResponse, err error) {
	ctx = log.Context(ctx, req)
	resp = &service.DownloaderServiceDownloaderSegmentResponse{
		TraceId: req.TraceId,
	}
	defer func() {
		if err != nil {
			resp.ErrMessage.ErrCode = service.ErrCode_ERR_CODE_ERROR
			resp.ErrMessage.ErrMsg = err.Error()
			log.CtxErrorw(ctx, "download segment failed", "error", err, "object", req.ObjectId, "segment idx", req.SegmentIdx)
		}
		log.CtxInfow(ctx, "download segment success", "object", req.ObjectId, "segment idx", req.SegmentIdx)
	}()
	if req.GetObjectId() == 0 {
		err = merrors.ErrObjectIdZero
		return
	}
	pieceKey := piecestore.EncodeSegmentPieceKey(req.GetObjectId(), req.GetSegmentIdx())
	resp.Data, err = downloader.pieceStore.GetPiece(ctx, pieceKey, 0, -1)
	return resp, nil
}

// DownloaderObject download the object data and return to client.
func (downloader *Downloader) DownloaderObject(req *service.DownloaderServiceDownloaderObjectRequest, stream service.DownloaderService_DownloaderObjectServer) (err error) {
	var (
		objectInfo *types.ObjectInfo
		size       int
		offset     uint64
		length     uint64
	)
	ctx := log.Context(context.Background(), req)
	resp := &service.DownloaderServiceDownloaderObjectResponse{
		TraceId: req.TraceId,
	}
	defer func() {
		if err != nil {
			resp.ErrMessage = merrors.MakeErrMsgResponse(err)
			err = stream.Send(resp)
		}
		log.CtxInfow(ctx, "download object completed", "error", err, "objectSize", objectInfo.Size, "sendSize", size)
	}()

	objectInfo, err = downloader.mockChain.QueryObjectByName(req.BucketName + "/" + req.ObjectName)
	if err != nil {
		return
	}
	// if length == 0, download all object data
	offset, length = req.GetOffset(), req.GetLength()
	if req.GetLength() == 0 {
		offset, length = 0, objectInfo.Size
	}
	var segmentInfo segments
	segmentInfo, err = DownloadPieceInfo(objectInfo.ObjectId, objectInfo.Size, offset, offset+length-1)
	if err != nil {
		return
	}
	for _, segment := range segmentInfo {
		resp.Data, err = downloader.pieceStore.GetPiece(ctx, segment.pieceKey, int64(segment.offset), int64(segment.offset)+int64(segment.length))
		if err != nil {
			return
		}
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
func DownloadPieceInfo(objectID, objectSize, start, end uint64) (pieceInfo segments, err error) {
	if objectSize == 0 || start > objectSize || end < start {
		return pieceInfo, fmt.Errorf("param error, object size: %d, start: %d, end: %d", objectSize, start, end)
	}
	segmentCount := int(objectSize / model.SegmentSize)
	if objectSize%model.SegmentSize != 0 {
		segmentCount++
	}
	for idx := 0; idx < segmentCount; idx++ {
		finish := false
		currentStart := uint64(idx) * model.SegmentSize
		currentEnd := uint64(idx+1)*model.SegmentSize - 1
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

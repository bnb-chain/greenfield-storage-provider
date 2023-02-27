package uploader

import (
	"context"
	"errors"
	"sync"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

type contextKey string

// JobMeta is Job Context, got from stone hub.
type JobMeta struct {
	objectID      uint64
	toUploadedIDs map[uint32]bool
	txHash        []byte
	pieceJob      *stypes.PieceJob
	done          bool
}

// reportJobProgress report done piece index to stone hub.
func (uploader *Uploader) reportJobProgress(ctx context.Context, jm *JobMeta, uploadID uint32, checkSum []byte) error {
	var (
		req      *stypes.StoneHubServiceDonePrimaryPieceJobRequest
		pieceJob *stypes.PieceJob
	)
	traceID, _ := ctx.Value("traceID").(string)
	pieceJob = &stypes.PieceJob{
		ObjectId:       jm.pieceJob.ObjectId,
		PayloadSize:    jm.pieceJob.PayloadSize,
		RedundancyType: jm.pieceJob.RedundancyType,
	}
	copy(pieceJob.TargetIdx, jm.pieceJob.TargetIdx)
	pieceJob.StorageProviderSealInfo = &stypes.StorageProviderSealInfo{
		StorageProviderId: uploader.config.StorageProvider,
		PieceIdx:          uploadID,
		PieceChecksum:     [][]byte{checkSum},
	}
	req = &stypes.StoneHubServiceDonePrimaryPieceJobRequest{
		TraceId:  traceID,
		PieceJob: pieceJob,
	}
	if _, err := uploader.stoneHub.DonePrimaryPieceJob(ctx, req); err != nil {
		return err
	}
	return nil
}

// UploadPayload merge CreateObject, SetObjectCreateInfo and BeginUploadPayload, special for heavy client use.
func (uploader *Uploader) UploadPayload(stream stypes.UploaderService_UploadPayloadServer) (err error) {
	var (
		txChan    = make(chan []byte)
		pieceChan = make(chan *SegmentContext, 500)
		wg        sync.WaitGroup
		resp      stypes.UploaderServiceUploadPayloadResponse
		waitDone  = make(chan bool)
		errChan   = make(chan error)
		ctx       = context.Background()
		sr        *streamReader
		jm        *JobMeta
	)

	defer func(resp *stypes.UploaderServiceUploadPayloadResponse, err error) {
		if err != nil {
			resp.ErrMessage = merrors.MakeErrMsgResponse(err)
			log.CtxErrorw(ctx, "failed to upload payload", "err", err)
		}
		if jm != nil {
			resp.ObjectId = jm.objectID
		}
		err = stream.SendAndClose(resp)
		log.Infow("upload object payload", "response", resp, "error", err)
	}(&resp, err)

	// fetch job meta, concurrently write payload's segments and report progresses.
	go func() {
		txHash, ok := <-txChan
		if !ok {
			return
		}
		if jm, err = uploader.checkAndPrepareMeta(sr, txHash); err != nil {
			errChan <- err
			return
		}
		for piece := range pieceChan {
			go func(segPiece *SegmentContext) {
				defer wg.Done()

				if _, ok := jm.toUploadedIDs[segPiece.Index]; !ok {
					// has uploaded, and skip.
					return
				}

				pieceKey := piecestore.EncodeSegmentPieceKey(jm.objectID, segPiece.Index)
				if err := uploader.store.PutPiece(pieceKey, segPiece.PieceData); err != nil {
					errChan <- err
					return
				}
				checksum := hash.GenerateChecksum(segPiece.PieceData)
				ctx := context.WithValue(ctx, contextKey("traceID"), sr.traceID)
				if err := uploader.reportJobProgress(ctx, jm, segPiece.Index, checksum); err != nil {
					errChan <- err
					return
				}
			}(piece)
		}
	}()

	// stream read and split segments
	sr = newStreamReader(stream, txChan)
	err = sr.splitSegment(model.SegmentSize, pieceChan, &wg)
	if err != nil {
		return
	}

	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		log.Info("succeed to upload")
		return
	case err = <-errChan:
		log.Errorw("failed to upload", "err", err)
		return
	}
}

// checkAndPrepareMeta get meta from chain and stone hub.
func (uploader *Uploader) checkAndPrepareMeta(sr *streamReader, txHash []byte) (*JobMeta, error) {
	objectInfo := &ptypes.ObjectInfo{
		BucketName:     sr.bucket,
		ObjectName:     sr.object,
		Size_:          sr.size,
		PrimarySp:      &ptypes.StorageProviderInfo{SpId: uploader.config.StorageProvider},
		RedundancyType: sr.redundancyType,
	}
	chainObjectInfo, err := uploader.chain.QueryObjectInfo(context.Background(), objectInfo.BucketName, objectInfo.ObjectName)
	if err != nil {
		log.Errorw("failed to query chain",
			"bucketName", objectInfo.BucketName, "objectName", objectInfo.ObjectName, "error", err)
		return nil, err
	}
	objectInfo.ObjectId = chainObjectInfo.Id.Uint64()

	resp, err := uploader.stoneHub.BeginUploadPayloadV2(context.Background(), &stypes.StoneHubServiceBeginUploadPayloadV2Request{
		TraceId:    sr.traceID,
		ObjectInfo: objectInfo,
	})
	if err != nil {
		return nil, err
	}
	if resp.GetPieceJob() == nil {
		return nil, errors.New("stone dispatch piece job is nil")
	}
	// if resp.GetPrimaryDone() {
	// }
	jm := &JobMeta{
		objectID:      resp.GetPieceJob().GetObjectId(),
		pieceJob:      resp.GetPieceJob(),
		done:          resp.GetPrimaryDone(),
		txHash:        objectInfo.TxHash,
		toUploadedIDs: make(map[uint32]bool),
	}
	if targetIdx := resp.PieceJob.GetTargetIdx(); targetIdx != nil {
		for idx := range targetIdx {
			jm.toUploadedIDs[uint32(idx)] = true
		}
	}
	return jm, err
}

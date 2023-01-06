package uploader

import (
	"context"
	"crypto/sha256"
	"sync"

	"github.com/bnb-chain/inscription-storage-provider/model"
	"github.com/bnb-chain/inscription-storage-provider/model/piecestore"
	pbService "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// uploaderImpl is a grpc server handler implement.
type uploaderImpl struct {
	pbService.UnimplementedUploaderServiceServer
	uploader *Uploader
}

// CreateObject handle grpc CreateObject request, include steps:
// 1.co-sign object tx by signerService;
// 2.register object job meta by stoneHub;
// 3.await co-sign object tx to on-chain;
// 4.update object job meta(tx's block height) by stoneHub.
func (ui *uploaderImpl) CreateObject(ctx context.Context, req *pbService.UploaderServiceCreateObjectRequest) (resp *pbService.UploaderServiceCreateObjectResponse, err error) {
	var (
		txHash   []byte
		height   uint64
		objectID uint64
	)

	resp = &pbService.UploaderServiceCreateObjectResponse{}
	defer func(r *pbService.UploaderServiceCreateObjectResponse, err error) {
		if err != nil {
			r.ErrMessage.ErrCode = pbService.ErrCode_ERR_CODE_ERROR
			r.ErrMessage.ErrMsg = err.Error()
		}
		log.Infow("create object tx", "response", resp, "error", err)
	}(resp, err)

	if txHash, err = ui.uploader.signer.coSignTx(ctx, req); err != nil {
		return resp, err
	}
	if err = ui.uploader.stoneHub.registerJobMeta(ctx, req, txHash); err != nil {
		return resp, err
	}
	if height, objectID, err = ui.uploader.eventWaiter.waitChainEvent(txHash); err != nil {
		return resp, err
	}
	if err = ui.uploader.stoneHub.updateJobMeta(txHash, height, objectID); err != nil {
		return resp, err
	}
	resp.TxHash = txHash
	return resp, err
}

// UploadPayload handle grpc UploadPayload request, include steps:
// 1.stream split grpc payload to segment data, and forward to upload goroutine;
// 2.concurrently upload:
//
//	2.1 fetch upload job meta from stone hub;
//	2.2 upload segment to piece store, and report job progress.
func (ui *uploaderImpl) UploadPayload(stream pbService.UploaderService_UploadPayloadServer) (err error) {
	var (
		txChan    = make(chan []byte)
		pieceChan = make(chan *SegmentContext, 500)
		wg        sync.WaitGroup
		resp      pbService.UploaderServiceUploadPayloadResponse
		waitDone  = make(chan bool)
		errChan   = make(chan error)
	)
	defer func(resp *pbService.UploaderServiceUploadPayloadResponse, err error) {
		if err != nil {
			resp.ErrMessage.ErrCode = pbService.ErrCode_ERR_CODE_ERROR
			resp.ErrMessage.ErrMsg = err.Error()
		}
		err = stream.SendAndClose(resp)
		log.Infow("upload object payload", "response", resp, "error", err)
	}(&resp, err)

	// fetch job meta, concurrently write payload's segments and report progresses.
	go func() {
		var jm *JobMeta
		select {
		case txHash, ok := <-txChan:
			if !ok {
				return
			}
			if jm, err = ui.uploader.stoneHub.fetchJobMeta(txHash); err != nil {
				errChan <- err
				return
			}
		}
		for piece := range pieceChan {
			go func() {
				defer wg.Done()
				if _, ok := jm.toUploadedIDs[piece.Index]; !ok {
					// has uploaded, and skip.
					return
				}
				pieceKey := piecestore.EncodeSegmentPieceKey(jm.objectID, uint64(piece.Index))
				if err := ui.uploader.store.putPiece(pieceKey, piece.PieceData); err != nil {
					errChan <- err
					return
				}
				hash := sha256.New()
				hash.Write(piece.PieceData)
				checkSum := hash.Sum(nil)
				if err := ui.uploader.stoneHub.reportJobProgress(jm, piece.Index, checkSum); err != nil {
					errChan <- err
					return
				}
			}()
		}
	}()

	// stream read and split segments
	sr := newStreamReader(stream, txChan)
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
		log.Warnw("failed to upload", "err", err)
		return
	}
}

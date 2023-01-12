package uploader

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/bnb-chain/inscription-storage-provider/mock"
	"github.com/bnb-chain/inscription-storage-provider/model"
	merrors "github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/model/piecestore"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	pbService "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/store/metadb"
	"github.com/bnb-chain/inscription-storage-provider/util/hash"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

var (
	CreateObjectTimeout = time.Second * 5
)

// uploaderImpl is a grpc server handler implement.
type uploaderImpl struct {
	pbService.UnimplementedUploaderServiceServer
	uploader *Uploader
}

// CreateObject handle grpc CreateObject request, send create object tx to chain
func (ui *uploaderImpl) CreateObject(ctx context.Context, req *pbService.UploaderServiceCreateObjectRequest) (resp *pbService.UploaderServiceCreateObjectResponse, err error) {
	ctx = log.Context(ctx, req)
	resp = &pbService.UploaderServiceCreateObjectResponse{TraceId: req.GetTraceId()}
	defer func(r *pbService.UploaderServiceCreateObjectResponse, err error) {
		if err != nil {
			r.ErrMessage.ErrCode = pbService.ErrCode_ERR_CODE_ERROR
			r.ErrMessage.ErrMsg = err.Error()
			log.CtxErrorw(ctx, "create object failed", "error", err)
		}
		log.CtxInfow(ctx, "create object success")
	}(resp, err)

	// TODO:: 1. query object from inscription chain
	// 2. send create object tx to inscription chain
	txHash := ui.uploader.signer.BroadcastCreateObjectMessage(req.ObjectInfo)
	// 3.1 subscribe inscription chain create object event
	createObjectCh := ui.uploader.eventWaiter.SubscribeEvent(mock.CreateObject)
	// 3.2 register the object info to stone hub
	if _, err = ui.uploader.stoneHub.CreateObject(ctx, &pbService.StoneHubServiceCreateObjectRequest{
		TraceId:    req.TraceId,
		TxHash:     txHash,
		ObjectInfo: req.ObjectInfo,
	}); err != nil {
		return
	}
	// 3.3 wait create object tx to inscription chain
	createObjectTimer := time.After(CreateObjectTimeout)
	var objectInfo *types.ObjectInfo
	for {
		if objectInfo != nil {
			break
		}
		select {
		case event := <-createObjectCh:
			object := event.(*types.ObjectInfo)
			if bytes.Equal(object.TxHash, txHash) {
				objectInfo = object
			}
		case <-createObjectTimer:
			err = errors.New("create object to chain timeout")
			return
		}
	}
	if objectInfo == nil {
		err = errors.New("create object to chain failed")
		return
	}
	// 4. update object height and object id to stone hub
	if _, err = ui.uploader.stoneHub.SetObjectCreateInfo(ctx, &pbService.StoneHubServiceSetObjectCreateInfoRequest{
		TraceId:  req.TraceId,
		TxHash:   txHash,
		TxHeight: objectInfo.Height,
		ObjectId: objectInfo.ObjectId,
	}); err != nil {
		return
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
		ctx       = context.Background()
		sr        *streamReader
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
			ctx := context.WithValue(ctx, "traceID", sr.traceID)
			if jm, err = ui.fetchJobMeta(ctx, txHash); err != nil {
				errChan <- err
				return
			}
		}
		for piece := range pieceChan {
			go func(segPiece *SegmentContext) {
				defer wg.Done()
				if _, ok := jm.toUploadedIDs[segPiece.Index]; !ok {
					// has uploaded, and skip.
					return
				}
				pieceKey := piecestore.EncodeSegmentPieceKey(jm.objectID, segPiece.Index)
				if err := ui.uploader.store.PutPiece(pieceKey, segPiece.PieceData); err != nil {
					errChan <- err
					return
				}
				checksum := hash.GenerateChecksum(segPiece.PieceData)
				ctx := context.WithValue(ctx, "traceID", sr.traceID)
				if err := ui.reportJobProgress(ctx, jm, segPiece.Index, checksum); err != nil {
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
		log.Warnw("failed to upload", "err", err)
		return
	}
}

// JobMeta is Job Context, got from stone hub.
type JobMeta struct {
	objectID      uint64
	toUploadedIDs map[uint32]bool
	txHash        []byte
	pieceJob      *pbService.PieceJob
	done          bool
}

// fetchJobMeta fetch job meta from stone hub.
func (ui *uploaderImpl) fetchJobMeta(ctx context.Context, txHash []byte) (*JobMeta, error) {
	traceID, _ := ctx.Value("traceID").(string)
	resp, err := ui.uploader.stoneHub.BeginUploadPayload(ctx, &pbService.StoneHubServiceBeginUploadPayloadRequest{
		TraceId: traceID,
		TxHash:  txHash,
	})
	if err != nil {
		return nil, err
	}
	if resp.GetPieceJob() == nil {
		return nil, errors.New("stone dispatch piece job is nil")
	}
	if resp.GetPrimaryDone() {

	}
	jm := &JobMeta{
		objectID:      resp.GetPieceJob().GetObjectId(),
		pieceJob:      resp.GetPieceJob(),
		done:          resp.GetPrimaryDone(),
		txHash:        txHash,
		toUploadedIDs: make(map[uint32]bool),
	}
	if targetIdx := resp.PieceJob.GetTargetIdx(); targetIdx != nil {
		for idx := range targetIdx {
			jm.toUploadedIDs[uint32(idx)] = true
		}
	}
	return jm, err
}

// reportJobProgress report done piece index to stone hub.
func (ui *uploaderImpl) reportJobProgress(ctx context.Context, jm *JobMeta, uploadID uint32, checkSum []byte) error {
	var (
		req      *pbService.StoneHubServiceDonePrimaryPieceJobRequest
		pieceJob pbService.PieceJob
	)
	traceID, _ := ctx.Value("traceID").(string)
	pieceJob = *jm.pieceJob
	pieceJob.StorageProviderSealInfo = &pbService.StorageProviderSealInfo{
		StorageProviderId: ui.uploader.config.StorageProvider,
		PieceIdx:          uploadID,
		PieceChecksum:     [][]byte{checkSum},
	}
	req = &pbService.StoneHubServiceDonePrimaryPieceJobRequest{
		TraceId:  traceID,
		TxHash:   jm.txHash,
		PieceJob: &pieceJob,
	}
	if _, err := ui.uploader.stoneHub.DonePrimaryPieceJob(ctx, req); err != nil {
		return err
	}
	return nil
}

// GetAuthentication get auth info, currently PreSignature is mocked.
func (ui *uploaderImpl) GetAuthentication(ctx context.Context, req *pbService.UploaderServiceGetAuthenticationRequest) (resp *pbService.UploaderServiceGetAuthenticationResponse, err error) {
	ctx = log.Context(ctx, req)
	defer func() {
		if err != nil {
			resp.ErrMessage = merrors.MakeErrMsgResponse(err)
			log.CtxErrorw(ctx, "failed to get authentication", "err", err)
		} else {
			log.CtxInfow(ctx, "succeed to get authentication")
		}
	}()

	resp = &pbService.UploaderServiceGetAuthenticationResponse{TraceId: req.TraceId}
	meta := &metadb.UploadPayloadAskingMeta{
		BucketName: req.Bucket,
		ObjectName: req.Object,
		Timeout:    time.Now().Add(1 * time.Hour).Unix(),
	}
	if err = ui.uploader.metaDB.SetUploadPayloadAskingMeta(meta); err != nil {
		log.Errorw("failed to insert metaDB")
		return
	}
	log.CtxInfow(ctx, "insert to metadb", "bucket", req.Bucket, "object", req.Object)
	// mock
	resp.PreSignature = hash.GenerateChecksum([]byte(time.Now().String()))
	return resp, nil
}

// UploadPayloadV2 merge CreateObject, SetObjectCreateInfo and BeginUploadPayload, special for heavy client use.
func (ui *uploaderImpl) UploadPayloadV2(stream pbService.UploaderService_UploadPayloadV2Server) (err error) {
	var (
		txChan    = make(chan []byte)
		pieceChan = make(chan *SegmentContext, 500)
		wg        sync.WaitGroup
		resp      pbService.UploaderServiceUploadPayloadV2Response
		waitDone  = make(chan bool)
		errChan   = make(chan error)
		ctx       = context.Background()
		sr        *streamReader
	)
	defer func(resp *pbService.UploaderServiceUploadPayloadV2Response, err error) {
		if err != nil {
			resp.ErrMessage.ErrCode = pbService.ErrCode_ERR_CODE_ERROR
			resp.ErrMessage.ErrMsg = err.Error()
		}
		err = stream.SendAndClose(resp)
		log.Infow("upload object payload v2", "response", resp, "error", err)
	}(&resp, err)

	// fetch job meta, concurrently write payload's segments and report progresses.
	go func() {
		var jm *JobMeta
		select {
		case txHash, ok := <-txChan:
			if !ok {
				return
			}
			if jm, err = ui.checkAndPrepareMeta(sr, txHash); err != nil {
				errChan <- err
				return
			}
		}
		for piece := range pieceChan {
			go func(segPiece *SegmentContext) {
				defer wg.Done()

				if _, ok := jm.toUploadedIDs[segPiece.Index]; !ok {
					// has uploaded, and skip.
					return
				}

				pieceKey := piecestore.EncodeSegmentPieceKey(jm.objectID, segPiece.Index)
				if err := ui.uploader.store.PutPiece(pieceKey, segPiece.PieceData); err != nil {
					errChan <- err
					return
				}
				checksum := hash.GenerateChecksum(segPiece.PieceData)
				ctx := context.WithValue(ctx, "traceID", sr.traceID)
				if err := ui.reportJobProgress(ctx, jm, segPiece.Index, checksum); err != nil {
					errChan <- err
					return
				}
			}(piece)
		}
	}()

	// stream read and split segments
	sr = newStreamReaderV2(stream, txChan)
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

// checkAndPrepareMeta check auth by metaDB, and then get meta from stoneHub.
func (ui *uploaderImpl) checkAndPrepareMeta(sr *streamReader, txHash []byte) (*JobMeta, error) {
	objectID, height := ui.uploader.eventWaiter.GenerateObjectIDAndHeight()
	redundancyType := types.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED
	if sr.size <= model.InlineSize {
		redundancyType = types.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE
	}
	objectInfo := &types.ObjectInfo{
		Owner:          "",
		BucketName:     sr.bucket,
		ObjectName:     sr.object,
		Size:           sr.size,
		Checksum:       nil,
		IsPrivate:      false,
		ContentType:    "",
		PrimarySp:      &types.StorageProviderInfo{SpId: ui.uploader.config.StorageProvider},
		JobId:          0,
		Height:         height,
		TxHash:         txHash,
		ObjectId:       objectID,
		RedundancyType: redundancyType,
		SecondarySps:   nil,
	}

	meta, err := ui.uploader.metaDB.GetUploadPayloadAskingMeta(objectInfo.BucketName, objectInfo.ObjectName)
	if err != nil {
		log.Errorw("failed to query metaDB", "bucket", objectInfo.BucketName, "object", objectInfo.ObjectName, "error", err)
		return nil, err
	}
	if time.Now().Unix() > meta.Timeout {
		err = errors.New("auth info has timeout")
		return nil, err
	}
	resp, err := ui.uploader.stoneHub.BeginUploadPayloadV2(context.Background(), &pbService.StoneHubServiceBeginUploadPayloadV2Request{
		TraceId:    sr.traceID,
		ObjectInfo: objectInfo,
	})
	if err != nil {
		return nil, err
	}
	if resp.GetPieceJob() == nil {
		return nil, errors.New("stone dispatch piece job is nil")
	}
	if resp.GetPrimaryDone() {
	}
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

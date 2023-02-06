package uploader

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/mock"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

var (
	CreateObjectTimeout = time.Second * 5
)

// CreateObject handle grpc CreateObject request, send create object tx to chain
func (uploader *Uploader) CreateObject(ctx context.Context, req *stypes.UploaderServiceCreateObjectRequest) (
	resp *stypes.UploaderServiceCreateObjectResponse, err error) {
	ctx = log.Context(ctx, req, req.GetObjectInfo())
	resp = &stypes.UploaderServiceCreateObjectResponse{TraceId: req.GetTraceId()}
	defer func(r *stypes.UploaderServiceCreateObjectResponse, err error) {
		if err != nil {
			r.ErrMessage.ErrCode = stypes.ErrCode_ERR_CODE_ERROR
			r.ErrMessage.ErrMsg = err.Error()
			log.CtxErrorw(ctx, "create object failed", "error", err)
		}
		log.CtxInfow(ctx, "create object success")
	}(resp, err)

	// TODO:: 1. query object from inscription chain
	// 2. send create object tx to inscription chain
	txHash := uploader.signer.BroadcastCreateObjectMessage(req.ObjectInfo)
	// 3.1 subscribe inscription chain create object event
	createObjectCh := uploader.eventWaiter.SubscribeEvent(mock.CreateObject)
	// 3.2 register the object info to stone hub
	if _, err = uploader.stoneHub.CreateObject(ctx, &stypes.StoneHubServiceCreateObjectRequest{
		TraceId:    req.TraceId,
		TxHash:     txHash,
		ObjectInfo: req.ObjectInfo,
	}); err != nil {
		return
	}
	// 3.3 wait create object tx to inscription chain
	createObjectTimer := time.After(CreateObjectTimeout)
	var objectInfo *ptypes.ObjectInfo
	for {
		if objectInfo != nil {
			break
		}
		select {
		case event := <-createObjectCh:
			object := event.(*ptypes.ObjectInfo)
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
	if _, err = uploader.stoneHub.SetObjectCreateInfo(ctx, &stypes.StoneHubServiceSetObjectCreateInfoRequest{
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
	)
	defer func(resp *stypes.UploaderServiceUploadPayloadResponse, err error) {
		if err != nil {
			resp.ErrMessage.ErrCode = stypes.ErrCode_ERR_CODE_ERROR
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
			if jm, err = uploader.fetchJobMeta(ctx, txHash); err != nil {
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
				if err := uploader.store.PutPiece(pieceKey, segPiece.PieceData); err != nil {
					errChan <- err
					return
				}
				checksum := hash.GenerateChecksum(segPiece.PieceData)
				ctx := context.WithValue(ctx, "traceID", sr.traceID)
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
		log.Warnw("failed to upload", "err", err)
		return
	}
}

// JobMeta is Job Context, got from stone hub.
type JobMeta struct {
	objectID      uint64
	toUploadedIDs map[uint32]bool
	txHash        []byte
	pieceJob      *stypes.PieceJob
	done          bool
}

// fetchJobMeta fetch job meta from stone hub.
func (uploader *Uploader) fetchJobMeta(ctx context.Context, txHash []byte) (*JobMeta, error) {
	traceID, _ := ctx.Value("traceID").(string)
	resp, err := uploader.stoneHub.BeginUploadPayload(ctx, &stypes.StoneHubServiceBeginUploadPayloadRequest{
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
func (uploader *Uploader) reportJobProgress(ctx context.Context, jm *JobMeta, uploadID uint32, checkSum []byte) error {
	var (
		req      *stypes.StoneHubServiceDonePrimaryPieceJobRequest
		pieceJob stypes.PieceJob
	)
	traceID, _ := ctx.Value("traceID").(string)
	pieceJob = *jm.pieceJob
	pieceJob.StorageProviderSealInfo = &stypes.StorageProviderSealInfo{
		StorageProviderId: uploader.config.StorageProvider,
		PieceIdx:          uploadID,
		PieceChecksum:     [][]byte{checkSum},
	}
	req = &stypes.StoneHubServiceDonePrimaryPieceJobRequest{
		TraceId:  traceID,
		PieceJob: &pieceJob,
	}
	if _, err := uploader.stoneHub.DonePrimaryPieceJob(ctx, req); err != nil {
		return err
	}
	return nil
}

// GetAuthentication get auth info, currently PreSignature is mocked.
func (uploader *Uploader) GetAuthentication(ctx context.Context, req *stypes.UploaderServiceGetAuthenticationRequest) (
	resp *stypes.UploaderServiceGetAuthenticationResponse, err error) {
	ctx = log.Context(ctx, req)
	defer func() {
		if err != nil {
			resp.ErrMessage = merrors.MakeErrMsgResponse(err)
			log.CtxErrorw(ctx, "failed to get authentication", "err", err)
		} else {
			log.CtxInfow(ctx, "succeed to get authentication")
		}
	}()

	resp = &stypes.UploaderServiceGetAuthenticationResponse{TraceId: req.TraceId}
	meta := &metadb.UploadPayloadAskingMeta{
		BucketName: req.Bucket,
		ObjectName: req.Object,
		Timeout:    time.Now().Add(1 * time.Hour).Unix(),
	}
	if err = uploader.metaDB.SetUploadPayloadAskingMeta(meta); err != nil {
		log.Errorw("failed to insert metaDB")
		return
	}
	log.CtxInfow(ctx, "insert authentication info to metadb", "bucket", req.Bucket, "object", req.Object)
	// mock
	resp.PreSignature = hash.GenerateChecksum([]byte(time.Now().String()))
	return resp, nil
}

// UploadPayloadV2 merge CreateObject, SetObjectCreateInfo and BeginUploadPayload, special for heavy client use.
func (uploader *Uploader) UploadPayloadV2(stream stypes.UploaderService_UploadPayloadV2Server) (err error) {
	var (
		txChan    = make(chan []byte)
		pieceChan = make(chan *SegmentContext, 500)
		wg        sync.WaitGroup
		resp      stypes.UploaderServiceUploadPayloadV2Response
		waitDone  = make(chan bool)
		errChan   = make(chan error)
		ctx       = context.Background()
		sr        *streamReader
	)
	defer func(resp *stypes.UploaderServiceUploadPayloadV2Response, err error) {
		if err != nil {
			resp.ErrMessage.ErrCode = stypes.ErrCode_ERR_CODE_ERROR
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
			if jm, err = uploader.checkAndPrepareMeta(sr, txHash); err != nil {
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
				if err := uploader.store.PutPiece(pieceKey, segPiece.PieceData); err != nil {
					errChan <- err
					return
				}
				checksum := hash.GenerateChecksum(segPiece.PieceData)
				ctx := context.WithValue(ctx, "traceID", sr.traceID)
				if err := uploader.reportJobProgress(ctx, jm, segPiece.Index, checksum); err != nil {
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
func (uploader *Uploader) checkAndPrepareMeta(sr *streamReader, txHash []byte) (*JobMeta, error) {
	objectInfo := &ptypes.ObjectInfo{
		BucketName:     sr.bucket,
		ObjectName:     sr.object,
		Size:           sr.size,
		PrimarySp:      &ptypes.StorageProviderInfo{SpId: uploader.config.StorageProvider},
		RedundancyType: sr.redundancyType,
	}
	uploader.eventWaiter.CreateObjectByName(txHash, objectInfo)

	meta, err := uploader.metaDB.GetUploadPayloadAskingMeta(objectInfo.BucketName, objectInfo.ObjectName)
	if err != nil {
		log.Errorw("failed to query metaDB", "bucket", objectInfo.BucketName, "object", objectInfo.ObjectName, "error", err)
		return nil, err
	}
	if time.Now().Unix() > meta.Timeout {
		err = errors.New("auth info has timeout")
		return nil, err
	}
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

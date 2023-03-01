package uploader

import (
	"context"
	"io"
	"math"

	"github.com/bnb-chain/greenfield-common/go/hash"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	payloadstream "github.com/bnb-chain/greenfield-storage-provider/pkg/stream"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	types "github.com/bnb-chain/greenfield-storage-provider/service/uploader/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
)

var _ types.UploaderServiceServer = &Uploader{}

// UploadObject upload an object payload data with object info.
func (uploader *Uploader) UploadObject(
	stream types.UploaderService_UploadObjectServer) (err error) {
	var (
		resp          types.UploadObjectResponse
		pstream       = payloadstream.NewAsyncPayloadStream()
		traceInfo     = &servicetypes.SegmentInfo{}
		checkSum      [][]byte
		integrityMeta = &sqldb.IntegrityMeta{}
		errCh         = make(chan error, 10)
	)
	defer func(resp *types.UploadObjectResponse, err error) {
		if err != nil {
			log.Errorw("failed to replicate payload", "err", err)
			uploader.spDB.UpdateJobState(servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_ERROR,
				traceInfo.GetObjectInfo().Id.Uint64())
			return
		}
		integrityHash, signature, err := uploader.signer.SignIntegrityHash(context.Background(), checkSum)
		if err != nil {
			log.Errorw("failed to sign integrity hash", "err", err)
			uploader.spDB.UpdateJobState(servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_ERROR,
				traceInfo.GetObjectInfo().Id.Uint64())
			return
		}
		integrityMeta.Checksum = checkSum
		integrityMeta.IntegrityHash = integrityHash
		integrityMeta.Signature = signature
		err = uploader.spDB.SetObjectIntegrity(integrityMeta)
		if err != nil {
			log.Errorw("failed to write integrity hash to db", "error", err)
			uploader.spDB.UpdateJobState(servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_ERROR,
				traceInfo.GetObjectInfo().Id.Uint64())
			return
		}
		traceInfo.IntegrityHash = integrityHash
		traceInfo.Signature = signature
		uploader.cache.Add(traceInfo.ObjectInfo.Id.Uint64(), traceInfo)
		err = stream.SendAndClose(resp)
		pstream.Close()
		uploader.spDB.UpdateJobState(servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_DONE,
			traceInfo.GetObjectInfo().Id.Uint64())
		log.Infow("finish to upload payload", "error", err)
	}(&resp, err)
	params, err := uploader.spDB.GetStorageParams()
	if err != nil {
		return
	}
	segmentSize := params.GetMaxSegmentSize()

	// read payload from gRPC stream
	go func() {
		init := true
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				pstream.StreamClose()
				return
			}
			if err != nil {
				log.Debugw("receive payload exception", "error", err)
				pstream.StreamCloseWithError(err)
				errCh <- err
				return
			}
			if init {
				pstream.InitAsyncPayloadStream(
					req.GetObjectInfo().Id.Uint64(),
					math.MaxUint32,
					segmentSize,
					storagetypes.REDUNDANCY_REPLICA_TYPE)
				integrityMeta.ObjectID = req.GetObjectInfo().Id.Uint64()
				traceInfo.ObjectInfo = req.GetObjectInfo()
				uploader.cache.Add(req.GetObjectInfo().Id.Uint64(), traceInfo)
				uploader.spDB.CreateUploadJob(req.GetObjectInfo())
				init = false
			}
			pstream.StreamWrite(req.GetPayload())
		}
	}()

	// read payload from stream, the payload is spilt to segment size
	for {
		uploader.spDB.UpdateJobState(servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_DOING,
			traceInfo.GetObjectInfo().Id.Uint64())
		select {
		case entry := <-pstream.AsyncStreamRead():
			log.Debugw("read segment from stream", "segment_key", entry.Key(), "error", entry.Error())
			if entry.Error() == io.EOF {
				errCh <- nil
				return
			}
			if entry.Error() != nil {
				errCh <- entry.Error()
				return
			}
			checkSum = append(checkSum, hash.GenerateChecksum(entry.Data()))
			traceInfo.Checksum = checkSum
			traceInfo.Completed++
			uploader.cache.Add(entry.ID(), traceInfo)
			go func() {
				if err := uploader.pieceStore.PutSegment(entry.Key(), entry.Data()); err != nil {
					errCh <- err
				}
			}()
		case err = <-errCh:
			return
		}
	}
}

// QueryUploadingObject query an uploading object with object id from cache
func (uploader *Uploader) QueryUploadingObject(
	ctx context.Context,
	req *types.QueryUploadingObjectRequest) (
	resp *types.QueryUploadingObjectResponse, err error) {
	ctx = log.Context(ctx, req)
	objectId := req.GetObjectId()
	val, ok := uploader.cache.Get(objectId)
	if !ok {
		err = merrors.ErrCacheMiss
		return
	}
	resp.SegmentInfo = val.(*servicetypes.SegmentInfo)
	return
}

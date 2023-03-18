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
	"github.com/bnb-chain/greenfield-storage-provider/service/uploader/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
)

var _ types.UploaderServiceServer = &Uploader{}

// PutObject upload an object payload data with object info.
func (uploader *Uploader) PutObject(stream types.UploaderService_PutObjectServer) (err error) {
	var (
		resp          types.PutObjectResponse
		pstream       = payloadstream.NewAsyncPayloadStream()
		traceInfo     = &servicetypes.SegmentInfo{}
		checksum      [][]byte
		integrityMeta = &sqldb.IntegrityMeta{}
		errCh         = make(chan error, 10)
	)
	defer func(resp *types.PutObjectResponse, err error) {
		if err != nil {
			log.Errorw("failed to replicate payload", "err", err)
			uploader.spDB.UpdateJobState(traceInfo.GetObjectInfo().Id.Uint64(),
				servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_ERROR)
			return
		}
		integrityHash, signature, err := uploader.signer.SignIntegrityHash(context.Background(),
			integrityMeta.ObjectID, checksum)
		if err != nil {
			log.Errorw("failed to sign integrity hash", "err", err)
			uploader.spDB.UpdateJobState(traceInfo.GetObjectInfo().Id.Uint64(),
				servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_ERROR)
			return
		}
		integrityMeta.Checksum = checksum
		integrityMeta.IntegrityHash = integrityHash
		integrityMeta.Signature = signature
		err = uploader.spDB.SetObjectIntegrity(integrityMeta)
		if err != nil {
			log.Errorw("failed to write integrity hash to db", "error", err)
			uploader.spDB.UpdateJobState(traceInfo.GetObjectInfo().Id.Uint64(),
				servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_ERROR)
			return
		}
		traceInfo.IntegrityHash = integrityHash
		traceInfo.Signature = signature
		uploader.cache.Add(traceInfo.ObjectInfo.Id.Uint64(), traceInfo)
		err = uploader.taskNode.ReplicateObject(context.Background(), traceInfo.GetObjectInfo())
		if err != nil {
			log.Errorw("failed to notify task node to replicate object", "error", err)
			uploader.spDB.UpdateJobState(traceInfo.GetObjectInfo().Id.Uint64(),
				servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_ERROR)
			return
		}
		err = stream.SendAndClose(resp)
		pstream.Close()
		uploader.spDB.UpdateJobState(traceInfo.GetObjectInfo().Id.Uint64(),
			servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_DONE)
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
				uploader.spDB.UpdateJobState(traceInfo.GetObjectInfo().Id.Uint64(),
					servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_DOING)
				init = false
			}
			pstream.StreamWrite(req.GetPayload())
		}
	}()

	// read payload from stream, the payload is spilt to segment size
	for {
		select {
		case entry := <-pstream.AsyncStreamRead():
			log.Debugw("read segment from stream", "segment_key", entry.Key(), "error", entry.Error())
			if entry.Error() == io.EOF {
				err = nil
				return
			}
			if entry.Error() != nil {
				err = entry.Error()
				return
			}
			checksum = append(checksum, hash.GenerateChecksum(entry.Data()))
			traceInfo.Checksum = checksum
			traceInfo.Completed++
			uploader.cache.Add(entry.ID(), traceInfo)
			if err = uploader.pieceStore.PutSegment(entry.Key(), entry.Data()); err != nil {
				return
			}
		case err = <-errCh:
			return
		}
	}
}

// QueryPuttingObject query an uploading object with object id from cache
func (uploader *Uploader) QueryPuttingObject(ctx context.Context, req *types.QueryPuttingObjectRequest) (
	resp *types.QueryPuttingObjectResponse, err error) {
	ctx = log.Context(ctx, req)
	objectID := req.GetObjectId()
	log.CtxDebugw(ctx, "query putting object", "objectID", objectID)
	val, ok := uploader.cache.Get(objectID)
	if !ok {
		err = merrors.ErrCacheMiss
		return
	}
	resp.SegmentInfo = val.(*servicetypes.SegmentInfo)
	return
}

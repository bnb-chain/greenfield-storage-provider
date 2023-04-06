package uploader

import (
	"bytes"
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
		init        bool
		checksum    [][]byte
		objectID    uint64
		objectInfo  *storagetypes.ObjectInfo
		params      *storagetypes.Params
		req         *types.PutObjectRequest
		resp        = &types.PutObjectResponse{}
		pstream     = payloadstream.NewAsyncPayloadStream()
		ctx, cancel = context.WithCancel(context.Background())
		errCh       = make(chan error)
	)
	defer func() {
		defer cancel()
		if err != nil {
			if init {
				uploader.spDB.UpdateJobState(objectID, servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_ERROR)
			}
			log.CtxErrorw(ctx, "failed to replicate payload", "error", err)
			return
		}
		uploader.spDB.UpdateJobState(objectID, servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_DONE)
		err = uploader.signIntegrityHash(ctx, objectID, objectInfo.GetChecksums()[0], checksum)
		if err != nil {
			return
		}
		err = uploader.taskNode.ReplicateObject(ctx, objectInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to notify task node to replicate object", "error", err)
			uploader.spDB.UpdateJobState(objectID, servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_ERROR)
			return
		}
		err = stream.SendAndClose(resp)
		pstream.Close()
		log.CtxInfow(ctx, "succeed to put object", "error", err)
	}()
	params, err = uploader.spDB.GetStorageParams()
	if err != nil {
		return
	}
	segmentPieceSize := params.GetMaxSegmentSize()
	// read payload from gRPC stream
	go func() {
		for {
			req, err = stream.Recv()
			if err == io.EOF {
				pstream.StreamClose()
				return
			}
			if err != nil {
				log.CtxErrorw(ctx, "receive payload exception", "error", err)
				pstream.StreamCloseWithError(err)
				errCh <- err
				return
			}
			if !init {
				if req.GetObjectInfo() == nil {
					errCh <- merrors.ErrDanglingPointer
					return
				}
				objectInfo = req.GetObjectInfo()
				if int(params.GetRedundantDataChunkNum()+params.GetRedundantParityChunkNum()+1) !=
					len(objectInfo.GetChecksums()) {
					errCh <- merrors.ErrMismatchChecksumNum
				}
				objectID = req.GetObjectInfo().Id.Uint64()
				ctx = log.WithValue(ctx, "object_id", req.GetObjectInfo().Id.String())
				pstream.InitAsyncPayloadStream(
					objectID,
					storagetypes.REDUNDANCY_REPLICA_TYPE,
					segmentPieceSize,
					math.MaxUint32, /*useless*/
				)
				uploader.spDB.CreateUploadJob(objectInfo)
				uploader.spDB.UpdateJobState(objectID, servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_DOING)
				init = true
			}
			pstream.StreamWrite(req.GetPayload())
		}
	}()

	// read payload from stream, the payload is spilt to pieces by piece size
	for {
		select {
		case entry, ok := <-pstream.AsyncStreamRead():
			if !ok { // has finished
				return
			}
			log.CtxDebugw(ctx, "get piece entry from stream", "piece_key", entry.PieceKey(),
				"piece_len", len(entry.Data()), "error", entry.Error())
			if entry.Error() != nil {
				err = entry.Error()
				return
			}
			checksum = append(checksum, hash.GenerateChecksum(entry.Data()))
			if err = uploader.pieceStore.PutPiece(entry.PieceKey(), entry.Data()); err != nil {
				return
			}
		case err = <-errCh:
			return
		}
	}
}

func (uploader *Uploader) signIntegrityHash(ctx context.Context, objectID uint64, rootHash []byte, checksum [][]byte) (err error) {
	var (
		integrityMeta = &sqldb.IntegrityMeta{ObjectID: objectID, Checksum: checksum}
	)
	uploader.spDB.UpdateJobState(objectID, servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_DOING)
	defer func() {
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign the integrity hash", "error", err)
			uploader.spDB.UpdateJobState(objectID, servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_ERROR)
			return
		}
		uploader.spDB.UpdateJobState(objectID, servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_DONE)
		log.CtxInfow(ctx, "succeed to sign the integrity hash", "error", err)
	}()
	integrityMeta.IntegrityHash, integrityMeta.Signature, err = uploader.signer.SignIntegrityHash(ctx, objectID, checksum)
	if err != nil {
		return
	}
	if !bytes.Equal(rootHash, integrityMeta.IntegrityHash) {
		err = merrors.ErrMismatchIntegrityHash
		return
	}
	err = uploader.spDB.SetObjectIntegrity(integrityMeta)
	if err != nil {
		return
	}
	return
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

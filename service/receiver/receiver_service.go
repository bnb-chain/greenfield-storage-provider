package receiver

import (
	"bytes"
	"context"
	"encoding/hex"
	"io"

	"github.com/bnb-chain/greenfield-common/go/hash"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	payloadstream "github.com/bnb-chain/greenfield-storage-provider/pkg/stream"
	"github.com/bnb-chain/greenfield-storage-provider/service/receiver/types"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
)

var _ types.ReceiverServiceServer = &Receiver{}

// ReceiveObjectPiece an object payload to storage provider.
func (receiver *Receiver) ReceiveObjectPiece(stream types.ReceiverService_ReceiveObjectPieceServer) (err error) {
	var (
		req                   *types.ReceiveObjectPieceRequest
		expectedIntegrityHash []byte
		pstream               = payloadstream.NewAsyncPayloadStream()
		traceInfo             = &servicetypes.PieceInfo{}
		checksum              [][]byte
		integrityMeta         = &sqldb.IntegrityMeta{}
		errCh                 = make(chan error, 10)
		objectInfo            *storagetypes.ObjectInfo
	)

	defer func() {
		if err != nil {
			log.Errorw("failed to replicate payload", "error", err)
			return
		}
		resp := &types.ReceiveObjectPieceResponse{}
		if resp.IntegrityHash, resp.Signature, err = receiver.signer.SignIntegrityHash(context.Background(),
			integrityMeta.ObjectID, checksum); err != nil {
			log.Errorw("failed to sign integrity hash", "error", err)
			return
		}

		if !bytes.Equal(expectedIntegrityHash, resp.IntegrityHash) {
			err = merrors.ErrMismatchIntegrityHash
			log.Errorw("failed to check root hash",
				"expected", hex.EncodeToString(expectedIntegrityHash),
				"actual", hex.EncodeToString(resp.IntegrityHash), "error", err)
			return
		}
		integrityMeta.Checksum = checksum
		integrityMeta.IntegrityHash = resp.IntegrityHash
		integrityMeta.Signature = resp.Signature
		if err = receiver.spDB.SetObjectIntegrity(integrityMeta); err != nil {
			log.Errorw("failed to write integrity hash to db", "error", err)
			return
		}
		traceInfo.IntegrityHash = resp.IntegrityHash
		traceInfo.Signature = resp.Signature
		receiver.cache.Add(traceInfo.ObjectInfo.Id.Uint64(), traceInfo)
		if err = stream.SendAndClose(resp); err != nil {
			log.Errorw("failed to send and close stream", "error", err)
			return
		}
		pstream.Close()
		log.Infow("succeed to replicate payload", "response", resp)
	}()

	// TODO:: add flow control, syncing one object request cost 4 parallel goroutine at least

	// read payload from gRPC
	go func() {
		var isInited bool
		for {
			req, err = stream.Recv()
			if err == io.EOF {
				err = nil
				pstream.StreamClose()
				return
			}
			if err != nil {
				log.Debugw("receive payload exception", "error", err)
				pstream.StreamCloseWithError(err)
				errCh <- err
				return
			}
			if !isInited {
				if objectInfo = req.GetObjectInfo(); objectInfo == nil {
					err = merrors.ErrDanglingPointer
					errCh <- err
					return
				}
				if int(req.GetRedundancyIdx())+1 > len(objectInfo.GetChecksums()) {
					err = merrors.ErrInvalidParams
					errCh <- err
					return
				}
				expectedIntegrityHash = objectInfo.GetChecksums()[int(req.GetRedundancyIdx())+1]
				pstream.InitAsyncPayloadStream(
					objectInfo.Id.Uint64(),
					objectInfo.GetRedundancyType(),
					req.GetPieceSize(),
					uint32(req.GetRedundancyIdx()))
				integrityMeta.ObjectID = objectInfo.Id.Uint64()
				traceInfo.ObjectInfo = objectInfo
				receiver.cache.Add(objectInfo.Id.Uint64(), traceInfo)
				isInited = true
			}

			pstream.StreamWrite(req.GetPieceStreamData())
		}
	}()

	// read payload from stream, the payload is spilt to segment size
	for {
		select {
		case entry, ok := <-pstream.AsyncStreamRead():
			if !ok { // has finished
				return
			}
			log.Debugw("get piece entry from stream", "piece_key", entry.PieceKey(),
				"piece_len", len(entry.Data()), "error", entry.Error())
			if entry.Error() != nil {
				errCh <- entry.Error()
				return
			}
			checksum = append(checksum, hash.GenerateChecksum(entry.Data()))
			traceInfo.Checksum = checksum
			traceInfo.CompletedNum++
			receiver.cache.Add(entry.ObjectID(), traceInfo)
			go func() {
				if err = receiver.pieceStore.PutPiece(entry.PieceKey(), entry.Data()); err != nil {
					log.Errorw("receiver failed to put piece to piece store", "error", err)
					errCh <- err
				}
			}()
		case err = <-errCh:
			return
		}
	}
}

// QueryReceivingObject query a receiving object info by object id.
func (receiver *Receiver) QueryReceivingObject(ctx context.Context, req *types.QueryReceivingObjectRequest) (
	resp *types.QueryReceivingObjectResponse, err error) {
	ctx = log.Context(ctx, req)
	log.CtxDebug(ctx, "query receiving object")
	cached, ok := receiver.cache.Get(req.GetObjectId())
	if !ok {
		err = merrors.ErrCacheMiss
		return
	}
	resp = &types.QueryReceivingObjectResponse{}
	resp.PieceInfo = cached.(*servicetypes.PieceInfo)
	return
}

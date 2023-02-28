package syncer

import (
	"context"
	"io"

	"github.com/bnb-chain/greenfield-common/go/hash"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	payloadstream "github.com/bnb-chain/greenfield-storage-provider/pkg/stream"
	"github.com/bnb-chain/greenfield-storage-provider/service/syncer/types"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	"github.com/bnb-chain/greenfield-storage-provider/store"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

var _ types.SyncerServiceServer = &Syncer{}

// SyncObject an object payload to storage provider.
func (syncer *Syncer) SyncObject(stream types.SyncerService_SyncObjectServer) (err error) {
	var (
		resp      types.SyncObjectResponse
		pstream   = payloadstream.NewAsyncPayloadStream()
		traceInfo = &servicetypes.SegmentInfo{}

		checksum      [][]byte
		integrityMeta = &store.IntegrityMeta{}

		errCh = make(chan error, 10)
	)

	defer func(resp *types.SyncObjectResponse, err error) {
		if err != nil {
			log.Errorw("failed to replicate payload", "err", err)
			return
		}
		resp.IntegrityHash, resp.Signature, err = syncer.signer.SignIntegrityHash(context.Background(), checksum)
		if err != nil {
			log.Errorw("failed to sign integrity hash", "err", err)
			return
		}
		integrityMeta.Checksum = checksum
		integrityMeta.IntegrityHash = resp.IntegrityHash
		integrityMeta.Signature = resp.Signature
		err = syncer.spdb.SetObjectIntegrity(integrityMeta)
		if err != nil {
			log.Errorw("fail to write integrity hash to db", "error", err)
			return
		}
		traceInfo.IntegrityHash = resp.IntegrityHash
		traceInfo.Signature = resp.Signature
		syncer.cache.Add(traceInfo.ObjectInfo.Id.Uint64(), traceInfo)
		err = stream.SendAndClose(resp)
		pstream.Close()
		log.Infow("replicate payload", "response", resp, "error", err)
	}(&resp, err)

	//TODO:: add flow control, syncing one object request cost 4 parallel goroutine at least

	// read payload from gRPC
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
					req.GetReplicateIdx(),
					req.GetSegmentSize(),
					req.GetObjectInfo().GetRedundancyType())
				integrityMeta.ObjectId = req.GetObjectInfo().Id.Uint64()
				traceInfo.ObjectInfo = req.GetObjectInfo()
				syncer.cache.Add(req.GetObjectInfo().Id.Uint64(), traceInfo)
				init = false
			}

			pstream.StreamWrite(req.GetReplicateData())
		}
	}()

	// read payload from stream, the payload is spilt to segment size
	for {
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
			checksum = append(checksum, hash.GenerateChecksum(entry.Data()))
			traceInfo.CheckSum = checksum
			traceInfo.Completed++
			syncer.cache.Add(entry.ID(), traceInfo)
			go func() {
				if err := syncer.pieceStore.PutSegment(entry.Key(), entry.Data()); err != nil {
					errCh <- err
				}
				syncer.cache.Add(entry.ID(), traceInfo)
			}()
		case err = <-errCh:
			return
		}
	}
}

// QuerySyncingObject query a syncing object info by object id.
func (syncer *Syncer) QuerySyncingObject(
	ctx context.Context,
	req *types.QuerySyncingObjectRequest) (
	resp *types.QuerySyncingObjectResponse, err error) {
	ctx = log.Context(ctx, req)
	objectId := req.GetObjectId()
	cached, ok := syncer.cache.Get(objectId)
	if !ok {
		err = merrors.ErrCacheMiss
		return
	}
	resp = &types.QuerySyncingObjectResponse{}
	resp.SegmentInfo = cached.(*servicetypes.SegmentInfo)
	return
}

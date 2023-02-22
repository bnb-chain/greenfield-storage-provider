package syncer

import (
	"context"
	"io"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// SyncPiece syncs piece data to secondary storage provider
func (s *Syncer) SyncPiece(stream stypes.SyncerService_SyncPieceServer) error {
	var count uint32
	var integrityMeta *spdb.IntegrityMeta
	var spID string
	var traceID string
	var value []byte
	pieceHash := make([][]byte, 0)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			if count != integrityMeta.PieceCount {
				log.Errorw("syncer service received piece count is wrong")
				return merrors.ErrReceivedPieceCount
			}
			integrityMeta.PieceHash = pieceHash
			sealInfo, err := s.generateSealInfo(spID, integrityMeta)
			if err != nil {
				log.Errorw("syncer generate seal info failed", "error", err)
				return err
			}
			integrityMeta.IntegrityHash = sealInfo.GetIntegrityHash()
			integrityMeta.Signature = sealInfo.GetSignature()
			if err := s.metaDB.SetIntegrityMeta(integrityMeta); err != nil {
				log.Errorw("set integrity meta error", "error", err)
				return err
			}
			resp := &stypes.SyncerServiceSyncPieceResponse{
				TraceId:         traceID,
				SecondarySpInfo: sealInfo,
				ErrMessage: &stypes.ErrMessage{
					ErrCode: stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED,
					ErrMsg:  "success",
				},
			}
			ctx := log.Context(context.Background(), resp)
			log.CtxInfow(ctx, "receive piece data success", "integrity_hash", sealInfo.GetIntegrityHash())
			return stream.SendAndClose(resp)
		}
		if err != nil {
			log.Errorw("stream recv failed", "error", err)
			return err
		}
		count++
		spID = req.GetSyncerInfo().GetStorageProviderId()
		integrityMeta, value, err = s.handlePieceData(req, count)
		if err != nil {
			return err
		}
		traceID = req.GetTraceId()
		pieceHash = append(pieceHash, hash.GenerateChecksum(value))
	}
}

func (s *Syncer) generateSealInfo(spID string, integrityMeta *spdb.IntegrityMeta) (*stypes.StorageProviderSealInfo, error) {
	var err error
	resp := &stypes.StorageProviderSealInfo{
		StorageProviderId: spID,
		PieceIdx:          integrityMeta.EcIdx,
		PieceChecksum:     integrityMeta.PieceHash,
	}
	resp.IntegrityHash, resp.Signature, err = s.signer.SignIntegrityHash(context.Background(), resp.PieceChecksum)
	if err != nil {
		log.Warnw("failed to sign integrity hash", "error", err)
		return nil, err
	}
	return resp, nil
}

func (s *Syncer) handlePieceData(req *stypes.SyncerServiceSyncPieceRequest, count uint32) (*spdb.IntegrityMeta, []byte, error) {
	redundancyType := req.GetSyncerInfo().GetRedundancyType()
	objectID := req.GetSyncerInfo().GetObjectId()
	integrityMeta := &spdb.IntegrityMeta{
		ObjectID:       objectID,
		PieceCount:     req.GetSyncerInfo().GetPieceCount(),
		IsPrimary:      false,
		RedundancyType: redundancyType,
	}
	key, pieceIndex, err := encodePieceKey(redundancyType, objectID, count, req.GetSyncerInfo().GetPieceIndex())
	if err != nil {
		return nil, nil, err
	}
	integrityMeta.EcIdx = pieceIndex

	// put piece data into piece store
	value := req.GetPieceData()
	if err = s.store.PutPiece(key, value); err != nil {
		log.Errorw("put piece failed", "error", err)
		return nil, nil, err
	}
	return integrityMeta, value, nil
}

func encodePieceKey(redundancyType ptypes.RedundancyType, objectID uint64, segmentIndex, pieceIndex uint32) (
	string, uint32, error) {
	switch redundancyType {
	case ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE, ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
		return piecestore.EncodeSegmentPieceKey(objectID, segmentIndex), pieceIndex, nil
	case ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED:
		return piecestore.EncodeECPieceKey(objectID, segmentIndex, pieceIndex), pieceIndex, nil
	default:
		return "", 0, merrors.ErrRedundancyType
	}
}

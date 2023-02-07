package syncer

import (
	"context"
	"io"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// SyncPiece syncs piece data to secondary storage provider
func (s *Syncer) SyncPiece(stream stypes.SyncerService_SyncPieceServer) error {
	var count uint32
	var integrityMeta *metadb.IntegrityMeta
	var spID string
	var value []byte
	pieceHash := make([][]byte, 0)
	//defer func() {
	//	if err != nil && err != io.EOF {
	//		log.Info("entry defer func")
	//		err = stream.SendAndClose(&service.SyncerServiceSyncPieceResponse{
	//			ErrMessage: &service.ErrMessage{
	//				ErrCode: service.ErrCode_ERR_CODE_ERROR,
	//				ErrMsg:  err.Error(),
	//			},
	//		})
	//	}
	//}()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			if count != integrityMeta.PieceCount {
				log.Errorw("syncer service received piece count is wrong")
				return merrors.ErrReceivedPieceCount
			}

			integrityMeta.PieceHash = pieceHash
			sealInfo := generateSealInfo(spID, integrityMeta)
			integrityMeta.IntegrityHash = sealInfo.GetIntegrityHash()
			if err := s.setIntegrityMeta(s.metaDB, integrityMeta); err != nil {
				return err
			}
			resp := &stypes.SyncerServiceSyncPieceResponse{
				TraceId:         req.GetTraceId(),
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
		log.Infow("integrity meta", "piece count", integrityMeta.PieceCount, "count", count)
		pieceHash = append(pieceHash, hash.GenerateChecksum(value))
	}
}

func (s *Syncer) setIntegrityMeta(db metadb.MetaDB, meta *metadb.IntegrityMeta) error {
	if err := db.SetIntegrityMeta(meta); err != nil {
		log.Errorw("set integrity meta error", "error", err)
		return err
	}
	return nil
}

func generateSealInfo(spID string, integrityMeta *metadb.IntegrityMeta) *stypes.StorageProviderSealInfo {
	pieceHash := integrityMeta.PieceHash
	integrityHash := hash.GenerateIntegrityHash(pieceHash)
	resp := &stypes.StorageProviderSealInfo{
		StorageProviderId: spID,
		PieceIdx:          integrityMeta.PieceIdx,
		PieceChecksum:     pieceHash,
		IntegrityHash:     integrityHash,
		Signature:         nil, // TODO(mock)
	}
	return resp
}

func (s *Syncer) handlePieceData(req *stypes.SyncerServiceSyncPieceRequest, count uint32) (*metadb.IntegrityMeta, []byte, error) {
	redundancyType := req.GetSyncerInfo().GetRedundancyType()
	objectID := req.GetSyncerInfo().GetObjectId()
	integrityMeta := &metadb.IntegrityMeta{
		ObjectID:       objectID,
		PieceCount:     req.GetSyncerInfo().GetPieceCount(),
		IsPrimary:      false,
		RedundancyType: redundancyType,
	}
	key, pieceIndex, err := encodePieceKey(redundancyType, objectID, count, req.GetSyncerInfo().GetPieceIndex())
	if err != nil {
		return nil, nil, err
	}
	// put piece data into piece store
	value := req.GetPieceData()
	if err = s.store.PutPiece(key, value); err != nil {
		log.Errorw("put piece failed", "error", err)
		return nil, nil, err
	}
	integrityMeta.PieceIdx = pieceIndex
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

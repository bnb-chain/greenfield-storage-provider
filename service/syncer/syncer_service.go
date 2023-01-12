package syncer

import (
	"io"

	merrors "github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/model/piecestore"
	ptypes "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util"
	"github.com/bnb-chain/inscription-storage-provider/util/hash"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// UploadECPiece uploads piece data encoded using the ec algorithm to secondary storage provider
func (s *Syncer) UploadECPiece(stream service.SyncerService_UploadECPieceServer) error {
	var sealInfo *service.StorageProviderSealInfo
	var count uint32
	//var pieceIndex uint32
	var pieceCount uint32
	var err error

	defer func() {
		if err != nil && err != io.EOF {
			log.Info("entry defer func")
			err = stream.SendAndClose(&service.SyncerServiceUploadECPieceResponse{
				ErrMessage: &service.ErrMessage{
					ErrCode: service.ErrCode_ERR_CODE_ERROR,
					ErrMsg:  err.Error(),
				},
			})
		}
	}()

	for {
		req, err := stream.Recv()
		log.Infow("first", "object_id", req.GetSyncerInfo().GetObjectId(), "tx_hash", req.GetSyncerInfo().GetTxHash(),
			"storage_provider_id", req.GetSyncerInfo().GetStorageProviderId(), "rType", req.GetSyncerInfo().GetRedundancyType(), "traceID", req.GetTraceId())
		if err == io.EOF {
			log.Infow("upload ec piece closed", "error", err, "storage_provider_id", sealInfo.GetStorageProviderId(),
				"piece_idx", sealInfo.GetPieceIdx(), "count", count)
			if count != pieceCount {
				log.Errorw("syncer service received piece count is wrong")
				return merrors.ErrReceivedPieceCount
			}
			checksumList := sealInfo.GetPieceChecksum()
			integrityHash := hash.GenerateIntegrityHash(checksumList)
			sealInfo.IntegrityHash = integrityHash
			return stream.SendAndClose(&service.SyncerServiceUploadECPieceResponse{
				TraceId:         req.GetTraceId(),
				SecondarySpInfo: sealInfo,
				ErrMessage: &service.ErrMessage{
					ErrCode: service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED,
					ErrMsg:  "success",
				},
			})
		}
		if err != nil {
			log.Errorw("stream recv failed", "error", err)
			return err
		}
		pieceCount = req.GetSyncerInfo().GetPieceCount()
		sealInfo, _, err = s.handleUploadPiece(req)
		if err != nil {
			log.Errorw("handle upload piece error", "error", err)
			return err
		}
		count++
	}
}

// handleUploadPiece store piece data to piece store and compute integrity hash.
func (s *Syncer) handleUploadPiece(req *service.SyncerServiceUploadECPieceRequest) (
	*service.StorageProviderSealInfo, uint32, error) {
	var (
		pieceIndex uint32
		err        error
	)
	log.Infow("second", "req", req.GetSyncerInfo(), "rType", req.GetSyncerInfo().GetRedundancyType(),
		"traceID", req.GetTraceId())
	pieceChecksumList := make([][]byte, 0)
	keys := util.GenericSortedKeys(req.GetPieceData())
	log.Info("SortedKeys", "keys", keys)
	for _, key := range keys {
		// if redundancyType is ec, check all pieceIndex is equal
		pieceIndex, err = parsePieceIndex(req.GetSyncerInfo().GetRedundancyType(), key)
		if err != nil {
			return nil, 0, err
		}
		value := req.GetPieceData()[key]
		checksum := hash.GenerateChecksum(value)
		pieceChecksumList = append(pieceChecksumList, checksum)
		if err = s.store.PutPiece(key, value); err != nil {
			log.Errorw("put piece failed", "error", err)
			return nil, 0, err
		}
	}

	spID := req.GetSyncerInfo().GetStorageProviderId()
	//integrityHash := hash.GenerateIntegrityHash(pieceChecksumList)
	log.Infow("handleUploadPiece", "spID", spID, "pieceIndex", pieceIndex)
	resp := &service.StorageProviderSealInfo{
		StorageProviderId: spID,
		PieceIdx:          pieceIndex,
		PieceChecksum:     pieceChecksumList,
		Signature:         nil, // TODO(mock)
	}
	return resp, pieceIndex, nil
}

func parsePieceIndex(redundancyType ptypes.RedundancyType, key string) (uint32, error) {
	var (
		err        error
		pieceIndex uint32
	)
	log.Infow("parsePieceIndex", "rType", redundancyType, "key", key)
	switch redundancyType {
	case ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE, ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
		_, pieceIndex, err = piecestore.DecodeSegmentPieceKey(key)
	default: // ec type
		_, _, pieceIndex, err = piecestore.DecodeECPieceKey(key)
	}
	if err != nil {
		log.Errorw("decode piece key failed", "error", err)
		return 0, err
	}
	return pieceIndex, nil
}

//func writeIntegrityMetaToMetaDb(syncerInfo *service.SyncerInfo, pieceIndex uint32, sealInfo *service.StorageProviderSealInfo) error {
//	type s struct {
//		m metadb.MetaDB
//	}
//	integritaMeta := &metadb.IntegrityMeta{
//		ObjectID:       syncerInfo.GetObjectId(),
//		PieceIdx:       pieceIndex,
//		PieceCount:     syncerInfo.GetPieceCount(),
//		IsPrimary:      false,
//		RedundancyType: syncerInfo.GetRedundancyType(),
//		IntegrityHash:  sealInfo.GetIntegrityHash(),
//		//PieceHash:      sealInfo.GetPieceChecksum(),
//	}
//	a := s{}
//	a.m.SetIntegrityMeta()
//}

package syncer

import (
	"errors"
	"io"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// SyncPiece syncs piece data to secondary storage provider
func (s *Syncer) SyncPiece(stream service.SyncerService_SyncPieceServer) error {
	var count uint32
	var integrityMeta *metadb.IntegrityMeta
	var key string
	var spID string
	var value []byte
	pieceHash := make(map[string][]byte)
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
			return stream.SendAndClose(&service.SyncerServiceSyncPieceResponse{
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
		spID = req.GetSyncerInfo().GetStorageProviderId()
		integrityMeta, key, value, err = s.handlePieceData(req)
		if err != nil {
			return err
		}
		pieceHash[key] = hash.GenerateChecksum(value)
		count++
	}
}

func (s *Syncer) setIntegrityMeta(db metadb.MetaDB, meta *metadb.IntegrityMeta) error {
	if err := db.SetIntegrityMeta(meta); err != nil {
		log.Errorw("set integrity meta error", "error", err)
		return err
	}
	return nil
}

func generateSealInfo(spID string, integrityMeta *metadb.IntegrityMeta) *service.StorageProviderSealInfo {
	keys := util.GenericSortedKeys(integrityMeta.PieceHash)
	pieceChecksumList := make([][]byte, 0)
	var integrityHash []byte
	for _, key := range keys {
		value := integrityMeta.PieceHash[key]
		pieceChecksumList = append(pieceChecksumList, value)
	}
	integrityHash = hash.GenerateIntegrityHash(pieceChecksumList)
	resp := &service.StorageProviderSealInfo{
		StorageProviderId: spID,
		PieceIdx:          integrityMeta.PieceIdx,
		PieceChecksum:     pieceChecksumList,
		IntegrityHash:     integrityHash,
		Signature:         nil, // TODO(mock)
	}
	return resp
}

func (s *Syncer) handlePieceData(req *service.SyncerServiceSyncPieceRequest) (*metadb.IntegrityMeta, string, []byte, error) {
	if len(req.GetPieceData()) != 1 {
		return nil, "", nil, errors.New("the length of piece data map is not equal to 1")
	}

	var key string
	var value []byte
	redundancyType := req.GetSyncerInfo().GetRedundancyType()
	integrityMeta := &metadb.IntegrityMeta{
		ObjectID:       req.GetSyncerInfo().GetObjectId(),
		PieceCount:     req.GetSyncerInfo().GetPieceCount(),
		IsPrimary:      false,
		RedundancyType: redundancyType,
	}
	for k, v := range req.GetPieceData() {
		key = k
		value = v
		pieceIndex, err := parsePieceIndex(redundancyType, key)
		if err != nil {
			return nil, "", nil, err
		}
		integrityMeta.PieceIdx = pieceIndex

		// put piece data into piece store
		if err = s.store.PutPiece(key, value); err != nil {
			log.Errorw("put piece failed", "error", err)
			return nil, "", nil, err
		}
	}
	return integrityMeta, key, value, nil
}

func parsePieceIndex(redundancyType ptypes.RedundancyType, key string) (uint32, error) {
	var (
		err        error
		pieceIndex uint32
	)
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

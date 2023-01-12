package syncer

import (
	"errors"
	"io"

	merrors "github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/model/piecestore"
	ptypes "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/store/metadb"
	"github.com/bnb-chain/inscription-storage-provider/util"
	"github.com/bnb-chain/inscription-storage-provider/util/hash"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// SyncPiece syncs piece data to secondary storage provider
func (s *Syncer) SyncPiece(stream service.SyncerService_SyncPieceServer) error {
	var sealInfo *service.StorageProviderSealInfo
	var count uint32
	//var pieceIndex uint32
	//var pieceCount uint32
	//var err error
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
			log.Infow("upload ec piece closed", "error", err, "storage_provider_id", sealInfo.GetStorageProviderId(),
				"piece_idx", sealInfo.GetPieceIdx(), "count", count)
			if count != integrityMeta.PieceCount {
				log.Errorw("syncer service received piece count is wrong")
				return merrors.ErrReceivedPieceCount
			}

			integrityMeta.PieceHash = pieceHash
			sealInfo = generateSealInfo(spID, integrityMeta)
			integrityMeta.IntegrityHash = sealInfo.GetIntegrityHash()
			if err := s.setIntegrityMeta(s.metaDB, integrityMeta); err != nil {
				log.Errorw("setIntegrityMeta error", "error", err)
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
		integrityMeta, key, value, err = s.gatherPieceData(req)
		if err != nil {
			return err
		}
		pieceHash[key] = hash.GenerateChecksum(value)
		count++
		//sealInfo, _, err = s.handleUploadPiece(req)
		//if err != nil {
		//	log.Errorw("handle upload piece error", "error", err)
		//	return err
		//}
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
	log.Info("SortedKeys", "keys", keys)
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

func (s *Syncer) gatherPieceData(req *service.SyncerServiceSyncPieceRequest) (*metadb.IntegrityMeta, string, []byte, error) {
	if len(req.GetPieceData()) != 1 {
		return nil, "", nil, errors.New("the length of piece data map is not equal to 1")
	}

	var integrityMeta *metadb.IntegrityMeta
	var key string
	var value []byte
	for k, v := range req.GetPieceData() {
		key = k
		value = v
		integrityMeta.ObjectID = req.GetSyncerInfo().GetObjectId()
		integrityMeta.PieceCount = req.GetSyncerInfo().GetPieceCount()
		integrityMeta.IsPrimary = false
		redundancyType := req.GetSyncerInfo().GetRedundancyType()
		integrityMeta.RedundancyType = redundancyType
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

// handleUploadPiece store piece data to piece store and compute integrity hash.
func (s *Syncer) handleUploadPiece(req *service.SyncerServiceSyncPieceRequest) (
	*service.StorageProviderSealInfo, uint32, error) {
	var (
		pieceIndex uint32
		err        error
	)
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

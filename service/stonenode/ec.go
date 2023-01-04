package stonenode

import (
	"context"
	"sync"

	"github.com/bnb-chain/inscription-storage-provider/model"
	"github.com/bnb-chain/inscription-storage-provider/model/piecestore"
	"github.com/bnb-chain/inscription-storage-provider/pkg/redundancy"
	ptypes "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

type seg struct {
	segIndex int
	segData  []byte
}

func (s *StoneNodeService) doEC(ctx context.Context) error {
	allocResp, err := s.AllocStoneJob(ctx)
	if err != nil {
		log.Errorw("doEC calls AllocStoneJob failed", "error", err, "traceID", allocResp.GetTraceId())
		return err
	}

	objectID := allocResp.GetPieceJob().GetObjectId()
	payloadSize := allocResp.GetPieceJob().GetPayloadSize()
	segNumber := getSegmentNumber(payloadSize)
	segChan := make(chan seg, segNumber)
	defer func() {
		close(segChan)
	}()
	// 1. get data from primary storage provider
	for i := 0; i <= int(segNumber); i++ {
		go func(i int) {
			pieceKey := piecestore.EncodeSegmentPieceKey(objectID, int(segNumber))
			data, err := s.store.getPiece(pieceKey, 0, 0)
			if err != nil {
				log.Errorw("stone node gets segment data from piece store failed", "error", err, "piece key", pieceKey)
				s.errChan <- err
			}
			se := seg{
				segIndex: i,
				segData:  data,
			}
			segChan <- se
		}(i)
	}

	// 2. send data to secondary storage provider
	sInfoChan := make(chan *service.StorageProviderSealInfo)
	for v := range segChan {
		pieceMap := &sync.Map{}
		go func(sd seg) {
			rType := allocResp.GetPieceJob().GetRedundancyType()
			pieceMap, err = ecOrReplicaOrInline(rType, objectID, uint64(sd.segIndex), sd.segData)
			if err != nil {
				log.Errorw("stone node ecOrReplicaOrInline failed", "error", err)
				s.errChan <- err
			}
			pieceMap.Range(func(key, value any) bool {
				k := key.(string)
				val := value.([]byte)
				go func() {
					resp, err := s.UploadECPiece(ctx, segNumber, &service.SyncerInfo{
						ObjectId:          objectID,
						StorageProviderId: mockGetStorageProviderID()[sd.segIndex],
						RedundancyType:    allocResp.GetPieceJob().GetRedundancyType(),
					}, k, val)
					if err != nil {
						log.Errorw("UploadECPiece failed", "error", err)
						s.errChan <- err
					}
					sInfoChan <- resp.GetSecondarySpInfo()
				}()
				return true
			})
		}(v)
	}

	// 3. notify stone hub when a segment is done
	txHash := allocResp.GetTxHash()
	for v := range sInfoChan {
		if err := s.DoneSecondaryPieceJob(ctx, txHash, v); err != nil {
			log.Errorw("done secondary piece job failed", "error", err)
			s.errChan <- err
		}
	}

	select {
	case err := <-s.errChan:
		return err
	}
}

func mockGetStorageProviderID() []string {
	return []string{"sp1", "sp2", "sp3", "sp4", "sp5", "sp6"}
}

// 拿到map里的数据是一个segment对应的ec，从ec0到ec5
func ecOrReplicaOrInline(rType ptypes.RedundancyType, objectID, segIndex uint64, data []byte) (*sync.Map, error) {
	pieceMap := &sync.Map{}
	switch rType {
	case ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED:
		ecData, err := redundancy.EncodeRawSegment(data)
		if err != nil {
			log.Errorw("stone node EncodeRawSegment failed", "error", err)
			return nil, err
		}
		for i, v := range ecData {
			ecPieceKey := piecestore.EncodeECPieceKey(objectID, int(segIndex), i)
			pieceMap.Store(ecPieceKey, v)
		}
	default:
		pieceMap.Store(piecestore.EncodeSegmentPieceKey(objectID, int(segIndex)), data)
	}
	return pieceMap, nil
}

func getSegmentNumber(payloadSize uint64) uint64 {
	var segmentNumber uint64
	if payloadSize >= model.SegmentSize {
		if payloadSize%model.SegmentSize == 0 {
			segmentNumber = payloadSize / model.SegmentSize
		} else {
			segmentNumber = payloadSize/model.SegmentSize + 1
		}
	} else {
		segmentNumber = 0
	}
	return segmentNumber
}

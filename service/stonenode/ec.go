package stonenode

import (
	"context"
	"fmt"
	"sync"

	"github.com/bnb-chain/inscription-storage-provider/model"
	"github.com/bnb-chain/inscription-storage-provider/model/piecestore"
	"github.com/bnb-chain/inscription-storage-provider/pkg/redundancy"
	ptypes "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

func (s *StoneNodeService) doEC(ctx context.Context) error {
	allocResp, err := s.AllocStoneJob(ctx)
	if err != nil {
		log.Errorw("doEC calls AllocStoneJob failed", "error", err, "traceID", allocResp.GetTraceId())
		return err
	}

	var (
		objectID    = allocResp.GetPieceJob().GetObjectId()
		payloadSize = allocResp.GetPieceJob().GetPayloadSize()
		traceID     = allocResp.GetTraceId()
		txHash      = allocResp.GetTxHash()
		rType       = allocResp.GetPieceJob().GetRedundancyType()
		segNumber   = getSegmentNumber(payloadSize)

		wg        = new(sync.WaitGroup)
		errChan   = make(chan error)
		pieceChan = make(chan map[string]map[string][]byte, segNumber)
		bigMap    = make(map[string]map[string][]byte)
	)

	// 1. get data from primary storage provider and ec
	wg.Add(int(segNumber))
	for i := 0; i <= int(segNumber); i++ {
		go func(i int) {
			defer func() {
				close(pieceChan)
				wg.Done()
			}()
			pieceKey := piecestore.EncodeSegmentPieceKey(objectID, int(segNumber))
			data, err := s.store.getPiece(ctx, pieceKey, 0, 0)
			if err != nil {
				log.Errorw("stone node gets segment data from piece store failed", "error", err, "piece key", pieceKey)
				errChan <- err
			}

			pieceMap, err := ecOrReplicaOrInline(rType, objectID, uint64(i), data)
			if err != nil {
				log.Errorw("stone node ecOrReplicaOrInline failed", "error", err)
				errChan <- err
			}
			pieceChan <- pieceMap
		}(i)
	}
	wg.Wait()

	for pm := range pieceChan {
		for k, v := range pm {
			bigMap[k] = v
		}
	}

	// 2. send data to secondary storage provider and notify stone hub when a segment is done
	for _, v := range bigMap {
		go func(value map[string][]byte) {
			resp, err := s.UploadECPiece(ctx, segNumber, &service.SyncerInfo{
				ObjectId:          objectID,
				TxHash:            txHash,
				StorageProviderId: mockGetStorageProviderID()[1],
				RedundancyType:    allocResp.GetPieceJob().GetRedundancyType(),
			}, value, traceID)
			if err != nil {
				log.Errorw("UploadECPiece failed", "error", err)
				errChan <- err
			}

			// notify stone hub when an ec segment is done
			pieceJob := &service.PieceJob{
				BucketName:              allocResp.GetPieceJob().GetBucketName(),
				ObjectName:              allocResp.GetPieceJob().GetObjectName(),
				TxHash:                  txHash,
				ObjectId:                objectID,
				PayloadSize:             payloadSize,
				TargetIdx:               allocResp.GetPieceJob().GetTargetIdx(),
				RedundancyType:          rType,
				StorageProviderSealInfo: resp.GetSecondarySpInfo(),
			}
			if err := s.DoneSecondaryPieceJob(ctx, pieceJob, traceID); err != nil {
				log.Errorw("done secondary piece job failed", "error", err)
				errChan <- err
			}
		}(v)
	}

	select {
	case err := <-errChan:
		return err
	}
}

func mockGetStorageProviderID() []string {
	return []string{"sp1", "sp2", "sp3", "sp4", "sp5", "sp6"}
}

func ecOrReplicaOrInline(rType ptypes.RedundancyType, objectID, segIndex uint64, data []byte) (map[string]map[string][]byte, error) {
	var pieceMap map[string][]byte
	var bigMap map[string]map[string][]byte
	switch rType {
	case ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED:
		ecData, err := redundancy.EncodeRawSegment(data)
		if err != nil {
			log.Errorw("stone node EncodeRawSegment failed", "error", err)
			return nil, err
		}
		for i, j := range ecData {
			ecPieceKey := piecestore.EncodeECPieceKey(objectID, int(segIndex), i)
			pieceMap[ecPieceKey] = j
			bigMap[encodeEC(i)] = pieceMap
		}
	default:
		pieceMap[piecestore.EncodeSegmentPieceKey(objectID, int(segIndex))] = data
		bigMap[encodeSegment(int(segIndex))] = pieceMap
	}
	return bigMap, nil
}

func encodeEC(ecIndex int) string {
	return fmt.Sprintf("p%d", ecIndex)
}

func encodeSegment(segIndex int) string {
	return fmt.Sprintf("s%d", segIndex)
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

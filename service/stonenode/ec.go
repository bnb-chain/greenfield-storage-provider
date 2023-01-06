package stonenode

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/bnb-chain/inscription-storage-provider/model"
	"github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/model/piecestore"
	"github.com/bnb-chain/inscription-storage-provider/pkg/redundancy"
	ptypes "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/hash"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

func (s *StoneNodeService) doEC(ctx context.Context, storageProvider string) error {
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

		wg       = new(sync.WaitGroup)
		mu       = new(sync.RWMutex)
		errChan  = make(chan error, 10)
		pieceMap = make(map[string]map[string][]byte)
	)

	// 1. get data from primary storage provider and ec
	for i := 0; i <= int(segNumber); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			pieceKey := piecestore.EncodeSegmentPieceKey(objectID, segNumber)
			data, err := s.store.getPiece(ctx, pieceKey, 0, 0)
			if err != nil {
				log.Errorw("stone node gets segment data from piece store failed", "error", err, "piece key", pieceKey)
				errChan <- err
			}

			dataSlice, err := ecOrReplicaOrInline(rType, objectID, uint32(i), data)
			if err != nil {
				log.Errorw("stone node ecOrReplicaOrInline failed", "error", err)
				errChan <- err
			}

			for k, v := range dataSlice {
				mu.Lock()
				pieceMap[idKey(k)] = v
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	// 2. send data to secondary storage provider and notify stone hub when a segment is done
	for _, v := range pieceMap {
		go func(value map[string][]byte) {
			resp, err := s.UploadECPiece(ctx, segNumber, &service.SyncerInfo{
				ObjectId:          objectID,
				TxHash:            txHash,
				StorageProviderId: storageProvider,
				RedundancyType:    allocResp.GetPieceJob().GetRedundancyType(),
			}, value, traceID)
			if err != nil {
				log.Errorw("UploadECPiece failed", "error", err)
				errChan <- err
			}

			// check integrity hash of secondary sp if is equal to integrity hash of primary sp
			slice := mapValueToSlice(value, len(value))
			integrityHash := hash.GenerateIntegrityHash(slice, "bnb-sp")
			if equal := bytes.Equal(integrityHash, resp.GetSecondarySpInfo().GetIntegrityHash()); !equal {
				errChan <- errors.ErrIntegrityHash
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

func mapValueToSlice(m map[string][]byte, length int) [][]byte {
	slice := make([][]byte, 0, length)
	for _, v := range m {
		slice = append(slice, v)
	}
	return slice
}

func ecOrReplicaOrInline(rType ptypes.RedundancyType, objectID uint64, segIndex uint32, data []byte) ([]map[string][]byte, error) {
	mapSlice := make([]map[string][]byte, 0)
	switch rType {
	case ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED:
		ecData, err := redundancy.EncodeRawSegment(data)
		if err != nil {
			log.Errorw("stone node EncodeRawSegment failed", "error", err)
			return nil, err
		}
		for i, j := range ecData {
			pieceMap := make(map[string][]byte)
			ecPieceKey := piecestore.EncodeECPieceKey(objectID, segIndex, uint32(i))
			pieceMap[ecPieceKey] = j
			mapSlice = append(mapSlice, pieceMap)
		}
	case ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE, ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
		pieceMap := make(map[string][]byte)
		pieceMap[piecestore.EncodeSegmentPieceKey(objectID, segIndex)] = data
		mapSlice = append(mapSlice, pieceMap)
	default:
		return nil, fmt.Errorf("Unknown redundancy type")
	}
	return mapSlice, nil
}

func idKey(index int) string {
	return fmt.Sprintf("i%d", index)
}

func getSegmentNumber(payloadSize uint64) uint32 {
	var segmentNumber uint32
	if payloadSize >= model.SegmentSize {
		if payloadSize%model.SegmentSize == 0 {
			segmentNumber = uint32(payloadSize / model.SegmentSize)
		} else {
			segmentNumber = uint32(payloadSize/model.SegmentSize + 1)
		}
	} else {
		segmentNumber = 0
	}
	return segmentNumber
}

package stonenode

import (
	"bytes"
	"context"
	"errors"
	"sync"

	"github.com/bnb-chain/inscription-storage-provider/mock"
	"github.com/bnb-chain/inscription-storage-provider/model/piecestore"
	"github.com/bnb-chain/inscription-storage-provider/pkg/redundancy"
	ptypes "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util"
	"github.com/bnb-chain/inscription-storage-provider/util/hash"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// doSyncToSecondarySP load segment data from primary and sync to secondary.
func (node *StoneNodeService) doSyncToSecondarySP(ctx context.Context, allocResp *service.StoneHubServiceAllocStoneJobResponse) error {
	// TBD:: check secondarySPs count by redundancyType.
	// EC_TYPE need EC_M + EC_K + backup
	// REPLICA_TYPE and INLINE_TYPE need segments
	secondarySPs := mock.AllocUploadSecondarySP()
	pieceData, err := node.loadSegmentsData(ctx, allocResp)
	if err != nil {
		return err
	}
	redundancyType := allocResp.GetPieceJob().GetRedundancyType()
	secondaryPieceData, err := node.dispatchSecondarySP(pieceData, redundancyType, secondarySPs)
	if err != nil {
		return err
	}
	node.syncPieceToSecondarySP(ctx, allocResp, secondaryPieceData)
	return nil
}

// loadSegmentsData load segment data from primary storage provider.
func (node *StoneNodeService) loadSegmentsData(ctx context.Context, allocResp *service.StoneHubServiceAllocStoneJobResponse) (map[string][][]byte, error) {
	type segment struct {
		pieceKey  string
		pieceData [][]byte
		pieceErr  error
	}
	var (
		segmentCh      = make(chan *segment)
		objectID       = allocResp.GetPieceJob().GetObjectId()
		payloadSize    = allocResp.GetPieceJob().GetPayloadSize()
		redundancyType = allocResp.GetPieceJob().GetRedundancyType()
		segmentCount   = util.ComputeSegmentCount(payloadSize)
		pieces         = make(map[string][][]byte)
	)
	for i := 0; i < int(segmentCount); i++ {
		go func(segmentIdx int) {
			segment := &segment{}
			pieceKey := piecestore.EncodeSegmentPieceKey(objectID, segmentIdx)
			data, err := node.store.getPiece(ctx, pieceKey, 0, 0)
			if err != nil {
				log.CtxErrorw(ctx, "stone node gets segment data from piece store failed", "error", err, "piece key", pieceKey)
				segment.pieceErr = err
				segmentCh <- segment
				return
			}

			pieceData, err := node.spiltSegmentData(ctx, redundancyType, data)
			if err != nil {
				log.CtxErrorw(ctx, "stone node ec failed", "error", err, "piece key", pieceKey)
				segment.pieceErr = err
				segmentCh <- segment
				return
			}
			segment.pieceData = pieceData
			segmentCh <- segment
			return
		}(i)
	}

	var mu sync.Mutex
	for {
		mu.Lock()
		if len(pieces) == segmentCount {
			mu.Unlock()
			break
		}
		mu.Unlock()

		select {
		case segment := <-segmentCh:
			if segment.pieceErr != nil {
				return pieces, segment.pieceErr
			}
			mu.Lock()
			pieces[segment.pieceKey] = segment.pieceData
			mu.Unlock()
		}
	}
	return pieces, nil
}

// spiltSegmentData spilt segment data into pieces data.
func (node *StoneNodeService) spiltSegmentData(ctx context.Context, redundancyType ptypes.RedundancyType, segmentData []byte) ([][]byte, error) {
	var (
		pieceData [][]byte
		err       error
	)
	switch redundancyType {
	case ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED:
		pieceData, err = redundancy.EncodeRawSegment(segmentData)
		if err != nil {
			log.CtxErrorw(ctx, "ec encode failed", "error", err)
			return pieceData, err
		}
	default:
		pieceData = append(pieceData, segmentData)
	}
	return pieceData, nil
}

// dispatchSecondarySP dispatch piece data to secondary storage provider.
func (node *StoneNodeService) dispatchSecondarySP(pieceData map[string][][]byte, redundancyType ptypes.RedundancyType, secondarySPs []string) (map[string]map[string][]byte, error) {
	var secondaryPieceData map[string]map[string][]byte
	for pieceKey, pieceData := range pieceData {
		for idx, data := range pieceData {
			if idx >= len(secondarySPs) {
				return secondaryPieceData, errors.New("secondary sp is not enough")
			}
			if redundancyType == ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED {
				if _, ok := secondaryPieceData[secondarySPs[idx]]; !ok {
					secondaryPieceData[secondarySPs[idx]] = make(map[string][]byte)
				}
				key := piecestore.EncodeECPieceKeyBySegmentKey(pieceKey, idx)
				secondaryPieceData[secondarySPs[idx]][key] = data
			} else {
				if _, ok := secondaryPieceData[secondarySPs[idx]]; !ok {
					secondaryPieceData[secondarySPs[idx]] = make(map[string][]byte)
				}
				secondaryPieceData[secondarySPs[idx]][pieceKey] = data
			}
		}
	}
	return secondaryPieceData, nil
}

// syncPieceToSecondarySP send piece data to the secondary
func (node *StoneNodeService) syncPieceToSecondarySP(ctx context.Context, resp *service.StoneHubServiceAllocStoneJobResponse,
	secondaryPieceData map[string]map[string][]byte) error {
	var (
		objectID       = resp.GetPieceJob().GetObjectId()
		payloadSize    = resp.GetPieceJob().GetPayloadSize()
		redundancyType = resp.GetPieceJob().GetRedundancyType()
		segmentCount   = util.ComputeSegmentCount(payloadSize)
		txHash         = resp.GetTxHash()
	)

	for secondary, pieceData := range secondaryPieceData {
		go func(secondary string, pieceData map[string][]byte) {
			errMsg := &service.ErrMessage{}
			pieceJob := &service.PieceJob{
				BucketName:     resp.GetPieceJob().GetBucketName(),
				ObjectName:     resp.GetPieceJob().GetObjectName(),
				TxHash:         txHash,
				ObjectId:       objectID,
				PayloadSize:    payloadSize,
				RedundancyType: redundancyType,
			}

			defer func() {
				// notify stone hub when an ec segment is done
				if err := node.DoneSecondaryPieceJob(ctx, resp.TraceId, pieceJob, errMsg); err != nil {
					log.CtxErrorw(ctx, "done secondary piece job to stone hub failed", "error", err)
					return
				}
				log.CtxInfow(ctx, "upload secondary piece job secondary", "secondary sp", secondary)
			}()

			syncResp, err := node.UploadECPiece(ctx, segmentCount, &service.SyncerInfo{
				ObjectId:          objectID,
				TxHash:            txHash,
				StorageProviderId: secondary,
				RedundancyType:    redundancyType,
			}, pieceData, resp.TraceId)
			// TBD:: retry alloc secondary sp and rat again.
			if err != nil {
				log.CtxErrorw(ctx, "sync to secondary piece job failed", "error", err)
				errMsg.ErrCode = service.ErrCode_ERR_CODE_ERROR
				errMsg.ErrMsg = err.Error()
				return
			}
			var pieceHash [][]byte
			for _, data := range pieceData {
				pieceHash = append(pieceHash, hash.GenerateChecksum(data))
			}
			integrityHash := hash.GenerateIntegrityHash(pieceHash, secondary)
			if syncResp.GetSecondarySpInfo() == nil ||
				syncResp.GetSecondarySpInfo().GetIntegrityHash() == nil ||
				bytes.Equal(integrityHash, syncResp.GetSecondarySpInfo().GetIntegrityHash()) {
				log.CtxErrorw(ctx, "secondary integrity hash check error")
				errMsg.ErrCode = service.ErrCode_ERR_CODE_ERROR
				errMsg.ErrMsg = errors.New("secondary integrity hash check error").Error()
				return
			}
			pieceJob.StorageProviderSealInfo = syncResp.GetSecondarySpInfo()
			return
		}(secondary, pieceData)
	}
	return nil
}

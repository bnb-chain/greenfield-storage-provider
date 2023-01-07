package stonenode

import (
	"bytes"
	"context"
	"sync"
	"sync/atomic"

	"github.com/bnb-chain/inscription-storage-provider/mock"
	merrors "github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/model/piecestore"
	"github.com/bnb-chain/inscription-storage-provider/pkg/redundancy"
	ptypes "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util"
	"github.com/bnb-chain/inscription-storage-provider/util/hash"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// syncPieceToSecondarySP load segment data from primary and sync to secondary.
func (node *StoneNodeService) syncPieceToSecondarySP(ctx context.Context, allocResp *service.StoneHubServiceAllocStoneJobResponse) error {
	// TBD:: check secondarySPs count by redundancyType.
	// EC_TYPE need EC_M + EC_K + backup
	// REPLICA_TYPE and INLINE_TYPE need segments
	secondarySPs := mock.AllocUploadSecondarySP()
	pieceData, err := node.loadSegmentsData(ctx, allocResp)
	if err != nil {
		node.reportErrToStoneHub(ctx, allocResp, err)
		return err
	}
	redundancyType := allocResp.GetPieceJob().GetRedundancyType()
	secondaryPieceData, err := node.dispatchSecondarySP(pieceData, redundancyType, secondarySPs)
	if err != nil {
		node.reportErrToStoneHub(ctx, allocResp, err)
		return err
	}
	node.doSyncToSecondarySP(ctx, allocResp, secondaryPieceData)
	return nil
}

// loadSegmentsData load segment data from primary storage provider.
func (node *StoneNodeService) loadSegmentsData(ctx context.Context, allocResp *service.StoneHubServiceAllocStoneJobResponse) (map[string][][]byte, error) {
	type segment struct {
		objectID       uint64
		pieceKey       string
		segmentData    []byte
		pieceData      [][]byte
		pieceErr       error
		redundancyType ptypes.RedundancyType
	}
	var (
		doneSegments   int64
		loadSegmentErr error
		segmentCh      = make(chan *segment)
		interruptCh    = make(chan struct{})
		pieces         = make(map[string][][]byte)
		objectID       = allocResp.GetPieceJob().GetObjectId()
		payloadSize    = allocResp.GetPieceJob().GetPayloadSize()
		redundancyType = allocResp.GetPieceJob().GetRedundancyType()
		segmentCount   = util.ComputeSegmentCount(payloadSize)
	)

	loadFunc := func(ctx context.Context, segment *segment) error {
		select {
		case <-interruptCh:
			break
		default:
			data, err := node.store.getPiece(ctx, segment.pieceKey, 0, 0)
			if err != nil {
				log.CtxErrorw(ctx, "stone node gets segment data from piece store failed", "error", err,
					"piece key", segment.pieceKey)
				return err
			}
			segment.segmentData = data
		}
		return nil
	}

	spiltFunc := func(ctx context.Context, segment *segment) error {
		select {
		case <-interruptCh:
			break
		default:
			pieceData, err := node.spiltSegmentData(ctx, redundancyType, segment.segmentData)
			if err != nil {
				log.CtxErrorw(ctx, "stone node ec failed", "error", err, "piece key", segment.pieceKey)
				return err
			}
			segment.pieceData = pieceData
		}
		return nil
	}

	for i := 0; i < int(segmentCount); i++ {
		go func(segmentIdx int) {
			seg := &segment{
				objectID:       objectID,
				pieceKey:       piecestore.EncodeSegmentPieceKey(objectID, uint32(segmentIdx)),
				redundancyType: redundancyType,
			}
			defer func() {
				if seg.pieceErr != nil || atomic.AddInt64(&doneSegments, 1) == int64(segmentCount) {
					close(interruptCh)
					close(segmentCh)
				}
			}()
			if loadSegmentErr = loadFunc(ctx, seg); loadSegmentErr != nil {
				return
			}
			if loadSegmentErr = spiltFunc(ctx, seg); loadSegmentErr != nil {
				return
			}
			select {
			case <-interruptCh:
				return
			default:
				segmentCh <- seg
			}
		}(i)
	}

	var mu sync.Mutex
	for seg := range segmentCh {
		mu.Lock()
		pieces[seg.pieceKey] = seg.pieceData
		mu.Unlock()
	}
	return pieces, loadSegmentErr
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
	case ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE, ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
		pieceData = append(pieceData, segmentData)
	default:
		return nil, merrors.ErrRedundancyType
	}
	return pieceData, nil
}

// dispatchSecondarySP dispatch piece data to secondary storage provider.
func (node *StoneNodeService) dispatchSecondarySP(pieceData map[string][][]byte, redundancyType ptypes.RedundancyType,
	secondarySPs []string) (map[string]map[string][]byte, error) {
	var secondaryPieceData map[string]map[string][]byte
	for pieceKey, pData := range pieceData {
		for idx, data := range pData {
			if idx >= len(secondarySPs) {
				return secondaryPieceData, merrors.ErrSecondarySPNumber
			}
			if redundancyType == ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED {
				if _, ok := secondaryPieceData[secondarySPs[idx]]; !ok {
					secondaryPieceData[secondarySPs[idx]] = make(map[string][]byte)
				}
				key := piecestore.EncodeECPieceKeyBySegmentKey(pieceKey, idx)
				secondaryPieceData[secondarySPs[idx]][key] = data
			} else if redundancyType == ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE ||
				redundancyType == ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE {
				if _, ok := secondaryPieceData[secondarySPs[idx]]; !ok {
					secondaryPieceData[secondarySPs[idx]] = make(map[string][]byte)
				}
				secondaryPieceData[secondarySPs[idx]][pieceKey] = data
			}
		}
	}
	return secondaryPieceData, nil
}

// doSyncToSecondarySP send piece data to the secondary.
func (node *StoneNodeService) doSyncToSecondarySP(ctx context.Context, resp *service.StoneHubServiceAllocStoneJobResponse,
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

			syncResp, err := node.UploadECPiece(ctx, int(segmentCount), &service.SyncerInfo{
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
				errMsg.ErrMsg = merrors.ErrIntegrityHash.Error()
				return
			}
			pieceJob.StorageProviderSealInfo = syncResp.GetSecondarySpInfo()
			return
		}(secondary, pieceData)
	}
	return nil
}

// reportErrToStoneHub send error message to stone hub.
func (node *StoneNodeService) reportErrToStoneHub(ctx context.Context, resp *service.StoneHubServiceAllocStoneJobResponse, reportErr error) {
	if reportErr == nil {
		return
	}
	var (
		objectID       = resp.GetPieceJob().GetObjectId()
		payloadSize    = resp.GetPieceJob().GetPayloadSize()
		redundancyType = resp.GetPieceJob().GetRedundancyType()
		txHash         = resp.GetTxHash()
	)
	pieceJob := &service.PieceJob{
		BucketName:     resp.GetPieceJob().GetBucketName(),
		ObjectName:     resp.GetPieceJob().GetObjectName(),
		TxHash:         txHash,
		ObjectId:       objectID,
		PayloadSize:    payloadSize,
		RedundancyType: redundancyType,
	}
	errMsg := &service.ErrMessage{
		ErrCode: service.ErrCode_ERR_CODE_ERROR,
		ErrMsg:  reportErr.Error(),
	}
	if err := node.DoneSecondaryPieceJob(ctx, resp.TraceId, pieceJob, errMsg); err != nil {
		log.CtxErrorw(ctx, "report stone hub err msg failed", "error", err)
		return
	}
	log.CtxInfow(ctx, "report stone hub err msg success")
}

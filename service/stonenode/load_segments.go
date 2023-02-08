package stonenode

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/redundancy"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// loadSegmentsData load segment data from primary storage provider
func (node *StoneNodeService) loadSegmentsData(ctx context.Context, allocResp *stypes.StoneHubServiceAllocStoneJobResponse) (
	[][][]byte, error) {
	type segment struct {
		objectID       uint64
		pieceKey       string
		segmentData    []byte
		pieceData      [][]byte
		pieceErr       error
		segmentIndex   int
		redundancyType ptypes.RedundancyType
	}
	var (
		doneSegments   int64
		loadSegmentErr error
		segmentCh      = make(chan *segment)
		interruptCh    = make(chan struct{})
		pieces         = make(map[int][][]byte)
		objectID       = allocResp.GetPieceJob().GetObjectId()
		payloadSize    = allocResp.GetPieceJob().GetPayloadSize()
		redundancyType = allocResp.GetPieceJob().GetRedundancyType()
		segmentCount   = util.ComputeSegmentCount(payloadSize)
	)

	loadFunc := func(ctx context.Context, seg *segment) error {
		select {
		case <-interruptCh:
			break
		default:
			data, err := node.store.GetPiece(ctx, seg.pieceKey, 0, 0)
			if err != nil {
				log.CtxErrorw(ctx, "gets segment data from piece store failed", "error", err, "piece key",
					seg.pieceKey)
				return err
			}
			seg.segmentData = data
		}
		return nil
	}

	spiltFunc := func(ctx context.Context, seg *segment) error {
		select {
		case <-interruptCh:
			break
		default:
			pieceData, err := node.generatePieceData(redundancyType, seg.segmentData)
			if err != nil {
				log.CtxErrorw(ctx, "ec encode failed", "error", err, "piece key", seg.pieceKey)
				return err
			}
			seg.pieceData = pieceData
		}
		return nil
	}

	for i := 0; i < int(segmentCount); i++ {
		go func(segmentIdx int) {
			seg := &segment{
				objectID:       objectID,
				pieceKey:       piecestore.EncodeSegmentPieceKey(objectID, uint32(segmentIdx)),
				redundancyType: redundancyType,
				segmentIndex:   segmentIdx,
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
		pieces[seg.segmentIndex] = seg.pieceData
		mu.Unlock()
	}
	return util.MapValueToSlice(pieces), loadSegmentErr
}

// generatePieceData generates piece data from segment data
// pieceData contains one []byte when redundancyType is replica or inline
// pieceData contains six []byte when redundancyType is ec
func (node *StoneNodeService) generatePieceData(redundancyType ptypes.RedundancyType, segmentData []byte) (
	pieceData [][]byte, err error) {
	switch redundancyType {
	case ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE, ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
		pieceData = append(pieceData, segmentData)
	default: // ec type
		pieceData, err = redundancy.EncodeRawSegment(segmentData)
		if err != nil {
			return nil, err
		}
	}
	return pieceData, nil
}

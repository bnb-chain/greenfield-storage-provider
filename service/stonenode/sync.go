package stonenode

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/bnb-chain/greenfield-storage-provider/mock"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/redundancy"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// syncPieceToSecondarySP load segment data from primary and sync to secondary.
func (node *StoneNodeService) syncPieceToSecondarySP(ctx context.Context, allocResp *service.StoneHubServiceAllocStoneJobResponse) error {
	// TBD:: check secondarySPs count by redundancyType.
	// EC_TYPE need EC_M + EC_K + backup
	// REPLICA_TYPE and INLINE_TYPE need segments count + backup
	secondarySPs := mock.AllocUploadSecondarySP()

	// 1. load all segments data from primary piece store and do ec or not
	pieceData, err := node.loadSegmentsData(ctx, allocResp)
	if err != nil {
		node.reportErrToStoneHub(ctx, allocResp, err)
		return err
	}

	// check redundancyType and targetIdx is valid
	redundancyType := allocResp.GetPieceJob().GetRedundancyType()
	if err := checkRedundancyType(redundancyType); err != nil {
		node.reportErrToStoneHub(ctx, allocResp, err)
		return err
	}
	targetIdx := allocResp.GetPieceJob().GetTargetIdx()
	if len(targetIdx) == 0 {
		node.reportErrToStoneHub(ctx, allocResp, merrors.ErrEmptyTargetIdx)
		return merrors.ErrEmptyTargetIdx
	}

	// 2. dispatch the piece data to different secondary
	secondaryPieceData, err := node.dispatchSecondarySP(pieceData, redundancyType, secondarySPs, targetIdx)
	if err != nil {
		log.CtxErrorw(ctx, "dispatch piece data to secondary sp error")
		node.reportErrToStoneHub(ctx, allocResp, err)
		return err
	}

	// 3. send piece data to the secondary
	node.doSyncToSecondarySP(ctx, allocResp, secondaryPieceData)
	return nil
}

func checkRedundancyType(redundancyType ptypes.RedundancyType) error {
	switch redundancyType {
	case ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED:
		return nil
	case ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE, ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
		return nil
	default:
		return merrors.ErrRedundancyType
	}
}

// loadSegmentsData load segment data from primary storage provider.
// returned map key is segmentKey, value is corresponding ec data from ec1 to ec6, or segment data
func (node *StoneNodeService) loadSegmentsData(ctx context.Context, allocResp *service.StoneHubServiceAllocStoneJobResponse) (
	map[string][][]byte, error) {
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
func (node *StoneNodeService) generatePieceData(redundancyType ptypes.RedundancyType, segmentData []byte) (
	pieceData [][]byte, err error) {
	switch redundancyType {
	case ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED:
		pieceData, err = redundancy.EncodeRawSegment(segmentData)
		if err != nil {
			return
		}
	case ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE, ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
		pieceData = append(pieceData, segmentData)
	default:
		return nil, merrors.ErrRedundancyType
	}
	return
}

// dispatchSecondarySP dispatch piece data to secondary storage provider.
// returned map key is spID, value map key is ec piece key or segment key, value map's value is piece data
func (node *StoneNodeService) dispatchSecondarySP(pieceDataBySegment map[string][][]byte, redundancyType ptypes.RedundancyType,
	secondarySPs []string, targetIdx []uint32) (map[string]map[string][]byte, error) {
	pieceDataBySecondary := make(map[string]map[string][]byte)

	if redundancyType == ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE {
		return fillReplicaOrInlineData(pieceDataBySegment, secondarySPs, targetIdx)
	}
	// pieceDataBySegment key is segment key; if redundancyType is EC, value is [][]byte type,
	// a two-dimensional array which contains ec data from ec1 []byte data to ec6 []byte data
	// if redundancyType is replica or inline, value is [][]byte type, a two-dimensional array
	// which only contains one []byte data
	var err error
	pieceDataBySecondary, err = fillECData(pieceDataBySegment, secondarySPs, targetIdx, redundancyType)
	if err != nil {
		log.Errorw("fill piece data by secondary error", "error", err)
		return nil, err
	}

	/*
		switch redundancyType {
		case ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE, ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
			pieceDataBySecondary, err = fillReplicaOrInlineData(pieceDataBySegment, secondarySPs, targetIdx)
		default: // ec type
			pieceDataBySecondary, err = fillECData(pieceDataBySegment, secondarySPs, targetIdx, redundancyType)
		}
		if err != nil {
			log.Errorw("fill piece data by secondary error", "error", err)
			return nil, err
		}
	*/
	return pieceDataBySecondary, nil
}

func fillECData(pieceDataBySegment map[string][][]byte, secondarySPs []string, targetIdx []uint32,
	redundancyType ptypes.RedundancyType) (map[string]map[string][]byte, error) {
	ecPieceDataMap := make(map[string]map[string][]byte)
	for pieceKey, pieceData := range pieceDataBySegment {
		//if len(pieceData) != 6 {
		//	return map[string]map[string][]byte{}, merrors.ErrInvalidECData
		//}

		for idx, data := range pieceData {
			if idx >= len(secondarySPs) {
				return map[string]map[string][]byte{}, merrors.ErrSecondarySPNumber
			}
			// initialize data map
			sp := secondarySPs[idx]
			if len(targetIdx) != 0 {
				for _, j := range targetIdx {
					if int(j) == idx {
						if _, ok := ecPieceDataMap[sp]; !ok {
							ecPieceDataMap[sp] = make(map[string][]byte)
						}
					}
				}
			}

			if len(targetIdx) != 0 {
				for _, index := range targetIdx {
					if int(index) == idx {
						switch redundancyType {
						case ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE, ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
							ecPieceDataMap[sp][pieceKey] = data
						default:
							key := piecestore.EncodeECPieceKeyBySegmentKey(pieceKey, uint32(idx))
							ecPieceDataMap[sp][key] = data
						}
					}
				}
			}
		}
	}
	return ecPieceDataMap, nil
}

func fillReplicaOrInlineData(pieceDataBySegment map[string][][]byte, secondarySPs []string, targetIdx []uint32) (
	map[string]map[string][]byte, error) {
	replicaOrInlineDataMap := make(map[string]map[string][]byte)
	if len(targetIdx) > len(secondarySPs) {
		return replicaOrInlineDataMap, merrors.ErrSecondarySPNumber
	}

	for _, pieceIdx := range targetIdx {
		replicaOrInlineDataMap[secondarySPs[pieceIdx]] = make(map[string][]byte)
		for segmentKey, data := range pieceDataBySegment {
			if len(data) != 1 {
				return replicaOrInlineDataMap, merrors.ErrInvalidSegmentData
			}
			replicaOrInlineDataMap[secondarySPs[pieceIdx]][segmentKey] = data[0]
		}
	}
	return replicaOrInlineDataMap, nil
	//// iterate map in order
	//keys := util.GenericSortedKeys(pieceDataBySegment)
	//for i := 0; i < len(keys); i++ {
	//	pieceKey := keys[i]
	//	pieceData := pieceDataBySegment[pieceKey]
	//	if len(pieceData) != 1 {
	//		return nil, merrors.ErrInvalidSegmentData
	//	}
	//
	//	// each segment piece data writes to different sp
	//	sp := secondarySPs[i]
	//	if len(targetIdx) != 0 {
	//		for _, index := range targetIdx {
	//			if int(index) == i {
	//				if _, ok := replicaOrInlineDataMap[sp]; !ok {
	//					replicaOrInlineDataMap[sp] = make(map[string][]byte)
	//				}
	//			}
	//		}
	//	}
	//
	//	if len(targetIdx) != 0 {
	//		for _, index := range targetIdx {
	//			if int(index) == i {
	//				replicaOrInlineDataMap[sp][pieceKey] = pieceData[0]
	//			}
	//		}
	//	}
	//}
	//return replicaOrInlineDataMap, nil
}

// doSyncToSecondarySP send piece data to the secondary.
func (node *StoneNodeService) doSyncToSecondarySP(ctx context.Context, resp *service.StoneHubServiceAllocStoneJobResponse,
	pieceDataBySecondary map[string]map[string][]byte) error {
	var (
		objectID       = resp.GetPieceJob().GetObjectId()
		payloadSize    = resp.GetPieceJob().GetPayloadSize()
		redundancyType = resp.GetPieceJob().GetRedundancyType()
		//segmentCount   = util.ComputeSegmentCount(payloadSize)
		txHash = resp.GetTxHash()
	)
	for secondary, pieceData := range pieceDataBySecondary {
		go func(secondary string, pieceData map[string][]byte) {
			errMsg := &service.ErrMessage{}
			pieceJob := &service.PieceJob{
				TxHash:         txHash,
				ObjectId:       objectID,
				PayloadSize:    payloadSize,
				RedundancyType: redundancyType,
			}

			defer func() {
				// notify stone hub when an ec segment is done
				req := &service.StoneHubServiceDoneSecondaryPieceJobRequest{
					TraceId:    resp.GetTraceId(),
					TxHash:     pieceJob.GetTxHash(),
					PieceJob:   pieceJob,
					ErrMessage: errMsg,
				}
				// TBD:: according to the secondary to pick up the address
				if _, err := node.stoneHub.DoneSecondaryPieceJob(ctx, req); err != nil {
					log.CtxErrorw(ctx, "done secondary piece job to stone hub failed", "error", err)
					return
				}
			}()

			syncResp, err := node.syncPiece(ctx, &service.SyncerInfo{
				ObjectId:          objectID,
				TxHash:            txHash,
				StorageProviderId: secondary,
				PieceCount:        uint32(len(pieceData)),
				RedundancyType:    redundancyType,
			}, pieceData, resp.GetTraceId())
			// TBD:: retry alloc secondary sp and rat again.
			if err != nil {
				log.CtxErrorw(ctx, "sync to secondary piece job failed", "error", err)
				errMsg.ErrCode = service.ErrCode_ERR_CODE_ERROR
				errMsg.ErrMsg = err.Error()
				return
			}

			var pieceHash [][]byte
			keys := util.GenericSortedKeys(pieceData)
			log.Debugw("sorted keys", "keys", keys)
			for _, key := range keys {
				pieceHash = append(pieceHash, hash.GenerateChecksum(pieceData[key]))
			}
			integrityHash := hash.GenerateIntegrityHash(pieceHash)
			if syncResp.GetSecondarySpInfo() == nil || syncResp.GetSecondarySpInfo().GetIntegrityHash() == nil ||
				!bytes.Equal(integrityHash, syncResp.GetSecondarySpInfo().GetIntegrityHash()) {
				log.CtxErrorw(ctx, "secondary integrity hash check error")
				errMsg.ErrCode = service.ErrCode_ERR_CODE_ERROR
				errMsg.ErrMsg = merrors.ErrIntegrityHash.Error()
				return
			}
			pieceJob.StorageProviderSealInfo = syncResp.GetSecondarySpInfo()
			log.CtxDebugw(ctx, "sync piece data to secondary", "secondary_provider", secondary,
				"local_integrity_hash", integrityHash, "remote_integrity_hash", syncResp.GetSecondarySpInfo().GetIntegrityHash())
			return
		}(secondary, pieceData)
	}
	return nil
}

// reportErrToStoneHub send error message to stone hub.
func (node *StoneNodeService) reportErrToStoneHub(ctx context.Context, resp *service.StoneHubServiceAllocStoneJobResponse,
	reportErr error) {
	if reportErr == nil {
		return
	}
	req := &service.StoneHubServiceDoneSecondaryPieceJobRequest{
		TraceId: resp.GetTraceId(),
		TxHash:  resp.GetTxHash(),
		ErrMessage: &service.ErrMessage{
			ErrCode: service.ErrCode_ERR_CODE_ERROR,
			ErrMsg:  reportErr.Error(),
		},
	}
	if _, err := node.stoneHub.DoneSecondaryPieceJob(ctx, req); err != nil {
		log.CtxErrorw(ctx, "report stone hub err msg failed", "error", err)
		return
	}
	log.CtxInfow(ctx, "report stone hub err msg success")
}

// SyncPiece send rpc request to secondary storage provider to sync the piece data.
func (node *StoneNodeService) syncPiece(ctx context.Context, syncerInfo *service.SyncerInfo,
	pieceData map[string][]byte, traceID string) (*service.SyncerServiceSyncPieceResponse, error) {
	log.CtxInfow(ctx, "stone node upload piece data", "redundancy_type", syncerInfo.GetRedundancyType(),
		"spID", util.SpReadable(syncerInfo.GetStorageProviderId()), "length", len(pieceData))
	stream, err := node.syncer.SyncPiece(ctx)
	if err != nil {
		log.Errorw("sync secondary piece job error", "err", err)
		return nil, err
	}

	// send data one by one to avoid exceeding rpc max msg size
	for key, value := range pieceData {
		innerMap := make(map[string][]byte)
		innerMap[key] = value
		if err := stream.Send(&service.SyncerServiceSyncPieceRequest{
			TraceId:    traceID,
			SyncerInfo: syncerInfo,
			PieceData:  innerMap,
		}); err != nil {
			log.Errorw("client send request error", "error", err)
			return nil, err
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Errorw("client close error", "error", err, "traceID", resp.GetTraceId())
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.Errorw("sync piece sends to stone node response code is not success", "error", err, "traceID", resp.GetTraceId())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

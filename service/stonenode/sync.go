package stonenode

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"io"
	"sort"
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
	// REPLICA_TYPE and INLINE_TYPE need segments count + backup
	log.Info("20")
	secondarySPs := mock.AllocUploadSecondarySP()

	// 1. load all segments data from primary piece store and do ec or not
	log.Info("21")
	pieceData, err := node.loadSegmentsData(ctx, allocResp)
	if err != nil {
		node.reportErrToStoneHub(ctx, allocResp, err)
		return err
	}

	// 2. dispatch the piece data to different secondary
	redundancyType := allocResp.GetPieceJob().GetRedundancyType()
	targetIdx := allocResp.GetPieceJob().GetTargetIdx()
	secondaryPieceData, err := node.dispatchSecondarySP(pieceData, redundancyType, secondarySPs, targetIdx)
	if err != nil {
		node.reportErrToStoneHub(ctx, allocResp, err)
		return err
	}

	// 3. send piece data to the secondary
	node.doSyncToSecondarySP(ctx, allocResp, secondaryPieceData)
	return nil
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

	log.Info("22")
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

	log.Info("23")
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

	log.Info("24")
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
		log.Info("generatePieceData length: ", len(pieceData))
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
// returned map key is spID, value map key is piece key or segment key, value map's value is piece data
func (node *StoneNodeService) dispatchSecondarySP(pieceDataBySegment map[string][][]byte, redundancyType ptypes.RedundancyType,
	secondarySPs []string, targetIdx []uint32) (map[string]map[string][]byte, error) {
	pieceDataBySecondary := make(map[string]map[string][]byte)

	// pieceDataBySegment key is segment key, value is ec data from ec1 to ec6
	var err error
	switch redundancyType {
	case ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED:
		pieceDataBySecondary, err = fillECData(pieceDataBySegment, secondarySPs, targetIdx)
		if err != nil {
			return map[string]map[string][]byte{}, err
		}
		return pieceDataBySecondary, nil
	case ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE, ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
		pieceDataBySecondary, err = fillReplicaOrInlineData(pieceDataBySegment, secondarySPs, targetIdx)
		if err != nil {
			return map[string]map[string][]byte{}, err
		}
		return pieceDataBySecondary, nil
	default:
		return map[string]map[string][]byte{}, merrors.ErrRedundancyType
	}
}

func fillECData(pieceDataBySegment map[string][][]byte, secondarySPs []string, targetIdx []uint32) (
	map[string]map[string][]byte, error) {
	ecPieceDataMap := make(map[string]map[string][]byte)

	// iterate map in order
	//keys := util.SortedKeys(pieceDataBySegment)
	keys := sortedKeys(pieceDataBySegment)
	for _, pieceKey := range keys {
		pieceData := pieceDataBySegment[pieceKey]
		for idx, data := range pieceData {
			if idx >= len(secondarySPs) {
				return map[string]map[string][]byte{}, merrors.ErrSecondarySPNumber
			}
			// initialize data map
			sp := secondarySPs[idx]
			if len(targetIdx) == 0 {
				if _, ok := ecPieceDataMap[sp]; !ok {
					ecPieceDataMap[sp] = make(map[string][]byte)
				}
			} else {
				for _, j := range targetIdx {
					if int(j-1) == idx {
						if _, ok := ecPieceDataMap[sp]; !ok {
							ecPieceDataMap[sp] = make(map[string][]byte)
						}
					}
				}
			}

			var key string
			// if targetIdx is not equal to zero, retry to get data which idx is equal to targetIdx
			if len(targetIdx) != 0 {
				for _, j := range targetIdx {
					if int(j-1) == idx {
						key = piecestore.EncodeECPieceKeyBySegmentKey(pieceKey, idx)
						ecPieceDataMap[sp][key] = data
					}
				}
			} else {
				key = piecestore.EncodeECPieceKeyBySegmentKey(pieceKey, idx)
				ecPieceDataMap[sp][key] = data
			}
		}
	}
	return ecPieceDataMap, nil
}

func fillReplicaOrInlineData(pieceDataBySegment map[string][][]byte, secondarySPs []string, targetIdx []uint32) (
	map[string]map[string][]byte, error) {
	replicaOrInlineDataMap := make(map[string]map[string][]byte)
	if len(pieceDataBySegment) >= len(secondarySPs) {
		return map[string]map[string][]byte{}, merrors.ErrSecondarySPNumber
	}

	// iterate map in order
	keys := sortedKeys(pieceDataBySegment)
	for i := 0; i < len(keys); i++ {
		pieceKey := keys[i]
		pieceData := pieceDataBySegment[pieceKey]
		if len(pieceData) != 1 {
			return nil, merrors.ErrInvalidSegmentData
		}

		sp := secondarySPs[i]
		if len(targetIdx) == 0 {
			if _, ok := replicaOrInlineDataMap[sp]; !ok {
				replicaOrInlineDataMap[sp] = make(map[string][]byte)
			}
		} else {
			for _, index := range targetIdx {
				if int(index) == i {
					if _, ok := replicaOrInlineDataMap[sp]; !ok {
						replicaOrInlineDataMap[sp] = make(map[string][]byte)
					}
				}
			}
		}

		var key string
		if len(targetIdx) != 0 {
			for _, index := range targetIdx {
				if int(index) == i {
					key = pieceKey
					replicaOrInlineDataMap[sp][key] = pieceData[0]
				}
			}
		} else {
			key = pieceKey
			replicaOrInlineDataMap[sp][key] = pieceData[0]
		}
	}
	return replicaOrInlineDataMap, nil
}

func sortedKeys(dataMap map[string][][]byte) []string {
	keys := make([]string, 0, len(dataMap))
	for k := range dataMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
				BucketName:     resp.GetPieceJob().GetBucketName(),
				ObjectName:     resp.GetPieceJob().GetObjectName(),
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
				log.CtxInfow(ctx, "upload secondary piece job secondary", "secondary sp", secondary)
			}()

			syncResp, err := node.UploadECPiece(ctx, &service.SyncerInfo{
				ObjectId:          objectID,
				TxHash:            txHash,
				StorageProviderId: secondary,
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
			//for _, data := range pieceData {
			//	pieceHash = append(pieceHash, hash.GenerateChecksum(data))
			//}
			var pieceIndex uint32
			keys := util.SortedKeys(pieceData)
			log.Infow("sorted keys", "keys", keys)
			for _, key := range keys {
				pieceHash = append(pieceHash, hash.GenerateChecksum(pieceData[key]))
				_, _, pieceIndex, _ = piecestore.DecodeECPieceKey(key)
			}
			integrityHash := hash.GenerateIntegrityHash(pieceHash)
			log.Infow("compute locally", "integrityHash", hex.EncodeToString(integrityHash)[:10], "secondary", secondary,
				"pieceIndex", pieceIndex)

			log.CtxInfow(ctx, "syncResp", "GetIntegrityHash", syncResp.GetSecondarySpInfo().GetIntegrityHash(),
				"secondary", syncResp.GetSecondarySpInfo().GetStorageProviderId())

			if syncResp.GetSecondarySpInfo() == nil ||
				syncResp.GetSecondarySpInfo().GetIntegrityHash() == nil ||
				!bytes.Equal(integrityHash, syncResp.GetSecondarySpInfo().GetIntegrityHash()) {
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

// UploadECPiece send rpc request to secondary storage provider to sync the piece data.
func (node *StoneNodeService) UploadECPiece(ctx context.Context, sInfo *service.SyncerInfo,
	pieceData map[string][]byte, traceID string) (*service.SyncerServiceUploadECPieceResponse, error) {
	log.CtxInfow(ctx, "stone node UploadECPiece", "rType", sInfo.GetRedundancyType(),
		"spID", sInfo.GetStorageProviderId(), "traceID", traceID, "length", len(pieceData))
	stream, err := node.syncer.UploadECPiece(ctx)
	if err != nil {
		log.Errorw("upload secondary job piece job error", "err", err)
		return nil, err
	}

	//for k, v := range pieceData {
	//	log.Infow("pieceData", "key", k, "length", len(pieceData))
	//	innerMap := make(map[string][]byte)
	//	innerMap[k] = v
	//	if err := stream.Send(&service.SyncerServiceUploadECPieceRequest{
	//		TraceId:    traceID,
	//		SyncerInfo: sInfo,
	//		PieceData:  innerMap,
	//	}); err != nil {
	//		log.Errorw("client send request error", "error", err)
	//		return nil, err
	//	}
	//}

	if err := stream.Send(&service.SyncerServiceUploadECPieceRequest{
		TraceId:    traceID,
		SyncerInfo: sInfo,
		PieceData:  pieceData,
	}); err != nil {
		log.Errorw("client send request error", "error", err)
		return nil, err
	}

	resp, err := stream.CloseAndRecv()
	if err != nil && err != io.EOF {
		log.Errorw("client close error", "error", err, "traceID", resp.GetTraceId())
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != service.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.Errorw("upload ec piece sends to stone node response code is not success", "error", err, "traceID", resp.GetTraceId())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	log.Infof("traceID: %s", resp.GetTraceId())
	log.CtxInfow(ctx, "UploadECPiece", "spInfo", resp.GetSecondarySpInfo(), "GetIntegrityHash",
		resp.GetSecondarySpInfo().GetIntegrityHash(), "traceID", resp.GetTraceId())
	return resp, nil
}

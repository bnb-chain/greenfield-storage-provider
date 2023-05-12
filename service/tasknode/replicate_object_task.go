package tasknode

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"sync"
	"sync/atomic"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-common/go/redundancy"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	gatewayclient "github.com/bnb-chain/greenfield-storage-provider/service/gateway/client"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	p2ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/types"
	"github.com/bnb-chain/greenfield-storage-provider/util/maps"
)

// PieceDataReader defines [][]pieceData Reader.
type PieceDataReader struct {
	pieceData [][]byte
	outerIdx  int
	innerIdx  int
}

// Read populates the given byte slice with data and returns the number of bytes populated and an error value.
// It returns an io.EOF error when the stream ends.
func (p *PieceDataReader) Read(buf []byte) (n int, err error) {
	if len(buf) == 0 {
		return 0, fmt.Errorf("failed to read due to invalid args")
	}

	readLen := 0
	for p.outerIdx < len(p.pieceData) {
		curReadLen := copy(buf[readLen:], p.pieceData[p.outerIdx][p.innerIdx:])
		p.innerIdx += curReadLen
		if p.innerIdx == len(p.pieceData[p.outerIdx]) {
			p.outerIdx += 1
			p.innerIdx = 0
		}
		readLen = readLen + curReadLen
		if readLen == len(buf) {
			break
		}
	}
	if readLen != 0 {
		return readLen, nil
	}
	return 0, io.EOF
}

// pieceDataReplicator replicates a piece data to the target sp.
type pieceDataReplicator struct {
	task                  *replicateObjectTask
	redundancyIndex       uint32
	expectedIntegrityHash []byte
	pieceDataReader       *PieceDataReader
	sp                    *sptypes.StorageProvider
	approval              *p2ptypes.GetApprovalResponse
}

// replicate is used to start replicate the piece stream
func (r *pieceDataReplicator) replicate() (integrityHash []byte, signature []byte, err error) {
	var (
		gwClient      *gatewayclient.GatewayClient
		originMsgHash []byte
		approvalAddr  sdk.AccAddress
	)

	gwClient, err = gatewayclient.NewGatewayClient(r.sp.GetEndpoint())
	if err != nil {
		log.Errorw("failed to create gateway client",
			"sp_endpoint", r.sp.GetEndpoint(), "error", err)
		return
	}
	integrityHash, signature, err = gwClient.ReplicateObjectPieceStream(r.task.objectInfo.Id.Uint64(),
		uint32(r.task.pieceSize), r.task.replicateDataSize, r.redundancyIndex, r.approval, r.pieceDataReader)
	if err != nil {
		log.Errorw("failed to replicate object piece stream",
			"endpoint", r.sp.GetEndpoint(), "error", err)
		return
	}
	if !bytes.Equal(r.expectedIntegrityHash, integrityHash) {
		err = merrors.ErrMismatchIntegrityHash
		log.Errorw("failed to check root hash",
			"expected", hex.EncodeToString(r.expectedIntegrityHash),
			"actual", hex.EncodeToString(integrityHash), "error", err)
		return
	}

	// verify secondary signature
	originMsgHash = storagetypes.NewSecondarySpSignDoc(r.sp.GetOperator(), sdkmath.NewUint(r.task.objectInfo.Id.Uint64()), integrityHash).GetSignBytes()
	approvalAddr, err = sdk.AccAddressFromHexUnsafe(r.sp.GetApprovalAddress())
	if err != nil {
		log.Errorw("failed to parse sp operator address",
			"sp", r.sp.GetApprovalAddress(), "endpoint", r.sp.GetEndpoint(),
			"error", err)
		return
	}
	err = storagetypes.VerifySignature(approvalAddr, sdk.Keccak256(originMsgHash), signature)
	if err != nil {
		log.Errorw("failed to verify sp signature",
			"sp", r.sp.GetApprovalAddress(), "endpoint", r.sp.GetEndpoint(), "error", err)
		return
	}

	return integrityHash, signature, nil
}

// replicateObjectTask represents the background object replicate task, include replica/ec redundancy type.
// The task loads all segment piece data to memory and then replicates, thus consumes more memory than stream-task.
type replicateObjectTask struct {
	ctx                 context.Context
	taskNode            *TaskNode
	objectInfo          *storagetypes.ObjectInfo
	approximateMemSize  int
	storageParams       *storagetypes.Params
	segmentPieceNumber  int
	redundancyNumber    int
	replicateDataSize   int64
	pieceSize           int64
	mux                 sync.Mutex
	spMap               map[string]*sptypes.StorageProvider
	approvalResponseMap map[string]*p2ptypes.GetApprovalResponse
	sortedSpEndpoints   []string
}

// newReplicateObjectTask returns a ReplicateObjectTask instance.
func newReplicateObjectTask(ctx context.Context, task *TaskNode, object *storagetypes.ObjectInfo) (*replicateObjectTask, error) {
	if ctx == nil || task == nil || object == nil {
		return nil, merrors.ErrInvalidParams
	}
	return &replicateObjectTask{
		ctx:                 ctx,
		taskNode:            task,
		objectInfo:          object,
		spMap:               make(map[string]*sptypes.StorageProvider),
		approvalResponseMap: make(map[string]*p2ptypes.GetApprovalResponse),
	}, nil
}

// updateTaskState is used to update task state.
func (t *replicateObjectTask) updateTaskState(state servicetypes.JobState) error {
	return t.taskNode.spDB.UpdateJobState(t.objectInfo.Id.Uint64(), state)
}

// init is used to synchronize the resources which is needed to initialize the task.
func (t *replicateObjectTask) init() error {
	var err error
	t.storageParams, err = t.taskNode.spDB.GetStorageParams()
	if err != nil {
		log.CtxErrorw(t.ctx, "failed to query sp params", "error", err)
		return err
	}
	t.segmentPieceNumber = int(piecestore.ComputeSegmentCount(t.objectInfo.GetPayloadSize(),
		t.storageParams.VersionedParams.GetMaxSegmentSize()))
	t.redundancyNumber = int(t.storageParams.VersionedParams.GetRedundantDataChunkNum() + t.storageParams.VersionedParams.GetRedundantParityChunkNum())
	if t.redundancyNumber+1 != len(t.objectInfo.GetChecksums()) {
		log.CtxError(t.ctx, "failed to init due to redundancy number is not equal to checksums")
		return merrors.ErrInvalidParams
	}
	t.spMap, t.approvalResponseMap, err = t.taskNode.getApproval(
		t.objectInfo, t.redundancyNumber, t.redundancyNumber*ReplicateFactor, GetApprovalTimeout)
	if err != nil {
		log.CtxErrorw(t.ctx, "failed to get approvals", "error", err)
		return err
	}
	t.sortedSpEndpoints = maps.SortKeys(t.approvalResponseMap)

	// calculate the reserve memory, which is used in execute time
	if t.objectInfo.GetRedundancyType() == storagetypes.REDUNDANCY_REPLICA_TYPE {
		t.approximateMemSize = (int(t.storageParams.VersionedParams.GetRedundantDataChunkNum()) +
			int(t.storageParams.VersionedParams.GetRedundantParityChunkNum())) *
			int(t.objectInfo.GetPayloadSize())
	} else {
		t.approximateMemSize = int(math.Ceil(
			((float64(t.storageParams.VersionedParams.GetRedundantDataChunkNum())+float64(t.storageParams.VersionedParams.GetRedundantParityChunkNum()))/
				float64(t.storageParams.VersionedParams.GetRedundantDataChunkNum()) + 1) *
				float64(t.objectInfo.GetPayloadSize())))
	}

	return nil
}

// execute is used to start the task, and waitCh is used to wait runtime initialization.
func (t *replicateObjectTask) execute(waitCh chan error) {
	var (
		startTime            time.Time
		succeedIndexMapMutex sync.RWMutex
		succeedIndexMap      map[int]bool
		err                  error
		scopeSpan            rcmgr.ResourceScopeSpan
		sealMsg              *storagetypes.MsgSealObject
		progressInfo         *servicetypes.ReplicatePieceInfo
	)

	// runtime initialization
	startTime = time.Now()
	succeedIndexMap = make(map[int]bool, t.redundancyNumber)
	getNeedReplicateNumber := func(inputIndexMap map[int]bool) int {
		var needReplicateNumber int
		for i := 0; i < t.redundancyNumber; i++ {
			if !inputIndexMap[i] {
				needReplicateNumber++
			}
		}
		return needReplicateNumber
	}
	metrics.ReplicateObjectTaskGauge.WithLabelValues(model.TaskNodeService).Inc()
	t.updateTaskState(servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_DOING)

	if scopeSpan, err = t.taskNode.rcScope.BeginSpan(); err != nil {
		log.CtxErrorw(t.ctx, "failed to begin span", "error", err)
		waitCh <- err
		return
	}
	if err = scopeSpan.ReserveMemory(t.approximateMemSize, rcmgr.ReservationPriorityAlways); err != nil {
		log.CtxErrorw(t.ctx, "failed to reserve memory from resource manager",
			"reserve_size", t.approximateMemSize, "error", err)
		waitCh <- err
		return
	}
	log.CtxDebugw(t.ctx, "reserve memory from resource manager",
		"reserve_size", t.approximateMemSize, "resource_state", rcmgr.GetServiceState(model.TaskNodeService))
	replicateData, err := t.encodeReplicateData()
	if err != nil {
		log.CtxErrorw(t.ctx, "failed to encode replicate data", "error", err)
		waitCh <- err
		return
	}
	waitCh <- nil

	// defer func
	defer func() {
		close(waitCh)
		metrics.ReplicateObjectTaskGauge.WithLabelValues(model.TaskNodeService).Dec()
		observer := metrics.SealObjectTimeHistogram.WithLabelValues(model.TaskNodeService)
		observer.Observe(time.Since(startTime).Seconds())

		if getNeedReplicateNumber(succeedIndexMap) == 0 {
			metrics.SealObjectTotalCounter.WithLabelValues("success").Inc()
			log.CtxInfo(t.ctx, "succeed to seal object on chain")
		} else {
			t.updateTaskState(servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_ERROR)
			metrics.SealObjectTotalCounter.WithLabelValues("failure").Inc()
			log.CtxErrorw(t.ctx, "failed to replicate object data to sp", "error", err, "succeed_index_map", succeedIndexMap)
		}

		if scopeSpan != nil {
			scopeSpan.Done()
			log.CtxDebugw(t.ctx, "release memory to resource manager",
				"release_size", t.approximateMemSize, "resource_state", rcmgr.GetServiceState(model.TaskNodeService))
		}
	}()

	// execution
	pickSp := func() (sp *sptypes.StorageProvider, approval *p2ptypes.GetApprovalResponse, err error) {
		t.mux.Lock()
		defer t.mux.Unlock()
		if len(t.approvalResponseMap) == 0 {
			log.CtxError(t.ctx, "backup storage providers exhausted")
			err = merrors.ErrExhaustedSP
			return
		}
		endpoint := t.sortedSpEndpoints[0]
		sp = t.spMap[endpoint]
		approval = t.approvalResponseMap[endpoint]
		t.sortedSpEndpoints = t.sortedSpEndpoints[1:]
		delete(t.spMap, endpoint)
		delete(t.approvalResponseMap, endpoint)
		return
	}
	sealMsg = &storagetypes.MsgSealObject{
		Operator:              t.taskNode.config.SpOperatorAddress,
		BucketName:            t.objectInfo.GetBucketName(),
		ObjectName:            t.objectInfo.GetObjectName(),
		SecondarySpAddresses:  make([]string, t.redundancyNumber),
		SecondarySpSignatures: make([][]byte, t.redundancyNumber),
	}
	t.objectInfo.SecondarySpAddresses = make([]string, t.redundancyNumber)
	progressInfo = &servicetypes.ReplicatePieceInfo{
		PieceInfos: make([]*servicetypes.PieceInfo, t.redundancyNumber),
	}
	for {
		if getNeedReplicateNumber(succeedIndexMap) == 0 {
			log.CtxInfo(t.ctx, "succeed to replicate object data")
			break
		}
		if getNeedReplicateNumber(succeedIndexMap) > len(t.sortedSpEndpoints) {
			log.CtxError(t.ctx, "failed to replicate due to sp is not enough")
			err = merrors.ErrExhaustedSP
			return
		}

		var wg sync.WaitGroup
		for rIdx := 0; rIdx < t.redundancyNumber; rIdx++ {
			succeedIndexMapMutex.Lock()
			_, hasReplicated := succeedIndexMap[rIdx]
			succeedIndexMapMutex.Unlock()

			if !hasReplicated {
				wg.Add(1)
				log.CtxDebugw(t.ctx, "start to replicate object", "redundancy_index", rIdx)
				go func(rIdx int) {
					defer wg.Done()

					sp, approval, innerErr := pickSp()
					if innerErr != nil {
						log.CtxErrorw(t.ctx, "failed to pick a secondary sp", "redundancy_index", rIdx, "error", innerErr)
						return
					}
					var toReplicateData [][]byte
					for idx := 0; idx < t.segmentPieceNumber; idx++ {
						toReplicateData = append(toReplicateData, replicateData[idx][rIdx])
					}
					r := &pieceDataReplicator{
						task:                  t,
						redundancyIndex:       uint32(rIdx),
						expectedIntegrityHash: t.objectInfo.GetChecksums()[rIdx+1],
						pieceDataReader: &PieceDataReader{
							pieceData: toReplicateData,
						},
						sp:       sp,
						approval: approval,
					}
					integrityHash, signature, innerErr := r.replicate()
					if innerErr != nil {
						log.CtxErrorw(t.ctx, "failed to replicate piece stream", "redundancy_index", rIdx, "error", innerErr)
						return
					}

					succeedIndexMapMutex.Lock()
					succeedIndexMap[rIdx] = true
					succeedIndexMapMutex.Unlock()

					sealMsg.GetSecondarySpAddresses()[rIdx] = sp.GetOperator().String()
					sealMsg.GetSecondarySpSignatures()[rIdx] = signature
					progressInfo.PieceInfos[rIdx] = &servicetypes.PieceInfo{
						ObjectInfo:    t.objectInfo,
						Signature:     signature,
						IntegrityHash: integrityHash,
					}
					t.objectInfo.SecondarySpAddresses[rIdx] = sp.GetOperator().String()
					t.taskNode.spDB.SetObjectInfo(t.objectInfo.Id.Uint64(), t.objectInfo)
					t.taskNode.cache.Add(t.objectInfo.Id.Uint64(), progressInfo)
					log.CtxInfow(t.ctx, "succeed to replicate object piece stream to the target sp",
						"sp", sp.GetOperator(), "endpoint", sp.GetEndpoint(), "redundancy_index", rIdx)

				}(rIdx)
			}

		}
		wg.Wait()
	}

	// seal onto the greenfield chain
	if getNeedReplicateNumber(succeedIndexMap) == 0 { // succeed
		retry := 0
		for {
			if retry >= MaxSealRetryNumber {
				log.CtxErrorw(t.ctx, "failed to seal object", "error", err, "retry", retry)
				break
			}
			retry++
			t.updateTaskState(servicetypes.JobState_JOB_STATE_SIGN_OBJECT_DOING)
			_, err = t.taskNode.signer.SealObjectOnChain(context.Background(), sealMsg)
			if err != nil {
				t.updateTaskState(servicetypes.JobState_JOB_STATE_SIGN_OBJECT_ERROR)
				log.CtxErrorw(t.ctx, "failed to sign object by signer", "error", err, "retry", retry)
				continue
			}
			t.updateTaskState(servicetypes.JobState_JOB_STATE_SEAL_OBJECT_DOING)
			err = t.taskNode.chain.ListenObjectSeal(context.Background(), t.objectInfo.GetBucketName(),
				t.objectInfo.GetObjectName(), 10)
			if err != nil {
				t.updateTaskState(servicetypes.JobState_JOB_STATE_SEAL_OBJECT_ERROR)
				log.CtxErrorw(t.ctx, "failed to seal object on chain", "error", err, "retry", retry)
				continue
			}
			t.updateTaskState(servicetypes.JobState_JOB_STATE_SEAL_OBJECT_DONE)
			break
		}
	} // the else failed case is in defer func
}

// encodeReplicateData load segment data and encode according to redundancy type.
func (t *replicateObjectTask) encodeReplicateData() (data [][][]byte, err error) {
	var (
		wg            sync.WaitGroup
		succeedNumber int64
	)
	for i := 0; i < t.segmentPieceNumber; i++ {
		data = append(data, make([][]byte, t.redundancyNumber))
	}

	log.CtxDebugw(t.ctx, "start to encode replicate data", "segment_number", t.segmentPieceNumber,
		"redundancy_number", t.redundancyNumber)
	for segIdx := 0; segIdx < t.segmentPieceNumber; segIdx++ {
		wg.Add(1)
		go func(segIdx int) {
			defer wg.Done()
			key := piecestore.EncodeSegmentPieceKey(t.objectInfo.Id.Uint64(), uint32(segIdx))
			segmentPieceData, innerErr := t.taskNode.pieceStore.GetPiece(t.ctx, key, 0, 0)
			if innerErr != nil {
				log.CtxErrorw(t.ctx, "failed to get segment piece", "key", key)
				return
			}
			if t.objectInfo.GetRedundancyType() == storagetypes.REDUNDANCY_EC_TYPE {
				encodeData, innerErr := redundancy.EncodeRawSegment(segmentPieceData,
					int(t.storageParams.VersionedParams.GetRedundantDataChunkNum()),
					int(t.storageParams.VersionedParams.GetRedundantParityChunkNum()))
				if innerErr != nil {
					log.CtxErrorw(t.ctx, "failed to encode ec data", "key", key)
					return
				}
				copy(data[segIdx], encodeData)
				t.replicateDataSize += int64(len(encodeData[0]))

			} else {
				for rIdx := 0; rIdx < t.redundancyNumber; rIdx++ {
					data[segIdx][rIdx] = segmentPieceData
				}
				t.replicateDataSize += int64(len(segmentPieceData))
			}
			atomic.AddInt64(&succeedNumber, 1)
			log.CtxDebugw(t.ctx, "succeed to encode payload", "key", key)
		}(segIdx)
	}
	wg.Wait()
	if int(succeedNumber) != t.segmentPieceNumber {
		return data, fmt.Errorf("failed to encode replicate data")
	}
	t.pieceSize = int64(len(data[0][0]))
	log.CtxDebugw(t.ctx, "finish to encode replicate data",
		"redundancy_number", t.redundancyNumber,
		"segment_number", t.segmentPieceNumber,
		"object_size", t.objectInfo.GetPayloadSize(),
		"replicate_size", t.replicateDataSize,
		"piece_size", t.pieceSize,
	)
	return data, nil
}

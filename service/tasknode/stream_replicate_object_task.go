package tasknode

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"io"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-common/go/redundancy"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	p2ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	gatewayclient "github.com/bnb-chain/greenfield-storage-provider/service/gateway/client"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	"github.com/bnb-chain/greenfield-storage-provider/util/maps"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ReplicateFactor defines the redundancy of replication
	// TODO:: will update to （1, 2] on main net
	ReplicateFactor = 1
	// GetApprovalTimeout defines the timeout of getting secondary sp approval
	GetApprovalTimeout = 10
	// MaxSealRetryNumber defines max number of retrying seal object
	MaxSealRetryNumber = 3
)

// streamReader is used to stream produce/consume piece data stream.
type streamReader struct {
	pRead  *io.PipeReader
	pWrite *io.PipeWriter
}

// Read populates the given byte slice with data and returns the number of bytes populated and an error value.
// It returns an io.EOF error when the stream ends.
func (s *streamReader) Read(buf []byte) (n int, err error) {
	return s.pRead.Read(buf)
}

// streamReaderGroup is used to the primary sp replicate object piece data to other secondary sps.
type streamReaderGroup struct {
	task            *streamReplicateObjectTask
	pieceSize       int
	streamReaderMap map[int]*streamReader
}

// newStreamReaderGroup returns a streamReaderGroup
func newStreamReaderGroup(t *streamReplicateObjectTask, excludeIndexMap map[int]bool) (*streamReaderGroup, error) {
	var atLeastHasOne bool
	sg := &streamReaderGroup{
		task:            t,
		streamReaderMap: make(map[int]*streamReader),
	}
	for segmentPieceIdx := 0; segmentPieceIdx < t.segmentPieceNumber; segmentPieceIdx++ {
		for idx := 0; idx < t.redundancyNumber; idx++ {
			if excludeIndexMap[idx] {
				continue
			}
			sg.streamReaderMap[idx] = &streamReader{}
			sg.streamReaderMap[idx].pRead, sg.streamReaderMap[idx].pWrite = io.Pipe()
			atLeastHasOne = true
		}
	}
	if !atLeastHasOne {
		return nil, merrors.ErrInvalidParams
	}
	return sg, nil
}

// produceStreamPieceData produce stream piece data
func (sg *streamReaderGroup) produceStreamPieceData() {
	ch := make(chan int)
	go func(pieceSizeCh chan int) {
		defer close(pieceSizeCh)
		gotPieceSize := false

		for segmentPieceIdx := 0; segmentPieceIdx < sg.task.segmentPieceNumber; segmentPieceIdx++ {
			segmentPieceKey := piecestore.EncodeSegmentPieceKey(sg.task.objectInfo.Id.Uint64(), uint32(segmentPieceIdx))
			segmentPieceData, err := sg.task.taskNode.pieceStore.GetPiece(context.Background(), segmentPieceKey, 0, 0)
			if err != nil {
				for idx := range sg.streamReaderMap {
					sg.streamReaderMap[idx].pWrite.CloseWithError(err)
				}
				log.Errorw("failed to get piece data", "piece_key", segmentPieceKey, "error", err)
				return
			}
			if sg.task.objectInfo.GetRedundancyType() == types.REDUNDANCY_EC_TYPE {
				ecPieceData, err := redundancy.EncodeRawSegment(segmentPieceData,
					int(sg.task.storageParams.GetRedundantDataChunkNum()),
					int(sg.task.storageParams.GetRedundantParityChunkNum()))
				if err != nil {
					for idx := range sg.streamReaderMap {
						sg.streamReaderMap[idx].pWrite.CloseWithError(err)
					}
					log.Errorw("failed to encode ec piece data", "error", err)
					return
				}
				if !gotPieceSize {
					pieceSizeCh <- len(ecPieceData[0])
					gotPieceSize = true
				}
				for idx := range sg.streamReaderMap {
					sg.streamReaderMap[idx].pWrite.Write(ecPieceData[idx])
					log.Debugw("succeed to produce an ec piece data", "piece_len", len(ecPieceData[idx]), "redundancy_index", idx)
				}
			} else {
				if !gotPieceSize {
					pieceSizeCh <- len(segmentPieceData)
					gotPieceSize = true
				}
				for idx := range sg.streamReaderMap {
					sg.streamReaderMap[idx].pWrite.Write(segmentPieceData)
					log.Debugw("succeed to produce an segment piece data", "piece_len", len(segmentPieceData), "redundancy_index", idx)
				}
			}
		}
		for idx := range sg.streamReaderMap {
			sg.streamReaderMap[idx].pWrite.Close()
			log.Debugw("succeed to finish a piece stream",
				"redundancy_index", idx, "redundancy_type", sg.task.objectInfo.GetRedundancyType())
		}
	}(ch)
	sg.pieceSize = <-ch
}

// streamPieceDataReplicator replicates a piece stream to the target sp
type streamPieceDataReplicator struct {
	task                  *streamReplicateObjectTask
	dataSize              int64
	pieceSize             uint32
	redundancyIndex       uint32
	expectedIntegrityHash []byte
	streamReader          *streamReader
	sp                    *sptypes.StorageProvider
	approval              *p2ptypes.GetApprovalResponse
}

// replicate is used to start replicate the piece stream
func (r *streamPieceDataReplicator) replicate() (integrityHash []byte, signature []byte, err error) {
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
		r.pieceSize, r.dataSize, r.redundancyIndex, r.approval, r.streamReader)
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

// streamReplicateObjectTask represents the background object replicate task, include replica/ec redundancy type.
// TODO: Streaming takes up less memory, but in complex network conditions, the transmission is slower.
// The specific reason needs to be further located and optimized in the future.
type streamReplicateObjectTask struct {
	ctx                 context.Context
	taskNode            *TaskNode
	objectInfo          *types.ObjectInfo
	approximateMemSize  int
	storageParams       *storagetypes.Params
	segmentPieceNumber  int
	redundancyNumber    int
	replicateDataSize   int64
	mux                 sync.Mutex
	spMap               map[string]*sptypes.StorageProvider
	approvalResponseMap map[string]*p2ptypes.GetApprovalResponse
	sortedSpEndpoints   []string
}

// newStreamReplicateObjectTask returns a ReplicateObjectTask instance.
func newStreamReplicateObjectTask(ctx context.Context, task *TaskNode, object *types.ObjectInfo) (*streamReplicateObjectTask, error) {
	if ctx == nil || task == nil || object == nil {
		return nil, merrors.ErrInvalidParams
	}
	return &streamReplicateObjectTask{
		ctx:                 ctx,
		taskNode:            task,
		objectInfo:          object,
		spMap:               make(map[string]*sptypes.StorageProvider),
		approvalResponseMap: make(map[string]*p2ptypes.GetApprovalResponse),
	}, nil
}

// updateTaskState is used to update task state.
func (t *streamReplicateObjectTask) updateTaskState(state servicetypes.JobState) error {
	return t.taskNode.spDB.UpdateJobState(t.objectInfo.Id.Uint64(), state)
}

// init is used to synchronize the resources which is needed to initialize the task.
func (t *streamReplicateObjectTask) init() error {
	var err error
	t.storageParams, err = t.taskNode.spDB.GetStorageParams()
	if err != nil {
		log.CtxErrorw(t.ctx, "failed to query sp params", "error", err)
		return err
	}
	t.segmentPieceNumber = int(piecestore.ComputeSegmentCount(t.objectInfo.GetPayloadSize(),
		t.storageParams.GetMaxSegmentSize()))
	t.redundancyNumber = int(t.storageParams.GetRedundantDataChunkNum() + t.storageParams.GetRedundantParityChunkNum())
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
	t.approximateMemSize = int(float64(t.storageParams.GetMaxSegmentSize()) *
		(float64(t.redundancyNumber)/float64(t.storageParams.GetRedundantDataChunkNum()) + 1))
	if t.objectInfo.GetPayloadSize() < t.storageParams.GetMaxSegmentSize() {
		t.approximateMemSize = int(float64(t.objectInfo.GetPayloadSize()) *
			(float64(t.redundancyNumber)/float64(t.storageParams.GetRedundantDataChunkNum()) + 1))
	}
	err = t.ComputeReplicatePiecesSizeForOneSP()
	if err != nil {
		log.CtxErrorw(t.ctx, "failed to compute replicate pieces data size per sp", "error", err)
		return err
	}
	return nil
}

// execute is used to start the task, and waitCh is used to wait runtime initialization.
func (t *streamReplicateObjectTask) execute(waitCh chan error) {
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
	isAllSucceed := func(inputIndexMap map[int]bool) bool {
		for i := 0; i < t.redundancyNumber; i++ {
			if !inputIndexMap[i] {
				return false
			}
		}
		return true
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
	waitCh <- nil

	// defer func
	defer func() {
		close(waitCh)
		metrics.ReplicateObjectTaskGauge.WithLabelValues(model.TaskNodeService).Dec()
		observer := metrics.SealObjectTimeHistogram.WithLabelValues(model.TaskNodeService)
		observer.Observe(time.Since(startTime).Seconds())

		if isAllSucceed(succeedIndexMap) {
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
		if isAllSucceed(succeedIndexMap) {
			log.CtxInfo(t.ctx, "succeed to replicate object data")
			break
		}
		var sg *streamReaderGroup
		sg, err = newStreamReaderGroup(t, succeedIndexMap)
		if err != nil {
			log.CtxErrorw(t.ctx, "failed to new stream reader group", "error", err)
			return
		}
		if len(sg.streamReaderMap) > len(t.sortedSpEndpoints) {
			log.CtxError(t.ctx, "failed to replicate due to sp is not enough")
			err = merrors.ErrExhaustedSP
			return
		}
		sg.produceStreamPieceData()

		var wg sync.WaitGroup
		for redundancyIdx := range sg.streamReaderMap {
			wg.Add(1)
			go func(rIdx int) {
				defer wg.Done()

				sp, approval, innerErr := pickSp()
				if innerErr != nil {
					log.CtxErrorw(t.ctx, "failed to pick a secondary sp", "redundancy_index", rIdx, "error", innerErr)
					return
				}
				r := &streamPieceDataReplicator{
					task:                  t,
					dataSize:              t.replicateDataSize,
					pieceSize:             uint32(sg.pieceSize),
					redundancyIndex:       uint32(rIdx),
					expectedIntegrityHash: sg.task.objectInfo.GetChecksums()[rIdx+1],
					streamReader:          sg.streamReaderMap[rIdx],
					sp:                    sp,
					approval:              approval,
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
			}(redundancyIdx)
		}
		wg.Wait()
	}

	// seal onto the greenfield chain
	if isAllSucceed(succeedIndexMap) {
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

func (t *streamReplicateObjectTask) ComputeReplicatePiecesSizeForOneSP() error {
	if t.segmentPieceNumber <= 0 {
		return errors.New("segment piece number invalid")
	}

	pieceSize := func(idx int) (int, error) {
		segmentPieceKey := piecestore.EncodeSegmentPieceKey(t.objectInfo.Id.Uint64(), uint32(idx))
		segmentPieceData, err := t.taskNode.pieceStore.GetPiece(context.Background(), segmentPieceKey, 0, 0)
		if err != nil {
			return 0, err
		}
		if t.objectInfo.GetRedundancyType() == types.REDUNDANCY_EC_TYPE {
			ecPieceData, err := redundancy.EncodeRawSegment(segmentPieceData,
				int(t.storageParams.GetRedundantDataChunkNum()),
				int(t.storageParams.GetRedundantParityChunkNum()))
			if err != nil {
				return 0, err
			}
			return len(ecPieceData[0]), nil
		} else {
			return len(segmentPieceData), nil
		}
	}
	tailIdx := t.segmentPieceNumber - 1
	tailSize, err := pieceSize(tailIdx)
	if err != nil {
		return err
	}
	integralSize := 0
	if t.segmentPieceNumber > 1 {
		size, err := pieceSize(1)
		if err != nil {
			return err
		}
		integralSize = size * (t.segmentPieceNumber - 1)
	}
	t.replicateDataSize = int64(tailSize + integralSize)
	return nil
}

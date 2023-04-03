package tasknode

import (
	"context"
	"errors"
	"io"
	"sync"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-common/go/redundancy"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	p2ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/types"
	gatewayclient "github.com/bnb-chain/greenfield-storage-provider/service/gateway/client"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	"github.com/bnb-chain/greenfield-storage-provider/util/maps"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// streamReader is used to the specified redundancy index piece stream.
type streamReader struct {
	redundancyIndex int
	pRead           *io.PipeReader
	pWrite          *io.PipeWriter
}

// Read populates the given byte slice with data and returns the number of bytes populated and an error value.
// It returns an io.EOF error when the stream ends.
func (s *streamReader) Read(buf []byte) (n int, err error) {
	return s.pRead.Read(buf)
}

// streamReaderGroup is used to the primary sp replicate object data to other secondary sps.
type streamReaderGroup struct {
	task            *replicateObjectTask
	pieceSize       int
	streamReaderMap map[int]*streamReader
}

// newStreamReaderGroup returns a streamReaderGroup
func newStreamReaderGroup(t *replicateObjectTask, excludeIndexMap map[int]bool) (*streamReaderGroup, error) {
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
			sg.streamReaderMap[idx] = &streamReader{
				redundancyIndex: idx,
			}
			sg.streamReaderMap[idx].pRead, sg.streamReaderMap[idx].pWrite = io.Pipe()
			atLeastHasOne = true
		}
	}
	if !atLeastHasOne {
		return nil, merrors.ErrInvalidParams
	}
	return sg, nil
}

func (sg *streamReaderGroup) streamProducePieceData() {
	ch := make(chan int)
	go func(pieceSizeCh chan int) {
		defer close(pieceSizeCh)

		for segmentPieceIdx := 0; segmentPieceIdx < sg.task.segmentPieceNumber; segmentPieceIdx++ {
			segmentPiecekey := piecestore.EncodeSegmentPieceKey(sg.task.objectInfo.Id.Uint64(), uint32(segmentPieceIdx))
			segmentPieceData, err := sg.task.taskNode.pieceStore.GetPiece(context.Background(), segmentPiecekey, 0, 0)
			if err != nil {
				for idx := range sg.streamReaderMap {
					sg.streamReaderMap[idx].pWrite.CloseWithError(err)
				}
				log.Errorw("failed to get piece data", "piece_key", segmentPiecekey, "error", err)
				return
			}
			// TODO: support replica type
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
			pieceSizeCh <- len(ecPieceData[0])
			for idx := range sg.streamReaderMap {
				sg.streamReaderMap[idx].pWrite.Write(ecPieceData[idx])
				log.Debugw("succeed to produce an ec piece data", "piece_len", len(ecPieceData[idx]), "redundancy_index", idx)
			}
		}
		for idx := range sg.streamReaderMap {
			sg.streamReaderMap[idx].pWrite.Close()
			log.Debugw("succeed to finish an ec piece stream", "redundancy_index", idx)
		}
	}(ch)
	sg.pieceSize = <-ch
}

// replicateObjectTask represents the background object replicate task, include replica/ec redundancy type.
type replicateObjectTask struct {
	taskNode           *TaskNode
	objectInfo         *types.ObjectInfo
	storageParams      *storagetypes.Params
	segmentPieceNumber int
	redundancyNumber   int

	mux                 sync.Mutex
	spMap               map[string]*sptypes.StorageProvider
	approvalResponseMap map[string]*p2ptypes.GetApprovalResponse
	sortedSpEndpoints   []string
}

// newReplicateObjectTask returns a ReplicateObjectTask instance
func newReplicateObjectTask(t *TaskNode, o *types.ObjectInfo) (*replicateObjectTask, error) {
	if t == nil || o == nil {
		return nil, merrors.ErrInvalidParams
	}
	if o.GetRedundancyType() != types.REDUNDANCY_EC_TYPE {
		log.Errorw("failed to new replicate object task due to unsupported redundancy type",
			"redundancy_type", o.GetRedundancyType())
		return nil, merrors.ErrInvalidParams
	}
	return &replicateObjectTask{
		taskNode:            t,
		objectInfo:          o,
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
		log.Errorw("failed to query sp params", "error", err)
		return err
	}
	t.segmentPieceNumber = int(piecestore.ComputeSegmentCount(t.objectInfo.GetPayloadSize(),
		t.storageParams.GetMaxSegmentSize()))
	t.redundancyNumber = int(t.storageParams.GetRedundantDataChunkNum() + t.storageParams.GetRedundantParityChunkNum())
	t.spMap, t.approvalResponseMap, err = t.taskNode.getApproval(
		t.objectInfo, t.redundancyNumber, t.redundancyNumber*ReplicateFactor, GetApprovalTimeout)
	t.sortedSpEndpoints = maps.SortKeys(t.approvalResponseMap)
	if err != nil {
		log.Errorw("failed to get approvals", "error", err)
		return err
	}
	return nil
}

// execute is used to execute the task.
func (t *replicateObjectTask) execute() {
	var (
		sealMsg         *storagetypes.MsgSealObject
		processInfo     *servicetypes.ReplicateSegmentInfo
		succeedIndexMap map[int]bool
	)
	succeedIndexMap = make(map[int]bool, t.redundancyNumber)
	isAllSucceed := func(inputIndexMap map[int]bool) bool {
		for i := 0; i < t.redundancyNumber; i++ {
			if !inputIndexMap[i] {
				return false
			}
		}
		return true
	}
	getSpFunc := func() (sp *sptypes.StorageProvider, approval *p2ptypes.GetApprovalResponse, err error) {
		t.mux.Lock()
		defer t.mux.Unlock()
		if len(t.approvalResponseMap) == 0 {
			log.Error("backup storage providers depleted")
			err = errors.New("no backup sp to pick up")
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
	processInfo = &servicetypes.ReplicateSegmentInfo{
		SegmentInfos: make([]*servicetypes.SegmentInfo, t.redundancyNumber),
	}

	for retry := 0; retry < 2; retry++ {
		if isAllSucceed(succeedIndexMap) {
			log.Infow("succeed to replicate object data")
			break
		}
		sg, err := newStreamReaderGroup(t, succeedIndexMap)
		if err != nil {
			log.Errorw("failed to new stream reader group", "error", err)
			return
		}
		sg.streamProducePieceData()

		var wg sync.WaitGroup
		for redundancyIdx := range sg.streamReaderMap {
			wg.Add(1)
			go func(rIdx int) {
				defer wg.Done()

				sp, approval, innerErr := getSpFunc()
				if innerErr != nil {
					log.Errorw("failed to get secondary sp", "redundancy_index", rIdx, "error", innerErr)
					return
				}
				gatewayClient, innerErr := gatewayclient.NewGatewayClient(sp.GetEndpoint())
				if innerErr != nil {
					log.Errorw("failed to create gateway client",
						"sp_endpoint", sp.GetEndpoint(), "redundancy_index", rIdx, "error", innerErr)
					return
				}
				integrityHash, signature, innerErr := gatewayClient.ReplicateObjectData(t.objectInfo, uint32(sg.pieceSize),
					uint32(rIdx), approval, sg.streamReaderMap[rIdx])
				if innerErr != nil {
					log.Errorw("failed to sync piece data",
						"endpoint", sp.GetEndpoint(), "redundancy_index", rIdx, "error", innerErr)
					return
				}

				msg := storagetypes.NewSecondarySpSignDoc(sp.GetOperator(), sdkmath.NewUint(t.objectInfo.Id.Uint64()), integrityHash).GetSignBytes()
				approvalAddr, innerErr := sdk.AccAddressFromHexUnsafe(sp.GetApprovalAddress())
				if innerErr != nil {
					log.Errorw("failed to parse sp operator address",
						"sp", sp.GetApprovalAddress(), "endpoint", sp.GetEndpoint(),
						"redundancy_index", rIdx, "error", innerErr)
					return
				}
				innerErr = storagetypes.VerifySignature(approvalAddr, sdk.Keccak256(msg), signature)
				if innerErr != nil {
					log.Errorw("failed to verify sp signature",
						"sp", sp.GetApprovalAddress(), "endpoint", sp.GetEndpoint(),
						"redundancy_index", rIdx, "error", innerErr)
					return
				}
				succeedIndexMap[rIdx] = true
				sealMsg.GetSecondarySpAddresses()[rIdx] = sp.GetOperator().String()
				sealMsg.GetSecondarySpSignatures()[rIdx] = signature
				processInfo.SegmentInfos[rIdx] = &servicetypes.SegmentInfo{
					ObjectInfo:    t.objectInfo,
					Signature:     signature,
					IntegrityHash: integrityHash,
				}
				t.objectInfo.SecondarySpAddresses[rIdx] = sp.GetOperator().String()
				t.taskNode.spDB.SetObjectInfo(t.objectInfo.Id.Uint64(), t.objectInfo)
				t.taskNode.cache.Add(t.objectInfo.Id.Uint64(), processInfo)
				log.Infow("succeed to replicate object data to sp",
					"sp", sp.GetOperator(), "endpoint", sp.GetEndpoint(), "redundancy_index", rIdx)

			}(redundancyIdx)
		}
		wg.Wait()
	}

	// seal info
	if isAllSucceed(succeedIndexMap) {
		t.updateTaskState(servicetypes.JobState_JOB_STATE_SIGN_OBJECT_DOING)
		_, err := t.taskNode.signer.SealObjectOnChain(context.Background(), sealMsg)
		if err != nil {
			t.taskNode.spDB.UpdateJobState(t.objectInfo.Id.Uint64(), servicetypes.JobState_JOB_STATE_SIGN_OBJECT_ERROR)
			log.Errorw("failed to sign object by signer", "error", err)
			return
		}
		t.updateTaskState(servicetypes.JobState_JOB_STATE_SEAL_OBJECT_TX_DOING)
		err = t.taskNode.chain.ListenObjectSeal(context.Background(), t.objectInfo.GetBucketName(),
			t.objectInfo.GetObjectName(), 10)
		if err != nil {
			t.updateTaskState(servicetypes.JobState_JOB_STATE_SEAL_OBJECT_ERROR)
			log.Errorw("failed to seal object on chain", "error", err)
			return
		}
		t.updateTaskState(servicetypes.JobState_JOB_STATE_SEAL_OBJECT_DONE)
		log.Info("succeed to seal object on chain")
	} else {
		err := t.updateTaskState(servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_ERROR)
		log.Errorw("failed to replicate object data to sp", "error", err, "succeed_index_map", succeedIndexMap)
		return
	}
}

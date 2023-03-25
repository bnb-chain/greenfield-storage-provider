package tasknode

import (
	"context"
	"errors"
	"math"
	"sync"
	"sync/atomic"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/util/maps"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	p2ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/types"
	gatewayclient "github.com/bnb-chain/greenfield-storage-provider/service/gateway/client"
	"github.com/bnb-chain/greenfield-storage-provider/service/tasknode/types"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
)

const (
	ReplicateFactor    = 2
	GetApprovalTimeout = 10
)

var _ types.TaskNodeServiceServer = &TaskNode{}

// ReplicateObject call AsyncReplicateObject non-blocking upstream services
func (taskNode *TaskNode) ReplicateObject(ctx context.Context, req *types.ReplicateObjectRequest) (
	resp *types.ReplicateObjectResponse, err error) {
	resp = &types.ReplicateObjectResponse{}
	taskNode.spDB.UpdateJobState(req.GetObjectInfo().Id.Uint64(), servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_DOING)
	go taskNode.AsyncReplicateObject(req)
	log.Debugw("receive the replicate object task", "object_id", req.GetObjectInfo().Id)
	return
}

// AsyncReplicateObject replicate an object payload to other storage providers and seal object.
func (taskNode *TaskNode) AsyncReplicateObject(req *types.ReplicateObjectRequest) (err error) {
	ctx := context.Background()
	processInfo := &servicetypes.ReplicateSegmentInfo{}
	sealMsg := &storagetypes.MsgSealObject{}
	objectInfo := req.GetObjectInfo()
	defer func() {
		if err != nil {
			taskNode.spDB.UpdateJobState(objectInfo.Id.Uint64(), servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_ERROR)
			log.CtxErrorw(ctx, "failed to replicate payload data to sp", "error", err)
			return
		}
		taskNode.spDB.UpdateJobState(objectInfo.Id.Uint64(), servicetypes.JobState_JOB_STATE_SIGN_OBJECT_DOING)
		_, err = taskNode.signer.SealObjectOnChain(ctx, sealMsg)
		if err != nil {
			taskNode.spDB.UpdateJobState(objectInfo.Id.Uint64(), servicetypes.JobState_JOB_STATE_SIGN_OBJECT_ERROR)
			log.CtxErrorw(ctx, "failed to sign object by signer", "error", err)
			return
		}
		taskNode.spDB.UpdateJobState(objectInfo.Id.Uint64(), servicetypes.JobState_JOB_STATE_SEAL_OBJECT_TX_DOING)
		err := taskNode.chain.ListenObjectSeal(ctx, objectInfo.GetBucketName(),
			objectInfo.GetObjectName(), 10)
		if err != nil {
			taskNode.spDB.UpdateJobState(objectInfo.Id.Uint64(), servicetypes.JobState_JOB_STATE_SEAL_OBJECT_ERROR)
			log.CtxErrorw(ctx, "failed to seal object on chain", "error", err)
			return
		}
		taskNode.spDB.UpdateJobState(objectInfo.Id.Uint64(), servicetypes.JobState_JOB_STATE_SEAL_OBJECT_DONE)
		log.CtxInfow(ctx, "succeed to seal object on chain")
	}()

	params, err := taskNode.spDB.GetStorageParams()
	if err != nil {
		log.CtxErrorw(ctx, "failed to query sp params", "error", err)
		return
	}
	segments := piecestore.ComputeSegmentCount(objectInfo.GetPayloadSize(),
		params.GetMaxSegmentSize())
	replicates := params.GetRedundantDataChunkNum() + params.GetRedundantParityChunkNum()
	replicateData, err := taskNode.EncodeReplicateSegments(ctx, objectInfo.Id.Uint64(),
		segments, int(replicates), objectInfo.GetRedundancyType())
	if err != nil {
		log.CtxErrorw(ctx, "failed to encode payload", "error", err)
		return
	}
	spList, approvals, err := taskNode.getApproval(objectInfo, int(replicates), int(replicates*ReplicateFactor), GetApprovalTimeout)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get storage providers to replicate", "error", err)
		return
	}

	// allocates memory from resource manager
	var memSize int
	if objectInfo.GetRedundancyType() == storagetypes.REDUNDANCY_REPLICA_TYPE {
		memSize = (int(params.GetRedundantDataChunkNum()) +
			int(params.GetRedundantParityChunkNum())) *
			int(objectInfo.GetPayloadSize())
	} else {
		memSize = int(math.Ceil(
			((float64(params.GetRedundantDataChunkNum()) + float64(params.GetRedundantParityChunkNum())) /
				float64(params.GetRedundantDataChunkNum())) *
				float64(objectInfo.GetPayloadSize())))
	}
	scope, err := taskNode.rcScope.BeginSpan()
	if err != nil {
		log.CtxErrorw(ctx, "failed to begin reserve resource", "error", err)
		return
	}
	stateFunc := func() string {
		var state string
		rcmgr.RcManager().ViewSystem(func(scope rcmgr.ResourceScope) error {
			state = scope.Stat().String()
			return nil
		})
		return state
	}
	err = scope.ReserveMemory(memSize, rcmgr.ReservationPriorityAlways)
	if err != nil {
		log.CtxErrorw(ctx, "failed to reserve memory from resource manager",
			"reserve_size", memSize, "resource_state", stateFunc(), "error", err)
		return
	}
	defer func() {
		scope.Done()
		log.Debugw("end replicate object request", "resource_state", stateFunc())
	}()

	spEndpoints := maps.SortKeys(approvals)
	var mux sync.Mutex
	getSpFunc := func() (sp *sptypes.StorageProvider, approval *p2ptypes.GetApprovalResponse, err error) {
		mux.Lock()
		defer mux.Unlock()
		if len(approvals) == 0 {
			log.CtxErrorw(ctx, "backup storage providers depleted")
			err = errors.New("no backup sp to pick up")
			return
		}
		endpoint := spEndpoints[0]
		sp = spList[endpoint]
		approval = approvals[endpoint]
		spEndpoints = spEndpoints[1:]
		delete(spList, endpoint)
		delete(approvals, endpoint)
		return
	}

	sealMsg.Operator = taskNode.config.SpOperatorAddress
	sealMsg.BucketName = objectInfo.GetBucketName()
	sealMsg.ObjectName = objectInfo.GetObjectName()
	sealMsg.SecondarySpAddresses = make([]string, replicates)
	sealMsg.SecondarySpSignatures = make([][]byte, replicates)
	objectInfo.SecondarySpAddresses = make([]string, replicates)

	processInfo.SegmentInfos = make([]*servicetypes.SegmentInfo, replicates)
	var done int64
	errCh := make(chan error, 10)
	for rIdx := 0; rIdx < int(replicates); rIdx++ {
		log.CtxDebugw(ctx, "start to replicate object", "object_id", objectInfo.Id, "replica_idx", rIdx)
		processInfo.SegmentInfos[rIdx] = &servicetypes.SegmentInfo{ObjectInfo: objectInfo}
		go func(rIdx int) {
			for {
				sp, approval, err := getSpFunc()
				if err != nil {
					errCh <- err
					return
				}
				var data [][]byte
				for idx := 0; idx < int(segments); idx++ {
					data = append(data, replicateData[idx][rIdx])
				}
				gatewayClient, err := gatewayclient.NewGatewayClient(sp.GetEndpoint())
				if err != nil {
					log.CtxErrorw(ctx, "failed to create gateway client",
						"sp_endpoint", sp.GetEndpoint(), "error", err)
					continue
				}
				// TODO:: add approval param to gateway, and secondary sp check the timeout, signature, spOpAddr of approval
				integrityHash, signature, err := gatewayClient.SyncPieceData(
					req.GetObjectInfo(), uint32(rIdx), uint32(len(replicateData[0][0])), approval, data)
				if err != nil {
					log.CtxErrorw(ctx, "failed to sync piece data", "endpoint", sp.GetEndpoint(), "error", err)
					continue
				}
				log.CtxDebugw(ctx, "receive the sp response", "replica_idx", rIdx, "integrity_hash",
					integrityHash, "endpoint", sp.GetEndpoint(), "signature", signature)

				msg := storagetypes.NewSecondarySpSignDoc(sp.GetOperator(), sdkmath.NewUint(objectInfo.Id.Uint64()), integrityHash).GetSignBytes()
				approvalAddr, err := sdk.AccAddressFromHexUnsafe(sp.GetApprovalAddress())
				if err != nil {
					log.CtxErrorw(ctx, "failed to parser sp operator address",
						"sp", sp.GetApprovalAddress(), "endpoint", sp.GetEndpoint(), "error", err)
					continue
				}
				err = storagetypes.VerifySignature(approvalAddr, sdk.Keccak256(msg), signature)
				if err != nil {
					log.CtxErrorw(ctx, "failed to verify sp signature",
						"sp", sp.GetApprovalAddress(), "endpoint", sp.GetEndpoint(), "error", err)
					continue
				}
				sealMsg.GetSecondarySpAddresses()[rIdx] = sp.GetOperator().String()
				sealMsg.GetSecondarySpSignatures()[rIdx] = signature
				processInfo.SegmentInfos[rIdx].Signature = signature
				processInfo.SegmentInfos[rIdx].IntegrityHash = integrityHash
				objectInfo.SecondarySpAddresses[rIdx] = sp.GetOperator().String()
				taskNode.spDB.SetObjectInfo(objectInfo.Id.Uint64(), objectInfo)
				taskNode.cache.Add(objectInfo.Id.Uint64(), processInfo)
				log.CtxInfow(ctx, "success to sync payload to sp",
					"sp", sp.GetOperator(), "endpoint", sp.GetEndpoint(), "replica_idx", rIdx)
				if atomic.AddInt64(&done, 1) == int64(replicates) {
					log.CtxInfow(ctx, "finish to sync all replicas")
					errCh <- nil
				}
				return
			}
		}(rIdx)
	}
	for err = range errCh {
		return
	}
	return
}

// QueryReplicatingObject query a replicating object information by object id
func (taskNode *TaskNode) QueryReplicatingObject(ctx context.Context, req *types.QueryReplicatingObjectRequest) (
	resp *types.QueryReplicatingObjectResponse, err error) {
	ctx = log.Context(ctx, req)
	objectID := req.GetObjectId()
	log.CtxDebugw(ctx, "query replicating object", "objectID", objectID)
	val, ok := taskNode.cache.Get(objectID)
	if !ok {
		err = merrors.ErrCacheMiss
		return
	}
	resp.ReplicateSegmentInfo = val.(*servicetypes.ReplicateSegmentInfo)
	return
}

func (taskNode *TaskNode) getApproval(
	objectInfo *storagetypes.ObjectInfo, low, high int, timeout int64) (
	map[string]*sptypes.StorageProvider, map[string]*p2ptypes.GetApprovalResponse, error) {
	var (
		spList       = make(map[string]*sptypes.StorageProvider)
		approvalList = make(map[string]*p2ptypes.GetApprovalResponse)
	)
	approvals, _, err := taskNode.p2p.GetApproval(context.Background(), objectInfo, int64(high), timeout)
	if err != nil {
		return spList, approvalList, err
	}
	if len(approvals) < low {
		return spList, approvalList, merrors.ErrSPApprovalNumber
	}
	for spOpAddr, approval := range approvals {
		sp, err := taskNode.spDB.GetSpByAddress(spOpAddr, sqldb.OperatorAddressType)
		if err != nil {
			continue
		}
		spList[sp.GetEndpoint()] = sp
		approvalList[sp.GetEndpoint()] = approval
	}
	if len(spList) < low {
		return spList, approvalList, merrors.ErrSPNumber
	}
	return spList, approvalList, nil
}

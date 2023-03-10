package stonenode

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	gatewayclient "github.com/bnb-chain/greenfield-storage-provider/service/gateway/client"
	"github.com/bnb-chain/greenfield-storage-provider/service/stonenode/types"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
)

var _ types.StoneNodeServiceServer = &StoneNode{}

// ReplicateObject call AsyncReplicateObject non-blocking upstream services
func (node *StoneNode) ReplicateObject(ctx context.Context, req *types.ReplicateObjectRequest) (
	resp *types.ReplicateObjectResponse, err error) {
	resp = &types.ReplicateObjectResponse{}
	node.spDB.UpdateJobState(req.GetObjectInfo().Id.String(), servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_DOING)
	go node.AsyncReplicateObject(req)
	log.Debugw("receive the replicate object task", "object_id", req.GetObjectInfo().Id)
	return
}

// AsyncReplicateObject replicate an object payload to other storage providers and seal object.
func (node *StoneNode) AsyncReplicateObject(req *types.ReplicateObjectRequest) (err error) {
	ctx := context.Background()
	processInfo := &servicetypes.ReplicateSegmentInfo{}
	sealMsg := &storagetypes.MsgSealObject{}
	objectInfo := req.GetObjectInfo()
	defer func() {
		if err != nil {
			node.spDB.UpdateJobState(objectInfo.Id.String(), servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_ERROR)
			log.CtxErrorw(ctx, "failed to replicate payload data to sp", "error", err)
			return
		}
		node.spDB.UpdateJobState(objectInfo.Id.String(), servicetypes.JobState_JOB_STATE_SIGN_OBJECT_DOING)
		_, err = node.signer.SealObjectOnChain(ctx, sealMsg)
		if err != nil {
			node.spDB.UpdateJobState(objectInfo.Id.String(), servicetypes.JobState_JOB_STATE_SIGN_OBJECT_ERROR)
			log.CtxErrorw(ctx, "failed to sign object by signer", "error", err)
			return
		}
		node.spDB.UpdateJobState(objectInfo.Id.String(), servicetypes.JobState_JOB_STATE_SEAL_OBJECT_TX_DOING)
		success, err := node.chain.ListenObjectSeal(ctx, objectInfo.GetBucketName(),
			objectInfo.GetObjectName(), 10)
		if err != nil {
			node.spDB.UpdateJobState(objectInfo.Id.String(), servicetypes.JobState_JOB_STATE_SEAL_OBJECT_ERROR)
			log.CtxErrorw(ctx, "failed to seal object on chain", "error", err)
			return
		}
		node.spDB.UpdateJobState(objectInfo.Id.String(), servicetypes.JobState_JOB_STATE_SEAL_OBJECT_DONE)
		log.CtxInfow(ctx, "seal object on chain", "success", success)
	}()

	params, err := node.spDB.GetStorageParams()
	if err != nil {
		log.CtxErrorw(ctx, "failed to query sp params", "error", err)
		return
	}
	segments := piecestore.ComputeSegmentCount(objectInfo.GetPayloadSize(),
		params.GetMaxSegmentSize())
	replicates := params.GetRedundantDataChunkNum() + params.GetRedundantParityChunkNum()
	replicateData, err := node.EncodeReplicateSegments(ctx, objectInfo.Id,
		segments, int(replicates), objectInfo.GetRedundancyType())
	if err != nil {
		log.CtxErrorw(ctx, "failed to encode payload", "error", err)
		return
	}
	spList, err := node.spDB.FetchAllSpWithoutOwnSp(sptypes.STATUS_IN_SERVICE)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get storage providers to replicate", "error", err)
		return
	}

	sealMsg.Operator = node.config.SpOperatorAddress
	sealMsg.BucketName = objectInfo.GetBucketName()
	sealMsg.ObjectName = objectInfo.GetObjectName()
	sealMsg.SecondarySpAddresses = make([]string, replicates)
	sealMsg.SecondarySpSignatures = make([][]byte, replicates)
	objectInfo.SecondarySpAddresses = make([]string, replicates)

	var mux sync.Mutex
	getSpFunc := func() (*sptypes.StorageProvider, error) {
		mux.Lock()
		defer mux.Unlock()
		if len(spList) == 0 {
			log.CtxErrorw(ctx, "backup storage providers depleted")
			return nil, errors.New("no backup sp to pick up")
		}
		sp := spList[0]
		spList = spList[1:]
		return sp, nil
	}
	processInfo.SegmentInfos = make([]*servicetypes.SegmentInfo, replicates)
	var done int64
	errCh := make(chan error, 10)
	for rIdx := 0; rIdx < int(replicates); rIdx++ {
		log.CtxDebugw(ctx, "start to replicate object", "object_id", objectInfo.Id, "replica_idx", rIdx)
		processInfo.SegmentInfos[rIdx] = &servicetypes.SegmentInfo{ObjectInfo: objectInfo}
		go func(rIdx int) {
			for {
				sp, err := getSpFunc()
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
				integrityHash, signature, err := gatewayClient.SyncPieceData(
					req.GetObjectInfo(), uint32(rIdx), uint32(len(replicateData[0][0])), data)
				if err != nil {
					log.CtxErrorw(ctx, "failed to sync piece data", "endpoint", sp.GetEndpoint(), "error", err)
					continue
				}
				log.CtxDebugw(ctx, "receive the sp response", "replica_idx", rIdx, "integrity_hash",
					integrityHash, "endpoint", sp.GetEndpoint(), "signature", signature)
				msg := storagetypes.NewSecondarySpSignDoc(sp.GetOperator(), integrityHash).GetSignBytes()
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
				node.spDB.SetObjectInfo(objectInfo.Id.String(), objectInfo)
				node.cache.Add(objectInfo.Id.Uint64(), processInfo)
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
func (node *StoneNode) QueryReplicatingObject(ctx context.Context, req *types.QueryReplicatingObjectRequest) (
	resp *types.QueryReplicatingObjectResponse, err error) {
	ctx = log.Context(ctx, req)
	objectID := req.GetObjectId()
	log.CtxDebugw(ctx, "query replicating object", "objectID", objectID)
	val, ok := node.cache.Get(objectID)
	if !ok {
		err = merrors.ErrCacheMiss
		return
	}
	resp.ReplicateSegmentInfo = val.(*servicetypes.ReplicateSegmentInfo)
	return
}

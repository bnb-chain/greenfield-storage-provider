package stonenode

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	xtypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/stonenode/types"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

	gatewayclient "github.com/bnb-chain/greenfield-storage-provider/service/gateway/client"
)

var _ types.StoneNodeServiceServer = &StoneNode{}

// ReplicateObject call AsyncReplicateObject non-blocking upstream services
func (node *StoneNode) ReplicateObject(
	ctx context.Context,
	req *types.ReplicateObjectRequest) (
	resp *types.ReplicateObjectResponse, err error) {
	resp = &types.ReplicateObjectResponse{}
	node.spDB.UpdateJobState(req.GetObjectInfo().Id.Uint64(), servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_DOING)
	go node.AsyncReplicateObject(context.Background(), req)
	return
}

// AsyncReplicateObject replicate an object payload to other storage providers and seal object.
func (node *StoneNode) AsyncReplicateObject(ctx context.Context,
	req *types.ReplicateObjectRequest) (err error) {
	ctx = log.Context(ctx, req)
	processInfo := &servicetypes.ReplicateSegmentInfo{}
	sealMsg := &storagetypes.MsgSealObject{}
	objectInfo := req.GetObjectInfo()
	defer func() {
		if err != nil {
			node.spDB.UpdateJobState(objectInfo.Id.Uint64(), servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_ERROR)
			log.CtxErrorw(ctx, "failed to replicate payload data to sp", "error", err)
			return
		}
		node.spDB.UpdateJobState(objectInfo.Id.Uint64(), servicetypes.JobState_JOB_STATE_SIGN_OBJECT_DOING)
		_, err = node.signer.SealObjectOnChain(ctx, sealMsg)
		if err != nil {
			node.spDB.UpdateJobState(objectInfo.Id.Uint64(), servicetypes.JobState_JOB_STATE_SIGN_OBJECT_ERROR)
			log.CtxErrorw(ctx, "failed to sign object by signer", "error", err)
			return
		}
		node.spDB.UpdateJobState(objectInfo.Id.Uint64(), servicetypes.JobState_JOB_STATE_SEAL_OBJECT_TX_DOING)
		success, err := node.chain.ListenObjectSeal(ctx, objectInfo.GetBucketName(),
			objectInfo.GetObjectName(), 10)
		if err != nil {
			node.spDB.UpdateJobState(objectInfo.Id.Uint64(), servicetypes.JobState_JOB_STATE_SEAL_OBJECT_ERROR)
			log.CtxErrorw(ctx, "failed to seal object on chain", "error", err)
			return
		}
		node.spDB.UpdateJobState(objectInfo.Id.Uint64(), servicetypes.JobState_JOB_STATE_SEAL_OBJECT_DONE)
		log.CtxInfow(ctx, "seal object on chain", "success", success)
		return
	}()

	params, err := node.spDB.GetStorageParams()
	if err != nil {
		log.CtxErrorw(ctx, "failed to query sp params", "error", err)
		return
	}
	segments := piecestore.ComputeSegmentCount(objectInfo.GetPayloadSize(),
		params.GetMaxSegmentSize())
	replicates := params.GetRedundantDataChunkNum() +
		params.GetRedundantDataChunkNum()
	replicateData, err := node.EncodeReplicateSegments(ctx, objectInfo.Id.Uint64(),
		segments, int(replicates), objectInfo.GetRedundancyType())
	if err != nil {
		log.CtxErrorw(ctx, "failed to encode payload", "error", err)
		return
	}
	spList, err := node.spDB.FetchAllSPWithoutOwnSP(sptypes.STATUS_IN_SERVICE)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get storage providers to replicate", "error", err)
		return
	}

	sealMsg.Operator = node.config.SpOperatorAddress
	sealMsg.BucketName = objectInfo.GetBucketName()
	sealMsg.ObjectName = objectInfo.GetObjectName()
	sealMsg.SecondarySpAddresses = make([]string, replicates)
	sealMsg.SecondarySpSignatures = make([][]byte, replicates)

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
		processInfo.SegmentInfos[rIdx] = &servicetypes.SegmentInfo{ObjectInfo: objectInfo}
		go func(rIdx int) {
			for {
				sp, err := getSpFunc()
				if err != nil {
					errCh <- err
					return
				}
				var data [][]byte
				for idx := 0; idx < len(replicateData[0]); idx++ {
					data[idx] = replicateData[idx][rIdx]
				}
				gatewayClient, err := gatewayclient.NewGatewayClient(sp.GetEndpoint())
				if err != nil {
					log.CtxErrorw(ctx, "failed to create gateway client", "sp_endpoint", sp.GetEndpoint(), "error", err)
					continue
				}
				integrityHash, signature, err := gatewayClient.SyncPieceData(req.GetObjectInfo(), uint32(rIdx), uint32(len(replicateData[0][0])), data)
				if err != nil {
					log.CtxErrorw(ctx, "failed to sync piece data", "error", err)
					continue
				}
				log.CtxDebugw(ctx, "receive the sp response", "replica_idx", rIdx, "integrity_hash", integrityHash, "signature", signature)
				msg := storagetypes.NewSecondarySpSignDoc(sp.GetOperator(), integrityHash).GetSignBytes()
				approvalAddr, err := sdk.AccAddressFromHexUnsafe(sp.GetApprovalAddress())
				if err != nil {
					log.CtxErrorw(ctx, "failed to parser sp operator address", "sp", sp.GetApprovalAddress(), "error", err)
					continue
				}
				err = xtypes.VerifySignature(approvalAddr, sdk.Keccak256(msg), signature)
				if err != nil {
					log.CtxErrorw(ctx, "failed to verify sp signature", "sp", sp.GetApprovalAddress(), "error", err)
					continue
				}
				log.CtxInfow(ctx, "success to sync payload to sp", "sp", sp.GetOperator(), "replica_idx", rIdx)
				if atomic.AddInt64(&done, 1) == int64(replicates) {
					log.CtxInfow(ctx, "finish to sync all replicas")
					errCh <- nil
					return
				}
				sealMsg.GetSecondarySpAddresses()[rIdx] = sp.GetOperator().String()
				sealMsg.GetSecondarySpSignatures()[rIdx] = signature
				processInfo.SegmentInfos[rIdx].Signature = signature
				processInfo.SegmentInfos[rIdx].IntegrityHash = integrityHash
				node.cache.Add(objectInfo.Id.Uint64(), processInfo)
			}
		}(rIdx)
	}
	for {
		select {
		case err = <-errCh:
			return
		}
	}
}

// QueryReplicatingObject query a replicating object information by object id
func (node *StoneNode) QueryReplicatingObject(
	ctx context.Context,
	req *types.QueryReplicatingObjectRequest) (
	resp *types.QueryReplicatingObjectResponse, err error) {
	ctx = log.Context(ctx, req)
	objectID := req.GetObjectId()
	val, ok := node.cache.Get(objectID)
	if !ok {
		err = merrors.ErrCacheMiss
		return
	}
	resp.ReplicateSegmentInfo = val.(*servicetypes.ReplicateSegmentInfo)
	return
}

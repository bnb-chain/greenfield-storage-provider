package tasknode

import (
	"context"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	p2ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/types"
	"github.com/bnb-chain/greenfield-storage-provider/service/tasknode/types"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var _ types.TaskNodeServiceServer = &TaskNode{}

const enableStreamReplicate = false

type replicateTask interface {
	init() error
	execute(waitCh chan error)
}

// ReplicateObject call AsyncReplicateObject non-blocking upstream services
func (taskNode *TaskNode) ReplicateObject(ctx context.Context, req *types.ReplicateObjectRequest) (
	*types.ReplicateObjectResponse, error) {
	if req.GetObjectInfo() == nil {
		return nil, merrors.ErrDanglingPointer
	}

	var (
		resp   *types.ReplicateObjectResponse
		err    error
		task   replicateTask
		waitCh chan error
	)

	ctx = log.WithValue(ctx, "object_id", req.GetObjectInfo().Id.String())
	if enableStreamReplicate {
		if task, err = newStreamReplicateObjectTask(ctx, taskNode, req.GetObjectInfo()); err != nil {
			log.CtxErrorw(ctx, "failed to new replicate object task", "error", err)
			return nil, err
		}
	} else {
		if task, err = newReplicateObjectTask(ctx, taskNode, req.GetObjectInfo()); err != nil {
			log.CtxErrorw(ctx, "failed to new replicate object task", "error", err)
			return nil, err
		}

	}
	if err = task.init(); err != nil {
		log.CtxErrorw(ctx, "failed to init replicate object task", "error", err)
		return nil, err
	}

	waitCh = make(chan error)
	go task.execute(waitCh)
	if err = <-waitCh; err != nil {
		return nil, err
	}

	resp = &types.ReplicateObjectResponse{}
	return resp, nil
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
	resp.ReplicatePieceInfo = val.(*servicetypes.ReplicatePieceInfo)
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

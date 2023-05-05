package tasknode

import (
	"context"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// sealObjectTask represents the background object seal task.
type sealObjectTask struct {
	ctx            context.Context
	taskNode       *TaskNode
	objectInfo     *storagetypes.ObjectInfo
	sealObjectInfo *storagetypes.MsgSealObject
}

// newSealObjectTask returns a sealObjectTask instance.
func newSealObjectTask(ctx context.Context, task *TaskNode, objectInfo *storagetypes.ObjectInfo, sealObjectInfo *storagetypes.MsgSealObject) (*sealObjectTask, error) {
	if ctx == nil || task == nil || objectInfo == nil || sealObjectInfo == nil {
		return nil, merrors.ErrInvalidParams
	}
	return &sealObjectTask{
		ctx:            ctx,
		taskNode:       task,
		objectInfo:     objectInfo,
		sealObjectInfo: sealObjectInfo,
	}, nil
}

// init is used to synchronize the resources which is needed to initialize the task.
func (t *sealObjectTask) init() error {
	return nil
}

// updateTaskState is used to update task state.
func (t *sealObjectTask) updateTaskState(state servicetypes.JobState) error {
	return t.taskNode.spDB.UpdateJobState(t.objectInfo.Id.Uint64(), state)
}

// execute is used to start the task, and waitCh is used to wait runtime initialization.
func (t *sealObjectTask) execute(waitCh chan error) {
	// TODO: refine it
	waitCh <- nil
	t.updateTaskState(servicetypes.JobState_JOB_STATE_SIGN_OBJECT_DOING)
	_, err := t.taskNode.signer.SealObjectOnChain(context.Background(), t.sealObjectInfo)
	if err != nil {
		t.updateTaskState(servicetypes.JobState_JOB_STATE_SIGN_OBJECT_ERROR)
		log.CtxErrorw(t.ctx, "failed to sign object by signer", "error", err)
		return
	}
	t.updateTaskState(servicetypes.JobState_JOB_STATE_SEAL_OBJECT_DOING)
	err = t.taskNode.chain.ListenObjectSeal(context.Background(), t.objectInfo.GetBucketName(),
		t.objectInfo.GetObjectName(), 10)
	if err != nil {
		t.updateTaskState(servicetypes.JobState_JOB_STATE_SEAL_OBJECT_ERROR)
		log.CtxErrorw(t.ctx, "failed to seal object on chain", "error", err)
		return
	}
	t.updateTaskState(servicetypes.JobState_JOB_STATE_SEAL_OBJECT_DONE)
	t.taskNode.manager.DoneSealObjectTask(context.Background(), t.objectInfo)
}

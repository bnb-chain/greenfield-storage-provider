package gfspapp

import (
	"context"
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGfSpBaseApp_GfSpAskSecondaryReplicatePieceApprovalSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockP2P(ctrl)
	g.p2p = m
	approval1 := &gfsptask.GfSpReplicatePieceApprovalTask{
		Task:                 &gfsptask.GfSpTask{},
		ObjectInfo:           mockObjectInfo,
		StorageParams:        mockStorageParams,
		AskSpOperatorAddress: "mockAskSpOperatorAddress1",
	}
	approvalList := make([]task.ApprovalReplicatePieceTask, 0)
	approvalList = append(approvalList, approval1)
	m.EXPECT().HandleReplicatePieceApproval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		approvalList, nil).Times(1)
	req := &gfspserver.GfSpAskSecondaryReplicatePieceApprovalRequest{
		ReplicatePieceApprovalTask: approval1,
		Min:                        1,
		Max:                        2,
		Timeout:                    0,
	}
	result, err := g.GfSpAskSecondaryReplicatePieceApproval(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, "mockObjectName", result.GetApprovedTasks()[0].GetObjectInfo().GetObjectName())
}

func TestGfSpBaseApp_GfSpAskSecondaryReplicatePieceApprovalFailure1(t *testing.T) {
	t.Log("Failure case description: dangling pointer task")
	g := setup(t)
	result, err := g.GfSpAskSecondaryReplicatePieceApproval(context.TODO(), nil)
	assert.Equal(t, ErrReplicatePieceApprovalTaskDangling, err)
	assert.Nil(t, result)
}

func TestGfSpBaseApp_GfSpAskSecondaryReplicatePieceApprovalFailure2(t *testing.T) {
	t.Log("Failure case description: failed to get replicate piece approval from p2p")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockP2P(ctrl)
	g.p2p = m
	m.EXPECT().HandleReplicatePieceApproval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		nil, mockErr).Times(1)
	req := &gfspserver.GfSpAskSecondaryReplicatePieceApprovalRequest{
		ReplicatePieceApprovalTask: &gfsptask.GfSpReplicatePieceApprovalTask{
			Task:                 &gfsptask.GfSpTask{},
			ObjectInfo:           mockObjectInfo,
			StorageParams:        mockStorageParams,
			AskSpOperatorAddress: "mockAskSpOperatorAddress1",
		},
		Min:     1,
		Max:     2,
		Timeout: 0,
	}
	result, err := g.GfSpAskSecondaryReplicatePieceApproval(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpQueryP2PBootstrapSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockP2P(ctrl)
	g.p2p = m
	m.EXPECT().HandleQueryBootstrap(gomock.Any()).Return([]string{"1", "2"}, nil).Times(1)
	result, err := g.GfSpQueryP2PBootstrap(context.TODO(), &gfspserver.GfSpQueryP2PNodeRequest{})
	assert.Nil(t, err)
	assert.Equal(t, []string{"1", "2"}, result.GetNodes())
}

func TestGfSpBaseApp_GfSpQueryP2PBootstrapFailure(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockP2P(ctrl)
	g.p2p = m
	m.EXPECT().HandleQueryBootstrap(gomock.Any()).Return(nil, mockErr).Times(1)
	result, err := g.GfSpQueryP2PBootstrap(context.TODO(), &gfspserver.GfSpQueryP2PNodeRequest{})
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

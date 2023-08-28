package gfspapp

import (
	"context"
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGfSpBaseApp_GfSpReplicatePieceSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockReceiver(ctrl)
	g.receiver = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()
	m.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(m1, nil).Times(1)
	m.EXPECT().HandleReceivePieceTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	req := &gfspserver.GfSpReplicatePieceRequest{
		ReceivePieceTask: &gfsptask.GfSpReceivePieceTask{
			Task:          &gfsptask.GfSpTask{},
			ObjectInfo:    mockObjectInfo,
			StorageParams: mockStorageParams,
		},
		PieceData: []byte("mockPieceData"),
	}
	result, err := g.GfSpReplicatePiece(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, &gfspserver.GfSpReplicatePieceResponse{}, result)
}

func TestGfSpBaseApp_GfSpReplicatePieceFailure1(t *testing.T) {
	t.Log("Failure case description: task pointer dangling")
	g := setup(t)
	result, err := g.GfSpReplicatePiece(context.TODO(), nil)
	assert.Nil(t, err)
	assert.Equal(t, ErrReceiveTaskDangling, result.GetErr())
}

func TestGfSpBaseApp_GfSpReplicatePieceFailure2(t *testing.T) {
	t.Log("Failure case description: failed to reserve resource")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockReceiver(ctrl)
	g.receiver = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()
	m.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpReplicatePieceRequest{
		ReceivePieceTask: &gfsptask.GfSpReceivePieceTask{
			Task:          &gfsptask.GfSpTask{},
			ObjectInfo:    mockObjectInfo,
			StorageParams: mockStorageParams,
		},
		PieceData: []byte("mockPieceData"),
	}
	result, err := g.GfSpReplicatePiece(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, ErrReceiveExhaustResource, result.GetErr())
}

func TestGfSpBaseApp_GfSpReplicatePieceFailure3(t *testing.T) {
	t.Log("Failure case description: failed to replicate piece")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockReceiver(ctrl)
	g.receiver = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()
	m.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(m1, nil).Times(1)
	m.EXPECT().HandleReceivePieceTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockErr).Times(1)
	req := &gfspserver.GfSpReplicatePieceRequest{
		ReceivePieceTask: &gfsptask.GfSpReceivePieceTask{
			Task:          &gfsptask.GfSpTask{},
			ObjectInfo:    mockObjectInfo,
			StorageParams: mockStorageParams,
		},
		PieceData: []byte("mockPieceData"),
	}
	result, err := g.GfSpReplicatePiece(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_GfSpDoneReplicatePieceSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockReceiver(ctrl)
	g.receiver = m
	m.EXPECT().HandleDoneReceivePieceTask(gomock.Any(), gomock.Any()).Return([]byte("mockSig"), nil).Times(1)
	req := &gfspserver.GfSpDoneReplicatePieceRequest{ReceivePieceTask: &gfsptask.GfSpReceivePieceTask{
		Task:          &gfsptask.GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}}
	result, err := g.GfSpDoneReplicatePiece(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, []byte("mockSig"), result.GetSignature())
}

func TestGfSpBaseApp_GfSpDoneReplicatePieceFailure1(t *testing.T) {
	t.Log("Failure case description: task pointer dangling")
	g := setup(t)
	result, err := g.GfSpDoneReplicatePiece(context.TODO(), nil)
	assert.Nil(t, err)
	assert.Equal(t, ErrReceiveTaskDangling, result.GetErr())
}

func TestGfSpBaseApp_GfSpDoneReplicatePieceFailure2(t *testing.T) {
	t.Log("Failure case description: failed to done replicate piece")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockReceiver(ctrl)
	g.receiver = m
	m.EXPECT().HandleDoneReceivePieceTask(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpDoneReplicatePieceRequest{ReceivePieceTask: &gfsptask.GfSpReceivePieceTask{
		Task:          &gfsptask.GfSpTask{},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}}
	result, err := g.GfSpDoneReplicatePiece(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

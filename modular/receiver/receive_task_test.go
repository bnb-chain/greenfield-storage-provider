package receiver

import (
	"context"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfsppieceop"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestErrPieceStoreWithDetail(t *testing.T) {
	mock := "mockDetail"
	result := ErrPieceStoreWithDetail(mock)
	assert.Equal(t, mock, result.Description)
}

func TestErrGfSpDBWithDetail(t *testing.T) {
	mock := "mockDetail"
	result := ErrGfSpDBWithDetail(mock)
	assert.Equal(t, mock, result.Description)
}

func TestHandleReceivePieceTask_RepeatedTask(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	q := taskqueue.NewMockTQueueOnStrategy(ctrl)
	r.receiveQueue = q
	q.EXPECT().Has(gomock.Any()).Return(true).Times(1)
	mockTask := &gfsptask.GfSpReceivePieceTask{
		Task: &gfsptask.GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:           sdkmath.NewUint(100),
			ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
			PayloadSize:  100,
		}}
	err := r.HandleReceivePieceTask(context.TODO(), mockTask, nil)
	assert.NotNil(t, err)
}

func TestHandleReceivePieceTask_PushTaskFailed(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	q := taskqueue.NewMockTQueueOnStrategy(ctrl)
	r.receiveQueue = q
	q.EXPECT().Has(gomock.Any()).Return(false).Times(1)
	q.EXPECT().Push(gomock.Any()).Return(fmt.Errorf("failed to push")).Times(1)
	mockTask := &gfsptask.GfSpReceivePieceTask{
		Task: &gfsptask.GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:           sdkmath.NewUint(100),
			ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
			PayloadSize:  100,
		}}
	err := r.HandleReceivePieceTask(context.TODO(), mockTask, nil)
	assert.NotNil(t, err)
}

func TestHandleReceivePieceTask_CheckChecksumFailed(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	q := taskqueue.NewMockTQueueOnStrategy(ctrl)
	r.receiveQueue = q
	q.EXPECT().Has(gomock.Any()).Return(false).Times(1)
	q.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	q.EXPECT().PopByKey(gomock.Any()).Return(nil).Times(1)
	mockTask := &gfsptask.GfSpReceivePieceTask{
		Task: &gfsptask.GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:           sdkmath.NewUint(100),
			ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
			PayloadSize:  100,
		}}
	err := r.HandleReceivePieceTask(context.TODO(), mockTask, nil)
	assert.NotNil(t, err)
}

func TestHandleReceivePieceTask_SetPieceChecksumFailed(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	q := taskqueue.NewMockTQueueOnStrategy(ctrl)
	r.baseApp.SetPieceOp(&gfsppieceop.GfSpPieceOp{})
	r.receiveQueue = q
	q.EXPECT().Has(gomock.Any()).Return(false).Times(1)
	q.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	q.EXPECT().PopByKey(gomock.Any()).Return(nil).Times(1)
	data := []byte{'a'}
	checksum := hash.GenerateChecksum(data)
	mockTask := &gfsptask.GfSpReceivePieceTask{
		Task: &gfsptask.GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:           sdkmath.NewUint(100),
			ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
			PayloadSize:  100,
		},
		PieceChecksum: checksum,
	}
	mockSPDB := spdb.NewMockSPDB(ctrl)
	r.baseApp.SetGfSpDB(mockSPDB)
	mockSPDB.EXPECT().SetReplicatePieceChecksum(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("failed to set piece checksum")).Times(1)
	err := r.HandleReceivePieceTask(context.TODO(), mockTask, data)
	assert.NotNil(t, err)
}

func TestHandleReceivePieceTask_PutPieceFailed(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	q := taskqueue.NewMockTQueueOnStrategy(ctrl)
	r.baseApp.SetPieceOp(&gfsppieceop.GfSpPieceOp{})
	r.receiveQueue = q
	q.EXPECT().Has(gomock.Any()).Return(false).Times(1)
	q.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	q.EXPECT().PopByKey(gomock.Any()).Return(nil).Times(1)
	data := []byte{'a'}
	checksum := hash.GenerateChecksum(data)
	mockTask := &gfsptask.GfSpReceivePieceTask{
		Task: &gfsptask.GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:           sdkmath.NewUint(100),
			ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
			PayloadSize:  100,
		},
		PieceChecksum: checksum,
	}
	mockSPDB := spdb.NewMockSPDB(ctrl)
	r.baseApp.SetGfSpDB(mockSPDB)
	mockSPDB.EXPECT().SetReplicatePieceChecksum(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	mockPieceStoreAPI := piecestore.NewMockPieceStore(ctrl)
	r.baseApp.SetPieceStore(mockPieceStoreAPI)
	mockPieceStoreAPI.EXPECT().PutPiece(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("failed to put piece")).Times(1)
	err := r.HandleReceivePieceTask(context.TODO(), mockTask, data)
	assert.NotNil(t, err)
}

func TestHandleReceivePieceTask_HandleReceivePieceTaskSucceed(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	q := taskqueue.NewMockTQueueOnStrategy(ctrl)
	r.baseApp.SetPieceOp(&gfsppieceop.GfSpPieceOp{})
	r.receiveQueue = q
	q.EXPECT().Has(gomock.Any()).Return(false).Times(1)
	q.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	q.EXPECT().PopByKey(gomock.Any()).Return(nil).Times(1)
	data := []byte{'a'}
	checksum := hash.GenerateChecksum(data)
	mockTask := &gfsptask.GfSpReceivePieceTask{
		Task: &gfsptask.GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:           sdkmath.NewUint(100),
			ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
			PayloadSize:  100,
		},
		PieceChecksum: checksum,
	}
	mockSPDB := spdb.NewMockSPDB(ctrl)
	r.baseApp.SetGfSpDB(mockSPDB)
	mockSPDB.EXPECT().SetReplicatePieceChecksum(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	mockPieceStoreAPI := piecestore.NewMockPieceStore(ctrl)
	r.baseApp.SetPieceStore(mockPieceStoreAPI)
	mockPieceStoreAPI.EXPECT().PutPiece(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	err := r.HandleReceivePieceTask(context.TODO(), mockTask, data)
	assert.Nil(t, err)
}

func TestHandleDoneReceivePieceTask_PushTaskFailed(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	q := taskqueue.NewMockTQueueOnStrategy(ctrl)
	r.receiveQueue = q
	q.EXPECT().Push(gomock.Any()).Return(fmt.Errorf("failed to push")).Times(1)
	mockTask := &gfsptask.GfSpReceivePieceTask{
		Task: &gfsptask.GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:           sdkmath.NewUint(100),
			ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
			PayloadSize:  100,
		}}
	_, err := r.HandleDoneReceivePieceTask(context.TODO(), mockTask)
	assert.NotNil(t, err)
}

func TestHandleDoneReceivePieceTask_GetAllPieceChecksumFailed(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	q := taskqueue.NewMockTQueueOnStrategy(ctrl)
	r.receiveQueue = q
	r.baseApp.SetPieceOp(&gfsppieceop.GfSpPieceOp{})
	q.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	q.EXPECT().PopByKey(gomock.Any()).Return(nil).Times(1)
	mockTask := &gfsptask.GfSpReceivePieceTask{
		Task: &gfsptask.GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:           sdkmath.NewUint(100),
			ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
			PayloadSize:  100,
		},
		StorageParams: &storagetypes.Params{
			VersionedParams: storagetypes.VersionedParams{
				MaxSegmentSize: 16 * 1024 * 1024,
			}}}
	mockSPDB := spdb.NewMockSPDB(ctrl)
	r.baseApp.SetGfSpDB(mockSPDB)
	mockSPDB.EXPECT().GetAllReplicatePieceChecksum(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("failed to get all piece checksum")).Times(1)
	_, err := r.HandleDoneReceivePieceTask(context.TODO(), mockTask)
	assert.NotNil(t, err)
}

func TestHandleDoneReceivePieceTask_PieceCountMismatch(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	q := taskqueue.NewMockTQueueOnStrategy(ctrl)
	r.receiveQueue = q
	r.baseApp.SetPieceOp(&gfsppieceop.GfSpPieceOp{})
	q.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	q.EXPECT().PopByKey(gomock.Any()).Return(nil).Times(1)
	mockTask := &gfsptask.GfSpReceivePieceTask{
		Task: &gfsptask.GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:           sdkmath.NewUint(100),
			ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
			PayloadSize:  100,
		},
		StorageParams: &storagetypes.Params{
			VersionedParams: storagetypes.VersionedParams{
				MaxSegmentSize: 16 * 1024 * 1024,
			}}}
	mockSPDB := spdb.NewMockSPDB(ctrl)
	r.baseApp.SetGfSpDB(mockSPDB)
	mockSPDB.EXPECT().GetAllReplicatePieceChecksum(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
	_, err := r.HandleDoneReceivePieceTask(context.TODO(), mockTask)
	assert.NotNil(t, err)
}

func TestHandleDoneReceivePieceTask_SignSecondarySealBlsFailed(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	q := taskqueue.NewMockTQueueOnStrategy(ctrl)
	r.receiveQueue = q
	r.baseApp.SetPieceOp(&gfsppieceop.GfSpPieceOp{})
	q.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	q.EXPECT().PopByKey(gomock.Any()).Return(nil).Times(1)
	mockTask := &gfsptask.GfSpReceivePieceTask{
		Task: &gfsptask.GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:           sdkmath.NewUint(100),
			ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
			PayloadSize:  100,
		},
		StorageParams: &storagetypes.Params{
			VersionedParams: storagetypes.VersionedParams{
				MaxSegmentSize: 16 * 1024 * 1024,
			}}}
	mockSPDB := spdb.NewMockSPDB(ctrl)
	r.baseApp.SetGfSpDB(mockSPDB)
	mockSPDB.EXPECT().GetAllReplicatePieceChecksum(gomock.Any(), gomock.Any(), gomock.Any()).Return([][]byte{[]byte("mock")}, nil).Times(1)
	mockGRPCAPI := gfspclient.NewMockGfSpClientAPI(ctrl)
	r.baseApp.SetGfSpClient(mockGRPCAPI)
	mockGRPCAPI.EXPECT().SignSecondarySealBls(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("failed to sign bls")).Times(1)
	_, err := r.HandleDoneReceivePieceTask(context.TODO(), mockTask)
	assert.NotNil(t, err)
}

func TestHandleDoneReceivePieceTask_SetObjectIntegrityFailed(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	q := taskqueue.NewMockTQueueOnStrategy(ctrl)
	r.receiveQueue = q
	r.baseApp.SetPieceOp(&gfsppieceop.GfSpPieceOp{})
	q.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	q.EXPECT().PopByKey(gomock.Any()).Return(nil).Times(1)
	mockTask := &gfsptask.GfSpReceivePieceTask{
		Task: &gfsptask.GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:           sdkmath.NewUint(100),
			ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
			PayloadSize:  100,
		},
		StorageParams: &storagetypes.Params{
			VersionedParams: storagetypes.VersionedParams{
				MaxSegmentSize: 16 * 1024 * 1024,
			}}}
	mockSPDB := spdb.NewMockSPDB(ctrl)
	r.baseApp.SetGfSpDB(mockSPDB)
	mockSPDB.EXPECT().GetAllReplicatePieceChecksum(gomock.Any(), gomock.Any(), gomock.Any()).Return([][]byte{[]byte("mock")}, nil).Times(1)
	mockSPDB.EXPECT().SetObjectIntegrity(gomock.Any()).Return(fmt.Errorf("failed to set integrity")).Times(1)
	mockGRPCAPI := gfspclient.NewMockGfSpClientAPI(ctrl)
	r.baseApp.SetGfSpClient(mockGRPCAPI)
	mockGRPCAPI.EXPECT().SignSecondarySealBls(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
	_, err := r.HandleDoneReceivePieceTask(context.TODO(), mockTask)
	assert.NotNil(t, err)
}

func TestHandleDoneReceivePieceTask_HandleDoneReceivePieceTaskSucceed(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	q := taskqueue.NewMockTQueueOnStrategy(ctrl)
	r.receiveQueue = q
	r.baseApp.SetPieceOp(&gfsppieceop.GfSpPieceOp{})
	q.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	q.EXPECT().PopByKey(gomock.Any()).Return(nil).Times(1)
	mockTask := &gfsptask.GfSpReceivePieceTask{
		Task: &gfsptask.GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:           sdkmath.NewUint(100),
			ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
			PayloadSize:  100,
		},
		StorageParams: &storagetypes.Params{
			VersionedParams: storagetypes.VersionedParams{
				MaxSegmentSize: 16 * 1024 * 1024,
			}}}
	mockSPDB := spdb.NewMockSPDB(ctrl)
	r.baseApp.SetGfSpDB(mockSPDB)
	mockSPDB.EXPECT().GetAllReplicatePieceChecksum(gomock.Any(), gomock.Any(), gomock.Any()).Return([][]byte{[]byte("mock")}, nil).Times(1)
	mockSPDB.EXPECT().SetObjectIntegrity(gomock.Any()).Return(nil).Times(1)
	mockSPDB.EXPECT().DeleteAllReplicatePieceChecksum(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("failed to delete all piece checksum")).Times(1)
	mockGRPCAPI := gfspclient.NewMockGfSpClientAPI(ctrl)
	r.baseApp.SetGfSpClient(mockGRPCAPI)
	mockGRPCAPI.EXPECT().SignSecondarySealBls(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
	mockGRPCAPI.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	_, err := r.HandleDoneReceivePieceTask(context.TODO(), mockTask)
	assert.Nil(t, err)
}

func TestQueryTasks(t *testing.T) {
	r := setup(t)
	ctrl := gomock.NewController(t)
	q := taskqueue.NewMockTQueueOnStrategy(ctrl)
	r.receiveQueue = q
	q.EXPECT().ScanTask(gomock.Any()).Times(1)
	r.QueryTasks(context.TODO(), "")
}

package uploader

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func TestUploadModular_PreUploadObject(t *testing.T) {
	cases := []struct {
		name             string
		uploadObjectTask coretask.UploadObjectTask
		wantedErr        error
	}{
		{
			name:             "1",
			uploadObjectTask: nil,
			wantedErr:        ErrDanglingUploadTask,
		},
		{
			name: "2",
			uploadObjectTask: &gfsptask.GfSpUploadObjectTask{
				ObjectInfo:    &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_DISCONTINUED},
				StorageParams: &storagetypes.Params{MaxPayloadSize: 1},
			},
			wantedErr: ErrNotCreatedState,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			u := setup(t)
			err := u.PreUploadObject(context.TODO(), tt.uploadObjectTask)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestUploadModular_PreUploadObjectSuccess(t *testing.T) {
	u := setup(t)
	ctrl := gomock.NewController(t)

	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.uploadQueue = m
	m.EXPECT().Has(gomock.Any()).Return(false).Times(1)

	m1 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m1)
	m1.EXPECT().CreateUploadObject(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	uploadObjectTask := &gfsptask.GfSpUploadObjectTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_CREATED},
		StorageParams: &storagetypes.Params{MaxPayloadSize: 1},
	}
	err := u.PreUploadObject(context.TODO(), uploadObjectTask)
	assert.Nil(t, err)
}

func TestUploadModular_PreUploadObjectFailure1(t *testing.T) {
	t.Log("Failure case description: task repeated")
	u := setup(t)
	ctrl := gomock.NewController(t)

	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.uploadQueue = m
	m.EXPECT().Has(gomock.Any()).Return(true).Times(1)

	uploadObjectTask := &gfsptask.GfSpUploadObjectTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_CREATED},
		StorageParams: &storagetypes.Params{MaxPayloadSize: 1},
	}
	err := u.PreUploadObject(context.TODO(), uploadObjectTask)
	assert.Equal(t, ErrRepeatedTask, err)
}

func TestUploadModular_PreUploadObjectFailure2(t *testing.T) {
	t.Log("Failure case description: failed to begin upload object task")
	u := setup(t)
	ctrl := gomock.NewController(t)

	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.uploadQueue = m
	m.EXPECT().Has(gomock.Any()).Return(false).Times(1)

	m1 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m1)
	m1.EXPECT().CreateUploadObject(gomock.Any(), gomock.Any()).Return(mockErr).Times(1)

	uploadObjectTask := &gfsptask.GfSpUploadObjectTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_CREATED},
		StorageParams: &storagetypes.Params{MaxPayloadSize: 1},
	}
	err := u.PreUploadObject(context.TODO(), uploadObjectTask)
	assert.Equal(t, mockErr, err)
}

func TestUploadModular_HandleUploadObjectTaskSuccess1(t *testing.T) {
	t.Log("Success case description: succeed to upload payload to piece store")
	u := setup(t)
	ctrl := gomock.NewController(t)

	m := gfspclient.NewMockstdLib(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(0, io.EOF).AnyTimes()

	m1 := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.uploadQueue = m1
	m1.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	m1.EXPECT().PopByKey(gomock.Any()).Return(&gfsptask.GfSpUploadObjectTask{}).AnyTimes()

	m2 := piecestore.NewMockPieceOp(ctrl)
	u.baseApp.SetPieceOp(m2)
	m2.EXPECT().MaxSegmentPieceSize(gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)

	m3 := corespdb.NewMockSPDB(ctrl)
	u.baseApp.SetGfSpDB(m3)
	m3.EXPECT().SetObjectIntegrity(gomock.Any()).Return(nil).AnyTimes()

	m4 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m4)
	m4.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	uploadObjectTask := &gfsptask.GfSpUploadObjectTask{
		Task: &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id: sdkmath.NewUint(1),
			Checksums: [][]byte{{227, 176, 196, 66, 152, 252, 28, 20, 154, 251, 244, 200, 153, 111, 185, 36, 39, 174, 65,
				228, 100, 155, 147, 76, 164, 149, 153, 27, 120, 82, 184, 85}},
		},
		StorageParams: &storagetypes.Params{},
	}
	err := u.HandleUploadObjectTask(context.TODO(), uploadObjectTask, m)
	assert.Nil(t, err)
}

func TestUploadModular_HandleUploadObjectTaskFailure1(t *testing.T) {
	t.Log("Failure case description: failed to push upload queue")
	u := setup(t)
	ctrl := gomock.NewController(t)

	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.uploadQueue = m
	m.EXPECT().Push(gomock.Any()).Return(mockErr).Times(1)
	m.EXPECT().PopByKey(gomock.Any()).Return(&gfsptask.GfSpUploadObjectTask{}).AnyTimes()

	err := u.HandleUploadObjectTask(context.TODO(), &gfsptask.GfSpUploadObjectTask{}, nil)
	assert.Equal(t, mockErr, err)
}

func TestUploadModular_HandleUploadObjectTaskFailure2(t *testing.T) {
	t.Log("Failure case description: failed to put segment piece to piece store")
	u := setup(t)
	ctrl := gomock.NewController(t)

	m := gfspclient.NewMockstdLib(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(1, io.EOF).AnyTimes()

	m1 := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.uploadQueue = m1
	m1.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	m1.EXPECT().PopByKey(gomock.Any()).Return(&gfsptask.GfSpUploadObjectTask{}).AnyTimes()

	m2 := piecestore.NewMockPieceOp(ctrl)
	u.baseApp.SetPieceOp(m2)
	m2.EXPECT().MaxSegmentPieceSize(gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
	m2.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").AnyTimes()

	m3 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m3)
	m3.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	m4 := piecestore.NewMockPieceStore(ctrl)
	u.baseApp.SetPieceStore(m4)
	m4.EXPECT().PutPiece(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockErr).AnyTimes()

	uploadObjectTask := &gfsptask.GfSpUploadObjectTask{
		Task:          &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo:    &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
		StorageParams: &storagetypes.Params{},
	}
	err := u.HandleUploadObjectTask(context.TODO(), uploadObjectTask, m)
	assert.Contains(t, err.Error(), mockErr.Error())
}

func TestUploadModular_HandleUploadObjectTaskFailure3(t *testing.T) {
	t.Log("Failure case description: invalid integrity hash")
	u := setup(t)
	ctrl := gomock.NewController(t)

	m := gfspclient.NewMockstdLib(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(0, io.EOF).AnyTimes()

	m1 := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.uploadQueue = m1
	m1.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	m1.EXPECT().PopByKey(gomock.Any()).Return(&gfsptask.GfSpUploadObjectTask{}).AnyTimes()

	m2 := piecestore.NewMockPieceOp(ctrl)
	u.baseApp.SetPieceOp(m2)
	m2.EXPECT().MaxSegmentPieceSize(gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)

	m3 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m3)
	m3.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	uploadObjectTask := &gfsptask.GfSpUploadObjectTask{
		Task:          &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo:    &storagetypes.ObjectInfo{Checksums: [][]byte{[]byte("test")}},
		StorageParams: &storagetypes.Params{},
	}
	err := u.HandleUploadObjectTask(context.TODO(), uploadObjectTask, m)
	assert.Equal(t, ErrInvalidIntegrity, err)
}

func TestUploadModular_HandleUploadObjectTaskFailure4(t *testing.T) {
	t.Log("Failure case description: failed to write integrity hash to db")
	u := setup(t)
	ctrl := gomock.NewController(t)

	m := gfspclient.NewMockstdLib(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(0, io.EOF).AnyTimes()

	m1 := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.uploadQueue = m1
	m1.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	m1.EXPECT().PopByKey(gomock.Any()).Return(&gfsptask.GfSpUploadObjectTask{}).AnyTimes()

	m2 := piecestore.NewMockPieceOp(ctrl)
	u.baseApp.SetPieceOp(m2)
	m2.EXPECT().MaxSegmentPieceSize(gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)

	m3 := corespdb.NewMockSPDB(ctrl)
	u.baseApp.SetGfSpDB(m3)
	m3.EXPECT().SetObjectIntegrity(gomock.Any()).Return(mockErr).AnyTimes()

	m4 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m4)
	m4.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	uploadObjectTask := &gfsptask.GfSpUploadObjectTask{
		Task: &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id: sdkmath.NewUint(1),
			Checksums: [][]byte{{227, 176, 196, 66, 152, 252, 28, 20, 154, 251, 244, 200, 153, 111, 185, 36, 39, 174, 65,
				228, 100, 155, 147, 76, 164, 149, 153, 27, 120, 82, 184, 85}},
		},
		StorageParams: &storagetypes.Params{},
	}
	err := u.HandleUploadObjectTask(context.TODO(), uploadObjectTask, m)
	assert.Contains(t, err.Error(), mockErr.Error())
}

func TestUploadModular_HandleUploadObjectTaskFailure5(t *testing.T) {
	t.Log("Failure case description: stream closed abnormally")
	u := setup(t)
	ctrl := gomock.NewController(t)

	m := gfspclient.NewMockstdLib(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(0, mockErr).AnyTimes()

	m1 := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.uploadQueue = m1
	m1.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	m1.EXPECT().PopByKey(gomock.Any()).Return(&gfsptask.GfSpUploadObjectTask{}).AnyTimes()

	m2 := piecestore.NewMockPieceOp(ctrl)
	u.baseApp.SetPieceOp(m2)
	m2.EXPECT().MaxSegmentPieceSize(gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)

	m3 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m3)
	m3.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	uploadObjectTask := &gfsptask.GfSpUploadObjectTask{
		Task: &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:        sdkmath.NewUint(1),
			Checksums: [][]byte{{227, 176, 196, 66, 152, 252, 28, 20, 154, 251, 244, 200, 153, 111, 185, 36, 39, 174, 65, 228, 100, 155, 147, 76, 164, 149, 153, 27, 120, 82, 184, 85}},
		},
		StorageParams: &storagetypes.Params{},
	}
	err := u.HandleUploadObjectTask(context.TODO(), uploadObjectTask, m)
	assert.Equal(t, ErrClosedStream, err)
}

func TestUploadModular_HandleUploadObjectTaskFailure6(t *testing.T) {
	t.Log("Failure case description: failed to put segment piece to piece store")
	u := setup(t)
	ctrl := gomock.NewController(t)

	m := gfspclient.NewMockstdLib(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(1, nil).AnyTimes()

	m1 := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.uploadQueue = m1
	m1.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	m1.EXPECT().PopByKey(gomock.Any()).Return(&gfsptask.GfSpUploadObjectTask{}).AnyTimes()

	m2 := piecestore.NewMockPieceOp(ctrl)
	u.baseApp.SetPieceOp(m2)
	m2.EXPECT().MaxSegmentPieceSize(gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
	m2.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").AnyTimes()

	m3 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m3)
	m3.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	m4 := piecestore.NewMockPieceStore(ctrl)
	u.baseApp.SetPieceStore(m4)
	m4.EXPECT().PutPiece(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockErr).AnyTimes()

	uploadObjectTask := &gfsptask.GfSpUploadObjectTask{
		Task: &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id: sdkmath.NewUint(1),
		},
		StorageParams: &storagetypes.Params{},
	}
	err := u.HandleUploadObjectTask(context.TODO(), uploadObjectTask, m)
	assert.Contains(t, err.Error(), mockErr.Error())
}

func TestUploadModular_PreResumableUploadObject(t *testing.T) {
	cases := []struct {
		name             string
		uploadObjectTask coretask.ResumableUploadObjectTask
		wantedErr        error
	}{
		{
			name:             "1",
			uploadObjectTask: nil,
			wantedErr:        ErrDanglingUploadTask,
		},
		{
			name: "2",
			uploadObjectTask: &gfsptask.GfSpResumableUploadObjectTask{
				ObjectInfo:    &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_DISCONTINUED},
				StorageParams: &storagetypes.Params{MaxPayloadSize: 1},
			},
			wantedErr: ErrNotCreatedState,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			u := setup(t)
			err := u.PreResumableUploadObject(context.TODO(), tt.uploadObjectTask)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestUploadModular_PreResumableUploadObjectSuccess(t *testing.T) {
	u := setup(t)
	ctrl := gomock.NewController(t)

	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.resumeableUploadQueue = m
	m.EXPECT().Has(gomock.Any()).Return(false).Times(1)

	m1 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m1)
	m1.EXPECT().CreateResumableUploadObject(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	uploadObjectTask := &gfsptask.GfSpResumableUploadObjectTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_CREATED},
		StorageParams: &storagetypes.Params{MaxPayloadSize: 1},
	}
	err := u.PreResumableUploadObject(context.TODO(), uploadObjectTask)
	assert.Nil(t, err)
}

func TestUploadModular_PreResumableUploadObjectFailure1(t *testing.T) {
	t.Log("Failure case description: task repeated")
	u := setup(t)
	ctrl := gomock.NewController(t)

	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.resumeableUploadQueue = m
	m.EXPECT().Has(gomock.Any()).Return(true).Times(1)

	uploadObjectTask := &gfsptask.GfSpResumableUploadObjectTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_CREATED},
		StorageParams: &storagetypes.Params{MaxPayloadSize: 1},
	}
	err := u.PreResumableUploadObject(context.TODO(), uploadObjectTask)
	assert.Equal(t, ErrRepeatedTask, err)
}

func TestUploadModular_PreResumableUploadObjectFailure2(t *testing.T) {
	t.Log("Failure case description: failed to begin upload object task")
	u := setup(t)
	ctrl := gomock.NewController(t)

	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.resumeableUploadQueue = m
	m.EXPECT().Has(gomock.Any()).Return(false).Times(1)

	m1 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m1)
	m1.EXPECT().CreateResumableUploadObject(gomock.Any(), gomock.Any()).Return(mockErr).Times(1)

	uploadObjectTask := &gfsptask.GfSpResumableUploadObjectTask{
		ObjectInfo:    &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_CREATED},
		StorageParams: &storagetypes.Params{MaxPayloadSize: 1},
	}
	err := u.PreResumableUploadObject(context.TODO(), uploadObjectTask)
	assert.Equal(t, mockErr, err)
}

func TestUploadModular_HandleResumableUploadObjectTaskSuccess(t *testing.T) {
	u := setup(t)
	ctrl := gomock.NewController(t)

	m := gfspclient.NewMockstdLib(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(0, io.EOF).AnyTimes()

	m1 := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.resumeableUploadQueue = m1
	m1.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	m1.EXPECT().PopByKey(gomock.Any()).Return(&gfsptask.GfSpResumableUploadObjectTask{}).AnyTimes()

	m2 := piecestore.NewMockPieceOp(ctrl)
	u.baseApp.SetPieceOp(m2)
	m2.EXPECT().MaxSegmentPieceSize(gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)

	m3 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m3)
	m3.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	task := &gfsptask.GfSpResumableUploadObjectTask{
		Task: &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id: sdkmath.NewUint(1),
		},
		StorageParams: &storagetypes.Params{},
	}
	err := u.HandleResumableUploadObjectTask(context.TODO(), task, m)
	assert.Nil(t, err)
}

func TestUploadModular_HandleResumableUploadObjectTaskFailure1(t *testing.T) {
	t.Log("Failure case description: failed to push upload queue")
	u := setup(t)
	ctrl := gomock.NewController(t)

	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.resumeableUploadQueue = m
	m.EXPECT().Push(gomock.Any()).Return(mockErr).Times(1)
	m.EXPECT().PopByKey(gomock.Any()).Return(&gfsptask.GfSpResumableUploadObjectTask{}).AnyTimes()

	err := u.HandleResumableUploadObjectTask(context.TODO(), &gfsptask.GfSpResumableUploadObjectTask{}, nil)
	assert.Equal(t, mockErr, err)
}

func TestUploadModular_HandleResumableUploadObjectTaskFailure2(t *testing.T) {
	t.Log("Failure case description: failed to put segment piece to piece store")
	u := setup(t)
	ctrl := gomock.NewController(t)

	m := gfspclient.NewMockstdLib(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(1, io.EOF).AnyTimes()

	m1 := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.resumeableUploadQueue = m1
	m1.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	m1.EXPECT().PopByKey(gomock.Any()).Return(&gfsptask.GfSpResumableUploadObjectTask{}).AnyTimes()

	m2 := piecestore.NewMockPieceOp(ctrl)
	u.baseApp.SetPieceOp(m2)
	m2.EXPECT().MaxSegmentPieceSize(gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
	m2.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").AnyTimes()

	m3 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m3)
	m3.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	m4 := piecestore.NewMockPieceStore(ctrl)
	m4.EXPECT().PutPiece(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockErr).AnyTimes()
	u.baseApp.SetPieceStore(m4)

	task := &gfsptask.GfSpResumableUploadObjectTask{
		Task: &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id: sdkmath.NewUint(1),
		},
		StorageParams: &storagetypes.Params{},
	}
	err := u.HandleResumableUploadObjectTask(context.TODO(), task, m)
	assert.Contains(t, err.Error(), mockErr.Error())
}

func TestUploadModular_HandleResumableUploadObjectTaskFailure3(t *testing.T) {
	t.Log("Failure case description: failed to append integrity checksum to db")
	u := setup(t)
	ctrl := gomock.NewController(t)
	m := gfspclient.NewMockstdLib(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(1, io.EOF).AnyTimes()

	m1 := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.resumeableUploadQueue = m1
	m1.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	m1.EXPECT().PopByKey(gomock.Any()).Return(&gfsptask.GfSpResumableUploadObjectTask{}).AnyTimes()

	m2 := piecestore.NewMockPieceOp(ctrl)
	u.baseApp.SetPieceOp(m2)
	m2.EXPECT().MaxSegmentPieceSize(gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
	m2.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").AnyTimes()

	m3 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m3)
	m3.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	m4 := piecestore.NewMockPieceStore(ctrl)
	m4.EXPECT().PutPiece(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	u.baseApp.SetPieceStore(m4)

	m5 := corespdb.NewMockSPDB(ctrl)
	u.baseApp.SetGfSpDB(m5)
	m5.EXPECT().UpdatePieceChecksum(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockErr).AnyTimes()

	task := &gfsptask.GfSpResumableUploadObjectTask{
		Task: &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id: sdkmath.NewUint(1),
		},
		StorageParams: &storagetypes.Params{},
	}
	err := u.HandleResumableUploadObjectTask(context.TODO(), task, m)
	assert.Contains(t, err.Error(), mockErr.Error())
}

func TestUploadModular_HandleResumableUploadObjectTaskFailure4(t *testing.T) {
	t.Log("Failure case description: failed to get object integrity hash")
	u := setup(t)
	ctrl := gomock.NewController(t)
	m := gfspclient.NewMockstdLib(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(0, io.EOF).AnyTimes()

	m1 := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.resumeableUploadQueue = m1
	m1.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	m1.EXPECT().PopByKey(gomock.Any()).Return(&gfsptask.GfSpResumableUploadObjectTask{}).AnyTimes()

	m2 := piecestore.NewMockPieceOp(ctrl)
	u.baseApp.SetPieceOp(m2)
	m2.EXPECT().MaxSegmentPieceSize(gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)

	m3 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m3)
	m3.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	m4 := corespdb.NewMockSPDB(ctrl)
	u.baseApp.SetGfSpDB(m4)
	m4.EXPECT().GetObjectIntegrity(gomock.Any(), gomock.Any()).Return(nil, mockErr).AnyTimes()

	task := &gfsptask.GfSpResumableUploadObjectTask{
		Task: &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id: sdkmath.NewUint(1),
		},
		StorageParams: &storagetypes.Params{},
		Completed:     true,
	}
	err := u.HandleResumableUploadObjectTask(context.TODO(), task, m)
	assert.Equal(t, err, mockErr)
}

func TestUploadModular_HandleResumableUploadObjectTaskFailure5(t *testing.T) {
	t.Log("Failure case description: invalid integrity hash")
	u := setup(t)
	ctrl := gomock.NewController(t)
	m := gfspclient.NewMockstdLib(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(0, io.EOF).AnyTimes()

	m1 := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.resumeableUploadQueue = m1
	m1.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	m1.EXPECT().PopByKey(gomock.Any()).Return(&gfsptask.GfSpResumableUploadObjectTask{}).AnyTimes()

	m2 := piecestore.NewMockPieceOp(ctrl)
	u.baseApp.SetPieceOp(m2)
	m2.EXPECT().MaxSegmentPieceSize(gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)

	m3 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m3)
	m3.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	m4 := corespdb.NewMockSPDB(ctrl)
	u.baseApp.SetGfSpDB(m4)
	m4.EXPECT().GetObjectIntegrity(gomock.Any(), gomock.Any()).Return(&corespdb.IntegrityMeta{
		PieceChecksumList: [][]byte{[]byte("test")}}, nil).AnyTimes()

	task := &gfsptask.GfSpResumableUploadObjectTask{
		Task: &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:        sdkmath.NewUint(1),
			Checksums: [][]byte{[]byte("mock")},
		},
		StorageParams: &storagetypes.Params{},
		Completed:     true,
	}
	err := u.HandleResumableUploadObjectTask(context.TODO(), task, m)
	assert.Equal(t, ErrInvalidIntegrity, err)
}

func TestUploadModular_HandleResumableUploadObjectTaskFailure6(t *testing.T) {
	t.Log("Failure case description: failed to write integrity hash to db")
	u := setup(t)
	ctrl := gomock.NewController(t)
	m := gfspclient.NewMockstdLib(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(0, io.EOF).AnyTimes()

	m1 := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.resumeableUploadQueue = m1
	m1.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	m1.EXPECT().PopByKey(gomock.Any()).Return(&gfsptask.GfSpResumableUploadObjectTask{}).AnyTimes()

	m2 := piecestore.NewMockPieceOp(ctrl)
	u.baseApp.SetPieceOp(m2)
	m2.EXPECT().MaxSegmentPieceSize(gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)

	m3 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m3)
	m3.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	m4 := corespdb.NewMockSPDB(ctrl)
	u.baseApp.SetGfSpDB(m4)
	m4.EXPECT().GetObjectIntegrity(gomock.Any(), gomock.Any()).Return(&corespdb.IntegrityMeta{
		PieceChecksumList: [][]byte{[]byte("test")}}, nil).AnyTimes()
	m4.EXPECT().UpdateIntegrityChecksum(gomock.Any()).Return(mockErr).AnyTimes()

	task := &gfsptask.GfSpResumableUploadObjectTask{
		Task: &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id: sdkmath.NewUint(1),
			Checksums: [][]byte{{159, 134, 208, 129, 136, 76, 125, 101, 154, 47, 234, 160, 197, 90, 208, 21, 163, 191,
				79, 27, 43, 11, 130, 44, 209, 93, 108, 21, 176, 240, 10, 8}},
		},
		StorageParams: &storagetypes.Params{},
		Completed:     true,
	}
	err := u.HandleResumableUploadObjectTask(context.TODO(), task, m)
	assert.Contains(t, err.Error(), mockErr.Error())
}

func TestUploadModular_HandleResumableUploadObjectTaskFailure7(t *testing.T) {
	t.Log("Failure case description: stream closed abnormally")
	u := setup(t)
	ctrl := gomock.NewController(t)
	m := gfspclient.NewMockstdLib(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(0, mockErr).AnyTimes()

	m1 := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.resumeableUploadQueue = m1
	m1.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	m1.EXPECT().PopByKey(gomock.Any()).Return(&gfsptask.GfSpResumableUploadObjectTask{}).AnyTimes()

	m2 := piecestore.NewMockPieceOp(ctrl)
	u.baseApp.SetPieceOp(m2)
	m2.EXPECT().MaxSegmentPieceSize(gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)

	task := &gfsptask.GfSpResumableUploadObjectTask{
		Task: &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id: sdkmath.NewUint(1),
		},
		StorageParams: &storagetypes.Params{},
	}
	err := u.HandleResumableUploadObjectTask(context.TODO(), task, m)
	assert.Equal(t, ErrClosedStream, err)
}

func TestUploadModular_HandleResumableUploadObjectTaskFailure8(t *testing.T) {
	t.Log("Failure case description: failed to put segment piece to piece store")
	u := setup(t)
	ctrl := gomock.NewController(t)
	m := gfspclient.NewMockstdLib(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(1, nil).AnyTimes()

	m1 := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.resumeableUploadQueue = m1
	m1.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	m1.EXPECT().PopByKey(gomock.Any()).Return(&gfsptask.GfSpResumableUploadObjectTask{}).AnyTimes()

	m2 := piecestore.NewMockPieceOp(ctrl)
	u.baseApp.SetPieceOp(m2)
	m2.EXPECT().MaxSegmentPieceSize(gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
	m2.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").AnyTimes()

	m3 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m3)
	m3.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	m4 := piecestore.NewMockPieceStore(ctrl)
	u.baseApp.SetPieceStore(m4)
	m4.EXPECT().PutPiece(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockErr).AnyTimes()

	task := &gfsptask.GfSpResumableUploadObjectTask{
		Task: &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id: sdkmath.NewUint(1),
		},
		StorageParams: &storagetypes.Params{},
	}
	err := u.HandleResumableUploadObjectTask(context.TODO(), task, m)
	assert.Contains(t, err.Error(), mockErr.Error())
}

func TestUploadModular_HandleResumableUploadObjectTaskFailure9(t *testing.T) {
	t.Log("Failure case description: failed to append integrity checksum to db")
	u := setup(t)
	ctrl := gomock.NewController(t)
	m := gfspclient.NewMockstdLib(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(1, nil).AnyTimes()

	m1 := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.resumeableUploadQueue = m1
	m1.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	m1.EXPECT().PopByKey(gomock.Any()).Return(&gfsptask.GfSpResumableUploadObjectTask{}).AnyTimes()

	m2 := piecestore.NewMockPieceOp(ctrl)
	u.baseApp.SetPieceOp(m2)
	m2.EXPECT().MaxSegmentPieceSize(gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
	m2.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").AnyTimes()

	m3 := gfspclient.NewMockGfSpClientAPI(ctrl)
	u.baseApp.SetGfSpClient(m3)
	m3.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	m4 := piecestore.NewMockPieceStore(ctrl)
	u.baseApp.SetPieceStore(m4)
	m4.EXPECT().PutPiece(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	m5 := corespdb.NewMockSPDB(ctrl)
	u.baseApp.SetGfSpDB(m5)
	m5.EXPECT().UpdatePieceChecksum(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockErr).AnyTimes()

	task := &gfsptask.GfSpResumableUploadObjectTask{
		Task: &gfsptask.GfSpTask{TaskPriority: 1},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id: sdkmath.NewUint(1),
		},
		StorageParams: &storagetypes.Params{},
	}
	err := u.HandleResumableUploadObjectTask(context.TODO(), task, m)
	assert.Contains(t, err.Error(), mockErr.Error())
}

func TestUploadModular_PostUploadObject(t *testing.T) {
	u := setup(t)
	u.PostUploadObject(context.TODO(), &gfsptask.GfSpUploadObjectTask{})
}

func TestUploadModular_QueryTasks(t *testing.T) {
	u := setup(t)
	ctrl := gomock.NewController(t)

	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	u.uploadQueue = m
	m.EXPECT().ScanTask(gomock.Any()).AnyTimes()

	result, err := u.QueryTasks(context.TODO(), "111")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(result))
}

func TestUploadModular_PostResumableUploadObject(t *testing.T) {
	u := setup(t)
	u.PostResumableUploadObject(context.TODO(), &gfsptask.GfSpResumableUploadObjectTask{})
}

func TestStreamReadAt(t *testing.T) {
	cases := []struct {
		name         string
		stream       io.Reader
		b            []byte
		wantedResult int
		wantedErr    error
	}{
		{
			name:         "1",
			stream:       nil,
			b:            nil,
			wantedResult: 0,
			wantedErr:    errors.New(("failed to read due to invalid args")),
		},
		{
			name:         "2",
			stream:       strings.NewReader("mock"),
			b:            []byte("test"),
			wantedResult: 4,
			wantedErr:    nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StreamReadAt(tt.stream, tt.b)
			assert.Equal(t, tt.wantedResult, result)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

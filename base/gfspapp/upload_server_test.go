package gfspapp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
)

func TestGfSpBaseApp_GfSpUploadObjectSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := NewMockgRPCUploadStream(ctrl)
	m.EXPECT().SendAndClose(gomock.Any()).Return(nil).AnyTimes()
	m.EXPECT().Recv().Return(&gfspserver.GfSpUploadObjectRequest{
		UploadObjectTask: &gfsptask.GfSpUploadObjectTask{
			Task:       &gfsptask.GfSpTask{},
			ObjectInfo: mockObjectInfo,
		},
	}, nil)

	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()

	m2 := coremodule.NewMockUploader(ctrl)
	g.uploader = m2
	m2.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(m1, nil).AnyTimes()
	m2.EXPECT().PreUploadObject(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().HandleUploadObjectTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().PostUploadObject(gomock.Any(), gomock.Any()).AnyTimes()

	err := g.GfSpUploadObject(m)
	assert.Nil(t, err)
}

func TestGfSpBaseApp_GfSpUploadObjectFailure1(t *testing.T) {
	t.Log("Failure case description: failed to close upload object stream")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := NewMockgRPCUploadStream(ctrl)
	m.EXPECT().SendAndClose(gomock.Any()).Return(mockErr).AnyTimes()
	m.EXPECT().Recv().Return(&gfspserver.GfSpUploadObjectRequest{
		UploadObjectTask: &gfsptask.GfSpUploadObjectTask{
			Task:       &gfsptask.GfSpTask{},
			ObjectInfo: mockObjectInfo,
		},
		Payload: mockSig,
	}, nil)

	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()

	m2 := coremodule.NewMockUploader(ctrl)
	g.uploader = m2
	m2.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(m1, nil).AnyTimes()
	m2.EXPECT().PreUploadObject(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().HandleUploadObjectTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().PostUploadObject(gomock.Any(), gomock.Any()).AnyTimes()

	err := g.GfSpUploadObject(m)
	assert.Nil(t, err)
}

func TestGfSpBaseApp_GfSpUploadObjectFailure2(t *testing.T) {
	t.Log("Failure case description: failed to receive object")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := NewMockgRPCUploadStream(ctrl)
	m.EXPECT().SendAndClose(gomock.Any()).Return(nil).AnyTimes()
	m.EXPECT().Recv().Return(nil, mockErr).AnyTimes()

	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()

	err := g.GfSpUploadObject(m)
	assert.Equal(t, ErrExceptionsStream, err)
}

func TestGfSpBaseApp_GfSpUploadObjectFailure3(t *testing.T) {
	t.Log("Failure case description: upload object task pointer dangling")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := NewMockgRPCUploadStream(ctrl)
	m.EXPECT().SendAndClose(gomock.Any()).Return(nil).AnyTimes()
	m.EXPECT().Recv().Return(nil, nil)

	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()

	m2 := coremodule.NewMockUploader(ctrl)
	g.uploader = m2
	m2.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(m1, nil).AnyTimes()
	m2.EXPECT().PreUploadObject(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().HandleUploadObjectTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().PostUploadObject(gomock.Any(), gomock.Any()).AnyTimes()

	err := g.GfSpUploadObject(m)
	assert.Equal(t, ErrUploadObjectDangling, err)
}

func TestGfSpBaseApp_GfSpUploadObjectFailure4(t *testing.T) {
	t.Log("Failure case description: failed to reserve resource")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := NewMockgRPCUploadStream(ctrl)
	m.EXPECT().SendAndClose(gomock.Any()).Return(nil).AnyTimes()
	m.EXPECT().Recv().Return(&gfspserver.GfSpUploadObjectRequest{
		UploadObjectTask: &gfsptask.GfSpUploadObjectTask{
			Task:       &gfsptask.GfSpTask{},
			ObjectInfo: mockObjectInfo,
		},
		Payload: mockSig,
	}, nil).AnyTimes()

	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()

	m2 := coremodule.NewMockUploader(ctrl)
	g.uploader = m2
	m2.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(nil, mockErr).AnyTimes()
	m2.EXPECT().PreUploadObject(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().HandleUploadObjectTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().PostUploadObject(gomock.Any(), gomock.Any()).AnyTimes()

	err := g.GfSpUploadObject(m)
	assert.Equal(t, ErrUploadExhaustResource, err)
}

func TestGfSpBaseApp_GfSpUploadObjectFailure5(t *testing.T) {
	t.Log("Failure case description: failed to pre upload object")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := NewMockgRPCUploadStream(ctrl)
	m.EXPECT().SendAndClose(gomock.Any()).Return(nil).AnyTimes()
	m.EXPECT().Recv().Return(&gfspserver.GfSpUploadObjectRequest{
		UploadObjectTask: &gfsptask.GfSpUploadObjectTask{
			Task:       &gfsptask.GfSpTask{},
			ObjectInfo: mockObjectInfo,
		},
		Payload: mockSig,
	}, nil).AnyTimes()

	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()

	m2 := coremodule.NewMockUploader(ctrl)
	g.uploader = m2
	m2.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(m1, nil).AnyTimes()
	m2.EXPECT().PreUploadObject(gomock.Any(), gomock.Any()).Return(mockErr).AnyTimes()
	m2.EXPECT().HandleUploadObjectTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().PostUploadObject(gomock.Any(), gomock.Any()).AnyTimes()

	err := g.GfSpUploadObject(m)
	assert.Equal(t, mockErr, err)
}

func TestGfSpBaseApp_GfSpUploadObjectFailure6(t *testing.T) {
	t.Log("Failure case description: failed to upload object data")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := NewMockgRPCUploadStream(ctrl)
	m.EXPECT().SendAndClose(gomock.Any()).Return(nil).AnyTimes()
	m.EXPECT().Recv().Return(&gfspserver.GfSpUploadObjectRequest{
		UploadObjectTask: &gfsptask.GfSpUploadObjectTask{
			Task:       &gfsptask.GfSpTask{},
			ObjectInfo: mockObjectInfo,
		},
		Payload: mockSig,
	}, nil).AnyTimes()

	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()

	m2 := coremodule.NewMockUploader(ctrl)
	g.uploader = m2
	m2.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(m1, nil).AnyTimes()
	m2.EXPECT().PreUploadObject(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().HandleUploadObjectTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockErr).AnyTimes()
	m2.EXPECT().PostUploadObject(gomock.Any(), gomock.Any()).AnyTimes()

	err := g.GfSpUploadObject(m)
	assert.Equal(t, mockErr, err)
}

func TestGfSpBaseApp_GfSpResumableUploadObjectSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := NewMockgRPCResumableUploadStream(ctrl)
	m.EXPECT().SendAndClose(gomock.Any()).Return(nil).AnyTimes()
	m.EXPECT().Recv().Return(&gfspserver.GfSpResumableUploadObjectRequest{
		ResumableUploadObjectTask: &gfsptask.GfSpResumableUploadObjectTask{
			Task:       &gfsptask.GfSpTask{},
			ObjectInfo: mockObjectInfo,
		},
	}, nil)

	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()

	m2 := coremodule.NewMockUploader(ctrl)
	g.uploader = m2
	m2.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(m1, nil).AnyTimes()
	m2.EXPECT().PreResumableUploadObject(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().HandleResumableUploadObjectTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().PostResumableUploadObject(gomock.Any(), gomock.Any()).AnyTimes()

	err := g.GfSpResumableUploadObject(m)
	assert.Nil(t, err)
}

func TestGfSpBaseApp_GfSpResumableUploadObjectFailure1(t *testing.T) {
	t.Log("Failure case description: failed to close resumable upload object stream")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := NewMockgRPCResumableUploadStream(ctrl)
	m.EXPECT().SendAndClose(gomock.Any()).Return(mockErr).AnyTimes()
	m.EXPECT().Recv().Return(&gfspserver.GfSpResumableUploadObjectRequest{
		ResumableUploadObjectTask: &gfsptask.GfSpResumableUploadObjectTask{
			Task:       &gfsptask.GfSpTask{},
			ObjectInfo: mockObjectInfo,
		},
		Payload: mockSig,
	}, nil)

	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()

	m2 := coremodule.NewMockUploader(ctrl)
	g.uploader = m2
	m2.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(m1, nil).AnyTimes()
	m2.EXPECT().PreResumableUploadObject(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().HandleResumableUploadObjectTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().PostResumableUploadObject(gomock.Any(), gomock.Any()).AnyTimes()

	err := g.GfSpResumableUploadObject(m)
	assert.Nil(t, err)
}

func TestGfSpBaseApp_GfSpResumableUploadObjectFailure2(t *testing.T) {
	t.Log("Failure case description: failed to receive resumable object")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := NewMockgRPCResumableUploadStream(ctrl)
	m.EXPECT().SendAndClose(gomock.Any()).Return(nil).AnyTimes()
	m.EXPECT().Recv().Return(nil, mockErr).AnyTimes()

	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()

	err := g.GfSpResumableUploadObject(m)
	assert.Equal(t, ErrExceptionsStream, err)
}

func TestGfSpBaseApp_GfSpResumableUploadObjectFailure3(t *testing.T) {
	t.Log("Failure case description: resumable upload object task pointer dangling")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := NewMockgRPCResumableUploadStream(ctrl)
	m.EXPECT().SendAndClose(gomock.Any()).Return(nil).AnyTimes()
	m.EXPECT().Recv().Return(nil, nil)

	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()

	m2 := coremodule.NewMockUploader(ctrl)
	g.uploader = m2
	m2.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(m1, nil).AnyTimes()
	m2.EXPECT().PreResumableUploadObject(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().HandleResumableUploadObjectTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().PostResumableUploadObject(gomock.Any(), gomock.Any()).AnyTimes()

	err := g.GfSpResumableUploadObject(m)
	assert.Equal(t, ErrUploadObjectDangling, err)
}

func TestGfSpBaseApp_GfSpResumableUploadObjectFailure4(t *testing.T) {
	t.Log("Failure case description: failed to reserve resource")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := NewMockgRPCResumableUploadStream(ctrl)
	m.EXPECT().SendAndClose(gomock.Any()).Return(nil).AnyTimes()
	m.EXPECT().Recv().Return(&gfspserver.GfSpResumableUploadObjectRequest{
		ResumableUploadObjectTask: &gfsptask.GfSpResumableUploadObjectTask{
			Task:       &gfsptask.GfSpTask{},
			ObjectInfo: mockObjectInfo,
		},
		Payload: mockSig,
	}, nil).AnyTimes()

	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()

	m2 := coremodule.NewMockUploader(ctrl)
	g.uploader = m2
	m2.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(nil, mockErr).AnyTimes()
	m2.EXPECT().PreResumableUploadObject(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().HandleResumableUploadObjectTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().PostResumableUploadObject(gomock.Any(), gomock.Any()).AnyTimes()

	err := g.GfSpResumableUploadObject(m)
	assert.Equal(t, ErrUploadExhaustResource, err)
}

func TestGfSpBaseApp_GfSpResumableUploadObject5(t *testing.T) {
	t.Log("Failure case description: failed to pre resumable upload object")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := NewMockgRPCResumableUploadStream(ctrl)
	m.EXPECT().SendAndClose(gomock.Any()).Return(nil).AnyTimes()
	m.EXPECT().Recv().Return(&gfspserver.GfSpResumableUploadObjectRequest{
		ResumableUploadObjectTask: &gfsptask.GfSpResumableUploadObjectTask{
			Task:       &gfsptask.GfSpTask{},
			ObjectInfo: mockObjectInfo,
		},
		Payload: mockSig,
	}, nil).AnyTimes()

	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()

	m2 := coremodule.NewMockUploader(ctrl)
	g.uploader = m2
	m2.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(m1, nil).AnyTimes()
	m2.EXPECT().PreResumableUploadObject(gomock.Any(), gomock.Any()).Return(mockErr).AnyTimes()
	m2.EXPECT().HandleResumableUploadObjectTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().PostResumableUploadObject(gomock.Any(), gomock.Any()).AnyTimes()

	err := g.GfSpResumableUploadObject(m)
	assert.Equal(t, mockErr, err)
}

func TestGfSpBaseApp_GfSpResumableUploadObjectFailure6(t *testing.T) {
	t.Log("Failure case description: failed to resumable upload object data")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := NewMockgRPCResumableUploadStream(ctrl)
	m.EXPECT().SendAndClose(gomock.Any()).Return(nil).AnyTimes()
	m.EXPECT().Recv().Return(&gfspserver.GfSpResumableUploadObjectRequest{
		ResumableUploadObjectTask: &gfsptask.GfSpResumableUploadObjectTask{
			Task:       &gfsptask.GfSpTask{},
			ObjectInfo: mockObjectInfo,
		},
		Payload: mockSig,
	}, nil).AnyTimes()

	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()

	m2 := coremodule.NewMockUploader(ctrl)
	g.uploader = m2
	m2.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(m1, nil).AnyTimes()
	m2.EXPECT().PreResumableUploadObject(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m2.EXPECT().HandleResumableUploadObjectTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockErr).AnyTimes()
	m2.EXPECT().PostResumableUploadObject(gomock.Any(), gomock.Any()).AnyTimes()

	err := g.GfSpResumableUploadObject(m)
	assert.Equal(t, mockErr, err)
}

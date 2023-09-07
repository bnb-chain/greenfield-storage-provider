package gfspapp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
)

func TestGfSpBaseApp_GfSpDownloadObjectSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockDownloader(ctrl)
	g.downloader = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()
	m.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(m1, nil).Times(1)
	m.EXPECT().PreDownloadObject(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	m.EXPECT().HandleDownloadObjectTask(gomock.Any(), gomock.Any()).Return([]byte("mockData"), nil).Times(1)
	m.EXPECT().PostDownloadObject(gomock.Any(), gomock.Any()).Return().Times(1)
	req := &gfspserver.GfSpDownloadObjectRequest{DownloadObjectTask: &gfsptask.GfSpDownloadObjectTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		ObjectInfo: mockObjectInfo,
	}}
	result, err := g.GfSpDownloadObject(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, []byte("mockData"), result.GetData())
}

func TestGfSpBaseApp_GfSpDownloadObjectFailure1(t *testing.T) {
	t.Log("Failure case description: download task dangling")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockDownloader(ctrl)
	g.downloader = m
	result, err := g.GfSpDownloadObject(context.TODO(), nil)
	assert.Nil(t, err)
	assert.Equal(t, ErrDownloadTaskDangling, result.GetErr())
}

func TestGfSpBaseApp_GfSpDownloadObjectFailure2(t *testing.T) {
	t.Log("Failure case description: failed to reserve resource")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockDownloader(ctrl)
	g.downloader = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()
	m.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpDownloadObjectRequest{DownloadObjectTask: &gfsptask.GfSpDownloadObjectTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		ObjectInfo: mockObjectInfo,
	}}
	result, err := g.GfSpDownloadObject(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, ErrDownloadExhaustResource, result.GetErr())
}

func TestGfSpBaseApp_OnDownloadObjectTaskSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockDownloader(ctrl)
	g.downloader = m
	m.EXPECT().PreDownloadObject(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	m.EXPECT().HandleDownloadObjectTask(gomock.Any(), gomock.Any()).Return([]byte("mockData"), nil).Times(1)
	m.EXPECT().PostDownloadObject(gomock.Any(), gomock.Any()).Return().Times(1)
	downloadTask := &gfsptask.GfSpDownloadObjectTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		ObjectInfo: mockObjectInfo,
	}
	result, err := g.OnDownloadObjectTask(context.TODO(), downloadTask)
	assert.Nil(t, err)
	assert.Equal(t, []byte("mockData"), result)
}

func TestGfSpBaseApp_OnDownloadObjectTaskFailure1(t *testing.T) {
	t.Log("Failure case description: download object task pointer dangling")
	g := setup(t)
	result, err := g.OnDownloadObjectTask(context.TODO(), nil)
	assert.Equal(t, ErrDownloadTaskDangling, err)
	assert.Nil(t, result)
}

func TestGfSpBaseApp_OnDownloadObjectTaskFailure2(t *testing.T) {
	t.Log("Failure case description: failed to pre download object")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockDownloader(ctrl)
	g.downloader = m
	m.EXPECT().PreDownloadObject(gomock.Any(), gomock.Any()).Return(mockErr).Times(1)
	downloadTask := &gfsptask.GfSpDownloadObjectTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		ObjectInfo: mockObjectInfo,
	}
	result, err := g.OnDownloadObjectTask(context.TODO(), downloadTask)
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestGfSpBaseApp_OnDownloadObjectTaskFailure3(t *testing.T) {
	t.Log("Failure case description: failed to download object")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockDownloader(ctrl)
	g.downloader = m
	m.EXPECT().PreDownloadObject(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	m.EXPECT().HandleDownloadObjectTask(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	downloadTask := &gfsptask.GfSpDownloadObjectTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		ObjectInfo: mockObjectInfo,
	}
	result, err := g.OnDownloadObjectTask(context.TODO(), downloadTask)
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestGfSpBaseApp_GfSpDownloadPieceSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockDownloader(ctrl)
	g.downloader = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()
	m.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(m1, nil).Times(1)
	m.EXPECT().PreDownloadPiece(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	m.EXPECT().HandleDownloadPieceTask(gomock.Any(), gomock.Any()).Return([]byte("mockData"), nil).Times(1)
	m.EXPECT().PostDownloadPiece(gomock.Any(), gomock.Any()).Return().Times(1)
	req := &gfspserver.GfSpDownloadPieceRequest{DownloadPieceTask: &gfsptask.GfSpDownloadPieceTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		ObjectInfo: mockObjectInfo,
	}}
	result, err := g.GfSpDownloadPiece(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, []byte("mockData"), result.GetData())
}

func TestGfSpBaseApp_GfSpDownloadPieceFailure1(t *testing.T) {
	t.Log("Failure case description: download piece dangling")
	g := setup(t)
	result, err := g.GfSpDownloadPiece(context.TODO(), nil)
	assert.Nil(t, err)
	assert.Equal(t, ErrDownloadTaskDangling, result.GetErr())
}

func TestGfSpBaseApp_GfSpDownloadPieceFailure2(t *testing.T) {
	t.Log("Failure case description: failed to reserve resource")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockDownloader(ctrl)
	g.downloader = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()
	m.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpDownloadPieceRequest{DownloadPieceTask: &gfsptask.GfSpDownloadPieceTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		ObjectInfo: mockObjectInfo,
	}}
	result, err := g.GfSpDownloadPiece(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, ErrDownloadExhaustResource, result.GetErr())
}

func TestGfSpBaseApp_OnDownloadPieceTaskFailure1(t *testing.T) {
	t.Log("Failure case description: download piece task pointer dangling")
	g := setup(t)
	result, err := g.OnDownloadPieceTask(context.TODO(), nil)
	assert.Equal(t, ErrDownloadTaskDangling, err)
	assert.Nil(t, result)
}

func TestGfSpBaseApp_OnDownloadPieceTaskFailure2(t *testing.T) {
	t.Log("Failure case description: failed to pre download piece")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockDownloader(ctrl)
	g.downloader = m
	m.EXPECT().PreDownloadPiece(gomock.Any(), gomock.Any()).Return(mockErr).Times(1)
	downloadTask := &gfsptask.GfSpDownloadPieceTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		ObjectInfo: mockObjectInfo,
	}
	result, err := g.OnDownloadPieceTask(context.TODO(), downloadTask)
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestGfSpBaseApp_OnDownloadPieceTaskFailure3(t *testing.T) {
	t.Log("Failure case description: failed to download piece")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockDownloader(ctrl)
	g.downloader = m
	m.EXPECT().PreDownloadPiece(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	m.EXPECT().HandleDownloadPieceTask(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	downloadTask := &gfsptask.GfSpDownloadPieceTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		ObjectInfo: mockObjectInfo,
	}
	result, err := g.OnDownloadPieceTask(context.TODO(), downloadTask)
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestGfSpBaseApp_GfSpGetChallengeInfoSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockDownloader(ctrl)
	g.downloader = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()
	m.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(m1, nil).Times(1)
	m.EXPECT().PreChallengePiece(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	m.EXPECT().HandleChallengePiece(gomock.Any(), gomock.Any()).Return([]byte{1}, [][]byte{[]byte{2}}, []byte("mockData"),
		nil).Times(1)
	m.EXPECT().PostChallengePiece(gomock.Any(), gomock.Any()).Return().Times(1)
	req := &gfspserver.GfSpGetChallengeInfoRequest{ChallengePieceTask: &gfsptask.GfSpChallengePieceTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}}
	result, err := g.GfSpGetChallengeInfo(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, []byte("mockData"), result.GetData())
}

func TestGfSpBaseApp_GfSpGetChallengeInfoFailure1(t *testing.T) {
	t.Log("Failure case description: challenge piece dangling")
	g := setup(t)
	result, err := g.GfSpGetChallengeInfo(context.TODO(), nil)
	assert.Nil(t, err)
	assert.Equal(t, ErrDownloadTaskDangling, result.GetErr())
}

func TestGfSpBaseApp_GfSpGetChallengeInfoFailure2(t *testing.T) {
	t.Log("Failure case description: failed to reserve resource")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockDownloader(ctrl)
	g.downloader = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()
	m.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GfSpGetChallengeInfoRequest{ChallengePieceTask: &gfsptask.GfSpChallengePieceTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}}
	result, err := g.GfSpGetChallengeInfo(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, ErrDownloadExhaustResource, result.GetErr())
}

func TestGfSpBaseApp_OnChallengePieceTaskFailure1(t *testing.T) {
	t.Log("Failure case description: dangling pointer")
	g := setup(t)
	result1, result2, result3, err := g.OnChallengePieceTask(context.TODO(), nil)
	assert.Equal(t, ErrDownloadTaskDangling, err)
	assert.Nil(t, result1)
	assert.Nil(t, result2)
	assert.Nil(t, result3)
}

func TestGfSpBaseApp_OnChallengePieceTaskFailure2(t *testing.T) {
	t.Log("Failure case description: failed to pre challenge piece")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockDownloader(ctrl)
	g.downloader = m
	m.EXPECT().PreChallengePiece(gomock.Any(), gomock.Any()).Return(mockErr).Times(1)
	ta := &gfsptask.GfSpChallengePieceTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	result1, result2, result3, err := g.OnChallengePieceTask(context.TODO(), ta)
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result1)
	assert.Nil(t, result2)
	assert.Nil(t, result3)
}

func TestGfSpBaseApp_OnChallengePieceTaskFailure3(t *testing.T) {
	t.Log("Failure case description: failed to handle challenge piece")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockDownloader(ctrl)
	g.downloader = m
	m.EXPECT().PreChallengePiece(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	m.EXPECT().HandleChallengePiece(gomock.Any(), gomock.Any()).Return(nil, nil, nil, mockErr).Times(1)
	ta := &gfsptask.GfSpChallengePieceTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		ObjectInfo:    mockObjectInfo,
		StorageParams: mockStorageParams,
	}
	result1, result2, result3, err := g.OnChallengePieceTask(context.TODO(), ta)
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result1)
	assert.Nil(t, result2)
	assert.Nil(t, result3)
}

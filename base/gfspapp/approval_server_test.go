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
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func TestGfSpBaseApp_GfSpAskApprovalSuccess1(t *testing.T) {
	t.Log("Success case description: create bucket approval task type")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockApprover(ctrl)
	g.approver = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()
	m.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
			return m1, nil
		}).AnyTimes()
	m.EXPECT().PreCreateBucketApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateBucketTask) error { return nil }).AnyTimes()
	m.EXPECT().HandleCreateBucketApprovalTask(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateBucketTask) (bool, error) { return true, nil }).AnyTimes()
	m.EXPECT().PostCreateBucketApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateBucketTask) {}).AnyTimes()
	req := &gfspserver.GfSpAskApprovalRequest{Request: &gfspserver.GfSpAskApprovalRequest_CreateBucketApprovalTask{
		CreateBucketApprovalTask: &gfsptask.GfSpCreateBucketApprovalTask{
			Task: &gfsptask.GfSpTask{
				Address: "mockAddress",
			},
			CreateBucketInfo: &storagetypes.MsgCreateBucket{
				Creator:    "mockCreator",
				BucketName: "mockBucketName",
			},
		},
	}}
	result, err := g.GfSpAskApproval(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, true, result.GetAllowed())
}

func TestGfSpBaseApp_GfSpAskApprovalSuccess2(t *testing.T) {
	t.Log("Success case description: migrate bucket approval task type")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockApprover(ctrl)
	g.approver = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()
	m.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
			return m1, nil
		}).AnyTimes()
	m.EXPECT().PreMigrateBucketApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalMigrateBucketTask) error { return nil }).AnyTimes()
	m.EXPECT().HandleMigrateBucketApprovalTask(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalMigrateBucketTask) (bool, error) { return true, nil }).AnyTimes()
	m.EXPECT().PostMigrateBucketApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalMigrateBucketTask) {}).AnyTimes()
	req := &gfspserver.GfSpAskApprovalRequest{Request: &gfspserver.GfSpAskApprovalRequest_MigrateBucketApprovalTask{
		MigrateBucketApprovalTask: &gfsptask.GfSpMigrateBucketApprovalTask{
			Task: &gfsptask.GfSpTask{
				Address: "mockAddress",
			},
			MigrateBucketInfo: &storagetypes.MsgMigrateBucket{
				BucketName: "mockBucketName",
			},
		},
	}}
	result, err := g.GfSpAskApproval(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, true, result.GetAllowed())
}

func TestGfSpBaseApp_GfSpAskApprovalSuccess3(t *testing.T) {
	t.Log("Success case description: create object approval task type")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockApprover(ctrl)
	g.approver = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()
	m.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
			return m1, nil
		}).AnyTimes()
	m.EXPECT().PreCreateObjectApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateObjectTask) error { return nil }).AnyTimes()
	m.EXPECT().HandleCreateObjectApprovalTask(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateObjectTask) (bool, error) { return true, nil }).AnyTimes()
	m.EXPECT().PostCreateObjectApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateObjectTask) {}).AnyTimes()
	req := &gfspserver.GfSpAskApprovalRequest{Request: &gfspserver.GfSpAskApprovalRequest_CreateObjectApprovalTask{
		CreateObjectApprovalTask: &gfsptask.GfSpCreateObjectApprovalTask{
			Task: &gfsptask.GfSpTask{
				Address: "mockAddress",
			},
			CreateObjectInfo: &storagetypes.MsgCreateObject{
				BucketName: "mockBucketName",
			},
		},
	}}
	result, err := g.GfSpAskApproval(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, true, result.GetAllowed())
}

func TestGfSpBaseApp_GfSpAskApprovalFailure1(t *testing.T) {
	t.Log("Failure case description: ask approval request dangling error")
	g := setup(t)
	req := &gfspserver.GfSpAskApprovalRequest{Request: nil}
	result, err := g.GfSpAskApproval(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, result.GetErr(), ErrApprovalTaskDangling)
}

func TestGfSpBaseApp_GfSpAskApprovalFailure2(t *testing.T) {
	t.Log("Failure case description: mock create bucket approval reserve resource returns error")

	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockApprover(ctrl)
	g.approver = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()
	m.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
			return m1, mockErr
		}).AnyTimes()
	req := &gfspserver.GfSpAskApprovalRequest{Request: &gfspserver.GfSpAskApprovalRequest_CreateBucketApprovalTask{
		CreateBucketApprovalTask: &gfsptask.GfSpCreateBucketApprovalTask{
			Task: &gfsptask.GfSpTask{
				Address: "mockAddress",
			},
			CreateBucketInfo: &storagetypes.MsgCreateBucket{
				Creator:    "mockCreator",
				BucketName: "mockBucketName",
			},
		},
	}}
	result, err := g.GfSpAskApproval(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, result.GetErr(), ErrApprovalExhaustResource)
}

func TestGfSpBaseApp_GfSpAskApprovalFailure3(t *testing.T) {
	t.Log("Failure case description: mock migrate bucket approval reserve resource returns error")

	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockApprover(ctrl)
	g.approver = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()
	m.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
			return m1, mockErr
		}).AnyTimes()
	req := &gfspserver.GfSpAskApprovalRequest{Request: &gfspserver.GfSpAskApprovalRequest_MigrateBucketApprovalTask{
		MigrateBucketApprovalTask: &gfsptask.GfSpMigrateBucketApprovalTask{
			Task: &gfsptask.GfSpTask{
				Address: "mockAddress",
			},
			MigrateBucketInfo: &storagetypes.MsgMigrateBucket{
				BucketName: "mockBucketName",
			},
		},
	}}
	result, err := g.GfSpAskApproval(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, result.GetErr(), ErrApprovalExhaustResource)
}

func TestGfSpBaseApp_GfSpAskApprovalFailure4(t *testing.T) {
	t.Log("Failure case description: mock create object approval reserve resource returns error")

	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockApprover(ctrl)
	g.approver = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m1.EXPECT().Done().AnyTimes()
	m.EXPECT().ReserveResource(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
			return m1, mockErr
		}).AnyTimes()
	req := &gfspserver.GfSpAskApprovalRequest{Request: &gfspserver.GfSpAskApprovalRequest_CreateObjectApprovalTask{
		CreateObjectApprovalTask: &gfsptask.GfSpCreateObjectApprovalTask{
			Task: &gfsptask.GfSpTask{
				Address: "mockAddress",
			},
			CreateObjectInfo: &storagetypes.MsgCreateObject{
				Creator:    "mockCreator",
				BucketName: "mockBucketName",
			},
		}}}
	result, err := g.GfSpAskApproval(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, result.GetErr(), ErrApprovalExhaustResource)
}

func TestGfSpBaseApp_OnAskCreateBucketApprovalSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockApprover(ctrl)
	g.approver = m
	m.EXPECT().PreCreateBucketApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateBucketTask) error { return nil }).AnyTimes()
	m.EXPECT().HandleCreateBucketApprovalTask(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateBucketTask) (bool, error) { return true, nil }).AnyTimes()
	m.EXPECT().PostCreateBucketApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateBucketTask) {}).AnyTimes()
	approvalTask := &gfsptask.GfSpCreateBucketApprovalTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		CreateBucketInfo: &storagetypes.MsgCreateBucket{
			Creator: "mockCreator",
		},
	}
	result, err := g.OnAskCreateBucketApproval(context.TODO(), approvalTask)
	assert.Nil(t, err)
	assert.Equal(t, true, result)
}

func TestGfSpBaseApp_OnAskCreateBucketApprovalFailure1(t *testing.T) {
	t.Log("Failure case description: create bucket approval task dangling error")
	g := setup(t)
	result, err := g.OnAskCreateBucketApproval(context.TODO(), nil)
	assert.Equal(t, ErrApprovalTaskDangling, err)
	assert.Equal(t, false, result)
}

func TestGfSpBaseApp_OnAskCreateBucketApprovalFailure2(t *testing.T) {
	t.Log("Failure case description: mock PreCreateBucketApproval returns error")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockApprover(ctrl)
	g.approver = m
	m.EXPECT().PreCreateBucketApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateBucketTask) error { return mockErr }).AnyTimes()
	approvalTask := &gfsptask.GfSpCreateBucketApprovalTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		CreateBucketInfo: &storagetypes.MsgCreateBucket{
			Creator: "mockCreator",
		},
	}
	result, err := g.OnAskCreateBucketApproval(context.TODO(), approvalTask)
	assert.Equal(t, mockErr, err)
	assert.Equal(t, false, result)
}

func TestGfSpBaseApp_OnAskCreateBucketApprovalFailure3(t *testing.T) {
	t.Log("Failure case description: mock HandleCreateBucketApprovalTask returns error")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockApprover(ctrl)
	g.approver = m
	m.EXPECT().PreCreateBucketApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateBucketTask) error { return nil }).AnyTimes()
	m.EXPECT().HandleCreateBucketApprovalTask(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateBucketTask) (bool, error) { return false, mockErr }).AnyTimes()
	approvalTask := &gfsptask.GfSpCreateBucketApprovalTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		CreateBucketInfo: &storagetypes.MsgCreateBucket{
			Creator: "mockCreator",
		},
	}
	result, err := g.OnAskCreateBucketApproval(context.TODO(), approvalTask)
	assert.Equal(t, mockErr, err)
	assert.Equal(t, false, result)
}

func TestGfSpBaseApp_OnAskMigrateBucketApprovalSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockApprover(ctrl)
	g.approver = m
	m.EXPECT().PreMigrateBucketApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalMigrateBucketTask) error { return nil }).AnyTimes()
	m.EXPECT().HandleMigrateBucketApprovalTask(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalMigrateBucketTask) (bool, error) { return true, nil }).AnyTimes()
	m.EXPECT().PostMigrateBucketApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalMigrateBucketTask) {}).AnyTimes()
	approvalTask := &gfsptask.GfSpMigrateBucketApprovalTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		MigrateBucketInfo: &storagetypes.MsgMigrateBucket{
			Operator: "mockOperator",
		},
	}
	result, err := g.OnAskMigrateBucketApproval(context.TODO(), approvalTask)
	assert.Nil(t, err)
	assert.Equal(t, true, result)
}

func TestGfSpBaseApp_OnAskMigrateBucketApprovalFailure1(t *testing.T) {
	t.Log("Failure case description: migrate bucket approval task dangling error")
	g := setup(t)
	result, err := g.OnAskMigrateBucketApproval(context.TODO(), nil)
	assert.Equal(t, ErrApprovalTaskDangling, err)
	assert.Equal(t, false, result)
}

func TestGfSpBaseApp_OnAskMigrateBucketApprovalFailure2(t *testing.T) {
	t.Log("Failure case description: mock PreMigrateBucketApproval returns error")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockApprover(ctrl)
	g.approver = m
	m.EXPECT().PreMigrateBucketApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalMigrateBucketTask) error { return mockErr }).AnyTimes()
	approvalTask := &gfsptask.GfSpMigrateBucketApprovalTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		MigrateBucketInfo: &storagetypes.MsgMigrateBucket{
			Operator: "mockOperator",
		},
	}
	result, err := g.OnAskMigrateBucketApproval(context.TODO(), approvalTask)
	assert.Equal(t, mockErr, err)
	assert.Equal(t, false, result)
}

func TestGfSpBaseApp_OnAskMigrateBucketApprovalFailure3(t *testing.T) {
	t.Log("Failure case description: mock HandleMigrateBucketApprovalTask returns error")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockApprover(ctrl)
	g.approver = m
	m.EXPECT().PreMigrateBucketApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalMigrateBucketTask) error { return nil }).AnyTimes()
	m.EXPECT().HandleMigrateBucketApprovalTask(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalMigrateBucketTask) (bool, error) { return false, mockErr }).AnyTimes()
	approvalTask := &gfsptask.GfSpMigrateBucketApprovalTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		MigrateBucketInfo: &storagetypes.MsgMigrateBucket{
			Operator: "mockOperator",
		},
	}
	result, err := g.OnAskMigrateBucketApproval(context.TODO(), approvalTask)
	assert.Equal(t, mockErr, err)
	assert.Equal(t, false, result)
}

func TestGfSpBaseApp_OnAskCreateObjectApprovalSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockApprover(ctrl)
	g.approver = m
	m.EXPECT().PreCreateObjectApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateObjectTask) error { return nil }).AnyTimes()
	m.EXPECT().HandleCreateObjectApprovalTask(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateObjectTask) (bool, error) { return true, nil }).AnyTimes()
	m.EXPECT().PostCreateObjectApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateObjectTask) {}).AnyTimes()
	approvalTask := &gfsptask.GfSpCreateObjectApprovalTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			Creator: "mockCreator",
		},
	}
	result, err := g.OnAskCreateObjectApproval(context.TODO(), approvalTask)
	assert.Nil(t, err)
	assert.Equal(t, true, result)
}

func TestGfSpBaseApp_OnAskCreateObjectApprovalFailure1(t *testing.T) {
	t.Log("Failure case description: create object approval task dangling error")
	g := setup(t)
	result, err := g.OnAskCreateObjectApproval(context.TODO(), nil)
	assert.Equal(t, ErrApprovalTaskDangling, err)
	assert.Equal(t, false, result)
}

func TestGfSpBaseApp_OnAskCreateObjectApprovalFailure2(t *testing.T) {
	t.Log("Failure case description: mock PreCreateObjectApproval returns error")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockApprover(ctrl)
	g.approver = m
	m.EXPECT().PreCreateObjectApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateObjectTask) error { return mockErr }).AnyTimes()
	approvalTask := &gfsptask.GfSpCreateObjectApprovalTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			Creator: "mockCreator",
		},
	}
	result, err := g.OnAskCreateObjectApproval(context.TODO(), approvalTask)
	assert.Equal(t, mockErr, err)
	assert.Equal(t, false, result)
}

func TestGfSpBaseApp_OnAskCreateObjectApprovalFailure3(t *testing.T) {
	t.Log("Failure case description: mock HandleCreateObjectApprovalTask returns error")
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockApprover(ctrl)
	g.approver = m
	m.EXPECT().PreCreateObjectApproval(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateObjectTask) error { return nil }).AnyTimes()
	m.EXPECT().HandleCreateObjectApprovalTask(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task task.ApprovalCreateObjectTask) (bool, error) { return false, mockErr }).AnyTimes()
	approvalTask := &gfsptask.GfSpCreateObjectApprovalTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			Creator: "mockCreator",
		},
	}
	result, err := g.OnAskCreateObjectApproval(context.TODO(), approvalTask)
	assert.Equal(t, mockErr, err)
	assert.Equal(t, false, result)
}

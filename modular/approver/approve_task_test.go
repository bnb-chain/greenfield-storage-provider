package approver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield/types/common"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func TestErrSignerWithDetail(t *testing.T) {
	mock := "mockDetail"
	result := ErrSignerWithDetail(mock)
	assert.Equal(t, mock, result.Description)
}

func TestApprovalModular_PreCreateBucketApproval(t *testing.T) {
	a := setup(t)
	err := a.PreCreateBucketApproval(context.TODO(), nil)
	assert.Nil(t, err)
}

func TestApprovalModular_HandleCreateBucketApprovalTaskSuccess1(t *testing.T) {
	t.Log("Success case description: repeated create bucket approval task")
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	a.bucketQueue = m
	m.EXPECT().Has(gomock.Any()).Return(true).Times(1)
	m.EXPECT().PopByKey(gomock.Any()).DoAndReturn(func(coretask.TKey) coretask.Task {
		return &gfsptask.GfSpCreateBucketApprovalTask{
			CreateBucketInfo: &storagetypes.MsgCreateBucket{
				Creator:    "mockCreator",
				BucketName: "mockBucketName",
			},
		}
	})
	m.EXPECT().Push(gomock.Any()).DoAndReturn(func(coretask.Task) error { return nil }).AnyTimes()
	approvalTask := &gfsptask.GfSpCreateBucketApprovalTask{
		Task: &gfsptask.GfSpTask{
			Address: "mockAddress",
		},
		CreateBucketInfo: &storagetypes.MsgCreateBucket{
			Creator: "mockCreator",
		},
	}
	result, err := a.HandleCreateBucketApprovalTask(context.TODO(), approvalTask)
	assert.Nil(t, err)
	assert.Equal(t, true, result)
}

func TestApprovalModular_HandleCreateBucketApprovalTaskSuccess2(t *testing.T) {
	t.Log("Success case description: create bucket approval task")
	a := setup(t)
	a.accountBucketNumber = 10
	ctrl := gomock.NewController(t)
	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	a.bucketQueue = m
	m.EXPECT().Has(gomock.Any()).Return(false).AnyTimes()
	m1 := gfspclient.NewMockGfSpClientAPI(ctrl)
	a.baseApp.SetGfSpClient(m1)
	m1.EXPECT().GetUserBucketsCount(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, account string,
		includeRemoved bool, opts ...grpc.DialOption) (int64, error) {
		return 1, nil
	}).Times(1)
	m1.EXPECT().PickVirtualGroupFamilyID(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context,
		task coretask.ApprovalCreateBucketTask) (uint32, error) {
		return 10, nil
	}).Times(1)
	m1.EXPECT().SignCreateBucketApproval(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context,
		bucket *storagetypes.MsgCreateBucket) ([]byte, error) {
		return []byte("mockSig"), nil
	}).Times(1)
	m.EXPECT().Push(gomock.Any()).DoAndReturn(func(coretask.Task) error { return nil }).AnyTimes()

	approvalTask := &gfsptask.GfSpCreateBucketApprovalTask{
		Task: &gfsptask.GfSpTask{Address: "mockAddress"},
		CreateBucketInfo: &storagetypes.MsgCreateBucket{
			Creator:           "mockCreator",
			PrimarySpApproval: &common.Approval{},
		},
	}
	result, err := a.HandleCreateBucketApprovalTask(context.TODO(), approvalTask)
	assert.Nil(t, err)
	assert.Equal(t, true, result)
}

func TestApprovalModular_HandleCreateBucketApprovalTaskFailure1(t *testing.T) {
	t.Log("Failure case description: mock create bucket approval dangling returns error")
	a := setup(t)
	result, err := a.HandleCreateBucketApprovalTask(context.TODO(), nil)
	assert.Equal(t, ErrDanglingPointer, err)
	assert.Equal(t, false, result)
}

func TestApprovalModular_HandleCreateBucketApprovalTaskFailure2(t *testing.T) {
	t.Log("Failure case description: failed to get account owns max bucket number")
	a := setup(t)
	a.accountBucketNumber = 10
	ctrl := gomock.NewController(t)
	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	a.bucketQueue = m
	m.EXPECT().Has(gomock.Any()).Return(false).Times(1)
	m1 := gfspclient.NewMockGfSpClientAPI(ctrl)
	a.baseApp.SetGfSpClient(m1)
	m1.EXPECT().GetUserBucketsCount(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), mockErr).Times(1)
	approvalTask := &gfsptask.GfSpCreateBucketApprovalTask{
		Task: &gfsptask.GfSpTask{Address: "mockAddress"},
		CreateBucketInfo: &storagetypes.MsgCreateBucket{
			Creator:           "mockCreator",
			PrimarySpApproval: &common.Approval{},
		},
	}
	result, err := a.HandleCreateBucketApprovalTask(context.TODO(), approvalTask)
	assert.Equal(t, mockErr, err)
	assert.Equal(t, false, result)
}

func TestApprovalModular_HandleCreateBucketApprovalTaskFailure3(t *testing.T) {
	t.Log("Failure case description: account owns bucket number exceed")
	a := setup(t)
	a.accountBucketNumber = 1
	ctrl := gomock.NewController(t)
	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	a.bucketQueue = m
	m.EXPECT().Has(gomock.Any()).Return(false).Times(1)
	m1 := gfspclient.NewMockGfSpClientAPI(ctrl)
	a.baseApp.SetGfSpClient(m1)
	m1.EXPECT().GetUserBucketsCount(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(2), nil).Times(1)
	approvalTask := &gfsptask.GfSpCreateBucketApprovalTask{
		Task: &gfsptask.GfSpTask{Address: "mockAddress"},
		CreateBucketInfo: &storagetypes.MsgCreateBucket{
			Creator:           "mockCreator",
			PrimarySpApproval: &common.Approval{},
		},
	}
	result, err := a.HandleCreateBucketApprovalTask(context.TODO(), approvalTask)
	assert.Equal(t, ErrExceedBucketNumber, err)
	assert.Equal(t, false, result)
}

func TestApprovalModular_HandleCreateBucketApprovalTaskFailure4(t *testing.T) {
	t.Log("Failure case description: failed to pick virtual group family")
	a := setup(t)
	a.accountBucketNumber = 10
	ctrl := gomock.NewController(t)
	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	a.bucketQueue = m
	m.EXPECT().Has(gomock.Any()).Return(false).Times(1)
	m1 := gfspclient.NewMockGfSpClientAPI(ctrl)
	a.baseApp.SetGfSpClient(m1)
	m1.EXPECT().GetUserBucketsCount(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), nil).Times(1)
	m1.EXPECT().PickVirtualGroupFamilyID(gomock.Any(), gomock.Any()).Return(uint32(0), mockErr).Times(1)
	approvalTask := &gfsptask.GfSpCreateBucketApprovalTask{
		Task: &gfsptask.GfSpTask{Address: "mockAddress"},
		CreateBucketInfo: &storagetypes.MsgCreateBucket{
			Creator:           "mockCreator",
			PrimarySpApproval: &common.Approval{},
		},
	}
	result, err := a.HandleCreateBucketApprovalTask(context.TODO(), approvalTask)
	assert.Equal(t, mockErr, err)
	assert.Equal(t, false, result)
}

func TestApprovalModular_HandleCreateBucketApprovalTaskFailure5(t *testing.T) {
	t.Log("Failure case description: failed to sign the create bucket approval")
	a := setup(t)
	a.accountBucketNumber = 10
	ctrl := gomock.NewController(t)
	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	a.bucketQueue = m
	m.EXPECT().Has(gomock.Any()).Return(false).Times(1)
	m1 := gfspclient.NewMockGfSpClientAPI(ctrl)
	a.baseApp.SetGfSpClient(m1)
	m1.EXPECT().GetUserBucketsCount(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), nil).Times(1)
	m1.EXPECT().PickVirtualGroupFamilyID(gomock.Any(), gomock.Any()).Return(uint32(0), nil).Times(1)
	m1.EXPECT().SignCreateBucketApproval(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	approvalTask := &gfsptask.GfSpCreateBucketApprovalTask{
		Task: &gfsptask.GfSpTask{Address: "mockAddress"},
		CreateBucketInfo: &storagetypes.MsgCreateBucket{
			Creator:           "mockCreator",
			PrimarySpApproval: &common.Approval{},
		},
	}
	result, err := a.HandleCreateBucketApprovalTask(context.TODO(), approvalTask)
	assert.Contains(t, err.Error(), mockErr.Error())
	assert.Equal(t, false, result)
}

func TestApprovalModular_PostCreateBucketApproval(t *testing.T) {
	a := setup(t)
	a.PostCreateBucketApproval(context.TODO(), nil)
}

func TestApprovalModular_PreMigrateBucketApproval(t *testing.T) {
	a := setup(t)
	err := a.PreMigrateBucketApproval(context.TODO(), nil)
	assert.Nil(t, err)
}

func TestApprovalModular_HandleMigrateBucketApprovalTaskSuccess1(t *testing.T) {
	t.Log("Success case description: repeated migrate bucket approval task")
	a := setup(t)
	a.accountBucketNumber = 10
	ctrl := gomock.NewController(t)
	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	a.bucketQueue = m
	m.EXPECT().Has(gomock.Any()).Return(true).Times(1)
	m.EXPECT().PopByKey(gomock.Any()).DoAndReturn(func(coretask.TKey) coretask.Task {
		return &gfsptask.GfSpMigrateBucketApprovalTask{
			MigrateBucketInfo: &storagetypes.MsgMigrateBucket{
				BucketName: "mockBucketName",
			},
		}
	}).Times(1)
	m.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	req := &gfsptask.GfSpMigrateBucketApprovalTask{
		Task: &gfsptask.GfSpTask{},
		MigrateBucketInfo: &storagetypes.MsgMigrateBucket{
			BucketName: "mockBucketName",
		},
	}
	result, err := a.HandleMigrateBucketApprovalTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, true, result)
}

func TestApprovalModular_HandleMigrateBucketApprovalTaskSuccess2(t *testing.T) {
	t.Log("Success case description: migrate bucket approval task")
	a := setup(t)
	a.accountBucketNumber = 10
	ctrl := gomock.NewController(t)
	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	a.bucketQueue = m
	m.EXPECT().Has(gomock.Any()).Return(false).Times(1)
	m1 := gfspclient.NewMockGfSpClientAPI(ctrl)
	a.baseApp.SetGfSpClient(m1)
	m1.EXPECT().SignMigrateBucketApproval(gomock.Any(), gomock.Any()).Return([]byte("mockSig"), nil).Times(1)
	m.EXPECT().Push(gomock.Any()).Return(nil).AnyTimes()
	req := &gfsptask.GfSpMigrateBucketApprovalTask{
		Task: &gfsptask.GfSpTask{},
		MigrateBucketInfo: &storagetypes.MsgMigrateBucket{
			BucketName:           "mockBucketName",
			DstPrimarySpApproval: &common.Approval{},
		},
	}
	result, err := a.HandleMigrateBucketApprovalTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, true, result)
}

func TestApprovalModular_HandleMigrateBucketApprovalTaskFailure1(t *testing.T) {
	t.Log("Failure case description: dangling pointer")
	a := setup(t)
	result, err := a.HandleMigrateBucketApprovalTask(context.TODO(), nil)
	assert.Equal(t, ErrDanglingPointer, err)
	assert.Equal(t, false, result)
}

func TestApprovalModular_HandleMigrateBucketApprovalTaskFailure2(t *testing.T) {
	t.Log("Failure case description: failed to sign migrate bucket approval")
	a := setup(t)
	a.accountBucketNumber = 10
	ctrl := gomock.NewController(t)
	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	a.bucketQueue = m
	m.EXPECT().Has(gomock.Any()).Return(false).Times(1)
	m1 := gfspclient.NewMockGfSpClientAPI(ctrl)
	a.baseApp.SetGfSpClient(m1)
	m1.EXPECT().SignMigrateBucketApproval(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfsptask.GfSpMigrateBucketApprovalTask{
		Task: &gfsptask.GfSpTask{},
		MigrateBucketInfo: &storagetypes.MsgMigrateBucket{
			BucketName:           "mockBucketName",
			DstPrimarySpApproval: &common.Approval{},
		},
	}
	result, err := a.HandleMigrateBucketApprovalTask(context.TODO(), req)
	assert.Contains(t, err.Error(), mockErr.Error())
	assert.Equal(t, false, result)
}

func TestApprovalModular_PostMigrateBucketApproval(t *testing.T) {
	a := setup(t)
	a.PostMigrateBucketApproval(context.TODO(), nil)
}

func TestApprovalModular_PreCreateObjectApproval(t *testing.T) {
	a := setup(t)
	err := a.PreCreateObjectApproval(context.TODO(), nil)
	assert.Nil(t, err)
}

func TestApprovalModular_HandleCreateObjectApprovalTaskSuccess1(t *testing.T) {
	t.Log("Success case description: repeated create object approval task")
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	a.objectQueue = m
	m.EXPECT().Has(gomock.Any()).Return(true).Times(1)
	m1 := gfspclient.NewMockGfSpClientAPI(ctrl)
	a.baseApp.SetGfSpClient(m1)
	m.EXPECT().PopByKey(gomock.Any()).DoAndReturn(func(coretask.TKey) coretask.Task {
		return &gfsptask.GfSpCreateObjectApprovalTask{
			CreateObjectInfo: &storagetypes.MsgCreateObject{
				BucketName: "mockBucketName",
				ObjectName: "mockObjectName",
			},
		}
	}).Times(1)
	m.EXPECT().Push(gomock.Any()).Return(nil).Times(1)
	req := &gfsptask.GfSpCreateObjectApprovalTask{
		Task: &gfsptask.GfSpTask{},
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			BucketName: "mockBucketName",
			ObjectName: "mockObjectName",
		},
	}
	result, err := a.HandleCreateObjectApprovalTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, true, result)
}

func TestApprovalModular_HandleCreateObjectApprovalTaskSuccess2(t *testing.T) {
	t.Log("Success case description: create object approval task")
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	a.objectQueue = m
	m.EXPECT().Has(gomock.Any()).Return(false).Times(1)
	m1 := gfspclient.NewMockGfSpClientAPI(ctrl)
	a.baseApp.SetGfSpClient(m1)
	m1.EXPECT().SignCreateObjectApproval(gomock.Any(), gomock.Any()).Return([]byte("mockSig"), nil).Times(1)
	m.EXPECT().Push(gomock.Any()).Return(nil).AnyTimes()
	req := &gfsptask.GfSpCreateObjectApprovalTask{
		Task: &gfsptask.GfSpTask{},
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			BucketName:        "mockBucketName",
			ObjectName:        "mockObjectName",
			PrimarySpApproval: &common.Approval{},
		},
	}
	result, err := a.HandleCreateObjectApprovalTask(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, true, result)
}

func TestApprovalModular_HandleCreateObjectApprovalTaskFailure1(t *testing.T) {
	t.Log("Failure case description: dangling pointer")
	a := setup(t)
	result, err := a.HandleCreateObjectApprovalTask(context.TODO(), nil)
	assert.Equal(t, ErrDanglingPointer, err)
	assert.Equal(t, false, result)
}

func TestApprovalModular_HandleCreateObjectApprovalTaskFailure2(t *testing.T) {
	t.Log("Success case description: failed to sign create object approval")
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	a.objectQueue = m
	m.EXPECT().Has(gomock.Any()).Return(false).Times(1)
	m1 := gfspclient.NewMockGfSpClientAPI(ctrl)
	a.baseApp.SetGfSpClient(m1)
	m1.EXPECT().SignCreateObjectApproval(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfsptask.GfSpCreateObjectApprovalTask{
		Task: &gfsptask.GfSpTask{},
		CreateObjectInfo: &storagetypes.MsgCreateObject{
			BucketName:        "mockBucketName",
			ObjectName:        "mockObjectName",
			PrimarySpApproval: &common.Approval{},
		},
	}
	result, err := a.HandleCreateObjectApprovalTask(context.TODO(), req)
	assert.Contains(t, err.Error(), mockErr.Error())
	assert.Equal(t, false, result)
}

func TestApprovalModular_PostCreateObjectApproval(t *testing.T) {
	a := setup(t)
	a.PostCreateObjectApproval(context.TODO(), nil)
}

func TestApprovalModular_QueryTasks(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	a.bucketQueue = m
	a.objectQueue = m
	m.EXPECT().ScanTask(gomock.Any()).Times(2)
	result, err := a.QueryTasks(context.TODO(), coretask.TKey("111"))
	assert.Nil(t, err)
	assert.Equal(t, 0, len(result))
}

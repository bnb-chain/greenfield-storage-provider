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
	m.EXPECT().Has(gomock.Any()).Return(true).AnyTimes()
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
	}).AnyTimes()
	m1.EXPECT().PickVirtualGroupFamilyID(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context,
		task coretask.ApprovalCreateBucketTask) (uint32, error) {
		return 10, nil
	}).AnyTimes()
	m1.EXPECT().SignCreateBucketApproval(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context,
		bucket *storagetypes.MsgCreateBucket) ([]byte, error) {
		return []byte("mockSig"), nil
	}).AnyTimes()
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

func TestApprovalModular_PostCreateBucketApproval(t *testing.T) {
	a := setup(t)
	a.PostCreateBucketApproval(context.TODO(), nil)
}

func TestApprovalModular_PreMigrateBucketApproval(t *testing.T) {
	a := setup(t)
	err := a.PreMigrateBucketApproval(context.TODO(), nil)
	assert.Nil(t, err)
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

func TestApprovalModular_PostCreateObjectApproval(t *testing.T) {
	a := setup(t)
	a.PostCreateObjectApproval(context.TODO(), nil)
}

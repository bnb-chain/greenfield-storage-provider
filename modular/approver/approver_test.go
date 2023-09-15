package approver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
)

var mockErr = errors.New("mock error")

func setup(t *testing.T) *ApprovalModular {
	t.Helper()
	return &ApprovalModular{baseApp: &gfspapp.GfSpBaseApp{}}
}

func TestApprovalModular_Name(t *testing.T) {
	a := setup(t)
	result := a.Name()
	assert.Equal(t, module.ApprovalModularName, result)
}

func TestApprovalModular_StartSuccess(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	a.bucketQueue = m
	a.objectQueue = m
	m.EXPECT().SetRetireTaskStrategy(gomock.Any()).AnyTimes()
	m1 := rcmgr.NewMockResourceManager(ctrl)
	a.baseApp.SetResourceManager(m1)
	m2 := rcmgr.NewMockResourceScope(ctrl)
	m1.EXPECT().OpenService(gomock.Any()).DoAndReturn(func(svc string) (rcmgr.ResourceScope, error) {
		return m2, nil
	}).Times(1)
	err := a.Start(context.TODO())
	assert.Nil(t, err)
}

func TestApprovalModular_StartFailure(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := taskqueue.NewMockTQueueOnStrategy(ctrl)
	a.bucketQueue = m
	a.objectQueue = m
	m.EXPECT().SetRetireTaskStrategy(gomock.Any()).AnyTimes()
	m1 := rcmgr.NewMockResourceManager(ctrl)
	a.baseApp.SetResourceManager(m1)
	m1.EXPECT().OpenService(gomock.Any()).DoAndReturn(func(svc string) (rcmgr.ResourceScope, error) {
		return nil, mockErr
	}).Times(1)
	err := a.Start(context.TODO())
	assert.Equal(t, mockErr, err)
}

func TestApprovalModular_Stop(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	a.scope = m
	m.EXPECT().Release().AnyTimes()
	err := a.Stop(context.TODO())
	assert.Nil(t, err)
}

func TestApprovalModular_ReserveResourceSuccess(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	a.scope = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().BeginSpan().DoAndReturn(func() (rcmgr.ResourceScopeSpan, error) {
		return m1, nil
	}).Times(1)
	m1.EXPECT().ReserveResources(gomock.Any()).DoAndReturn(func(st *rcmgr.ScopeStat) error { return nil }).AnyTimes()
	result, err := a.ReserveResource(context.TODO(), &rcmgr.ScopeStat{Memory: 1})
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestApprovalModular_ReserveResourceFailure1(t *testing.T) {
	t.Log("Failure case description: mock BeginSpan returns error")
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	a.scope = m
	m.EXPECT().BeginSpan().DoAndReturn(func() (rcmgr.ResourceScopeSpan, error) {
		return nil, mockErr
	}).Times(1)
	result, err := a.ReserveResource(context.TODO(), &rcmgr.ScopeStat{Memory: 1})
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestApprovalModular_ReserveResourceFailure2(t *testing.T) {
	t.Log("Failure case description: mock ReserveResources returns error")
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	a.scope = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().BeginSpan().DoAndReturn(func() (rcmgr.ResourceScopeSpan, error) {
		return m1, nil
	}).Times(1)
	m1.EXPECT().ReserveResources(gomock.Any()).DoAndReturn(func(st *rcmgr.ScopeStat) error { return mockErr }).AnyTimes()
	result, err := a.ReserveResource(context.TODO(), &rcmgr.ScopeStat{Memory: 1})
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestApprovalModular_ReleaseResource(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().Done().AnyTimes()
	a.ReleaseResource(context.TODO(), m)
}

// func TestApprovalModular_eventLoopSuccess(t *testing.T) {
// 	a := setup(t)
// 	ctrl := gomock.NewController(t)
// 	m := consensus.NewMockConsensus(ctrl)
// 	m.EXPECT().CurrentHeight(gomock.Any()).DoAndReturn(
// 		func(ctx context.Context) (uint64, error) { return 0, mockErr }).AnyTimes()
// 	a.baseApp.SetConsensus(m)
// 	ctx, cancel := context.WithCancel(context.TODO())
// 	go a.eventLoop(ctx)
// 	time.Sleep(4 * time.Second)
// 	cancel()
// }

func TestApprovalModular_eventLoopSuccess2(t *testing.T) {
	t.Log("Success case description: context calls cancel")
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := consensus.NewMockConsensus(ctrl)
	m.EXPECT().CurrentHeight(gomock.Any()).DoAndReturn(
		func(ctx context.Context) (uint64, error) { return 0, nil }).AnyTimes()
	a.baseApp.SetConsensus(m)
	ctx, cancel := context.WithCancel(context.TODO())
	go a.eventLoop(ctx)
	cancel()
}

func TestApprovalModular_GCApprovalQueue1(t *testing.T) {
	t.Log("expire approval task")
	a := setup(t)
	approvalTask := &gfsptask.GfSpCreateObjectApprovalTask{Task: &gfsptask.GfSpTask{}}
	approvalTask.SetCreateTime(10)
	result := a.GCApprovalQueue(approvalTask)
	assert.Equal(t, true, result)
}

func TestApprovalModular_GCApprovalQueue2(t *testing.T) {
	t.Log("not expired approval task")
	a := setup(t)
	approvalTask := &gfsptask.GfSpCreateObjectApprovalTask{Task: &gfsptask.GfSpTask{}}
	approvalTask.SetCreateTime(time.Now().Unix())
	result := a.GCApprovalQueue(approvalTask)
	assert.Equal(t, false, result)
}

func TestApprovalModular_GetCurrentBlockHeight(t *testing.T) {
	a := setup(t)
	result := a.GetCurrentBlockHeight()
	assert.Equal(t, uint64(0), result)
}

func TestApprovalModular_SetCurrentBlockHeight(t *testing.T) {
	a := setup(t)
	a.SetCurrentBlockHeight(10)
}

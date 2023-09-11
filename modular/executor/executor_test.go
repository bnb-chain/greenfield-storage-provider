package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var mockErr = errors.New("mock error")

func setup(t *testing.T) *ExecuteModular {
	t.Helper()
	return &ExecuteModular{baseApp: &gfspapp.GfSpBaseApp{}}
}

func TestExecuteModular_Name(t *testing.T) {
	e := setup(t)
	result := e.Name()
	assert.Equal(t, coremodule.ExecuteModularName, result)
}

func TestExecuteModular_StartSuccess(t *testing.T) {
	e := setup(t)
	e.statisticsOutputInterval = 1
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceManager(ctrl)
	e.baseApp.SetResourceManager(m)
	m1 := corercmgr.NewMockResourceScope(ctrl)
	m.EXPECT().OpenService(gomock.Any()).DoAndReturn(func(svc string) (corercmgr.ResourceScope, error) {
		return m1, nil
	}).Times(1)
	err := e.Start(context.TODO())
	assert.Nil(t, err)
}

func TestExecuteModular_StartFailure(t *testing.T) {
	e := setup(t)
	e.statisticsOutputInterval = 1
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceManager(ctrl)
	e.baseApp.SetResourceManager(m)
	m.EXPECT().OpenService(gomock.Any()).DoAndReturn(func(svc string) (corercmgr.ResourceScope, error) {
		return nil, mockErr
	}).Times(1)
	err := e.Start(context.TODO())
	assert.Equal(t, mockErr, err)
}

func TestExecuteModular_omitError(t *testing.T) {
	cases := []struct {
		name         string
		err          error
		wantedResult bool
	}{
		{
			name:         "1",
			err:          gfspapp.ErrNoTaskMatchLimit,
			wantedResult: true,
		},
		{
			name:         "2",
			err:          mockErr,
			wantedResult: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			e := setup(t)
			result := e.omitError(tt.err)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestExecuteModular_eventLoop1(t *testing.T) {
	t.Log("Case description: context cancel")
	e := setup(t)
	e.maxExecuteNum = 2
	e.statisticsOutputInterval = 1
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*2)
	cancel()
	e.eventLoop(ctx)
}

func TestExecuteModular_eventLoop2(t *testing.T) {
	t.Log("Case description: ask task")
	e := setup(t)
	e.maxExecuteNum = 1
	e.statisticsOutputInterval = 1
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceScope(ctrl)
	e.scope = m
	m.EXPECT().RemainingResource().Return(nil, mockErr).AnyTimes()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*10)
	e.eventLoop(ctx)
	defer cancel()
}

func TestExecuteModular_eventLoop3(t *testing.T) {
	t.Log("Case description: statistics ticker")
	e := setup(t)
	e.statisticsOutputInterval = 1

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*1)
	e.eventLoop(ctx)
	defer cancel()
}

func TestExecuteModular_AskTask(t *testing.T) {
	cases := []struct {
		name      string
		fn        func() *ExecuteModular
		wantedErr error
	}{
		{
			name: "failed to get remaining resource",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := corercmgr.NewMockResourceScope(ctrl)
				e.scope = m
				m.EXPECT().RemainingResource().Return(nil, mockErr).Times(1)
				return e
			},
			wantedErr: mockErr,
		},
		{
			name: "failed to ask task omitError",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				scopeMock := corercmgr.NewMockResourceScope(ctrl)
				e.scope = scopeMock
				limitMock := corercmgr.NewMockLimit(ctrl)
				limitMock.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
				limitMock.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
				limitMock.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(1).Times(1)
				limitMock.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(1).Times(1)
				limitMock.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(1).Times(1)
				scopeMock.EXPECT().RemainingResource().Return(limitMock, nil).Times(1)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().AskTask(gomock.Any(), gomock.Any()).Return(nil,
					gfspapp.ErrNoTaskMatchLimit).Times(1)
				e.baseApp.SetGfSpClient(clientMock)
				return e
			},
			wantedErr: gfspapp.ErrNoTaskMatchLimit,
		},
		{
			name: "failed to ask task",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				scopeMock := corercmgr.NewMockResourceScope(ctrl)
				e.scope = scopeMock
				limitMock := corercmgr.NewMockLimit(ctrl)
				limitMock.EXPECT().String().Return("1").Times(1)
				limitMock.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
				limitMock.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
				limitMock.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(1).Times(1)
				limitMock.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(1).Times(1)
				limitMock.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(1).Times(1)
				scopeMock.EXPECT().RemainingResource().Return(limitMock, nil).Times(1)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().AskTask(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				e.baseApp.SetGfSpClient(clientMock)
				return e
			},
			wantedErr: mockErr,
		},
		{
			name: "dangling pointer",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				scopeMock := corercmgr.NewMockResourceScope(ctrl)
				e.scope = scopeMock
				limitMock := corercmgr.NewMockLimit(ctrl)
				limitMock.EXPECT().String().Return("1").Times(1)
				limitMock.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
				limitMock.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
				limitMock.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(1).Times(1)
				limitMock.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(1).Times(1)
				limitMock.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(1).Times(1)
				scopeMock.EXPECT().RemainingResource().Return(limitMock, nil).Times(1)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().AskTask(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
				e.baseApp.SetGfSpClient(clientMock)
				return e
			},
			wantedErr: ErrDanglingPointer,
		},
		{
			name: "failed to reserve resource",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				scopeMock := corercmgr.NewMockResourceScope(ctrl)
				e.scope = scopeMock
				limitMock := corercmgr.NewMockLimit(ctrl)
				limitMock.EXPECT().String().Return("1").Times(1)
				limitMock.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
				limitMock.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
				limitMock.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(1).Times(1)
				limitMock.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(1).Times(1)
				limitMock.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(1).Times(1)
				scopeMock.EXPECT().RemainingResource().Return(limitMock, nil).Times(1)
				scopeMock.EXPECT().BeginSpan().DoAndReturn(func() (corercmgr.ResourceScopeSpan, error) {
					return nil, mockErr
				}).Times(1)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().AskTask(gomock.Any(), gomock.Any()).Return(&gfsptask.GfSpReplicatePieceTask{
					Task:          &gfsptask.GfSpTask{},
					ObjectInfo:    &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
					StorageParams: &storagetypes.Params{VersionedParams: storagetypes.VersionedParams{}},
				}, nil).Times(1)
				e.baseApp.SetGfSpClient(clientMock)
				return e
			},
			wantedErr: mockErr,
		},
		// {
		// 	name: "GfSpReplicatePieceTask",
		// 	fn: func() *ExecuteModular {
		// 		e := setup(t)
		// 		ctrl := gomock.NewController(t)
		// 		scopeMock := corercmgr.NewMockResourceScope(ctrl)
		// 		e.scope = scopeMock
		// 		limitMock := corercmgr.NewMockLimit(ctrl)
		// 		limitMock.EXPECT().String().Return("1").Times(1)
		// 		limitMock.EXPECT().GetMemoryLimit().Return(int64(1)).Times(1)
		// 		limitMock.EXPECT().GetTaskTotalLimit().Return(1).Times(1)
		// 		limitMock.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityHigh).Return(1).Times(1)
		// 		limitMock.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityMedium).Return(1).Times(1)
		// 		limitMock.EXPECT().GetTaskLimit(corercmgr.ReserveTaskPriorityLow).Return(1).Times(1)
		// 		scopeMock.EXPECT().RemainingResource().Return(limitMock, nil).Times(1)
		// 		spanMock := corercmgr.NewMockResourceScopeSpan(ctrl)
		// 		scopeMock.EXPECT().BeginSpan().DoAndReturn(func() (corercmgr.ResourceScopeSpan, error) {
		// 			return spanMock, nil
		// 		}).Times(1)
		// 		spanMock.EXPECT().ReserveResources(gomock.Any()).Return(nil).AnyTimes()
		// 		spanMock.EXPECT().Done().AnyTimes()
		//
		// 		clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
		// 		clientMock.EXPECT().AskTask(gomock.Any(), gomock.Any()).Return(&gfsptask.GfSpReplicatePieceTask{
		// 			Task:          &gfsptask.GfSpTask{},
		// 			ObjectInfo:    &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
		// 			StorageParams: &storagetypes.Params{VersionedParams: storagetypes.VersionedParams{}},
		// 		}, nil).Times(1)
		// 		e.baseApp.SetGfSpClient(clientMock)
		// 		return e
		// 	},
		// 	wantedErr: nil,
		// },
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn().AskTask(context.TODO())
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestExecuteModular_ReportTask(t *testing.T) {
	e := setup(t)
	ctrl := gomock.NewController(t)
	m := gfspclient.NewMockGfSpClientAPI(ctrl)
	m.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	e.baseApp.SetGfSpClient(m)
	err := e.ReportTask(context.TODO(), &gfsptask.GfSpReplicatePieceTask{
		Task:          &gfsptask.GfSpTask{},
		ObjectInfo:    &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
		StorageParams: &storagetypes.Params{VersionedParams: storagetypes.VersionedParams{}},
	})
	assert.Nil(t, err)
}

func TestExecuteModular_Stop(t *testing.T) {
	e := setup(t)
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceScope(ctrl)
	e.scope = m
	m.EXPECT().Release().AnyTimes()
	err := e.Stop(context.TODO())
	assert.Nil(t, err)
}

func TestExecuteModular_ReserveResourceSuccess(t *testing.T) {
	e := setup(t)
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceScope(ctrl)
	e.scope = m
	m1 := corercmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().BeginSpan().DoAndReturn(func() (corercmgr.ResourceScopeSpan, error) {
		return m1, nil
	}).Times(1)
	m1.EXPECT().ReserveResources(gomock.Any()).DoAndReturn(func(st *corercmgr.ScopeStat) error { return nil }).AnyTimes()
	result, err := e.ReserveResource(context.TODO(), &corercmgr.ScopeStat{Memory: 1})
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestAExecuteModular_ReserveResourceFailure1(t *testing.T) {
	t.Log("Failure case description: mock BeginSpan returns error")
	e := setup(t)
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceScope(ctrl)
	e.scope = m
	m.EXPECT().BeginSpan().DoAndReturn(func() (corercmgr.ResourceScopeSpan, error) {
		return nil, mockErr
	}).Times(1)
	result, err := e.ReserveResource(context.TODO(), &corercmgr.ScopeStat{Memory: 1})
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestExecuteModular_ReserveResourceFailure2(t *testing.T) {
	t.Log("Failure case description: mock ReserveResources returns error")
	e := setup(t)
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceScope(ctrl)
	e.scope = m
	m1 := corercmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().BeginSpan().DoAndReturn(func() (corercmgr.ResourceScopeSpan, error) {
		return m1, nil
	}).Times(1)
	m1.EXPECT().ReserveResources(gomock.Any()).DoAndReturn(func(st *corercmgr.ScopeStat) error { return mockErr }).AnyTimes()
	result, err := e.ReserveResource(context.TODO(), &corercmgr.ScopeStat{Memory: 1})
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestExecuteModular_ReleaseResource(t *testing.T) {
	e := setup(t)
	ctrl := gomock.NewController(t)
	m := corercmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().Done().AnyTimes()
	e.ReleaseResource(context.TODO(), m)
}

func TestExecuteModular_Statistics(t *testing.T) {
	e := setup(t)
	result := e.Statistics()
	assert.NotEmpty(t, result)
}

func TestExecuteModular_getSPID(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *ExecuteModular
		wantedResult uint32
		wantedErr    error
	}{
		{
			name: "1",
			fn: func() *ExecuteModular {
				e := setup(t)
				e.spID = 1
				return e
			},
			wantedResult: 1,
			wantedErr:    nil,
		},
		{
			name: "2",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := consensus.NewMockConsensus(ctrl)
				m.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				e.baseApp.SetConsensus(m)
				return e
			},
			wantedResult: 0,
			wantedErr:    mockErr,
		},
		{
			name: "3",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := consensus.NewMockConsensus(ctrl)
				m.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 2}, nil).Times(1)
				e.baseApp.SetConsensus(m)
				return e
			},
			wantedResult: 2,
			wantedErr:    nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn().getSPID()
			assert.Equal(t, tt.wantedErr, err)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

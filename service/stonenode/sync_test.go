package stonenode

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/service/client/mock"
	service "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
)

func TestInitClientFailed(t *testing.T) {
	node := &StoneNodeService{
		name:       ServiceNameStoneNode,
		stoneLimit: 0,
	}
	node.running.Store(true)
	err := node.initClient()
	assert.Equal(t, merrors.ErrStoneNodeStarted, err)
}

func Test_loadSegmentsDataSuccess(t *testing.T) {
	cases := []struct {
		name          string
		req1          uint64
		req2          uint64
		req3          ptypes.RedundancyType
		wantedResult1 string
		wantedResult2 int
		wantedErr     error
	}{
		{
			name:          "ec type: payload size greater than 16MB",
			req1:          20230109001,
			req2:          20 * 1024 * 1024,
			req3:          ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
			wantedResult1: "20230109001",
			wantedResult2: 2,
			wantedErr:     nil,
		},
		{
			name:          "ec type: payload size less than 16MB and greater than 1MB",
			req1:          20230109002,
			req2:          15 * 1024 * 1024,
			req3:          ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
			wantedResult1: "20230109002",
			wantedResult2: 1,
			wantedErr:     nil,
		},
		{
			name:          "replica type: payload size greater than 16MB",
			req1:          20230109003,
			req2:          20 * 1024 * 1024,
			req3:          ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE,
			wantedResult1: "20230109003",
			wantedResult2: 2,
			wantedErr:     nil,
		},
		{
			name:          "replica type: payload size less than 16MB and greater than 1MB",
			req1:          20230109004,
			req2:          15 * 1024 * 1024,
			req3:          ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE,
			wantedResult1: "20230109004",
			wantedResult2: 1,
			wantedErr:     nil,
		},
		{
			name:          "inline type: payload size less than 1MB",
			req1:          20230109005,
			req2:          1000 * 1024,
			req3:          ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE,
			wantedResult1: "20230109005",
			wantedResult2: 1,
			wantedErr:     nil,
		},
	}

	node := setup(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ps := mock.NewMockPieceStoreAPI(ctrl)
	node.store = ps
	ps.EXPECT().GetPiece(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string, offset, limit int64) ([]byte, error) {
			return []byte("1"), nil
		}).AnyTimes()

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			allocResp := mockAllocResp(tt.req1, tt.req2, tt.req3)
			result, err := node.loadSegmentsData(context.TODO(), allocResp)
			assert.Equal(t, nil, err)
			for k, _ := range result {
				assert.Contains(t, k, tt.wantedResult1)
			}
			assert.Equal(t, tt.wantedResult2, len(result))
		})
	}
}

func Test_loadSegmentsDataPieceStoreError(t *testing.T) {
	node := setup(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ps := mock.NewMockPieceStoreAPI(ctrl)
	node.store = ps
	ps.EXPECT().GetPiece(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string, offset, limit int64) ([]byte, error) {
			return nil, errors.New("piece store s3 network error")
		}).AnyTimes()

	result, err := node.loadSegmentsData(context.TODO(), mockAllocResp(20230109001, 20*1024*1024,
		ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED))
	assert.Equal(t, errors.New("piece store s3 network error"), err)
	assert.Equal(t, 0, len(result))
}

func Test_loadSegmentsDataUnknownRedundancyError(t *testing.T) {
	node := setup(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ps := mock.NewMockPieceStoreAPI(ctrl)
	node.store = ps
	ps.EXPECT().GetPiece(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string, offset, limit int64) ([]byte, error) {
			return []byte("1"), nil
		}).AnyTimes()

	result, err := node.loadSegmentsData(context.TODO(), mockAllocResp(20230109006, 20*1024*1024,
		ptypes.RedundancyType(-1)))
	assert.Equal(t, merrors.ErrRedundancyType, err)
	assert.Equal(t, 0, len(result))
}

func Test_generatePieceData(t *testing.T) {
	cases := []struct {
		name         string
		req1         ptypes.RedundancyType
		req2         []byte
		wantedResult int
		wantedErr    error
	}{
		{
			name:         "ec type",
			req1:         ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
			req2:         []byte("1"),
			wantedResult: 6,
			wantedErr:    nil,
		},
		{
			name:         "replica type",
			req1:         ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE,
			req2:         []byte("1"),
			wantedResult: 1,
			wantedErr:    nil,
		},
		{
			name:         "inline type",
			req1:         ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE,
			req2:         []byte("1"),
			wantedResult: 1,
			wantedErr:    nil,
		},
		{
			name:         "unknown redundancy type",
			req1:         ptypes.RedundancyType(-1),
			req2:         []byte("1"),
			wantedResult: 0,
			wantedErr:    merrors.ErrRedundancyType,
		},
	}

	node := setup(t)
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := node.generatePieceData(tt.req1, tt.req2)
			assert.Equal(t, err, tt.wantedErr)
			assert.Equal(t, len(result), tt.wantedResult)
		})
	}
}

func Test_dispatchSecondarySP(t *testing.T) {
	spList := []string{"sp1", "sp2", "sp3", "sp4", "sp5", "sp6"}
	cases := []struct {
		name         string
		req1         map[string][][]byte
		req2         ptypes.RedundancyType
		req3         []string
		req4         []uint32
		wantedResult int
		wantedErr    error
	}{
		{
			name:         "ec type dispatch",
			req1:         dispatchPieceMap(),
			req2:         ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
			req3:         spList,
			req4:         []uint32{0, 1, 2, 3, 4, 5},
			wantedResult: 6,
			wantedErr:    nil,
		},
		{
			name:         "replica type dispatch",
			req1:         dispatchSegmentMap(),
			req2:         ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE,
			req3:         spList,
			req4:         []uint32{0, 1, 2},
			wantedResult: 1,
			wantedErr:    nil,
		},
		{
			name:         "inline type dispatch",
			req1:         dispatchInlineMap(),
			req2:         ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE,
			req3:         spList,
			req4:         []uint32{0},
			wantedResult: 1,
			wantedErr:    nil,
		},
		{
			name:         "ec type data retransmission",
			req1:         dispatchPieceMap(),
			req2:         ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
			req3:         spList,
			req4:         []uint32{2, 3},
			wantedResult: 2,
			wantedErr:    nil,
		},
		{
			name:         "replica type data retransmission",
			req1:         dispatchSegmentMap(),
			req2:         ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE,
			req3:         spList,
			req4:         []uint32{1, 2},
			wantedResult: 0,
			wantedErr:    nil,
		},
		{
			name:         "wrong secondary sp number",
			req1:         dispatchPieceMap(),
			req2:         ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
			req3:         []string{},
			req4:         []uint32{0, 1, 2, 3, 4, 5},
			wantedResult: 0,
			wantedErr:    merrors.ErrSecondarySPNumber,
		},
		//{
		//	name:         "wrong ec segment data length",
		//	req1:         dispatchSegmentMap(),
		//	req2:         ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
		//	req3:         spList,
		//	req4:         []uint32{0, 1, 2, 3, 4, 5},
		//	wantedResult: 1,
		//	wantedErr:    merrors.ErrInvalidECData,
		//},
		//{
		//	name:         "wrong replica/inline segment data length",
		//	req1:         dispatchPieceMap(),
		//	req2:         ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE,
		//	req3:         spList,
		//	req4:         []uint32{0, 1, 2, 3, 4, 5},
		//	wantedResult: 1,
		//	wantedErr:    merrors.ErrInvalidSegmentData,
		//},
	}

	node := setup(t)
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := node.dispatchSecondarySP(tt.req1, tt.req2, tt.req3, tt.req4)
			assert.Equal(t, tt.wantedErr, err)
			assert.Equal(t, tt.wantedResult, len(result))
		})
	}
}

// TODO:need improved
func Test_doSyncToSecondarySP(t *testing.T) {
	data := map[string]map[string][]byte{
		"sp1": {
			"123456_s0_p0": []byte("test1"),
			"123456_s1_p0": []byte("test2"),
			"123456_s2_p0": []byte("test3"),
			"123456_s3_p0": []byte("test4"),
			"123456_s4_p0": []byte("test5"),
			"123456_s5_p0": []byte("test6"),
		},
	}
	cases := []struct {
		name string
		req1 *service.StoneHubServiceAllocStoneJobResponse
		req2 map[string]map[string][]byte
	}{
		{
			name: "1",
			req1: nil,
			req2: data,
		},
	}

	node := setup(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// stoneHub service stub
	//stoneHub := mock.NewMockStoneHubAPI(ctrl)
	//node.stoneHub = stoneHub
	//stoneHub.EXPECT().DoneSecondaryPieceJob(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
	//	func(ctx context.Context, in *service.StoneHubServiceDoneSecondaryPieceJobRequest, opts ...grpc.CallOption) (
	//		*service.StoneHubServiceDoneSecondaryPieceJobResponse, error) {
	//		return nil, nil
	//	})

	// syncer service stub
	streamClient := makeStreamMock()
	syncer := mock.NewMockSyncerAPI(ctrl)
	node.syncer = syncer
	syncer.EXPECT().SyncPiece(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, opts ...grpc.CallOption) (service.SyncerService_SyncPieceClient, error) {
			return streamClient, nil
		}).AnyTimes()

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			allocResp := mockAllocResp(123456, 20*1024*1024, ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED)
			err := node.doSyncToSecondarySP(context.TODO(), allocResp, tt.req2)
			assert.Equal(t, nil, err)
		})
	}
}

func TestSyncPieceSuccess(t *testing.T) {
	node := setup(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	streamClient := makeStreamMock()
	syncer := mock.NewMockSyncerAPI(ctrl)
	node.syncer = syncer
	syncer.EXPECT().SyncPiece(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, opts ...grpc.CallOption) (service.SyncerService_SyncPieceClient, error) {
			return streamClient, nil
		}).AnyTimes()

	sInfo := &service.SyncerInfo{
		ObjectId:          123456,
		TxHash:            []byte("i"),
		StorageProviderId: "sp1",
		RedundancyType:    ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
	}
	data := map[string][]byte{
		"123456_s0_p0": []byte("test1"),
		"123456_s1_p0": []byte("test2"),
		"123456_s2_p0": []byte("test3"),
		"123456_s3_p0": []byte("test4"),
		"123456_s4_p0": []byte("test5"),
		"123456_s5_p0": []byte("test6"),
	}
	resp, err := node.syncPiece(context.TODO(), sInfo, data, "test_traceID")
	assert.Equal(t, err, nil)
	assert.Equal(t, resp.GetTraceId(), "test_traceID")
	assert.Equal(t, resp.GetSecondarySpInfo().GetPieceIdx(), uint32(1))
}

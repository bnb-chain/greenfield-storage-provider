package stonenode

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	spmock "github.com/bnb-chain/greenfield-storage-provider/mock"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/service/client/mock"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
)

// TODO:need improved
func Test_doSyncToSecondarySP(t *testing.T) {
	cases := []struct {
		name string
		req1 *stypes.StoneHubServiceAllocStoneJobResponse
		req2 [][][]byte
	}{
		{
			name: "1",
			req1: nil,
			req2: dispatchECPiece(),
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
	syncer1 := mock.NewMockSyncerAPI(ctrl)
	node.syncer = append(node.syncer, syncer1)
	syncer1.EXPECT().SyncPiece(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, opts ...grpc.CallOption) (stypes.SyncerService_SyncPieceClient, error) {
			return streamClient, nil
		}).AnyTimes()

	syncer2 := mock.NewMockSyncerAPI(ctrl)
	node.syncer = append(node.syncer, syncer2)
	syncer2.EXPECT().SyncPiece(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, opts ...grpc.CallOption) (stypes.SyncerService_SyncPieceClient, error) {
			return streamClient, nil
		}).AnyTimes()

	syncer3 := mock.NewMockSyncerAPI(ctrl)
	node.syncer = append(node.syncer, syncer3)
	syncer3.EXPECT().SyncPiece(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, opts ...grpc.CallOption) (stypes.SyncerService_SyncPieceClient, error) {
			return streamClient, nil
		}).AnyTimes()

	syncer4 := mock.NewMockSyncerAPI(ctrl)
	node.syncer = append(node.syncer, syncer4)
	syncer4.EXPECT().SyncPiece(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, opts ...grpc.CallOption) (stypes.SyncerService_SyncPieceClient, error) {
			return streamClient, nil
		}).AnyTimes()

	syncer5 := mock.NewMockSyncerAPI(ctrl)
	node.syncer = append(node.syncer, syncer5)
	syncer5.EXPECT().SyncPiece(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, opts ...grpc.CallOption) (stypes.SyncerService_SyncPieceClient, error) {
			return streamClient, nil
		}).AnyTimes()

	syncer6 := mock.NewMockSyncerAPI(ctrl)
	node.syncer = append(node.syncer, syncer6)
	syncer6.EXPECT().SyncPiece(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, opts ...grpc.CallOption) (stypes.SyncerService_SyncPieceClient, error) {
			return streamClient, nil
		}).AnyTimes()

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			allocResp := mockAllocResp(123456, 20*1024*1024, ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED)
			err := node.doSyncToSecondarySP(context.TODO(), allocResp, tt.req2, spmock.AllocUploadSecondarySP())
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
	node.syncer = append(node.syncer, syncer)
	syncer.EXPECT().SyncPiece(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, opts ...grpc.CallOption) (stypes.SyncerService_SyncPieceClient, error) {
			return streamClient, nil
		}).AnyTimes()

	sInfo := &stypes.SyncerInfo{
		ObjectId:          123456,
		StorageProviderId: "440246a94fc4257096b8d4fa8db94a5655f455f88555f885b10da1466763f742",
		RedundancyType:    ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
	}
	data := [][]byte{
		[]byte("test1"),
		[]byte("test2"),
		[]byte("test3"),
		[]byte("test4"),
		[]byte("test5"),
		[]byte("test6"),
	}
	resp, err := node.syncPiece(context.TODO(), sInfo, data, 0, "test_traceID")
	assert.Equal(t, err, nil)
	assert.Equal(t, resp.GetTraceId(), "test_traceID")
	assert.Equal(t, resp.GetSecondarySpInfo().GetPieceIdx(), uint32(1))
}

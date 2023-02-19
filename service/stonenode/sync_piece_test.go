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

func Test_doSyncToSecondarySP(t *testing.T) {
	list := make([][][]byte, 0)
	ecList := [][]byte{[]byte("1"), []byte("2"), []byte("3"), []byte("4"), []byte("5"), []byte("6")}
	list = append(list, ecList)
	cases := []struct {
		name string
		req1 [][][]byte
	}{
		{
			name: "1",
			req1: list,
		},
	}

	node := setup(t)
	ctrl := gomock.NewController(t)

	// stoneHub service stub
	stoneHub := mock.NewMockStoneHubAPI(ctrl)
	node.stoneHub = stoneHub
	stoneHub.EXPECT().DoneSecondaryPieceJob(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, in *stypes.StoneHubServiceDoneSecondaryPieceJobRequest, opts ...grpc.CallOption) (
			*stypes.StoneHubServiceDoneSecondaryPieceJobResponse, error) {
			return nil, nil
		}).AnyTimes()

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
			err := node.doSyncToSecondarySP(context.TODO(), allocResp, tt.req1, spmock.AllocUploadSecondarySP(),
				spmock.AllocUploadSecondarySP())
			assert.Equal(t, nil, err)
		})
	}
}

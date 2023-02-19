package gateway

//import (
//	"context"
//	"testing"
//
//	"github.com/golang/mock/gomock"
//	"github.com/stretchr/testify/assert"
//	"google.golang.org/grpc"
//
//	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
//	"github.com/bnb-chain/greenfield-storage-provider/service/client/mock"
//	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
//)
//
//func TestSyncPieceSuccess(t *testing.T) {
//	node := setup(t)
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	streamClient := makeStreamMock()
//	syncer := mock.NewMockSyncerAPI(ctrl)
//	node.syncer = append(node.syncer, syncer)
//	syncer.EXPECT().SyncPiece(gomock.Any(), gomock.Any()).DoAndReturn(
//		func(ctx context.Context, opts ...grpc.CallOption) (stypes.SyncerService_SyncPieceClient, error) {
//			return streamClient, nil
//		}).AnyTimes()
//
//	sInfo := &stypes.SyncerInfo{
//		ObjectId:          123456,
//		StorageProviderId: "440246a94fc4257096b8d4fa8db94a5655f455f88555f885b10da1466763f742",
//		RedundancyType:    ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
//	}
//	data := [][]byte{
//		[]byte("test1"),
//		[]byte("test2"),
//		[]byte("test3"),
//		[]byte("test4"),
//		[]byte("test5"),
//		[]byte("test6"),
//	}
//	resp, err := node.syncPiece(context.TODO(), sInfo, data, 0, "test_traceID")
//	assert.Equal(t, err, nil)
//	assert.Equal(t, resp.GetTraceId(), "test_traceID")
//	assert.Equal(t, resp.GetSecondarySpInfo().GetPieceIdx(), uint32(1))
//}

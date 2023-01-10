package stonenode

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	merrors "github.com/bnb-chain/inscription-storage-provider/model/errors"
	ptypes "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/service/client/mock"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
)

func TestInitClientFailed(t *testing.T) {
	node := &StoneNodeService{
		name:       ServiceNameStoneNode,
		stoneLimit: 0,
	}
	node.running.Store(true)
	err := node.InitClient()
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
		name          string
		req1          map[string][][]byte
		req2          ptypes.RedundancyType
		req3          []string
		req4          []uint32
		wantedResult1 int
		wantedErr     error
	}{
		{
			name:          "ec type dispatch",
			req1:          dispatchPieceMap(),
			req2:          ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
			req3:          spList,
			req4:          []uint32{},
			wantedResult1: 6,
			wantedErr:     nil,
		},
		{
			name:          "replica type dispatch",
			req1:          dispatchSegmentMap(),
			req2:          ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE,
			req3:          spList,
			req4:          []uint32{},
			wantedResult1: 2,
			wantedErr:     nil,
		},
		{
			name:          "inline type dispatch",
			req1:          dispatchInlineMap(),
			req2:          ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE,
			req3:          spList,
			req4:          []uint32{},
			wantedResult1: 1,
			wantedErr:     nil,
		},
		{
			name:          "4",
			req1:          dispatchPieceMap(),
			req2:          ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
			req3:          spList,
			req4:          []uint32{2, 3},
			wantedResult1: 2,
			wantedErr:     nil,
		},
		{
			name:          "unknown redundancy type",
			req1:          dispatchPieceMap(),
			req2:          ptypes.RedundancyType(-1),
			req3:          spList,
			req4:          []uint32{},
			wantedResult1: 0,
			wantedErr:     merrors.ErrRedundancyType,
		},
		{
			name:          "wrong secondary sp number",
			req1:          dispatchPieceMap(),
			req2:          ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
			req3:          []string{},
			req4:          []uint32{},
			wantedResult1: 0,
			wantedErr:     merrors.ErrSecondarySPNumber,
		},
	}

	node := setup(t)
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := node.dispatchSecondarySP(tt.req1, tt.req2, tt.req3, tt.req4)
			assert.Equal(t, err, tt.wantedErr)
			assert.Equal(t, len(result), tt.wantedResult1)
		})
	}
}

func TestUploadECPiece(t *testing.T) {
	node := setup(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	syncer := mock.NewMockSyncerAPI(ctrl)
	node.syncer = syncer
	syncer.EXPECT().UploadECPiece(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, opts ...grpc.CallOption) (service.SyncerService_UploadECPieceClient, error) {
			return nil, nil
		}).AnyTimes()

}

func makeStreamMock() *StreamMock {
	return &StreamMock{
		ctx:            context.Background(),
		recvToServer:   make(chan *service.SyncerServiceUploadECPieceRequest, 10),
		sentFromServer: make(chan *service.SyncerServiceUploadECPieceResponse, 10),
	}
}

type StreamMock struct {
	grpc.ServerStream
	ctx            context.Context
	recvToServer   chan *service.SyncerServiceUploadECPieceRequest
	sentFromServer chan *service.SyncerServiceUploadECPieceResponse
}

func (m *StreamMock) Context() context.Context {
	return m.ctx
}

func (m *StreamMock) Send(resp *service.SyncerServiceUploadECPieceResponse) error {
	m.sentFromServer <- resp
	return nil
}

func (m *StreamMock) Recv() (*service.SyncerServiceUploadECPieceRequest, error) {
	req, more := <-m.recvToServer
	if !more {
		return nil, errors.New("empty")
	}
	return req, nil
}

func (m *StreamMock) SendFromClient(req *service.SyncerServiceUploadECPieceRequest) error {
	m.recvToServer <- req
	return nil
}

func (m *StreamMock) RecvToClient() (*service.SyncerServiceUploadECPieceResponse, error) {
	response, more := <-m.sentFromServer
	if !more {
		return nil, errors.New("empty")
	}
	return response, nil
}

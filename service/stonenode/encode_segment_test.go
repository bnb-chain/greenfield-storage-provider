package stonenode

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/service/client/mock"
)

func TestInitClientFailed(t *testing.T) {
	node := &StoneNodeService{
		name:       model.StoneNodeService,
		stoneLimit: 0,
	}
	node.running.Store(true)
	err := node.initClient()
	assert.Equal(t, merrors.ErrStoneNodeStarted, err)
}

func Test_encodeSegmentsDataSuccess(t *testing.T) {
	cases := []struct {
		name          string
		req1          uint64
		req2          uint64
		req3          ptypes.RedundancyType
		wantedResult1 int
		wantedErr     error
	}{
		{
			name:          "ec type: payload size greater than 16MB",
			req1:          20230109001,
			req2:          20 * 1024 * 1024,
			req3:          ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
			wantedResult1: 2,
			wantedErr:     nil,
		},
		{
			name:          "ec type: payload size less than 16MB and greater than 1MB",
			req1:          20230109002,
			req2:          15 * 1024 * 1024,
			req3:          ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
			wantedResult1: 1,
			wantedErr:     nil,
		},
		{
			name:          "replica type: payload size greater than 16MB",
			req1:          20230109003,
			req2:          20 * 1024 * 1024,
			req3:          ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE,
			wantedResult1: 2,
			wantedErr:     nil,
		},
		{
			name:          "replica type: payload size less than 16MB and greater than 1MB",
			req1:          20230109004,
			req2:          15 * 1024 * 1024,
			req3:          ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE,
			wantedResult1: 1,
			wantedErr:     nil,
		},
		{
			name:          "inline type: payload size less than 1MB",
			req1:          20230109005,
			req2:          1000 * 1024,
			req3:          ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE,
			wantedResult1: 1,
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
			result, err := node.encodeSegmentsData(context.TODO(), allocResp)
			assert.Equal(t, nil, err)
			assert.Equal(t, tt.wantedResult1, len(result))
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

	result, err := node.encodeSegmentsData(context.TODO(), mockAllocResp(20230109001, 20*1024*1024,
		ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED))
	assert.Equal(t, errors.New("piece store s3 network error"), err)
	assert.Equal(t, 0, len(result))
}

func Test_encodeSegmentData(t *testing.T) {
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
	}

	node := setup(t)
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := node.encode(tt.req1, tt.req2)
			assert.Equal(t, err, tt.wantedErr)
			assert.Equal(t, len(result), tt.wantedResult)
		})
	}
}

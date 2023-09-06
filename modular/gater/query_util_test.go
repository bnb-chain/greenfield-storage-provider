package gater

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func Test_getObjectChainMeta(t *testing.T) {
	cases := []struct {
		name        string
		fn          func() *GateModular
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name: "success",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				m := consensus.NewMockConsensus(ctrl)
				m.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{
					ObjectName: mockObjectName, CreateAt: 1}, nil).Times(1)
				m.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketName: mockBucketName}, nil).Times(1)
				m.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(&storagetypes.Params{
					MaxPayloadSize: DefaultMaxPayloadSize}, nil).Times(1)
				g.baseApp.SetConsensus(m)
				return g
			},
			wantedIsErr: false,
			wantedErr:   nil,
		},
		{
			name: "failed to get object info from consensus",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				m := consensus.NewMockConsensus(ctrl)
				g.baseApp.SetConsensus(m)
				m.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				return g
			},
			wantedIsErr: true,
			wantedErr:   mockErr,
		},
		{
			name: "failed to get bucket info from consensus",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				m := consensus.NewMockConsensus(ctrl)
				g.baseApp.SetConsensus(m)
				m.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{
					ObjectName: mockObjectName}, nil).Times(1)
				m.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				return g
			},
			wantedIsErr: true,
			wantedErr:   mockErr,
		},
		{
			name: "failed to get storage params",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				m := consensus.NewMockConsensus(ctrl)
				g.baseApp.SetConsensus(m)
				m.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{
					ObjectName: mockObjectName, CreateAt: 1}, nil).Times(1)
				m.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketName: mockBucketName}, nil).Times(1)
				m.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				return g
			},
			wantedIsErr: true,
			wantedErr:   mockErr,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result1, result2, result3, err := tt.fn().getObjectChainMeta(context.TODO(), mockObjectName, mockBucketName)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
				assert.Nil(t, result1)
				assert.Nil(t, result2)
				assert.Nil(t, result3)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result1)
				assert.NotNil(t, result2)
				assert.NotNil(t, result3)
			}
		})
	}
}

func Test_checkSPAndBucketStatus(t *testing.T) {
	cases := []struct {
		name        string
		fn          func() *GateModular
		bucketName  string
		creatorAddr string
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name: "success",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				m := consensus.NewMockConsensus(ctrl)
				g.baseApp.SetConsensus(m)
				m.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				m.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				return g
			},
			bucketName:  mockBucketName,
			creatorAddr: "",
			wantedIsErr: false,
			wantedErr:   nil,
		},
		{
			name: "failed to query sp by operator address",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				m := consensus.NewMockConsensus(ctrl)
				g.baseApp.SetConsensus(m)
				m.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				return g
			},
			bucketName:  mockBucketName,
			creatorAddr: "",
			wantedIsErr: true,
			wantedErr:   mockErr,
		},
		{
			name: "sp is not in service status",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				m := consensus.NewMockConsensus(ctrl)
				g.baseApp.SetConsensus(m)
				m.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_JAILED}, nil).Times(1)
				return g
			},
			bucketName:  mockBucketName,
			creatorAddr: "",
			wantedIsErr: true,
			wantedErr:   ErrSPUnavailable,
		},
		{
			name: "failed to query bucket info by bucket name",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				m := consensus.NewMockConsensus(ctrl)
				g.baseApp.SetConsensus(m)
				m.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				m.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				return g
			},
			bucketName:  mockBucketName,
			creatorAddr: "",
			wantedIsErr: true,
			wantedErr:   mockErr,
		},
		{
			name: "bucket is not in created status",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				m := consensus.NewMockConsensus(ctrl)
				g.baseApp.SetConsensus(m)
				m.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				m.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_MIGRATING}, nil).Times(1)
				return g
			},
			bucketName:  mockBucketName,
			creatorAddr: "",
			wantedIsErr: true,
			wantedErr:   ErrBucketUnavailable,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn().checkSPAndBucketStatus(context.TODO(), tt.bucketName, tt.creatorAddr)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_fromSpMaintenanceAcct(t *testing.T) {
	result := fromSpMaintenanceAcct(sptypes.STATUS_IN_SERVICE, "test", "mock")
	assert.Equal(t, false, result)
}

package executor

import (
	"context"
	"io"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	metadatatypes "github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtual_types "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func TestErrGfSpDBWithDetail(t *testing.T) {
	err := ErrGfSpDBWithDetail("test")
	assert.NotNil(t, err)
}

func TestErrPieceStoreWithDetail(t *testing.T) {
	err := ErrPieceStoreWithDetail("test")
	assert.NotNil(t, err)
}

func TestErrConsensusWithDetail(t *testing.T) {
	err := ErrConsensusWithDetail("test")
	assert.NotNil(t, err)
}

func TestExecuteModular_HandleSealObjectTask(t *testing.T) {
	cases := []struct {
		name      string
		task      coretask.SealObjectTask
		fn        func() *ExecuteModular
		wantedErr error
	}{
		{
			name:      "dangling pointer",
			task:      &gfsptask.GfSpSealObjectTask{Task: &gfsptask.GfSpTask{}},
			fn:        func() *ExecuteModular { return setup(t) },
			wantedErr: ErrDanglingPointer,
		},
		{
			name: "invalid secondary sig",
			task: &gfsptask.GfSpSealObjectTask{
				Task:                &gfsptask.GfSpTask{MaxRetry: 1},
				ObjectInfo:          &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
				StorageParams:       &storagetypes.Params{},
				SecondarySignatures: [][]byte{[]byte("test")},
			},
			fn:        func() *ExecuteModular { return setup(t) },
			wantedErr: ErrUnsealed,
		},
		{
			name: "object is sealed",
			task: &gfsptask.GfSpSealObjectTask{
				Task:          &gfsptask.GfSpTask{MaxRetry: 1},
				ObjectInfo:    &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
				StorageParams: &storagetypes.Params{},
				SecondarySignatures: [][]byte{
					{0xab, 0xb0, 0x12, 0x4c, 0x75, 0x74, 0xf2, 0x81, 0xa2, 0x93, 0xf4, 0x18, 0x5c, 0xad, 0x3c, 0xb2, 0x26, 0x81, 0xd5, 0x20, 0x91, 0x7c, 0xe4, 0x66, 0x65, 0x24, 0x3e, 0xac, 0xb0, 0x51, 0x00, 0x0d, 0x8b, 0xac, 0xf7, 0x5e, 0x14, 0x51, 0x87, 0x0c, 0xa6, 0xb3, 0xb9, 0xe6, 0xc9, 0xd4, 0x1a, 0x7b, 0x02, 0xea, 0xd2, 0x68, 0x5a, 0x84, 0x18, 0x8a, 0x4f, 0xaf, 0xd3, 0x82, 0x5d, 0xaf, 0x6a, 0x98, 0x96, 0x25, 0xd7, 0x19, 0xcc, 0xd2, 0xd8, 0x3a, 0x40, 0x10, 0x1f, 0x4a, 0x45, 0x3f, 0xca, 0x62, 0x87, 0x8c, 0x89, 0x0e, 0xca, 0x62, 0x23, 0x63, 0xf9, 0xdd, 0xb8, 0xf3, 0x67, 0xa9, 0x1e, 0x84},
					{0xb7, 0x86, 0xe5, 0x7, 0x43, 0xe2, 0x53, 0x6c, 0x15, 0x51, 0x9c, 0x6, 0x2a, 0xa7, 0xe5, 0x12, 0xf9, 0xb7, 0x77, 0x93, 0x3f, 0x55, 0xb3, 0xaf, 0x38, 0xf7, 0x39, 0xe4, 0x84, 0x6d, 0x88, 0x44, 0x52, 0x77, 0x65, 0x42, 0x95, 0xd9, 0x79, 0x93, 0x7e, 0xc8, 0x12, 0x60, 0xe3, 0x24, 0xea, 0x8, 0x10, 0x52, 0xcd, 0xd2, 0x7f, 0x5d, 0x25, 0x3a, 0xa8, 0x9b, 0xb7, 0x65, 0xa9, 0x31, 0xea, 0x7c, 0x85, 0x13, 0x53, 0xc0, 0xa3, 0x88, 0xd1, 0xa5, 0x54, 0x85, 0x2, 0x2d, 0xf8, 0xa1, 0xd7, 0xc1, 0x60, 0x58, 0x93, 0xec, 0x7c, 0xf9, 0x33, 0x43, 0x4, 0x48, 0x40, 0x97, 0xef, 0x67, 0x2a, 0x27},
					{0xb2, 0x12, 0xd0, 0xec, 0x46, 0x76, 0x6b, 0x24, 0x71, 0x91, 0x2e, 0xa8, 0x53, 0x9a, 0x48, 0xa3, 0x78, 0x30, 0xc, 0xe8, 0xf0, 0x86, 0xa3, 0x68, 0xec, 0xe8, 0x96, 0x43, 0x34, 0xda, 0xf, 0xf4, 0x65, 0x48, 0xbb, 0xe0, 0x92, 0xa1, 0x8, 0x12, 0x18, 0x46, 0xe6, 0x4a, 0xd6, 0x92, 0x88, 0xe, 0x2, 0xf5, 0xf3, 0x2a, 0x96, 0xb1, 0x4, 0xf1, 0x11, 0xa9, 0x92, 0x79, 0x52, 0x0, 0x64, 0x34, 0xeb, 0x25, 0xe, 0xf4, 0x29, 0x6b, 0x39, 0x4e, 0x28, 0x78, 0xfe, 0x25, 0xa3, 0xc0, 0x88, 0x5a, 0x40, 0xfd, 0x71, 0x37, 0x63, 0x79, 0xcd, 0x6b, 0x56, 0xda, 0xee, 0x91, 0x26, 0x72, 0xfc, 0xbc},
					{0x8f, 0xc0, 0xb4, 0x9e, 0x2e, 0xac, 0x50, 0x86, 0xe2, 0xe2, 0xaa, 0xf, 0xdc, 0x54, 0x23, 0x51, 0x6, 0xd8, 0x29, 0xf5, 0xae, 0x3, 0x5d, 0xb8, 0x31, 0x4d, 0x26, 0x3, 0x48, 0x18, 0xb9, 0x1f, 0x6b, 0xd7, 0x86, 0xb4, 0xa2, 0x69, 0xc7, 0xe7, 0xf5, 0xc0, 0x93, 0x19, 0x6e, 0xfd, 0x33, 0xb8, 0x1, 0xe1, 0x1f, 0x4e, 0xb4, 0xb1, 0xa0, 0x1, 0x30, 0x48, 0x8a, 0x6c, 0x97, 0x29, 0xd6, 0xcb, 0x1c, 0x45, 0xef, 0x87, 0xba, 0x4f, 0xce, 0x22, 0x84, 0x48, 0xad, 0x16, 0xf7, 0x5c, 0xb2, 0xa8, 0x34, 0xb9, 0xee, 0xb8, 0xbf, 0xe5, 0x58, 0x2c, 0x44, 0x7b, 0x1f, 0x9c, 0x22, 0x26, 0x3a, 0x22},
				},
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				e.maxListenSealRetry = 1
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().SealObject(gomock.Any(), gomock.Any()).Return("", mockErr).Times(2)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().ListenObjectSeal(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).Times(1)
				e.baseApp.SetConsensus(m1)
				return e
			},
			wantedErr: ErrUnsealed,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn().HandleSealObjectTask(context.TODO(), tt.task)
		})
	}
}

func TestExecuteModular_sealObject(t *testing.T) {
	cases := []struct {
		name      string
		fn        func() *ExecuteModular
		wantedErr error
	}{
		{
			name: "object unsealed",
			fn: func() *ExecuteModular {
				e := setup(t)
				e.maxListenSealRetry = 1
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().SealObject(gomock.Any(), gomock.Any()).Return("", mockErr).Times(2)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().ListenObjectSeal(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).Times(1)
				e.baseApp.SetConsensus(m1)
				return e
			},
			wantedErr: ErrUnsealed,
		},
		{
			name: "object is sealed",
			fn: func() *ExecuteModular {
				e := setup(t)
				e.maxListenSealRetry = 1
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().SealObject(gomock.Any(), gomock.Any()).Return("txHash", nil).Times(1)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().ListenObjectSeal(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).Times(1)
				e.baseApp.SetConsensus(m1)
				return e
			},
			wantedErr: nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			objectInfo := &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)}
			task := &gfsptask.GfSpUploadObjectTask{
				Task:                 &gfsptask.GfSpTask{MaxRetry: 1},
				VirtualGroupFamilyId: 1,
				ObjectInfo:           objectInfo,
				StorageParams:        &storagetypes.Params{},
			}
			sealMsg := &storagetypes.MsgSealObject{ObjectName: "mockObjectName"}
			err := tt.fn().sealObject(context.TODO(), task, sealMsg)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestExecuteModular_listenSealObject(t *testing.T) {
	cases := []struct {
		name      string
		fn        func() *ExecuteModular
		wantedErr error
	}{
		{
			name: "failed to listen object seal",
			fn: func() *ExecuteModular {
				e := setup(t)
				e.maxListenSealRetry = 1
				ctrl := gomock.NewController(t)
				m := consensus.NewMockConsensus(ctrl)
				m.EXPECT().ListenObjectSeal(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, mockErr).Times(1)
				e.baseApp.SetConsensus(m)
				return e
			},
			wantedErr: mockErr,
		},
		{
			name: "object unsealed",
			fn: func() *ExecuteModular {
				e := setup(t)
				e.maxListenSealRetry = 1
				ctrl := gomock.NewController(t)
				m := consensus.NewMockConsensus(ctrl)
				m.EXPECT().ListenObjectSeal(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).Times(1)
				e.baseApp.SetConsensus(m)
				return e
			},
			wantedErr: ErrUnsealed,
		},
		{
			name: "object is sealed",
			fn: func() *ExecuteModular {
				e := setup(t)
				e.maxListenSealRetry = 1
				ctrl := gomock.NewController(t)
				m := consensus.NewMockConsensus(ctrl)
				m.EXPECT().ListenObjectSeal(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).Times(1)
				e.baseApp.SetConsensus(m)
				return e
			},
			wantedErr: nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			objectInfo := &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)}
			task := &gfsptask.GfSpUploadObjectTask{
				Task:                 &gfsptask.GfSpTask{},
				VirtualGroupFamilyId: 1,
				ObjectInfo:           objectInfo,
				StorageParams:        &storagetypes.Params{},
			}
			err := tt.fn().listenSealObject(context.TODO(), task, objectInfo)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestExecuteModular_HandleReceivePieceTask(t *testing.T) {
	cases := []struct {
		name string
		task coretask.ReceivePieceTask
		fn   func() *ExecuteModular
	}{
		{
			name: "task pointer dangling",
			task: &gfsptask.GfSpReceivePieceTask{},
			fn:   func() *ExecuteModular { return setup(t) },
		},
		{
			name: "failed to get object info",
			task: &gfsptask.GfSpReceivePieceTask{
				Task:          &gfsptask.GfSpTask{},
				ObjectInfo:    &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
				StorageParams: &storagetypes.Params{MaxPayloadSize: 10},
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetObjectByID(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				e.baseApp.SetGfSpClient(m)
				return e
			},
		},
		{
			name: "failed to confirm receive task, object is unsealed",
			task: &gfsptask.GfSpReceivePieceTask{
				Task:          &gfsptask.GfSpTask{},
				ObjectInfo:    &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
				StorageParams: &storagetypes.Params{MaxPayloadSize: 10},
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetObjectByID(gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{
					ObjectStatus: storagetypes.OBJECT_STATUS_CREATED}, nil).Times(1)
				e.baseApp.SetGfSpClient(m)
				return e
			},
		},
		{
			name: "failed to get bucket by bucket name",
			task: &gfsptask.GfSpReceivePieceTask{
				Task:          &gfsptask.GfSpTask{},
				ObjectInfo:    &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
				StorageParams: &storagetypes.Params{MaxPayloadSize: 10},
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetObjectByID(gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{
					ObjectStatus: storagetypes.OBJECT_STATUS_SEALED}, nil).Times(1)
				m.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				e.baseApp.SetGfSpClient(m)
				return e
			},
		},
		{
			name: "failed to get global virtual group",
			task: &gfsptask.GfSpReceivePieceTask{
				Task:          &gfsptask.GfSpTask{},
				ObjectInfo:    &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
				StorageParams: &storagetypes.Params{MaxPayloadSize: 10},
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetObjectByID(gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{
					ObjectStatus: storagetypes.OBJECT_STATUS_SEALED}, nil).Times(1)
				m.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(&metadatatypes.Bucket{
					BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil).Times(1)
				m.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				e.baseApp.SetGfSpClient(m)
				return e
			},
		},
		{
			name: "replicate idx out of bounds",
			task: &gfsptask.GfSpReceivePieceTask{
				Task:          &gfsptask.GfSpTask{},
				ObjectInfo:    &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
				StorageParams: &storagetypes.Params{MaxPayloadSize: 10},
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetObjectByID(gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{
					ObjectStatus: storagetypes.OBJECT_STATUS_SEALED}, nil).Times(1)
				m.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(&metadatatypes.Bucket{
					BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil).Times(1)
				m.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(&virtual_types.GlobalVirtualGroup{
					SecondarySpIds: []uint32{}}, nil).Times(1)
				e.baseApp.SetGfSpClient(m)
				return e
			},
		},
		{
			name: "failed to get sp id",
			task: &gfsptask.GfSpReceivePieceTask{
				Task:          &gfsptask.GfSpTask{},
				ObjectInfo:    &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
				StorageParams: &storagetypes.Params{MaxPayloadSize: 10},
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetObjectByID(gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{
					ObjectStatus: storagetypes.OBJECT_STATUS_SEALED}, nil).Times(1)
				m.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(&metadatatypes.Bucket{
					BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil).Times(1)
				m.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(&virtual_types.GlobalVirtualGroup{
					SecondarySpIds: []uint32{1}}, nil).Times(1)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				e.baseApp.SetConsensus(m1)
				return e
			},
		},
		{
			name: "success",
			task: &gfsptask.GfSpReceivePieceTask{
				Task:          &gfsptask.GfSpTask{},
				ObjectInfo:    &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
				StorageParams: &storagetypes.Params{MaxPayloadSize: 10},
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetObjectByID(gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{
					ObjectStatus: storagetypes.OBJECT_STATUS_SEALED}, nil).Times(1)
				m.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(&metadatatypes.Bucket{
					BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil).Times(1)
				m.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(&virtual_types.GlobalVirtualGroup{
					SecondarySpIds: []uint32{1}}, nil).Times(1)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				e.baseApp.SetConsensus(m1)
				return e
			},
		},
		{
			name: "failed to delete integrity, REDUNDANCY_EC_TYPE and failed to delete piece data",
			task: &gfsptask.GfSpReceivePieceTask{
				Task: &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{
					Id:             sdkmath.NewUint(1),
					RedundancyType: storagetypes.REDUNDANCY_EC_TYPE,
				},
				StorageParams: &storagetypes.Params{MaxPayloadSize: 10},
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetObjectByID(gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{
					ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Id: sdkmath.NewUint(1)}, nil).Times(1)
				m.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(&metadatatypes.Bucket{
					BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil).Times(1)
				m.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(&virtual_types.GlobalVirtualGroup{
					SecondarySpIds: []uint32{2}}, nil).Times(1)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				e.baseApp.SetConsensus(m1)

				m2 := corespdb.NewMockSPDB(ctrl)
				m2.EXPECT().DeleteObjectIntegrity(gomock.Any(), gomock.Any()).Return(mockErr).Times(1)
				e.baseApp.SetGfSpDB(m2)

				m3 := piecestore.NewMockPieceOp(ctrl)
				m3.EXPECT().SegmentPieceCount(gomock.Any(), gomock.Any()).Return(uint32(1)).Times(1)
				m3.EXPECT().ECPieceKey(gomock.Any(), gomock.Any(), gomock.Any()).Return("test").Times(1)
				e.baseApp.SetPieceOp(m3)

				m4 := piecestore.NewMockPieceStore(ctrl)
				m4.EXPECT().DeletePiece(gomock.Any(), gomock.Any()).Return(mockErr).Times(1)
				e.baseApp.SetPieceStore(m4)
				return e
			},
		},
		{
			name: "non REDUNDANCY_EC_TYPE and succeed to delete piece data",
			task: &gfsptask.GfSpReceivePieceTask{
				Task: &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{
					Id:             sdkmath.NewUint(1),
					RedundancyType: storagetypes.REDUNDANCY_REPLICA_TYPE,
				},
				StorageParams: &storagetypes.Params{MaxPayloadSize: 10},
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetObjectByID(gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{
					ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Id: sdkmath.NewUint(1)}, nil).Times(1)
				m.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(&metadatatypes.Bucket{
					BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil).Times(1)
				m.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(&virtual_types.GlobalVirtualGroup{
					SecondarySpIds: []uint32{2}}, nil).Times(1)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				e.baseApp.SetConsensus(m1)

				m2 := spdb.NewMockSPDB(ctrl)
				m2.EXPECT().DeleteObjectIntegrity(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				e.baseApp.SetGfSpDB(m2)

				m3 := piecestore.NewMockPieceOp(ctrl)
				m3.EXPECT().SegmentPieceCount(gomock.Any(), gomock.Any()).Return(uint32(1)).Times(1)
				m3.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").Times(1)
				e.baseApp.SetPieceOp(m3)

				m4 := piecestore.NewMockPieceStore(ctrl)
				m4.EXPECT().DeletePiece(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				e.baseApp.SetPieceStore(m4)
				return e
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn().HandleReceivePieceTask(context.TODO(), tt.task)
		})
	}
}

func TestExecuteModular_HandleGCObjectTask(t *testing.T) {
	cases := []struct {
		name string
		task coretask.GCObjectTask
		fn   func() *ExecuteModular
	}{
		{
			name: "failed to query deleted object list",
			task: &gfsptask.GfSpGCObjectTask{
				Task: &gfsptask.GfSpTask{},
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().ListDeletedObjectsByBlockNumberRange(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(nil, uint64(0), mockErr).Times(1)
				m.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				e.baseApp.SetGfSpClient(m)
				return e
			},
		},
		{
			name: "metadata is not latest",
			task: &gfsptask.GfSpGCObjectTask{
				Task:             &gfsptask.GfSpTask{},
				StartBlockNumber: 1,
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().ListDeletedObjectsByBlockNumberRange(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(nil, uint64(0), nil).Times(1)
				m.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				e.baseApp.SetGfSpClient(m)
				return e
			},
		},
		{
			name: "no waiting gc objects",
			task: &gfsptask.GfSpGCObjectTask{
				Task: &gfsptask.GfSpTask{},
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().ListDeletedObjectsByBlockNumberRange(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(nil, uint64(0), nil).Times(1)
				m.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				e.baseApp.SetGfSpClient(m)
				return e
			},
		},
		{
			name: "failed to query storage params",
			task: &gfsptask.GfSpGCObjectTask{
				Task: &gfsptask.GfSpTask{},
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				waitingGCObjects := []*metadatatypes.Object{
					{
						ObjectInfo: &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
					},
				}
				m.EXPECT().ListDeletedObjectsByBlockNumberRange(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(waitingGCObjects, uint64(0), nil).Times(1)
				m.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				e.baseApp.SetConsensus(m1)
				return e
			},
		},
		{
			name: "failed to get bucket by bucket name",
			task: &gfsptask.GfSpGCObjectTask{
				Task:               &gfsptask.GfSpTask{},
				CurrentBlockNumber: 0,
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				waitingGCObjects := []*metadatatypes.Object{
					{
						ObjectInfo: &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
					},
				}
				m.EXPECT().ListDeletedObjectsByBlockNumberRange(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(waitingGCObjects, uint64(0), nil).Times(1)
				m.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				m.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(&storagetypes.Params{
					VersionedParams: storagetypes.VersionedParams{MaxSegmentSize: 10}}, nil).Times(1)
				e.baseApp.SetConsensus(m1)

				m2 := piecestore.NewMockPieceOp(ctrl)
				m2.EXPECT().SegmentPieceCount(gomock.Any(), gomock.Any()).Return(uint32(1)).Times(1)
				m2.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").Times(1)
				e.baseApp.SetPieceOp(m2)

				m3 := piecestore.NewMockPieceStore(ctrl)
				m3.EXPECT().DeletePiece(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				e.baseApp.SetPieceStore(m3)
				return e
			},
		},
		{
			name: "failed to get global virtual group",
			task: &gfsptask.GfSpGCObjectTask{
				Task:               &gfsptask.GfSpTask{},
				CurrentBlockNumber: 0,
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				waitingGCObjects := []*metadatatypes.Object{
					{
						ObjectInfo: &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
					},
				}
				m.EXPECT().ListDeletedObjectsByBlockNumberRange(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(waitingGCObjects, uint64(0), nil).Times(1)
				m.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				m.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(&metadatatypes.Bucket{
					BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil).Times(1)
				m.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(&storagetypes.Params{
					VersionedParams: storagetypes.VersionedParams{MaxSegmentSize: 10}}, nil).Times(1)
				e.baseApp.SetConsensus(m1)

				m2 := piecestore.NewMockPieceOp(ctrl)
				m2.EXPECT().SegmentPieceCount(gomock.Any(), gomock.Any()).Return(uint32(1)).Times(1)
				m2.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").Times(1)
				e.baseApp.SetPieceOp(m2)

				m3 := piecestore.NewMockPieceStore(ctrl)
				m3.EXPECT().DeletePiece(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				e.baseApp.SetPieceStore(m3)
				return e
			},
		},
		{
			name: "failed to get sp id",
			task: &gfsptask.GfSpGCObjectTask{
				Task:               &gfsptask.GfSpTask{},
				CurrentBlockNumber: 0,
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				waitingGCObjects := []*metadatatypes.Object{
					{
						ObjectInfo: &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
					},
				}
				m.EXPECT().ListDeletedObjectsByBlockNumberRange(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(waitingGCObjects, uint64(0), nil).Times(1)
				m.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				m.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(&metadatatypes.Bucket{
					BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil).Times(1)
				m.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(&virtual_types.GlobalVirtualGroup{
					SecondarySpIds: []uint32{1}}, nil).Times(1)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(&storagetypes.Params{
					VersionedParams: storagetypes.VersionedParams{MaxSegmentSize: 10}}, nil).Times(1)
				m1.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				e.baseApp.SetConsensus(m1)

				m2 := piecestore.NewMockPieceOp(ctrl)
				m2.EXPECT().SegmentPieceCount(gomock.Any(), gomock.Any()).Return(uint32(1)).Times(1)
				m2.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").Times(1)
				e.baseApp.SetPieceOp(m2)

				m3 := piecestore.NewMockPieceStore(ctrl)
				m3.EXPECT().DeletePiece(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				e.baseApp.SetPieceStore(m3)
				return e
			},
		},
		{
			name: "succeed to gc an object",
			task: &gfsptask.GfSpGCObjectTask{
				Task:               &gfsptask.GfSpTask{},
				CurrentBlockNumber: 0,
			},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				waitingGCObjects := []*metadatatypes.Object{
					{
						ObjectInfo: &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
					},
				}
				m.EXPECT().ListDeletedObjectsByBlockNumberRange(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(waitingGCObjects, uint64(0), nil).AnyTimes()
				m.EXPECT().ReportTask(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				m.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(&metadatatypes.Bucket{
					BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil).Times(1)
				m.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(&virtual_types.GlobalVirtualGroup{
					SecondarySpIds: []uint32{1}}, nil).Times(1)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(&storagetypes.Params{
					VersionedParams: storagetypes.VersionedParams{MaxSegmentSize: 10}}, nil).Times(1)
				m1.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				e.baseApp.SetConsensus(m1)

				m2 := piecestore.NewMockPieceOp(ctrl)
				m2.EXPECT().SegmentPieceCount(gomock.Any(), gomock.Any()).Return(uint32(1)).Times(1)
				m2.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").Times(1)
				m2.EXPECT().ECPieceKey(gomock.Any(), gomock.Any(), gomock.Any()).Return("test").Times(1)
				e.baseApp.SetPieceOp(m2)

				m3 := piecestore.NewMockPieceStore(ctrl)
				m3.EXPECT().DeletePiece(gomock.Any(), gomock.Any()).Return(nil).Times(2)
				e.baseApp.SetPieceStore(m3)

				m4 := corespdb.NewMockSPDB(ctrl)
				m4.EXPECT().DeleteObjectIntegrity(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				e.baseApp.SetGfSpDB(m4)
				return e
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn().HandleGCObjectTask(context.TODO(), tt.task)
		})
	}
}

// TODO add unit test
//func TestExecuteModular_HandleGCZombiePieceTask(t *testing.T) {
//	e := setup(t)
//	e.HandleGCZombiePieceTask(context.TODO(), nil)
//}
//
//func TestExecuteModular_HandleGCMetaTask(t *testing.T) {
//	e := setup(t)
//	e.HandleGCMetaTask(context.TODO(), nil)
//}

func TestExecuteModular_HandleRecoverPieceTaskFailure1(t *testing.T) {
	t.Log("Failure case description: ErrDanglingPointer")
	e := setup(t)
	task := &gfsptask.GfSpRecoverPieceTask{
		Task:          &gfsptask.GfSpTask{},
		StorageParams: &storagetypes.Params{},
	}
	e.HandleRecoverPieceTask(context.TODO(), task)
}

func TestExecuteModular_HandleRecoverPieceTaskFailure2(t *testing.T) {
	t.Log("Failure case description: ErrRecoveryRedundancyType")
	e := setup(t)
	task := &gfsptask.GfSpRecoverPieceTask{
		Task: &gfsptask.GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:             sdkmath.NewUint(1),
			RedundancyType: storagetypes.REDUNDANCY_REPLICA_TYPE,
		},
		StorageParams: &storagetypes.Params{},
	}
	e.HandleRecoverPieceTask(context.TODO(), task)
}

func TestExecuteModular_HandleRecoverPieceTaskFailure3(t *testing.T) {
	t.Log("Failure case description: ErrRecoveryPieceIndex")
	e := setup(t)
	task := &gfsptask.GfSpRecoverPieceTask{
		Task: &gfsptask.GfSpTask{},
		ObjectInfo: &storagetypes.ObjectInfo{
			Id:             sdkmath.NewUint(1),
			RedundancyType: storagetypes.REDUNDANCY_EC_TYPE,
		},
		StorageParams: &storagetypes.Params{},
		EcIdx:         -2,
	}
	e.HandleRecoverPieceTask(context.TODO(), task)
}

func TestExecuteModular_recoverByPrimarySPSuccess(t *testing.T) {
	e := setup(t)

	ctrl := gomock.NewController(t)
	m := gfspclient.NewMockGfSpClientAPI(ctrl)
	m.EXPECT().GetBucketMeta(gomock.Any(), gomock.Any(), true).Return(&metadatatypes.VGFInfoBucket{
		BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil, nil).Times(1)
	m.EXPECT().SignRecoveryTask(gomock.Any(), gomock.Any()).Return([]byte("mockSig"), nil).Times(1)
	m.EXPECT().GetPieceFromECChunks(gomock.Any(), gomock.Any(), gomock.Any()).Return(io.NopCloser(
		strings.NewReader("body")), nil).Times(1)
	e.baseApp.SetGfSpClient(m)

	m1 := consensus.NewMockConsensus(ctrl)
	m1.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtual_types.GlobalVirtualGroupFamily{
		PrimarySpId: 1}, nil).Times(1)
	m1.EXPECT().ListSPs(gomock.Any()).Return([]*sptypes.StorageProvider{
		{Id: 1, Endpoint: "endpoint"}}, nil).Times(1)
	e.baseApp.SetConsensus(m1)

	m2 := corespdb.NewMockSPDB(ctrl)
	m2.EXPECT().GetObjectIntegrity(gomock.Any(), gomock.Any()).Return(&corespdb.IntegrityMeta{
		PieceChecksumList: [][]byte{[]byte{35, 13, 131, 88, 220, 142, 136, 144, 180, 197, 141, 238, 182, 41, 18, 238, 47,
			32, 53, 122, 233, 42, 92, 200, 97, 185, 142, 104, 254, 49, 172, 181}}}, nil).Times(1)
	e.baseApp.SetGfSpDB(m2)

	m3 := piecestore.NewMockPieceOp(ctrl)
	m3.EXPECT().ECPieceKey(gomock.Any(), gomock.Any(), gomock.Any()).Return("test").Times(1)
	e.baseApp.SetPieceOp(m3)

	m4 := piecestore.NewMockPieceStore(ctrl)
	m4.EXPECT().PutPiece(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	e.baseApp.SetPieceStore(m4)

	task := &gfsptask.GfSpRecoverPieceTask{
		Task:          &gfsptask.GfSpTask{},
		ObjectInfo:    &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
		StorageParams: &storagetypes.Params{},
	}
	err := e.recoverByPrimarySP(context.TODO(), task)
	assert.Nil(t, err)
}

func TestExecuteModular_recoverBySecondarySPFailure1(t *testing.T) {
	e := setup(t)

	ctrl := gomock.NewController(t)
	m := gfspclient.NewMockGfSpClientAPI(ctrl)
	m.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), true).Return(nil, mockErr).Times(1)
	e.baseApp.SetGfSpClient(m)

	task := &gfsptask.GfSpRecoverPieceTask{
		Task:          &gfsptask.GfSpTask{},
		ObjectInfo:    &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
		StorageParams: &storagetypes.Params{},
	}
	err := e.recoverBySecondarySP(context.TODO(), task, true)
	assert.Equal(t, mockErr, err)
}

func TestExecuteModular_recoverBySecondarySPFailure2(t *testing.T) {
	e := setup(t)

	ctrl := gomock.NewController(t)
	m := gfspclient.NewMockGfSpClientAPI(ctrl)
	m.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), true).Return(
		&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil).Times(1)
	m.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&virtual_types.GlobalVirtualGroup{SecondarySpIds: []uint32{1}}, nil).Times(1)
	e.baseApp.SetGfSpClient(m)

	m1 := consensus.NewMockConsensus(ctrl)
	m1.EXPECT().ListSPs(gomock.Any()).Return([]*sptypes.StorageProvider{{Id: 1, Endpoint: "endpoint"}}, nil).Times(1)
	e.baseApp.SetConsensus(m1)

	task := &gfsptask.GfSpRecoverPieceTask{
		Task:          &gfsptask.GfSpTask{},
		ObjectInfo:    &storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)},
		StorageParams: &storagetypes.Params{},
	}
	err := e.recoverBySecondarySP(context.TODO(), task, true)
	assert.Equal(t, ErrRecoveryPieceNotEnough, err)
}

func TestExecuteModular_getECPieceBySegment(t *testing.T) {
	cases := []struct {
		name          string
		redundancyIdx int32
		objectInfo    *storagetypes.ObjectInfo
		params        *storagetypes.Params
		fn            func() *ExecuteModular
		wantedIsErr   bool
		wantedErrStr  string
	}{
		{
			name:          "invalid redundancyIdx",
			redundancyIdx: -1,
			objectInfo:    &storagetypes.ObjectInfo{},
			params:        &storagetypes.Params{},
			fn:            func() *ExecuteModular { return setup(t) },
			wantedIsErr:   true,
			wantedErrStr:  "invalid redundancyIdx",
		},
		{
			name:          "no error",
			redundancyIdx: 1,
			objectInfo:    &storagetypes.ObjectInfo{},
			params:        &storagetypes.Params{VersionedParams: storagetypes.VersionedParams{RedundantDataChunkNum: 4}},
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := piecestore.NewMockPieceOp(ctrl)
				m.EXPECT().ECPieceSize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
				e.baseApp.SetPieceOp(m)
				return e
			},
			wantedIsErr: false,
		},
		{
			name:          "no error",
			redundancyIdx: 4,
			objectInfo:    &storagetypes.ObjectInfo{},
			params: &storagetypes.Params{VersionedParams: storagetypes.VersionedParams{RedundantDataChunkNum: 4,
				RedundantParityChunkNum: 2}},
			fn:          func() *ExecuteModular { return setup(t) },
			wantedIsErr: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn().getECPieceBySegment(context.TODO(), tt.redundancyIdx, tt.objectInfo, tt.params,
				[]byte("test"), 1)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErrStr)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestExecuteModular_checkRecoveryChecksum(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *ExecuteModular
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name: "failed to get object integrity hash",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := corespdb.NewMockSPDB(ctrl)
				m.EXPECT().GetObjectIntegrity(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				e.baseApp.SetGfSpDB(m)
				return e
			},
			wantedIsErr:  true,
			wantedErrStr: "failed to get object integrity hash in db",
		},
		{
			name: "check integrity hash of recovery data err",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := corespdb.NewMockSPDB(ctrl)
				m.EXPECT().GetObjectIntegrity(gomock.Any(), gomock.Any()).Return(&corespdb.IntegrityMeta{
					PieceChecksumList: [][]byte{[]byte("a")}}, nil).Times(1)
				e.baseApp.SetGfSpDB(m)
				return e
			},
			wantedIsErr:  true,
			wantedErrStr: ErrRecoveryPieceChecksum.Error(),
		},
		{
			name: "success",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := corespdb.NewMockSPDB(ctrl)
				m.EXPECT().GetObjectIntegrity(gomock.Any(), gomock.Any()).Return(&corespdb.IntegrityMeta{
					PieceChecksumList: [][]byte{[]byte("test")}}, nil).Times(1)
				e.baseApp.SetGfSpDB(m)
				return e
			},
			wantedIsErr: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn().checkRecoveryChecksum(context.TODO(), &gfsptask.GfSpRecoverPieceTask{
				Task: &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{
					ObjectName: "mockObjectName",
					Id:         sdkmath.NewUint(1),
				}}, []byte{116, 101, 115, 116})
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErrStr)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestExecuteModular_doRecoveryPiece(t *testing.T) {
	cases := []struct {
		name        string
		fn          func() *ExecuteModular
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name: "failed to sign recovery task",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().SignRecoveryTask(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				e.baseApp.SetGfSpClient(m)
				return e
			},
			wantedIsErr: true,
			wantedErr:   mockErr,
		},
		{
			name: "failed to get piece from ec chunks",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().SignRecoveryTask(gomock.Any(), gomock.Any()).Return([]byte("mockSig"), nil).Times(1)
				m.EXPECT().GetPieceFromECChunks(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				e.baseApp.SetGfSpClient(m)
				return e
			},
			wantedIsErr: true,
			wantedErr:   mockErr,
		},
		{
			name: "failed to read recovery piece data from sp",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				e.baseApp.SetGfSpClient(m)
				m.EXPECT().SignRecoveryTask(gomock.Any(), gomock.Any()).Return([]byte("mockSig"), nil).Times(1)

				m1 := gfspclient.NewMockstdLib(ctrl)
				m1.EXPECT().Read(gomock.Any()).Return(0, mockErr).Times(1)
				m1.EXPECT().Close().AnyTimes()
				m.EXPECT().GetPieceFromECChunks(gomock.Any(), gomock.Any(), gomock.Any()).Return(m1, nil).Times(1)
				return e
			},
			wantedIsErr: true,
			wantedErr:   mockErr,
		},
		{
			name: "success",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				e.baseApp.SetGfSpClient(m)
				m.EXPECT().SignRecoveryTask(gomock.Any(), gomock.Any()).Return([]byte("mockSig"), nil).Times(1)
				m.EXPECT().GetPieceFromECChunks(gomock.Any(), gomock.Any(), gomock.Any()).Return(io.NopCloser(
					strings.NewReader("body")), nil).Times(1)
				return e
			},
			wantedIsErr: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn().doRecoveryPiece(context.TODO(), &gfsptask.GfSpRecoverPieceTask{
				Task: &gfsptask.GfSpTask{},
				ObjectInfo: &storagetypes.ObjectInfo{
					ObjectName: "mockObjectName",
					Id:         sdkmath.NewUint(1),
				}}, "mockEndpoint")
			if tt.wantedIsErr {
				assert.Equal(t, tt.wantedErr, err)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestExecuteModular_getObjectSecondaryEndpoints(t *testing.T) {
	cases := []struct {
		name        string
		fn          func() *ExecuteModular
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name: "failed to GetBucketByBucketName",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), true).Return(nil, mockErr).Times(1)
				e.baseApp.SetGfSpClient(m)
				return e
			},
			wantedIsErr: true,
			wantedErr:   mockErr,
		},
		{
			name: "failed to GetGlobalVirtualGroup",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), true).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil).Times(1)
				m.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				e.baseApp.SetGfSpClient(m)
				return e
			},
			wantedIsErr: true,
			wantedErr:   mockErr,
		},
		{
			name: "failed to ListSPs",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), true).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil).Times(1)
				m.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&virtual_types.GlobalVirtualGroup{SecondarySpIds: []uint32{1}}, nil).Times(1)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().ListSPs(gomock.Any()).Return(nil, mockErr).Times(1)
				e.baseApp.SetConsensus(m1)
				return e
			},
			wantedIsErr: true,
			wantedErr:   mockErr,
		},
		{
			name: "success",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), true).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil).Times(1)
				m.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&virtual_types.GlobalVirtualGroup{SecondarySpIds: []uint32{1}}, nil).Times(1)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().ListSPs(gomock.Any()).Return([]*sptypes.StorageProvider{{Id: 1, Endpoint: "endpoint"}}, nil).Times(1)
				e.baseApp.SetConsensus(m1)
				return e
			},
			wantedIsErr: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result1, result2, err := tt.fn().getObjectSecondaryEndpoints(context.TODO(), &storagetypes.ObjectInfo{
				BucketName: "mockBucketName", LocalVirtualGroupId: 1})
			if tt.wantedIsErr {
				assert.Equal(t, tt.wantedErr, err)
				assert.Nil(t, result1)
				assert.Equal(t, 0, result2)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, []string{"endpoint"}, result1)
				assert.Equal(t, 1, result2)
			}
		})
	}
}

func TestExecuteModular_getBucketPrimarySPEndpoint(t *testing.T) {
	cases := []struct {
		name        string
		fn          func() *ExecuteModular
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name: "failed to GetBucketMeta",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetBucketMeta(gomock.Any(), gomock.Any(), true).Return(nil, nil, mockErr).Times(1)
				e.baseApp.SetGfSpClient(m)
				return e
			},
			wantedIsErr: true,
			wantedErr:   mockErr,
		},
		{
			name: "failed to GetBucketPrimarySPID",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetBucketMeta(gomock.Any(), gomock.Any(), true).Return(&metadatatypes.VGFInfoBucket{
					BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil, nil).Times(1)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				e.baseApp.SetConsensus(m1)
				return e
			},
			wantedIsErr: true,
			wantedErr:   mockErr,
		},
		{
			name: "failed to ListSPs",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetBucketMeta(gomock.Any(), gomock.Any(), true).Return(&metadatatypes.VGFInfoBucket{
					BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil, nil).Times(1)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtual_types.GlobalVirtualGroupFamily{
					PrimarySpId: 1}, nil).Times(1)
				m1.EXPECT().ListSPs(gomock.Any()).Return(nil, mockErr).Times(1)
				e.baseApp.SetConsensus(m1)
				return e
			},
			wantedIsErr: true,
			wantedErr:   mockErr,
		},
		{
			name: "ErrPrimaryNotFound",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetBucketMeta(gomock.Any(), gomock.Any(), true).Return(&metadatatypes.VGFInfoBucket{
					BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil, nil).Times(1)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtual_types.GlobalVirtualGroupFamily{
					PrimarySpId: 1}, nil).Times(1)
				m1.EXPECT().ListSPs(gomock.Any()).Return([]*sptypes.StorageProvider{{Id: 2}}, nil).Times(1)
				e.baseApp.SetConsensus(m1)
				return e
			},
			wantedIsErr: true,
			wantedErr:   ErrPrimaryNotFound,
		},
		{
			name: "success",
			fn: func() *ExecuteModular {
				e := setup(t)
				ctrl := gomock.NewController(t)
				m := gfspclient.NewMockGfSpClientAPI(ctrl)
				m.EXPECT().GetBucketMeta(gomock.Any(), gomock.Any(), true).Return(&metadatatypes.VGFInfoBucket{
					BucketInfo: &storagetypes.BucketInfo{Id: sdkmath.NewUint(1)}}, nil, nil).Times(1)
				e.baseApp.SetGfSpClient(m)

				m1 := consensus.NewMockConsensus(ctrl)
				m1.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtual_types.GlobalVirtualGroupFamily{
					PrimarySpId: 1}, nil).Times(1)
				m1.EXPECT().ListSPs(gomock.Any()).Return([]*sptypes.StorageProvider{
					{Id: 1, Endpoint: "endpoint"}}, nil).Times(1)
				e.baseApp.SetConsensus(m1)
				return e
			},
			wantedIsErr: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn().getBucketPrimarySPEndpoint(context.TODO(), "bucketName")
			if tt.wantedIsErr {
				assert.Equal(t, tt.wantedErr, err)
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

package metadata

import (
	"context"
	"errors"
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	storetypes "github.com/bnb-chain/greenfield-storage-provider/store/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
)

func TestMetadataModular_GfSpQueryUploadProgress_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	mockSPDB := spdb.NewMockSPDB(ctrl)
	a.baseApp.SetGfSpDB(mockSPDB)
	mockSPDB.EXPECT().GetUploadState(gomock.Any()).DoAndReturn(
		func(uint642 uint64) (storetypes.TaskState, string, error) {
			return storetypes.TaskState_TASK_STATE_UPLOAD_OBJECT_DOING, "", nil
		},
	).Times(1)
	state, err := a.GfSpQueryUploadProgress(context.Background(), &types.GfSpQueryUploadProgressRequest{ObjectId: 1})
	assert.Nil(t, err)
	assert.Equal(t, "TASK_STATE_UPLOAD_OBJECT_DOING", state.State.String())
}

func TestMetadataModular_GfSpQueryUploadProgress_ErrRecordNotFound(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	mockSPDB := spdb.NewMockSPDB(ctrl)
	a.baseApp.SetGfSpDB(mockSPDB)
	mockSPDB.EXPECT().GetUploadState(gomock.Any()).DoAndReturn(
		func(uint642 uint64) (storetypes.TaskState, string, error) {
			return 0, gorm.ErrRecordNotFound.Error(), gorm.ErrRecordNotFound
		},
	).Times(1)
	state, err := a.GfSpQueryUploadProgress(context.Background(), &types.GfSpQueryUploadProgressRequest{ObjectId: 1})
	assert.Nil(t, err)
	assert.Equal(t, "no uploading record", state.Err.Description)
}

func TestMetadataModular_GfSpQueryUploadProgress_Err(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	mockSPDB := spdb.NewMockSPDB(ctrl)
	a.baseApp.SetGfSpDB(mockSPDB)
	mockSPDB.EXPECT().GetUploadState(gomock.Any()).DoAndReturn(
		func(uint642 uint64) (storetypes.TaskState, string, error) {
			return 0, "", errors.New("test error")
		},
	).Times(1)
	state, err := a.GfSpQueryUploadProgress(context.Background(), &types.GfSpQueryUploadProgressRequest{ObjectId: 1})
	assert.Nil(t, err)
	assert.Equal(t, "GfSpQueryUploadProgress error:"+"test error", state.Err.Description)
}

func TestMetadataModular_GfSpQueryResumableUploadSegment_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	mockSPDB := spdb.NewMockSPDB(ctrl)
	a.baseApp.SetGfSpDB(mockSPDB)
	mockSPDB.EXPECT().GetObjectIntegrity(gomock.Any(), gomock.Any()).DoAndReturn(
		func(uint64, int32) (*spdb.IntegrityMeta, error) {
			return &spdb.IntegrityMeta{
				ObjectID:          1,
				RedundancyIndex:   1,
				IntegrityChecksum: []byte{'a'},
				PieceChecksumList: [][]byte{
					[]byte("hello"),
					[]byte("world"),
					[]byte("golang"),
				},
			}, nil
		},
	).Times(1)
	state, err := a.GfSpQueryResumableUploadSegment(context.Background(), &types.GfSpQueryResumableUploadSegmentRequest{ObjectId: 1})
	assert.Nil(t, err)
	assert.Equal(t, uint32(3), state.SegmentCount)
}

func TestMetadataModular_GfSpQueryResumableUploadSegment_ErrRecordNotFound(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	mockSPDB := spdb.NewMockSPDB(ctrl)
	a.baseApp.SetGfSpDB(mockSPDB)
	mockSPDB.EXPECT().GetObjectIntegrity(gomock.Any(), gomock.Any()).DoAndReturn(
		func(uint64, int32) (*spdb.IntegrityMeta, error) {
			return nil, gorm.ErrRecordNotFound
		},
	).Times(1)
	state, err := a.GfSpQueryResumableUploadSegment(context.Background(), &types.GfSpQueryResumableUploadSegmentRequest{ObjectId: 1})
	assert.Nil(t, err)
	assert.Equal(t, "no uploading record", state.Err.Description)
}

func TestMetadataModular_GfSpQueryResumableUploadSegment_Err(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	mockSPDB := spdb.NewMockSPDB(ctrl)
	a.baseApp.SetGfSpDB(mockSPDB)
	mockSPDB.EXPECT().GetObjectIntegrity(gomock.Any(), gomock.Any()).DoAndReturn(
		func(uint64, int32) (*spdb.IntegrityMeta, error) {
			return nil, errors.New("test error")
		},
	).Times(1)
	state, err := a.GfSpQueryResumableUploadSegment(context.Background(), &types.GfSpQueryResumableUploadSegmentRequest{ObjectId: 1})
	assert.Nil(t, err)
	assert.Equal(t, "GfSpQueryResumableUploadSegment error: "+"test error", state.Err.Description)
}

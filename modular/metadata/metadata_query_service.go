package metadata

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
)

func (r *MetadataModular) GfSpQueryUploadProgress(ctx context.Context, req *types.GfSpQueryUploadProgressRequest) (
	*types.GfSpQueryUploadProgressResponse, error) {
	state, err := r.baseApp.GfSpDB().GetUploadState(req.GetObjectId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &types.GfSpQueryUploadProgressResponse{
				Err: ErrNoRecord,
			}, nil
		}
		return &types.GfSpQueryUploadProgressResponse{
			Err: ErrGfSpDB,
		}, nil
	}
	return &types.GfSpQueryUploadProgressResponse{
		State: state,
	}, nil
}

func (r *MetadataModular) GfSpQueryResumableUploadOffset(ctx context.Context, req *types.GfSpQueryResumableUploadOffsetRequest) (
	*types.GfSpQueryResumableUploadOffsetResponse, error) {
	integrityMeta, err := r.baseApp.GfSpDB().GetObjectIntegrity(req.GetObjectId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &types.GfSpQueryResumableUploadOffsetResponse{
				Err: ErrNoRecord,
			}, nil
		}
		return &types.GfSpQueryResumableUploadOffsetResponse{
			Err: ErrGfSpDB,
		}, nil
	}
	return &types.GfSpQueryResumableUploadOffsetResponse{
		SegmentCount: uint64(len(integrityMeta.PieceChecksumList)),
	}, nil
}

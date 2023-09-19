package metadata

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
)

func (r *MetadataModular) GfSpQueryUploadProgress(ctx context.Context, req *types.GfSpQueryUploadProgressRequest) (
	*types.GfSpQueryUploadProgressResponse, error) {
	state, errDescription, err := r.baseApp.GfSpDB().GetUploadState(req.GetObjectId())
	if err != nil {
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			return &types.GfSpQueryUploadProgressResponse{Err: ErrNoRecord}, nil
		}
		return &types.GfSpQueryUploadProgressResponse{
			Err: ErrGfSpDBWithDetail("GfSpQueryUploadProgress error:" + err.Error()),
		}, nil
	}
	return &types.GfSpQueryUploadProgressResponse{
		State:          state,
		ErrDescription: errDescription,
	}, nil
}

func (r *MetadataModular) GfSpQueryResumableUploadSegment(ctx context.Context, req *types.GfSpQueryResumableUploadSegmentRequest) (
	*types.GfSpQueryResumableUploadSegmentResponse, error) {
	integrityMeta, err := r.baseApp.GfSpDB().GetObjectIntegrity(req.GetObjectId(), -1)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &types.GfSpQueryResumableUploadSegmentResponse{
				Err: ErrNoRecord,
			}, nil
		}
		return &types.GfSpQueryResumableUploadSegmentResponse{
			Err: ErrGfSpDBWithDetail("GfSpQueryResumableUploadSegment error: " + err.Error()),
		}, nil
	}
	return &types.GfSpQueryResumableUploadSegmentResponse{
		SegmentCount: uint32(len(integrityMeta.PieceChecksumList)),
	}, nil
}

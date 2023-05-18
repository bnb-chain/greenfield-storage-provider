package retriever

import (
	"context"
	"errors"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/modular/retriever/types"
)

func (r *RetrieveModular) GfSpQueryUploadProgress(
	ctx context.Context,
	req *types.GfSpQueryUploadProgressRequest) (
	*types.GfSpQueryUploadProgressResponse, error) {
	job, err := r.baseApp.GfSpDB().GetJobByObjectID(req.GetObjectId())
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
		State: job.JobState,
	}, nil
}

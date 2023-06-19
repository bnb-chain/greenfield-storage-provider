package gater

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func getObjectChainMeta(reqCtx *RequestContext, baseApp *gfspapp.GfSpBaseApp) (*storagetypes.ObjectInfo, *storagetypes.BucketInfo, *storagetypes.Params, error) {
	objectInfo, err := baseApp.Consensus().QueryObjectInfo(reqCtx.Context(), reqCtx.bucketName, reqCtx.objectName)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		return nil, nil, nil, ErrConsensus
	}

	bucketInfo, err := baseApp.Consensus().QueryBucketInfo(reqCtx.Context(), objectInfo.GetBucketName())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket info from consensus", "error", err)
		return nil, nil, nil, ErrConsensus
	}

	params, err := baseApp.Consensus().QueryStorageParamsByTimestamp(
		reqCtx.Context(), objectInfo.GetCreateAt())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params", "error", err)
		return nil, nil, nil, ErrConsensus
	}

	return objectInfo, bucketInfo, params, nil
}

package gater

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func getObjectChainMeta(reqCtx *RequestContext, baseApp *gfspapp.GfSpBaseApp, objectName, bucketName string) (*storagetypes.ObjectInfo, *storagetypes.BucketInfo, *storagetypes.Params, error) {
	objectInfo, err := baseApp.Consensus().QueryObjectInfo(reqCtx.Context(), bucketName, objectName)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		return nil, nil, nil, ErrConsensusWithDetail("failed to get object info from consensus, error: " + err.Error())
	}

	bucketInfo, err := baseApp.Consensus().QueryBucketInfo(reqCtx.Context(), objectInfo.GetBucketName())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket info from consensus", "error", err)
		return nil, nil, nil, ErrConsensusWithDetail("failed to get bucket info from consensus, error: " + err.Error())
	}

	params, err := baseApp.Consensus().QueryStorageParamsByTimestamp(
		reqCtx.Context(), objectInfo.GetCreateAt())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params", "error", err)
		return nil, nil, nil, ErrConsensusWithDetail("failed to get storage params, error: " + err.Error())
	}

	return objectInfo, bucketInfo, params, nil
}

// checkSPAndBucketStatus check sp and bucket is in right status
func (g *GateModular) checkSPAndBucketStatus(ctx context.Context, bucketName string, creatorAddr string) error {
	spInfo, err := g.baseApp.Consensus().QuerySP(ctx, g.baseApp.OperatorAddress())
	if err != nil {
		log.Errorw("failed to query sp by operator address", "operator_address", g.baseApp.OperatorAddress(),
			"error", err)
		return ErrConsensusWithDetail("failed to query sp by operator address, operator_address: " + g.baseApp.OperatorAddress() +
			", error: " + err.Error())
	}
	spStatus := spInfo.GetStatus()
	if spStatus != sptypes.STATUS_IN_SERVICE && !fromSpMaintenanceAcct(spStatus, spInfo.MaintenanceAddress, creatorAddr) {
		log.Errorw("sp is not in service status", "operator_address", g.baseApp.OperatorAddress(),
			"sp_status", spStatus, "sp_id", spInfo.GetId(), "endpoint", spInfo.GetEndpoint())
		return ErrSPUnavailable
	}

	bucketInfo, err := g.baseApp.Consensus().QueryBucketInfo(ctx, bucketName)
	if err != nil {
		log.Errorw("failed to query bucket info by bucket name", "bucket_name", bucketName, "error", err)
		return ErrConsensusWithDetail("failed to query bucket info by bucket name, bucket_name: " + bucketName + ", error: " + err.Error())
	}
	bucketStatus := bucketInfo.GetBucketStatus()
	if bucketStatus != storagetypes.BUCKET_STATUS_CREATED {
		log.Errorw("bucket is not in created status", "bucket_name", bucketName, "bucket_status", bucketStatus,
			"bucket_id", bucketInfo.Id.String())
		return ErrBucketUnavailable
	}
	log.Info("sp and bucket status is right")
	return nil
}

func fromSpMaintenanceAcct(spStatus sptypes.Status, spMaintenanceAddr, creatorAddr string) bool {
	return spStatus == sptypes.STATUS_IN_MAINTENANCE && spMaintenanceAddr == creatorAddr
}

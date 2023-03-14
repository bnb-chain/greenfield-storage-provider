package service

import (
	"context"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	metatypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

// ListObjectsByBucketName list objects info by a bucket name
func (metadata *Metadata) ListObjectsByBucketName(ctx context.Context, req *metatypes.MetadataServiceListObjectsByBucketNameRequest) (resp *metatypes.MetadataServiceListObjectsByBucketNameResponse, err error) {
	ctx = log.Context(ctx, req)

	objects, err := metadata.spDB.ListObjectsByBucketName(req.BucketName)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list objects by bucket name", "error", err)
		return
	}

	res := make([]*metatypes.Object, 0)
	for _, object := range objects {
		res = append(res, &metatypes.Object{
			ObjectInfo: &types.ObjectInfo{
				Owner:                object.Owner.String(),
				BucketName:           object.BucketName,
				ObjectName:           object.ObjectName,
				Id:                   math.NewUint(uint64(object.ObjectID)),
				PayloadSize:          object.PayloadSize,
				IsPublic:             object.IsPublic,
				ContentType:          object.ContentType,
				CreateAt:             object.CreateAt,
				ObjectStatus:         types.ObjectStatus(types.ObjectStatus_value[object.ObjectStatus]),
				RedundancyType:       types.RedundancyType(types.RedundancyType_value[object.RedundancyType]),
				SourceType:           types.SourceType(types.SourceType_value[object.SourceType]),
				Checksums:            util.StringListToBytesSlice(object.CheckSums),
				SecondarySpAddresses: object.SecondarySpAddresses,
			},
			LockedBalance: object.LockedBalance,
			Removed:       object.Removed,
		})
	}

	resp = &metatypes.MetadataServiceListObjectsByBucketNameResponse{Objects: res}
	log.CtxInfow(ctx, "success to  list objects by bucket name")
	return resp, nil
}

// ListDeletedObjectsByBlockNumberRange list deleted objects info by a block number range
func (metadata *Metadata) ListDeletedObjectsByBlockNumberRange(ctx context.Context, req *metatypes.MetadataServiceListDeletedObjectsByBlockNumberRangeRequest) (resp *metatypes.MetadataServiceListDeletedObjectsByBlockNumberRangeResponse, err error) {
	ctx = log.Context(ctx, req)

	endBlockNumber, err := metadata.spDB.GetLatestBlockNumber()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get the latest block number", "error", err)
		return nil, err
	}

	if endBlockNumber > req.EndBlockNumber {
		endBlockNumber = req.EndBlockNumber
	}

	objects, err := metadata.spDB.ListDeletedObjectsByBlockNumberRange(req.StartBlockNumber, endBlockNumber, req.IsFullList)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list deleted objects by block number range", "error", err)
		return nil, err
	}

	res := make([]*metatypes.Object, 0)
	for _, object := range objects {
		res = append(res, &metatypes.Object{
			ObjectInfo: &types.ObjectInfo{
				Owner:                object.Owner.String(),
				BucketName:           object.BucketName,
				ObjectName:           object.ObjectName,
				Id:                   math.NewUint(uint64(object.ObjectID)),
				PayloadSize:          object.PayloadSize,
				IsPublic:             object.IsPublic,
				ContentType:          object.ContentType,
				CreateAt:             object.CreateAt,
				ObjectStatus:         types.ObjectStatus(types.ObjectStatus_value[object.ObjectStatus]),
				RedundancyType:       types.RedundancyType(types.RedundancyType_value[object.RedundancyType]),
				SourceType:           types.SourceType(types.SourceType_value[object.SourceType]),
				Checksums:            util.StringListToBytesSlice(object.CheckSums),
				SecondarySpAddresses: object.SecondarySpAddresses,
			},
			LockedBalance: object.LockedBalance,
			Removed:       object.Removed,
		})
	}

	resp = &metatypes.MetadataServiceListDeletedObjectsByBlockNumberRangeResponse{
		Objects:           res,
		LatestBlockNumber: endBlockNumber,
	}
	log.CtxInfow(ctx, "success to list deleted objects by block number range")
	return resp, nil
}

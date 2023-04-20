package service

import (
	"context"

	model "github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	"github.com/bnb-chain/greenfield/types/s3util"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	metatypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
)

// ListObjectsByBucketName list objects info by a bucket name
func (metadata *Metadata) ListObjectsByBucketName(ctx context.Context, req *metatypes.ListObjectsByBucketNameRequest) (resp *metatypes.ListObjectsByBucketNameResponse, err error) {
	ctx = log.Context(ctx, req)

	objects, err := metadata.bsDB.ListObjectsByBucketName(req.BucketName)
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
				Id:                   math.NewUintFromBigInt(object.ObjectID.Big()),
				PayloadSize:          object.PayloadSize,
				ContentType:          object.ContentType,
				CreateAt:             object.CreateTime,
				ObjectStatus:         types.ObjectStatus(types.ObjectStatus_value[object.ObjectStatus]),
				RedundancyType:       types.RedundancyType(types.RedundancyType_value[object.RedundancyType]),
				SourceType:           types.SourceType(types.SourceType_value[object.SourceType]),
				Checksums:            object.Checksums,
				SecondarySpAddresses: object.SecondarySpAddresses,
				Visibility:           types.VisibilityType(types.VisibilityType_value[object.Visibility]),
			},
			LockedBalance: object.LockedBalance.String(),
			Removed:       object.Removed,
		})
	}

	resp = &metatypes.ListObjectsByBucketNameResponse{Objects: res}
	log.CtxInfow(ctx, "succeed to list objects by bucket name")
	return resp, nil
}

// ListDeletedObjectsByBlockNumberRange list deleted objects info by a block number range
func (metadata *Metadata) ListDeletedObjectsByBlockNumberRange(ctx context.Context, req *metatypes.ListDeletedObjectsByBlockNumberRangeRequest) (resp *metatypes.ListDeletedObjectsByBlockNumberRangeResponse, err error) {
	ctx = log.Context(ctx, req)

	endBlockNumber, err := metadata.bsDB.GetLatestBlockNumber()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get the latest block number", "error", err)
		return nil, err
	}

	if endBlockNumber > req.EndBlockNumber {
		endBlockNumber = req.EndBlockNumber
	}

	objects, err := metadata.bsDB.ListDeletedObjectsByBlockNumberRange(req.StartBlockNumber, endBlockNumber, req.IsFullList)
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
				Id:                   math.NewUintFromBigInt(object.ObjectID.Big()),
				PayloadSize:          object.PayloadSize,
				ContentType:          object.ContentType,
				CreateAt:             object.CreateTime,
				ObjectStatus:         types.ObjectStatus(types.ObjectStatus_value[object.ObjectStatus]),
				RedundancyType:       types.RedundancyType(types.RedundancyType_value[object.RedundancyType]),
				SourceType:           types.SourceType(types.SourceType_value[object.SourceType]),
				Checksums:            object.Checksums,
				SecondarySpAddresses: object.SecondarySpAddresses,
				Visibility:           types.VisibilityType(types.VisibilityType_value[object.Visibility]),
			},
			LockedBalance: object.LockedBalance.String(),
			Removed:       object.Removed,
		})
	}

	resp = &metatypes.ListDeletedObjectsByBlockNumberRangeResponse{
		Objects:        res,
		EndBlockNumber: endBlockNumber,
	}
	log.CtxInfow(ctx, "succeed to list deleted objects by block number range")
	return resp, nil
}

// GetObjectByObjectNameAndBucketName get object info by an object name
func (metadata *Metadata) GetObjectByObjectNameAndBucketName(ctx context.Context, req *metatypes.GetObjectByObjectNameAndBucketNameRequest) (resp *metatypes.GetObjectByObjectNameAndBucketNameResponse, err error) {
	var (
		object *model.Object
		res    *metatypes.Object
	)

	ctx = log.Context(ctx, req)
	if err = s3util.CheckValidObjectName(req.ObjectName); err != nil {
		log.Errorw("failed to check object name", "object_name", req.ObjectName, "error", err)
		return nil, err
	}

	object, err = metadata.bsDB.GetObjectByName(req.ObjectName, req.BucketName, req.IsFullList)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get object by object name", "error", err)
		return nil, err
	}

	if object != nil {
		res = &metatypes.Object{
			ObjectInfo: &types.ObjectInfo{
				Owner:                object.Owner.String(),
				BucketName:           object.BucketName,
				ObjectName:           object.ObjectName,
				Id:                   math.NewUintFromBigInt(object.ObjectID.Big()),
				PayloadSize:          object.PayloadSize,
				ContentType:          object.ContentType,
				CreateAt:             object.CreateTime,
				ObjectStatus:         types.ObjectStatus(types.ObjectStatus_value[object.ObjectStatus]),
				RedundancyType:       types.RedundancyType(types.RedundancyType_value[object.RedundancyType]),
				SourceType:           types.SourceType(types.SourceType_value[object.SourceType]),
				Checksums:            object.Checksums,
				SecondarySpAddresses: object.SecondarySpAddresses,
				Visibility:           types.VisibilityType(types.VisibilityType_value[object.Visibility]),
			},
			LockedBalance: object.LockedBalance.String(),
			Removed:       object.Removed,
		}
	}
	resp = &metatypes.GetObjectByObjectNameAndBucketNameResponse{Object: res}
	log.CtxInfo(ctx, "succeed to get object by object name and bucket name")
	return resp, nil
}

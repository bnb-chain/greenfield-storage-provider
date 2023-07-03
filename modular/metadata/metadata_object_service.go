package metadata

import (
	"context"
	"encoding/base64"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/types/s3util"
	storage_types "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/forbole/juno/v4/common"

	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	model "github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

// GfSpListObjectsByBucketName list objects info by a bucket name
func (r *MetadataModular) GfSpListObjectsByBucketName(ctx context.Context, req *types.GfSpListObjectsByBucketNameRequest) (resp *types.GfSpListObjectsByBucketNameResponse, err error) {
	var (
		results               []*model.ListObjectsResult
		keyCount              uint64
		isTruncated           bool
		nextContinuationToken string
		maxKeys               uint64
		commonPrefixes        []string
		res                   []*types.Object
	)

	maxKeys = req.MaxKeys
	// if the user does not provide any input parameters, default values will be used
	if req.MaxKeys == 0 {
		maxKeys = model.ListObjectsDefaultMaxKeys
	}

	// returns some or all (up to 1000) of the objects in a bucket with each request
	if req.MaxKeys > model.ListObjectsLimitSize {
		maxKeys = model.ListObjectsLimitSize
	}

	ctx = log.Context(ctx, req)
	results, err = r.baseApp.GfBsDB().ListObjectsByBucketName(req.BucketName, req.ContinuationToken, req.Prefix, req.Delimiter, int(maxKeys), req.IncludeRemoved)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list objects by bucket name", "error", err)
		return
	}

	keyCount = uint64(len(results))
	// if keyCount is equal to req.MaxKeys+1 which means that we additionally return NextContinuationToken, and it is not counted in the keyCount
	// isTruncated set to false if all the results were returned, set to true if more keys are available to return
	// remove the returned NextContinuationToken object and separately return its object ID to the user for the next API call
	if keyCount == maxKeys+1 {
		isTruncated = true
		keyCount -= 1
		nextContinuationToken = results[len(results)-1].PathName
		if req.Delimiter == "" {
			nextContinuationToken = results[len(results)-1].ObjectName
		}
		results = results[:len(results)-1]
	}

	for _, object := range results {
		if object.ResultType == "common_prefix" {
			commonPrefixes = append(commonPrefixes, object.PathName)
		} else {
			res = append(res, &types.Object{
				ObjectInfo: &storage_types.ObjectInfo{
					Owner:               object.Owner.String(),
					Creator:             object.Creator.String(),
					BucketName:          object.BucketName,
					ObjectName:          object.ObjectName,
					Id:                  math.NewUintFromBigInt(object.ObjectID.Big()),
					LocalVirtualGroupId: object.LocalVirtualGroupId,
					PayloadSize:         object.PayloadSize,
					Visibility:          storage_types.VisibilityType(storage_types.VisibilityType_value[object.Visibility]),
					ContentType:         object.ContentType,
					CreateAt:            object.CreateTime,
					ObjectStatus:        storage_types.ObjectStatus(storage_types.ObjectStatus_value[object.ObjectStatus]),
					RedundancyType:      storage_types.RedundancyType(storage_types.RedundancyType_value[object.RedundancyType]),
					SourceType:          storage_types.SourceType(storage_types.SourceType_value[object.SourceType]),
					Checksums:           object.Checksums,
				},
				LockedBalance: object.LockedBalance.String(),
				Removed:       object.Removed,
				UpdateAt:      object.UpdateAt,
				DeleteAt:      object.DeleteAt,
				DeleteReason:  object.DeleteReason,
				Operator:      object.Operator.String(),
				CreateTxHash:  object.CreateTxHash.String(),
				UpdateTxHash:  object.UpdateTxHash.String(),
				SealTxHash:    object.SealTxHash.String(),
			})
		}
	}

	resp = &types.GfSpListObjectsByBucketNameResponse{
		Objects:               res,
		KeyCount:              keyCount,
		MaxKeys:               maxKeys,
		IsTruncated:           isTruncated,
		NextContinuationToken: base64.StdEncoding.EncodeToString([]byte(nextContinuationToken)),
		Name:                  req.BucketName,
		Prefix:                req.Prefix,
		Delimiter:             req.Delimiter,
		CommonPrefixes:        commonPrefixes,
		ContinuationToken:     base64.StdEncoding.EncodeToString([]byte(req.ContinuationToken)),
	}
	log.CtxInfo(ctx, "succeed to list objects by bucket name")
	return resp, nil
}

// GfSpListDeletedObjectsByBlockNumberRange list deleted objects info by a block number range
func (r *MetadataModular) GfSpListDeletedObjectsByBlockNumberRange(ctx context.Context, req *types.GfSpListDeletedObjectsByBlockNumberRangeRequest) (resp *types.GfSpListDeletedObjectsByBlockNumberRangeResponse, err error) {
	ctx = log.Context(ctx, req)

	endBlockNumber, err := r.baseApp.GfBsDB().GetLatestBlockNumber()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get the latest block number", "error", err)
		return nil, err
	}

	if endBlockNumber > req.EndBlockNumber {
		endBlockNumber = req.EndBlockNumber
	}

	objects, err := r.baseApp.GfBsDB().ListDeletedObjectsByBlockNumberRange(req.StartBlockNumber, endBlockNumber, req.IncludePrivate)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list deleted objects by block number range", "error", err)
		return nil, err
	}

	res := make([]*types.Object, 0)
	for _, object := range objects {
		res = append(res, &types.Object{
			ObjectInfo: &storage_types.ObjectInfo{
				Owner:               object.Owner.String(),
				Creator:             object.Creator.String(),
				BucketName:          object.BucketName,
				ObjectName:          object.ObjectName,
				Id:                  math.NewUintFromBigInt(object.ObjectID.Big()),
				LocalVirtualGroupId: object.LocalVirtualGroupId,
				PayloadSize:         object.PayloadSize,
				Visibility:          storage_types.VisibilityType(storage_types.VisibilityType_value[object.Visibility]),
				ContentType:         object.ContentType,
				CreateAt:            object.CreateTime,
				ObjectStatus:        storage_types.ObjectStatus(storage_types.ObjectStatus_value[object.ObjectStatus]),
				RedundancyType:      storage_types.RedundancyType(storage_types.RedundancyType_value[object.RedundancyType]),
				SourceType:          storage_types.SourceType(storage_types.SourceType_value[object.SourceType]),
				Checksums:           object.Checksums,
			},
			LockedBalance: object.LockedBalance.String(),
			Removed:       object.Removed,
			UpdateAt:      object.UpdateAt,
			DeleteAt:      object.DeleteAt,
			DeleteReason:  object.DeleteReason,
			Operator:      object.Operator.String(),
			CreateTxHash:  object.CreateTxHash.String(),
			UpdateTxHash:  object.UpdateTxHash.String(),
			SealTxHash:    object.SealTxHash.String(),
		})
	}

	resp = &types.GfSpListDeletedObjectsByBlockNumberRangeResponse{
		Objects:        res,
		EndBlockNumber: endBlockNumber,
	}
	log.CtxInfow(ctx, "succeed to list deleted objects by block number range")
	return resp, nil
}

// GfSpGetObjectMeta get object metadata
func (r *MetadataModular) GfSpGetObjectMeta(ctx context.Context, req *types.GfSpGetObjectMetaRequest) (resp *types.GfSpGetObjectMetaResponse, err error) {
	var (
		object *model.Object
		res    *types.Object
	)

	ctx = log.Context(ctx, req)
	if err = s3util.CheckValidObjectName(req.ObjectName); err != nil {
		log.Errorw("failed to check object name", "object_name", req.ObjectName, "error", err)
		return nil, err
	}

	object, err = r.baseApp.GfBsDB().GetObjectByName(req.ObjectName, req.BucketName, req.IncludePrivate)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get object by object name", "error", err)
		return nil, err
	}

	if object != nil {
		res = &types.Object{
			ObjectInfo: &storage_types.ObjectInfo{
				Owner:               object.Owner.String(),
				Creator:             object.Creator.String(),
				BucketName:          object.BucketName,
				ObjectName:          object.ObjectName,
				Id:                  math.NewUintFromBigInt(object.ObjectID.Big()),
				LocalVirtualGroupId: object.LocalVirtualGroupId,
				PayloadSize:         object.PayloadSize,
				Visibility:          storage_types.VisibilityType(storage_types.VisibilityType_value[object.Visibility]),
				ContentType:         object.ContentType,
				CreateAt:            object.CreateTime,
				ObjectStatus:        storage_types.ObjectStatus(storage_types.ObjectStatus_value[object.ObjectStatus]),
				RedundancyType:      storage_types.RedundancyType(storage_types.RedundancyType_value[object.RedundancyType]),
				SourceType:          storage_types.SourceType(storage_types.SourceType_value[object.SourceType]),
				Checksums:           object.Checksums,
			},
			LockedBalance: object.LockedBalance.String(),
			Removed:       object.Removed,
			DeleteAt:      object.DeleteAt,
			DeleteReason:  object.DeleteReason,
			Operator:      object.Operator.String(),
			CreateTxHash:  object.CreateTxHash.String(),
			UpdateTxHash:  object.UpdateTxHash.String(),
			SealTxHash:    object.SealTxHash.String(),
		}
	}
	resp = &types.GfSpGetObjectMetaResponse{Object: res}
	log.CtxInfo(ctx, "succeed to get object meta")
	return resp, nil
}

// GfSpListObjectsByObjectID list objects by object ids
func (r *MetadataModular) GfSpListObjectsByObjectID(ctx context.Context, req *types.GfSpListObjectsByObjectIDRequest) (resp *types.GfSpListObjectsByObjectIDResponse, err error) {
	var (
		objects    []*model.Object
		ids        []common.Hash
		objectsMap map[uint64]*types.Object
	)

	ids = make([]common.Hash, len(req.ObjectIds))
	for i, id := range req.ObjectIds {
		ids[i] = common.BigToHash(math.NewUint(id).BigInt())
	}

	objects, err = r.baseApp.GfBsDB().ListObjectsByObjectID(ids, req.IncludeRemoved)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list objects by object ids", "error", err)
		return nil, err
	}

	objectsMap = make(map[uint64]*types.Object)
	for _, id := range req.ObjectIds {
		objectsMap[id] = nil
	}

	for _, object := range objects {
		objectsMap[object.ObjectID.Big().Uint64()] = &types.Object{
			ObjectInfo: &storage_types.ObjectInfo{
				Owner:               object.Owner.String(),
				Creator:             object.Creator.String(),
				BucketName:          object.BucketName,
				ObjectName:          object.ObjectName,
				Id:                  math.NewUintFromBigInt(object.ObjectID.Big()),
				LocalVirtualGroupId: object.LocalVirtualGroupId,
				PayloadSize:         object.PayloadSize,
				Visibility:          storage_types.VisibilityType(storage_types.VisibilityType_value[object.Visibility]),
				ContentType:         object.ContentType,
				CreateAt:            object.CreateTime,
				ObjectStatus:        storage_types.ObjectStatus(storage_types.ObjectStatus_value[object.ObjectStatus]),
				RedundancyType:      storage_types.RedundancyType(storage_types.RedundancyType_value[object.RedundancyType]),
				SourceType:          storage_types.SourceType(storage_types.SourceType_value[object.SourceType]),
				Checksums:           object.Checksums,
			},
			LockedBalance: object.LockedBalance.String(),
			Removed:       object.Removed,
			DeleteAt:      object.DeleteAt,
			DeleteReason:  object.DeleteReason,
			Operator:      object.Operator.String(),
			CreateTxHash:  object.CreateTxHash.String(),
			UpdateTxHash:  object.UpdateTxHash.String(),
			SealTxHash:    object.SealTxHash.String(),
		}
	}
	resp = &types.GfSpListObjectsByObjectIDResponse{Objects: objectsMap}
	log.CtxInfo(ctx, "succeed to list objects by object ids")
	return resp, nil
}

// GfSpListPrimaryObjects list objects by primary sp id
func (r *MetadataModular) GfSpListPrimaryObjects(ctx context.Context, req *types.GfSpListPrimaryObjectsRequest) (resp *types.GfSpListPrimaryObjectsResponse, err error) {
	var (
		objects []*model.Object
		res     []*types.Object
		limit   int
	)

	ctx = log.Context(ctx, req)
	if req.Limit == 0 {
		limit = model.ListObjectsDefaultMaxKeys
	}

	if req.Limit > model.ListObjectsLimitSize {
		limit = model.ListObjectsLimitSize
	}

	objects, err = r.baseApp.GfBsDB().ListPrimaryObjects(req.SpId, common.BigToHash(math.NewUint(req.BucketId).BigInt()), common.BigToHash(math.NewUint(req.StartAfter).BigInt()), limit)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list objects by primary sp id", "error", err)
		return nil, err
	}

	res = make([]*types.Object, 0)
	for i, object := range objects {
		res[i] = &types.Object{
			ObjectInfo: &storage_types.ObjectInfo{
				Owner:               object.Owner.String(),
				Creator:             object.Creator.String(),
				BucketName:          object.BucketName,
				ObjectName:          object.ObjectName,
				Id:                  math.NewUintFromBigInt(object.ObjectID.Big()),
				LocalVirtualGroupId: object.LocalVirtualGroupId,
				PayloadSize:         object.PayloadSize,
				Visibility:          storage_types.VisibilityType(storage_types.VisibilityType_value[object.Visibility]),
				ContentType:         object.ContentType,
				CreateAt:            object.CreateTime,
				ObjectStatus:        storage_types.ObjectStatus(storage_types.ObjectStatus_value[object.ObjectStatus]),
				RedundancyType:      storage_types.RedundancyType(storage_types.RedundancyType_value[object.RedundancyType]),
				SourceType:          storage_types.SourceType(storage_types.SourceType_value[object.SourceType]),
				Checksums:           object.Checksums,
			},
			LockedBalance: object.LockedBalance.String(),
			Removed:       object.Removed,
			UpdateAt:      object.UpdateAt,
			DeleteAt:      object.DeleteAt,
			DeleteReason:  object.DeleteReason,
			Operator:      object.Operator.String(),
			CreateTxHash:  object.CreateTxHash.String(),
			UpdateTxHash:  object.UpdateTxHash.String(),
			SealTxHash:    object.SealTxHash.String(),
		}
	}

	resp = &types.GfSpListPrimaryObjectsResponse{Objects: res}
	log.CtxInfow(ctx, "succeed to list objects by primary sp id")
	return resp, nil
}

// GfSpListSecondaryObjects list objects by secondary sp id
func (r *MetadataModular) GfSpListSecondaryObjects(ctx context.Context, req *types.GfSpListSecondaryObjectsRequest) (resp *types.GfSpListSecondaryObjectsResponse, err error) {
	var (
		objects []*model.Object
		res     []*types.Object
		limit   int
	)

	ctx = log.Context(ctx, req)
	if req.Limit == 0 {
		limit = model.ListObjectsDefaultMaxKeys
	}

	if req.Limit > model.ListObjectsLimitSize {
		limit = model.ListObjectsLimitSize
	}
	objects, err = r.baseApp.GfBsDB().ListSecondaryObjects(req.SpId, common.BigToHash(math.NewUint(req.BucketId).BigInt()), common.BigToHash(math.NewUint(req.StartAfter).BigInt()), limit)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list objects by secondary sp id", "error", err)
		return nil, err
	}

	res = make([]*types.Object, 0)
	for i, object := range objects {
		res[i] = &types.Object{
			ObjectInfo: &storage_types.ObjectInfo{
				Owner:               object.Owner.String(),
				Creator:             object.Creator.String(),
				BucketName:          object.BucketName,
				ObjectName:          object.ObjectName,
				Id:                  math.NewUintFromBigInt(object.ObjectID.Big()),
				LocalVirtualGroupId: object.LocalVirtualGroupId,
				PayloadSize:         object.PayloadSize,
				Visibility:          storage_types.VisibilityType(storage_types.VisibilityType_value[object.Visibility]),
				ContentType:         object.ContentType,
				CreateAt:            object.CreateTime,
				ObjectStatus:        storage_types.ObjectStatus(storage_types.ObjectStatus_value[object.ObjectStatus]),
				RedundancyType:      storage_types.RedundancyType(storage_types.RedundancyType_value[object.RedundancyType]),
				SourceType:          storage_types.SourceType(storage_types.SourceType_value[object.SourceType]),
				Checksums:           object.Checksums,
			},
			LockedBalance: object.LockedBalance.String(),
			Removed:       object.Removed,
			UpdateAt:      object.UpdateAt,
			DeleteAt:      object.DeleteAt,
			DeleteReason:  object.DeleteReason,
			Operator:      object.Operator.String(),
			CreateTxHash:  object.CreateTxHash.String(),
			UpdateTxHash:  object.UpdateTxHash.String(),
			SealTxHash:    object.SealTxHash.String(),
		}
	}

	resp = &types.GfSpListSecondaryObjectsResponse{Objects: res}
	log.CtxInfow(ctx, "succeed to list objects by secondary sp id")
	return resp, nil
}

// GfSpListObjectsInGVG list objects by gvg and bucket id
func (r *MetadataModular) GfSpListObjectsInGVG(ctx context.Context, req *types.GfSpListObjectsInGVGRequest) (resp *types.GfSpListObjectsInGVGResponse, err error) {
	var (
		objects []*model.Object
		res     []*types.Object
		limit   int
	)

	ctx = log.Context(ctx, req)
	if req.Limit == 0 {
		limit = model.ListObjectsDefaultMaxKeys
	}

	if req.Limit > model.ListObjectsLimitSize {
		limit = model.ListObjectsLimitSize
	}
	objects, err = r.baseApp.GfBsDB().ListObjectsInGVG(common.BigToHash(math.NewUint(req.BucketId).BigInt()), req.GvgId, common.BigToHash(math.NewUint(req.StartAfter).BigInt()), limit)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list objects by gvg id", "error", err)
		return nil, err
	}

	res = make([]*types.Object, 0)
	for i, object := range objects {
		res[i] = &types.Object{
			ObjectInfo: &storage_types.ObjectInfo{
				Owner:               object.Owner.String(),
				Creator:             object.Creator.String(),
				BucketName:          object.BucketName,
				ObjectName:          object.ObjectName,
				Id:                  math.NewUintFromBigInt(object.ObjectID.Big()),
				LocalVirtualGroupId: object.LocalVirtualGroupId,
				PayloadSize:         object.PayloadSize,
				Visibility:          storage_types.VisibilityType(storage_types.VisibilityType_value[object.Visibility]),
				ContentType:         object.ContentType,
				CreateAt:            object.CreateTime,
				ObjectStatus:        storage_types.ObjectStatus(storage_types.ObjectStatus_value[object.ObjectStatus]),
				RedundancyType:      storage_types.RedundancyType(storage_types.RedundancyType_value[object.RedundancyType]),
				SourceType:          storage_types.SourceType(storage_types.SourceType_value[object.SourceType]),
				Checksums:           object.Checksums,
			},
			LockedBalance: object.LockedBalance.String(),
			Removed:       object.Removed,
			UpdateAt:      object.UpdateAt,
			DeleteAt:      object.DeleteAt,
			DeleteReason:  object.DeleteReason,
			Operator:      object.Operator.String(),
			CreateTxHash:  object.CreateTxHash.String(),
			UpdateTxHash:  object.UpdateTxHash.String(),
			SealTxHash:    object.SealTxHash.String(),
		}
	}

	resp = &types.GfSpListObjectsInGVGResponse{Objects: res}
	log.CtxInfow(ctx, "succeed to list objects by gvg id")
	return resp, nil
}

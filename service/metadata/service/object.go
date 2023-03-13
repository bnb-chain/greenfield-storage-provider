package service

import (
	"context"

	"cosmossdk.io/math"

	model "github.com/bnb-chain/greenfield-storage-provider/model/metadata"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	metatypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

// ListObjectsByBucketName list objects info by a bucket name
func (metadata *Metadata) ListObjectsByBucketName(ctx context.Context, req *metatypes.MetadataServiceListObjectsByBucketNameRequest) (resp *metatypes.MetadataServiceListObjectsByBucketNameResponse, err error) {
	ctx = log.Context(ctx, req)
	defer func() {
		if err != nil {
			log.CtxErrorw(ctx, "failed to list objects by bucket name", "err", err)
		} else {
			log.CtxInfow(ctx, "succeed to list objects by bucket name")
		}
	}()
	var objects []*model.Object
	//TODO:: cancel mock after impl db
	object1 := &model.Object{
		Owner:                "46765cbc-d30c-4f4a-a814-b68181fcab12",
		BucketName:           req.BucketName,
		ObjectName:           "test-object",
		ID:                   "1000",
		PayloadSize:          100,
		IsPublic:             false,
		ContentType:          "video",
		CreateAt:             1677143663461,
		ObjectStatus:         1,
		RedundancyType:       1,
		SourceType:           1,
		SecondarySpAddresses: nil,
		LockedBalance:        "1000",
	}
	object2 := &model.Object{
		Owner:                "0xdc4f0dba80cc3ee55aa1ad222a350c85a84261bd",
		BucketName:           req.BucketName,
		ObjectName:           "ETH",
		ID:                   "1001",
		PayloadSize:          500,
		IsPublic:             true,
		ContentType:          "image",
		CreateAt:             1677143880209,
		ObjectStatus:         2,
		RedundancyType:       2,
		SourceType:           2,
		SecondarySpAddresses: nil,
		LockedBalance:        "1000",
	}
	objects = append(objects, object1, object2)
	res := make([]*metatypes.Object, 0)

	for _, object := range objects {
		res = append(res, &metatypes.Object{
			ObjectInfo: &types.ObjectInfo{
				Owner:                object.Owner,
				BucketName:           object.BucketName,
				ObjectName:           object.ObjectName,
				Id:                   math.NewUintFromString(object.ID),
				PayloadSize:          object.PayloadSize,
				IsPublic:             object.IsPublic,
				ContentType:          object.ContentType,
				CreateAt:             object.CreateAt,
				ObjectStatus:         types.ObjectStatus(object.ObjectStatus),
				RedundancyType:       types.RedundancyType(object.RedundancyType),
				SourceType:           types.SourceType(object.SourceType),
				Checksums:            nil,
				SecondarySpAddresses: object.SecondarySpAddresses,
			},
			LockedBalance: object.LockedBalance,
		})
	}
	resp = &metatypes.MetadataServiceListObjectsByBucketNameResponse{Objects: res}
	return resp, nil
}

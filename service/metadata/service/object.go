package service

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
)

func (metadata *Metadata) ListObjectsByBucketName(ctx context.Context, req *stypes.MetadataServiceListObjectsByBucketNameRequest) (resp *stypes.MetadataServiceListObjectsByBucketNameResponse, err error) {
	ctx = log.Context(ctx, req)
	defer func() {
		if err != nil {
			log.CtxErrorw(ctx, "failed to list objects by name", "err", err)
		} else {
			log.CtxInfow(ctx, "succeed to list objects by name")
		}
	}()
	//objects, err := metadata.store.ListObjectsByBucketName(ctx, req.BucketName)
	var objects []*model.Object
	object1 := &model.Object{
		Owner:                "46765cbc-d30c-4f4a-a814-b68181fcab12",
		BucketName:           req.BucketName,
		ObjectName:           "test-object",
		Id:                   "1000",
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
		Id:                   "1001",
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
	res := make([]*stypes.Object, 0)
	for _, object := range objects {
		res = append(res, &stypes.Object{
			Owner:                object.Owner,
			BucketName:           object.BucketName,
			ObjectName:           object.ObjectName,
			Id:                   object.Id,
			PayloadSize:          object.PayloadSize,
			IsPublic:             object.IsPublic,
			ContentType:          object.ContentType,
			CreateAt:             object.CreateAt,
			ObjectStatus:         stypes.ObjectStatus(object.ObjectStatus),
			RedundancyType:       stypes.RedundancyType(object.RedundancyType),
			SourceType:           stypes.SourceType(object.SourceType),
			Checksums:            nil,
			SecondarySpAddresses: object.SecondarySpAddresses,
			LockedBalance:        object.LockedBalance,
		})
	}
	resp = &stypes.MetadataServiceListObjectsByBucketNameResponse{Objects: res}
	return resp, nil
}

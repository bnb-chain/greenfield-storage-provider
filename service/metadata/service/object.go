package service

import (
	"context"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

func (metadata *Metadata) ListObjectsByBucketName(ctx context.Context, req *stypes.MetadataServiceBucketNameRequest) (resp *stypes.MetadataServiceListObjectsByBucketNameResponse, err error) {
	ctx = log.Context(ctx, req)
	defer func() {
		if err != nil {
			resp.ErrMessage = merrors.MakeErrMsgResponse(err)
			log.CtxErrorw(ctx, "failed to get user buckets", "err", err)
		} else {
			log.CtxInfow(ctx, "succeed to get user buckets")
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
		CreateAt:             0,
		ObjectStatus:         1,
		RedundancyType:       1,
		SourceType:           1,
		SecondarySpAddresses: nil,
		LockedBalance:        "1000",
	}
	objects = append(objects, object1)
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

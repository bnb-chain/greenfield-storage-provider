package metadata

import (
	"context"
	"testing"

	"github.com/forbole/juno/v4/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

func TestMetadataModular_GfSpListObjectsByBucketName_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsByBucketName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, string, int, bool) ([]*bsdb.ListObjectsResult, error) {
			return []*bsdb.ListObjectsResult{&bsdb.ListObjectsResult{
				PathName:   "/folder1",
				ResultType: "Object",
				Object: &bsdb.Object{
					ID:                  1,
					Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
					Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
					Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
					LocalVirtualGroupId: 0,
					BucketName:          "barry",
					ObjectName:          "",
					ObjectID:            common.HexToHash("1"),
					BucketID:            common.HexToHash("1"),
					PayloadSize:         0,
					Visibility:          "",
					ContentType:         "",
					CreateAt:            0,
					CreateTime:          0,
					ObjectStatus:        "",
					RedundancyType:      "",
					SourceType:          "",
					Checksums:           nil,
					LockedBalance:       common.HexToHash("1"),
					Removed:             false,
					UpdateTime:          0,
					UpdateAt:            0,
					DeleteAt:            0,
					DeleteReason:        "",
					CreateTxHash:        common.HexToHash("1"),
					UpdateTxHash:        common.HexToHash("1"),
					SealTxHash:          common.HexToHash("1"),
				},
			}}, nil
		},
	).Times(1)
	objects, err := a.GfSpListObjectsByBucketName(context.Background(), &types.GfSpListObjectsByBucketNameRequest{
		BucketName:        "barry",
		AccountId:         "0xe978A9160BC061f602fa083e9C68539C549A421D",
		MaxKeys:           0,
		StartAfter:        "test",
		ContinuationToken: "",
		Delimiter:         "/",
		Prefix:            "/folder1",
		IncludeRemoved:    false,
	})
	assert.Nil(t, err)
	assert.Equal(t, "barry", objects.Objects[0].ObjectInfo.BucketName)
}

func TestMetadataModular_GfSpListObjectsByBucketName_Success2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsByBucketName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, string, int, bool) ([]*bsdb.ListObjectsResult, error) {
			return []*bsdb.ListObjectsResult{&bsdb.ListObjectsResult{
				PathName:   "/folder1",
				ResultType: "Object",
				Object: &bsdb.Object{
					ID:                  1,
					Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
					Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
					Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
					LocalVirtualGroupId: 0,
					BucketName:          "barry",
					ObjectName:          "",
					ObjectID:            common.HexToHash("1"),
					BucketID:            common.HexToHash("1"),
					PayloadSize:         0,
					Visibility:          "",
					ContentType:         "",
					CreateAt:            0,
					CreateTime:          0,
					ObjectStatus:        "",
					RedundancyType:      "",
					SourceType:          "",
					Checksums:           nil,
					LockedBalance:       common.HexToHash("1"),
					Removed:             false,
					UpdateTime:          0,
					UpdateAt:            0,
					DeleteAt:            0,
					DeleteReason:        "",
					CreateTxHash:        common.HexToHash("1"),
					UpdateTxHash:        common.HexToHash("1"),
					SealTxHash:          common.HexToHash("1"),
				},
			}}, nil
		},
	).Times(1)
	objects, err := a.GfSpListObjectsByBucketName(context.Background(), &types.GfSpListObjectsByBucketNameRequest{
		BucketName:        "barry",
		AccountId:         "0xe978A9160BC061f602fa083e9C68539C549A421D",
		MaxKeys:           11110,
		StartAfter:        "test",
		ContinuationToken: "",
		Delimiter:         "/",
		Prefix:            "/folder1",
		IncludeRemoved:    false,
	})
	assert.Nil(t, err)
	assert.Equal(t, "barry", objects.Objects[0].ObjectInfo.BucketName)
}

func TestMetadataModular_GfSpListObjectsByBucketName_Success3(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsByBucketName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, string, int, bool) ([]*bsdb.ListObjectsResult, error) {
			return []*bsdb.ListObjectsResult{
				&bsdb.ListObjectsResult{
					PathName:   "/folder1",
					ResultType: "object",
					Object: &bsdb.Object{
						ID:                  1,
						Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						LocalVirtualGroupId: 0,
						BucketName:          "barry",
						ObjectName:          "1111",
						ObjectID:            common.HexToHash("1"),
						BucketID:            common.HexToHash("1"),
						PayloadSize:         0,
						Visibility:          "",
						ContentType:         "",
						CreateAt:            0,
						CreateTime:          0,
						ObjectStatus:        "",
						RedundancyType:      "",
						SourceType:          "",
						Checksums:           nil,
						LockedBalance:       common.HexToHash("1"),
						Removed:             false,
						UpdateTime:          0,
						UpdateAt:            0,
						DeleteAt:            0,
						DeleteReason:        "",
						CreateTxHash:        common.HexToHash("1"),
						UpdateTxHash:        common.HexToHash("1"),
						SealTxHash:          common.HexToHash("1"),
					}},
				&bsdb.ListObjectsResult{
					PathName:   "/folder1",
					ResultType: "common_prefix",
					Object: &bsdb.Object{
						ID:                  1,
						Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						LocalVirtualGroupId: 0,
						BucketName:          "barry",
						ObjectName:          "2222",
						ObjectID:            common.HexToHash("2"),
						BucketID:            common.HexToHash("2"),
						PayloadSize:         0,
						Visibility:          "",
						ContentType:         "",
						CreateAt:            0,
						CreateTime:          0,
						ObjectStatus:        "",
						RedundancyType:      "",
						SourceType:          "",
						Checksums:           nil,
						LockedBalance:       common.HexToHash("2"),
						Removed:             false,
						UpdateTime:          0,
						UpdateAt:            0,
						DeleteAt:            0,
						DeleteReason:        "",
						CreateTxHash:        common.HexToHash("2"),
						UpdateTxHash:        common.HexToHash("2"),
						SealTxHash:          common.HexToHash("2"),
					}},
				&bsdb.ListObjectsResult{
					PathName:   "/folder1",
					ResultType: "common_prefix",
					Object: &bsdb.Object{
						ID:                  1,
						Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						LocalVirtualGroupId: 0,
						BucketName:          "barry",
						ObjectName:          "3333",
						ObjectID:            common.HexToHash("3"),
						BucketID:            common.HexToHash("3"),
						PayloadSize:         0,
						Visibility:          "",
						ContentType:         "",
						CreateAt:            0,
						CreateTime:          0,
						ObjectStatus:        "",
						RedundancyType:      "",
						SourceType:          "",
						Checksums:           nil,
						LockedBalance:       common.HexToHash("3"),
						Removed:             false,
						UpdateTime:          0,
						UpdateAt:            0,
						DeleteAt:            0,
						DeleteReason:        "",
						CreateTxHash:        common.HexToHash("3"),
						UpdateTxHash:        common.HexToHash("3"),
						SealTxHash:          common.HexToHash("3"),
					}},
			}, nil
		},
	).Times(1)
	objects, err := a.GfSpListObjectsByBucketName(context.Background(), &types.GfSpListObjectsByBucketNameRequest{
		BucketName:        "barry",
		AccountId:         "0xe978A9160BC061f602fa083e9C68539C549A421D",
		MaxKeys:           1,
		StartAfter:        "test",
		ContinuationToken: "",
		Delimiter:         "",
		Prefix:            "/folder1",
		IncludeRemoved:    false,
	})
	assert.Nil(t, err)
	assert.Equal(t, "barry", objects.Objects[0].ObjectInfo.BucketName)
}

func TestMetadataModular_GfSpListObjectsByBucketName_Success4(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsByBucketName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, string, int, bool) ([]*bsdb.ListObjectsResult, error) {
			return []*bsdb.ListObjectsResult{
				&bsdb.ListObjectsResult{
					PathName:   "/folder1",
					ResultType: "object",
					Object: &bsdb.Object{
						ID:                  1,
						Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						LocalVirtualGroupId: 0,
						BucketName:          "barry",
						ObjectName:          "1111",
						ObjectID:            common.HexToHash("1"),
						BucketID:            common.HexToHash("1"),
						PayloadSize:         0,
						Visibility:          "",
						ContentType:         "",
						CreateAt:            0,
						CreateTime:          0,
						ObjectStatus:        "",
						RedundancyType:      "",
						SourceType:          "",
						Checksums:           nil,
						LockedBalance:       common.HexToHash("1"),
						Removed:             false,
						UpdateTime:          0,
						UpdateAt:            0,
						DeleteAt:            0,
						DeleteReason:        "",
						CreateTxHash:        common.HexToHash("1"),
						UpdateTxHash:        common.HexToHash("1"),
						SealTxHash:          common.HexToHash("1"),
					}},
				&bsdb.ListObjectsResult{
					PathName:   "/folder1",
					ResultType: "common_prefix",
					Object: &bsdb.Object{
						ID:                  1,
						Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						LocalVirtualGroupId: 0,
						BucketName:          "barry",
						ObjectName:          "2222",
						ObjectID:            common.HexToHash("2"),
						BucketID:            common.HexToHash("2"),
						PayloadSize:         0,
						Visibility:          "",
						ContentType:         "",
						CreateAt:            0,
						CreateTime:          0,
						ObjectStatus:        "",
						RedundancyType:      "",
						SourceType:          "",
						Checksums:           nil,
						LockedBalance:       common.HexToHash("2"),
						Removed:             false,
						UpdateTime:          0,
						UpdateAt:            0,
						DeleteAt:            0,
						DeleteReason:        "",
						CreateTxHash:        common.HexToHash("2"),
						UpdateTxHash:        common.HexToHash("2"),
						SealTxHash:          common.HexToHash("2"),
					}},
			}, nil
		},
	).Times(1)
	objects, err := a.GfSpListObjectsByBucketName(context.Background(), &types.GfSpListObjectsByBucketNameRequest{
		BucketName:        "barry",
		AccountId:         "0xe978A9160BC061f602fa083e9C68539C549A421D",
		MaxKeys:           1,
		StartAfter:        "test",
		ContinuationToken: "",
		Delimiter:         "",
		Prefix:            "/folder1",
		IncludeRemoved:    false,
	})
	assert.Nil(t, err)
	assert.Equal(t, "barry", objects.Objects[0].ObjectInfo.BucketName)
}

func TestMetadataModular_GfSpListObjectsByBucketName_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsByBucketName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, string, int, bool) ([]*bsdb.ListObjectsResult, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListObjectsByBucketName(context.Background(), &types.GfSpListObjectsByBucketNameRequest{
		BucketName:        "barry",
		AccountId:         "0xe978A9160BC061f602fa083e9C68539C549A421D",
		MaxKeys:           11110,
		StartAfter:        "test",
		ContinuationToken: "",
		Delimiter:         "/",
		Prefix:            "/folder1",
		IncludeRemoved:    false,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpListDeletedObjectsByBlockNumberRange_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 10, nil
		},
	).Times(1)
	m.EXPECT().ListDeletedObjectsByBlockNumberRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(int64, int64, bool) ([]*bsdb.Object, error) {
			return []*bsdb.Object{
				&bsdb.Object{
					ID:                  1,
					Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
					Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
					Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
					LocalVirtualGroupId: 0,
					BucketName:          "barry",
					ObjectName:          "2222",
					ObjectID:            common.HexToHash("2"),
					BucketID:            common.HexToHash("2"),
					PayloadSize:         0,
					Visibility:          "",
					ContentType:         "",
					CreateAt:            0,
					CreateTime:          0,
					ObjectStatus:        "",
					RedundancyType:      "",
					SourceType:          "",
					Checksums:           nil,
					LockedBalance:       common.HexToHash("2"),
					Removed:             false,
					UpdateTime:          0,
					UpdateAt:            0,
					DeleteAt:            0,
					DeleteReason:        "",
					CreateTxHash:        common.HexToHash("2"),
					UpdateTxHash:        common.HexToHash("2"),
					SealTxHash:          common.HexToHash("2"),
				},
			}, nil
		},
	).Times(1)
	objects, err := a.GfSpListDeletedObjectsByBlockNumberRange(context.Background(), &types.GfSpListDeletedObjectsByBlockNumberRangeRequest{
		StartBlockNumber: 0,
		EndBlockNumber:   8,
		IncludePrivate:   false,
	})
	assert.Nil(t, err)
	assert.Equal(t, "barry", objects.Objects[0].ObjectInfo.BucketName)
}

func TestMetadataModular_GfSpListDeletedObjectsByBlockNumberRange_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 0, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListDeletedObjectsByBlockNumberRange(context.Background(), &types.GfSpListDeletedObjectsByBlockNumberRangeRequest{
		StartBlockNumber: 0,
		EndBlockNumber:   8,
		IncludePrivate:   false,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpListDeletedObjectsByBlockNumberRange_Failed2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 10, nil
		},
	).Times(1)
	m.EXPECT().ListDeletedObjectsByBlockNumberRange(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(int64, int64, bool) ([]*bsdb.Object, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListDeletedObjectsByBlockNumberRange(context.Background(), &types.GfSpListDeletedObjectsByBlockNumberRangeRequest{
		StartBlockNumber: 0,
		EndBlockNumber:   8,
		IncludePrivate:   false,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpGetObjectMeta_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetObjectByName(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, bool) (*bsdb.Object, error) {
			return &bsdb.Object{
				ID:                  1,
				Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				LocalVirtualGroupId: 0,
				BucketName:          "barry",
				ObjectName:          "2222",
				ObjectID:            common.HexToHash("2"),
				BucketID:            common.HexToHash("2"),
				PayloadSize:         0,
				Visibility:          "",
				ContentType:         "",
				CreateAt:            0,
				CreateTime:          0,
				ObjectStatus:        "",
				RedundancyType:      "",
				SourceType:          "",
				Checksums:           nil,
				LockedBalance:       common.HexToHash("2"),
				Removed:             false,
				UpdateTime:          0,
				UpdateAt:            0,
				DeleteAt:            0,
				DeleteReason:        "",
				CreateTxHash:        common.HexToHash("2"),
				UpdateTxHash:        common.HexToHash("2"),
				SealTxHash:          common.HexToHash("2"),
			}, nil
		},
	).Times(1)
	object, err := a.GfSpGetObjectMeta(context.Background(), &types.GfSpGetObjectMetaRequest{
		ObjectName:     "2222",
		BucketName:     "barry",
		IncludePrivate: false,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2222", object.Object.ObjectInfo.ObjectName)
}

func TestMetadataModular_GfSpGetObjectMeta_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetObjectByName(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, bool) (*bsdb.Object, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpGetObjectMeta(context.Background(), &types.GfSpGetObjectMetaRequest{
		ObjectName:     "2",
		BucketName:     "barry",
		IncludePrivate: false,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpGetObjectMeta_Failed2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	_, err := a.GfSpGetObjectMeta(context.Background(), &types.GfSpGetObjectMetaRequest{
		ObjectName:     "",
		BucketName:     "barry",
		IncludePrivate: false,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpListObjectsByIDs_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsByIDs(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Object, error) {
			return []*bsdb.Object{
				&bsdb.Object{
					ID:                  1,
					Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
					Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
					Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
					LocalVirtualGroupId: 0,
					BucketName:          "barry",
					ObjectName:          "2222",
					ObjectID:            common.HexToHash("2"),
					BucketID:            common.HexToHash("2"),
					PayloadSize:         0,
					Visibility:          "",
					ContentType:         "",
					CreateAt:            0,
					CreateTime:          0,
					ObjectStatus:        "",
					RedundancyType:      "",
					SourceType:          "",
					Checksums:           nil,
					LockedBalance:       common.HexToHash("2"),
					Removed:             false,
					UpdateTime:          0,
					UpdateAt:            0,
					DeleteAt:            0,
					DeleteReason:        "",
					CreateTxHash:        common.HexToHash("2"),
					UpdateTxHash:        common.HexToHash("2"),
					SealTxHash:          common.HexToHash("2"),
				},
			}, nil
		},
	).Times(1)
	objects, err := a.GfSpListObjectsByIDs(context.Background(), &types.GfSpListObjectsByIDsRequest{
		ObjectIds:      []uint64{1},
		IncludeRemoved: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2222", objects.Objects[2].ObjectInfo.ObjectName)
}

func TestMetadataModular_GfSpListObjectsByIDs_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsByIDs(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Object, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListObjectsByIDs(context.Background(), &types.GfSpListObjectsByIDsRequest{
		ObjectIds:      []uint64{1},
		IncludeRemoved: true,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpListObjectsInGVGAndBucket_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsInGVGAndBucket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32, common.Hash, int) ([]*bsdb.Object, *bsdb.Bucket, error) {
			return []*bsdb.Object{
					&bsdb.Object{
						ID:                  1,
						Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						LocalVirtualGroupId: 0,
						BucketName:          "barry",
						ObjectName:          "2222",
						ObjectID:            common.HexToHash("2"),
						BucketID:            common.HexToHash("2"),
						PayloadSize:         0,
						Visibility:          "",
						ContentType:         "",
						CreateAt:            0,
						CreateTime:          0,
						ObjectStatus:        "",
						RedundancyType:      "",
						SourceType:          "",
						Checksums:           nil,
						LockedBalance:       common.HexToHash("2"),
						Removed:             false,
						UpdateTime:          0,
						UpdateAt:            0,
						DeleteAt:            0,
						DeleteReason:        "",
						CreateTxHash:        common.HexToHash("2"),
						UpdateTxHash:        common.HexToHash("2"),
						SealTxHash:          common.HexToHash("2"),
					},
				},
				&bsdb.Bucket{
					ID:                         848,
					Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					BucketName:                 "barry",
					Visibility:                 "VISIBILITY_TYPE_PRIVATE",
					BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					SourceType:                 "SOURCE_TYPE_ORIGIN",
					CreateAt:                   0,
					CreateTime:                 0,
					CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
					PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					GlobalVirtualGroupFamilyID: 1,
					ChargedReadQuota:           0,
					PaymentPriceTime:           0,
					Removed:                    false,
					Status:                     "",
					DeleteAt:                   0,
					DeleteReason:               "",
					UpdateAt:                   0,
					UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
					UpdateTime:                 0,
				}, nil
		},
	).Times(1)
	m.EXPECT().GetGvgByBucketAndLvgID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32) (*bsdb.GlobalVirtualGroup, error) {
			return &bsdb.GlobalVirtualGroup{
				ID:                    1,
				GlobalVirtualGroupId:  1,
				FamilyId:              1,
				PrimarySpId:           1,
				SecondarySpIds:        []uint32{1},
				StoredSize:            0,
				VirtualPaymentAddress: common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				TotalDeposit:          nil,
				CreateAt:              0,
				CreateTxHash:          common.HexToHash("1"),
				CreateTime:            0,
				UpdateAt:              0,
				UpdateTxHash:          common.HexToHash("1"),
				UpdateTime:            0,
				Removed:               false,
			}, nil
		},
	).Times(1)
	objects, err := a.GfSpListObjectsInGVGAndBucket(context.Background(), &types.GfSpListObjectsInGVGAndBucketRequest{
		GvgId:      1,
		BucketId:   1,
		StartAfter: 0,
		Limit:      0,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2222", objects.Objects[0].Object.ObjectInfo.ObjectName)
}

func TestMetadataModular_GfSpListObjectsInGVGAndBucket_Success2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsInGVGAndBucket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32, common.Hash, int) ([]*bsdb.Object, *bsdb.Bucket, error) {
			return []*bsdb.Object{
					&bsdb.Object{
						ID:                  1,
						Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						LocalVirtualGroupId: 0,
						BucketName:          "barry",
						ObjectName:          "2222",
						ObjectID:            common.HexToHash("2"),
						BucketID:            common.HexToHash("2"),
						PayloadSize:         0,
						Visibility:          "",
						ContentType:         "",
						CreateAt:            0,
						CreateTime:          0,
						ObjectStatus:        "",
						RedundancyType:      "",
						SourceType:          "",
						Checksums:           nil,
						LockedBalance:       common.HexToHash("2"),
						Removed:             false,
						UpdateTime:          0,
						UpdateAt:            0,
						DeleteAt:            0,
						DeleteReason:        "",
						CreateTxHash:        common.HexToHash("2"),
						UpdateTxHash:        common.HexToHash("2"),
						SealTxHash:          common.HexToHash("2"),
					},
				},
				&bsdb.Bucket{
					ID:                         848,
					Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					BucketName:                 "barry",
					Visibility:                 "VISIBILITY_TYPE_PRIVATE",
					BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					SourceType:                 "SOURCE_TYPE_ORIGIN",
					CreateAt:                   0,
					CreateTime:                 0,
					CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
					PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					GlobalVirtualGroupFamilyID: 1,
					ChargedReadQuota:           0,
					PaymentPriceTime:           0,
					Removed:                    false,
					Status:                     "",
					DeleteAt:                   0,
					DeleteReason:               "",
					UpdateAt:                   0,
					UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
					UpdateTime:                 0,
				}, nil
		},
	).Times(1)
	m.EXPECT().GetGvgByBucketAndLvgID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32) (*bsdb.GlobalVirtualGroup, error) {
			return &bsdb.GlobalVirtualGroup{
				ID:                    1,
				GlobalVirtualGroupId:  1,
				FamilyId:              1,
				PrimarySpId:           1,
				SecondarySpIds:        []uint32{1},
				StoredSize:            0,
				VirtualPaymentAddress: common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				TotalDeposit:          nil,
				CreateAt:              0,
				CreateTxHash:          common.HexToHash("1"),
				CreateTime:            0,
				UpdateAt:              0,
				UpdateTxHash:          common.HexToHash("1"),
				UpdateTime:            0,
				Removed:               false,
			}, nil
		},
	).Times(1)
	objects, err := a.GfSpListObjectsInGVGAndBucket(context.Background(), &types.GfSpListObjectsInGVGAndBucketRequest{
		GvgId:      1,
		BucketId:   1,
		StartAfter: 0,
		Limit:      11111,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2222", objects.Objects[0].Object.ObjectInfo.ObjectName)
}

func TestMetadataModular_GfSpListObjectsInGVGAndBucket_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsInGVGAndBucket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32, common.Hash, int) ([]*bsdb.Object, *bsdb.Bucket, error) {
			return []*bsdb.Object{
					&bsdb.Object{
						ID:                  1,
						Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						LocalVirtualGroupId: 0,
						BucketName:          "barry",
						ObjectName:          "2222",
						ObjectID:            common.HexToHash("2"),
						BucketID:            common.HexToHash("2"),
						PayloadSize:         0,
						Visibility:          "",
						ContentType:         "",
						CreateAt:            0,
						CreateTime:          0,
						ObjectStatus:        "",
						RedundancyType:      "",
						SourceType:          "",
						Checksums:           nil,
						LockedBalance:       common.HexToHash("2"),
						Removed:             false,
						UpdateTime:          0,
						UpdateAt:            0,
						DeleteAt:            0,
						DeleteReason:        "",
						CreateTxHash:        common.HexToHash("2"),
						UpdateTxHash:        common.HexToHash("2"),
						SealTxHash:          common.HexToHash("2"),
					},
				},
				&bsdb.Bucket{
					ID:                         848,
					Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					BucketName:                 "barry",
					Visibility:                 "VISIBILITY_TYPE_PRIVATE",
					BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					SourceType:                 "SOURCE_TYPE_ORIGIN",
					CreateAt:                   0,
					CreateTime:                 0,
					CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
					PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					GlobalVirtualGroupFamilyID: 1,
					ChargedReadQuota:           0,
					PaymentPriceTime:           0,
					Removed:                    false,
					Status:                     "",
					DeleteAt:                   0,
					DeleteReason:               "",
					UpdateAt:                   0,
					UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
					UpdateTime:                 0,
				}, nil
		},
	).Times(1)
	m.EXPECT().GetGvgByBucketAndLvgID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32) (*bsdb.GlobalVirtualGroup, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListObjectsInGVGAndBucket(context.Background(), &types.GfSpListObjectsInGVGAndBucketRequest{
		GvgId:      1,
		BucketId:   1,
		StartAfter: 0,
		Limit:      11111,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpListObjectsInGVGAndBucket_Failed2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsInGVGAndBucket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32, common.Hash, int) ([]*bsdb.Object, *bsdb.Bucket, error) {
			return nil, nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListObjectsInGVGAndBucket(context.Background(), &types.GfSpListObjectsInGVGAndBucketRequest{
		GvgId:      1,
		BucketId:   1,
		StartAfter: 0,
		Limit:      11111,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpListObjectsByGVGAndBucketForGC_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsByGVGAndBucketForGC(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32, common.Hash, int) ([]*bsdb.Object, *bsdb.Bucket, error) {
			return []*bsdb.Object{
					&bsdb.Object{
						ID:                  1,
						Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						LocalVirtualGroupId: 0,
						BucketName:          "barry",
						ObjectName:          "2222",
						ObjectID:            common.HexToHash("2"),
						BucketID:            common.HexToHash("2"),
						PayloadSize:         0,
						Visibility:          "",
						ContentType:         "",
						CreateAt:            0,
						CreateTime:          0,
						ObjectStatus:        "",
						RedundancyType:      "",
						SourceType:          "",
						Checksums:           nil,
						LockedBalance:       common.HexToHash("2"),
						Removed:             false,
						UpdateTime:          0,
						UpdateAt:            0,
						DeleteAt:            0,
						DeleteReason:        "",
						CreateTxHash:        common.HexToHash("2"),
						UpdateTxHash:        common.HexToHash("2"),
						SealTxHash:          common.HexToHash("2"),
					},
				},
				&bsdb.Bucket{
					ID:                         848,
					Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					BucketName:                 "barry",
					Visibility:                 "VISIBILITY_TYPE_PRIVATE",
					BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					SourceType:                 "SOURCE_TYPE_ORIGIN",
					CreateAt:                   0,
					CreateTime:                 0,
					CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
					PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					GlobalVirtualGroupFamilyID: 1,
					ChargedReadQuota:           0,
					PaymentPriceTime:           0,
					Removed:                    false,
					Status:                     "",
					DeleteAt:                   0,
					DeleteReason:               "",
					UpdateAt:                   0,
					UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
					UpdateTime:                 0,
				}, nil
		},
	).Times(1)
	m.EXPECT().GetGvgByBucketAndLvgID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32) (*bsdb.GlobalVirtualGroup, error) {
			return &bsdb.GlobalVirtualGroup{
				ID:                    1,
				GlobalVirtualGroupId:  1,
				FamilyId:              1,
				PrimarySpId:           1,
				SecondarySpIds:        []uint32{1},
				StoredSize:            0,
				VirtualPaymentAddress: common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				TotalDeposit:          nil,
				CreateAt:              0,
				CreateTxHash:          common.HexToHash("1"),
				CreateTime:            0,
				UpdateAt:              0,
				UpdateTxHash:          common.HexToHash("1"),
				UpdateTime:            0,
				Removed:               false,
			}, nil
		},
	).Times(1)
	objects, err := a.GfSpListObjectsByGVGAndBucketForGC(context.Background(), &types.GfSpListObjectsByGVGAndBucketForGCRequest{
		DstGvgId:   1,
		BucketId:   1,
		StartAfter: 0,
		Limit:      0,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2222", objects.Objects[0].Object.ObjectInfo.ObjectName)
}

func TestMetadataModular_GfSpListObjectsByGVGAndBucketForGC_Success2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsByGVGAndBucketForGC(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32, common.Hash, int) ([]*bsdb.Object, *bsdb.Bucket, error) {
			return []*bsdb.Object{
					&bsdb.Object{
						ID:                  1,
						Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						LocalVirtualGroupId: 0,
						BucketName:          "barry",
						ObjectName:          "2222",
						ObjectID:            common.HexToHash("2"),
						BucketID:            common.HexToHash("2"),
						PayloadSize:         0,
						Visibility:          "",
						ContentType:         "",
						CreateAt:            0,
						CreateTime:          0,
						ObjectStatus:        "",
						RedundancyType:      "",
						SourceType:          "",
						Checksums:           nil,
						LockedBalance:       common.HexToHash("2"),
						Removed:             false,
						UpdateTime:          0,
						UpdateAt:            0,
						DeleteAt:            0,
						DeleteReason:        "",
						CreateTxHash:        common.HexToHash("2"),
						UpdateTxHash:        common.HexToHash("2"),
						SealTxHash:          common.HexToHash("2"),
					},
				},
				&bsdb.Bucket{
					ID:                         848,
					Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					BucketName:                 "barry",
					Visibility:                 "VISIBILITY_TYPE_PRIVATE",
					BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					SourceType:                 "SOURCE_TYPE_ORIGIN",
					CreateAt:                   0,
					CreateTime:                 0,
					CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
					PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					GlobalVirtualGroupFamilyID: 1,
					ChargedReadQuota:           0,
					PaymentPriceTime:           0,
					Removed:                    false,
					Status:                     "",
					DeleteAt:                   0,
					DeleteReason:               "",
					UpdateAt:                   0,
					UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
					UpdateTime:                 0,
				}, nil
		},
	).Times(1)
	m.EXPECT().GetGvgByBucketAndLvgID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32) (*bsdb.GlobalVirtualGroup, error) {
			return &bsdb.GlobalVirtualGroup{
				ID:                    1,
				GlobalVirtualGroupId:  1,
				FamilyId:              1,
				PrimarySpId:           1,
				SecondarySpIds:        []uint32{1},
				StoredSize:            0,
				VirtualPaymentAddress: common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				TotalDeposit:          nil,
				CreateAt:              0,
				CreateTxHash:          common.HexToHash("1"),
				CreateTime:            0,
				UpdateAt:              0,
				UpdateTxHash:          common.HexToHash("1"),
				UpdateTime:            0,
				Removed:               false,
			}, nil
		},
	).Times(1)
	objects, err := a.GfSpListObjectsByGVGAndBucketForGC(context.Background(), &types.GfSpListObjectsByGVGAndBucketForGCRequest{
		DstGvgId:   1,
		BucketId:   1,
		StartAfter: 0,
		Limit:      11111,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2222", objects.Objects[0].Object.ObjectInfo.ObjectName)
}

func TestMetadataModular_GfSpListObjectsByGVGAndBucketForGC_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsByGVGAndBucketForGC(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32, common.Hash, int) ([]*bsdb.Object, *bsdb.Bucket, error) {
			return []*bsdb.Object{
					&bsdb.Object{
						ID:                  1,
						Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						LocalVirtualGroupId: 0,
						BucketName:          "barry",
						ObjectName:          "2222",
						ObjectID:            common.HexToHash("2"),
						BucketID:            common.HexToHash("2"),
						PayloadSize:         0,
						Visibility:          "",
						ContentType:         "",
						CreateAt:            0,
						CreateTime:          0,
						ObjectStatus:        "",
						RedundancyType:      "",
						SourceType:          "",
						Checksums:           nil,
						LockedBalance:       common.HexToHash("2"),
						Removed:             false,
						UpdateTime:          0,
						UpdateAt:            0,
						DeleteAt:            0,
						DeleteReason:        "",
						CreateTxHash:        common.HexToHash("2"),
						UpdateTxHash:        common.HexToHash("2"),
						SealTxHash:          common.HexToHash("2"),
					},
				},
				&bsdb.Bucket{
					ID:                         848,
					Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					BucketName:                 "barry",
					Visibility:                 "VISIBILITY_TYPE_PRIVATE",
					BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					SourceType:                 "SOURCE_TYPE_ORIGIN",
					CreateAt:                   0,
					CreateTime:                 0,
					CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
					PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					GlobalVirtualGroupFamilyID: 1,
					ChargedReadQuota:           0,
					PaymentPriceTime:           0,
					Removed:                    false,
					Status:                     "",
					DeleteAt:                   0,
					DeleteReason:               "",
					UpdateAt:                   0,
					UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
					UpdateTime:                 0,
				}, nil
		},
	).Times(1)
	m.EXPECT().GetGvgByBucketAndLvgID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32) (*bsdb.GlobalVirtualGroup, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListObjectsByGVGAndBucketForGC(context.Background(), &types.GfSpListObjectsByGVGAndBucketForGCRequest{
		DstGvgId:   1,
		BucketId:   1,
		StartAfter: 0,
		Limit:      11111,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpListObjectsByGVGAndBucketForGC_Failed2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsByGVGAndBucketForGC(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32, common.Hash, int) ([]*bsdb.Object, *bsdb.Bucket, error) {
			return nil, nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListObjectsByGVGAndBucketForGC(context.Background(), &types.GfSpListObjectsByGVGAndBucketForGCRequest{
		DstGvgId:   1,
		BucketId:   1,
		StartAfter: 0,
		Limit:      11111,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpListObjectsInGVG_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsInGVG(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(uint32, common.Hash, int) ([]*bsdb.Object, *bsdb.Bucket, error) {
			return []*bsdb.Object{
					&bsdb.Object{
						ID:                  1,
						Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						LocalVirtualGroupId: 0,
						BucketName:          "barry",
						ObjectName:          "2222",
						ObjectID:            common.HexToHash("2"),
						BucketID:            common.HexToHash("2"),
						PayloadSize:         0,
						Visibility:          "",
						ContentType:         "",
						CreateAt:            0,
						CreateTime:          0,
						ObjectStatus:        "",
						RedundancyType:      "",
						SourceType:          "",
						Checksums:           nil,
						LockedBalance:       common.HexToHash("2"),
						Removed:             false,
						UpdateTime:          0,
						UpdateAt:            0,
						DeleteAt:            0,
						DeleteReason:        "",
						CreateTxHash:        common.HexToHash("2"),
						UpdateTxHash:        common.HexToHash("2"),
						SealTxHash:          common.HexToHash("2"),
					},
				},
				&bsdb.Bucket{
					ID:                         848,
					Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					BucketName:                 "barry",
					Visibility:                 "VISIBILITY_TYPE_PRIVATE",
					BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					SourceType:                 "SOURCE_TYPE_ORIGIN",
					CreateAt:                   0,
					CreateTime:                 0,
					CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
					PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					GlobalVirtualGroupFamilyID: 1,
					ChargedReadQuota:           0,
					PaymentPriceTime:           0,
					Removed:                    false,
					Status:                     "",
					DeleteAt:                   0,
					DeleteReason:               "",
					UpdateAt:                   0,
					UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
					UpdateTime:                 0,
				}, nil
		},
	).Times(1)
	m.EXPECT().GetGvgByBucketAndLvgID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32) (*bsdb.GlobalVirtualGroup, error) {
			return &bsdb.GlobalVirtualGroup{
				ID:                    1,
				GlobalVirtualGroupId:  1,
				FamilyId:              1,
				PrimarySpId:           1,
				SecondarySpIds:        []uint32{1},
				StoredSize:            0,
				VirtualPaymentAddress: common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				TotalDeposit:          nil,
				CreateAt:              0,
				CreateTxHash:          common.HexToHash("1"),
				CreateTime:            0,
				UpdateAt:              0,
				UpdateTxHash:          common.HexToHash("1"),
				UpdateTime:            0,
				Removed:               false,
			}, nil
		},
	).Times(1)
	objects, err := a.GfSpListObjectsInGVG(context.Background(), &types.GfSpListObjectsInGVGRequest{
		GvgId:      1,
		StartAfter: 0,
		Limit:      0,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2222", objects.Objects[0].Object.ObjectInfo.ObjectName)
}

func TestMetadataModular_GfSpListObjectsInGVG_Success2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsInGVG(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(uint32, common.Hash, int) ([]*bsdb.Object, []*bsdb.Bucket, error) {
			return []*bsdb.Object{
					&bsdb.Object{
						ID:                  1,
						Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						LocalVirtualGroupId: 0,
						BucketName:          "barry",
						ObjectName:          "2222",
						ObjectID:            common.HexToHash("2"),
						BucketID:            common.HexToHash("1"),
						PayloadSize:         0,
						Visibility:          "",
						ContentType:         "",
						CreateAt:            0,
						CreateTime:          0,
						ObjectStatus:        "",
						RedundancyType:      "",
						SourceType:          "",
						Checksums:           nil,
						LockedBalance:       common.HexToHash("2"),
						Removed:             false,
						UpdateTime:          0,
						UpdateAt:            0,
						DeleteAt:            0,
						DeleteReason:        "",
						CreateTxHash:        common.HexToHash("2"),
						UpdateTxHash:        common.HexToHash("2"),
						SealTxHash:          common.HexToHash("2"),
					},
				}, []*bsdb.Bucket{
					&bsdb.Bucket{
						ID:                         848,
						Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
						Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
						BucketName:                 "barry",
						Visibility:                 "VISIBILITY_TYPE_PRIVATE",
						BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
						SourceType:                 "SOURCE_TYPE_ORIGIN",
						CreateAt:                   0,
						CreateTime:                 0,
						CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
						PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
						GlobalVirtualGroupFamilyID: 1,
						ChargedReadQuota:           0,
						PaymentPriceTime:           0,
						Removed:                    false,
						Status:                     "",
						DeleteAt:                   0,
						DeleteReason:               "",
						UpdateAt:                   0,
						UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
						UpdateTime:                 0,
					}}, nil
		},
	).Times(1)
	m.EXPECT().GetGvgByBucketAndLvgID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32) (*bsdb.GlobalVirtualGroup, error) {
			return &bsdb.GlobalVirtualGroup{
				ID:                    1,
				GlobalVirtualGroupId:  1,
				FamilyId:              1,
				PrimarySpId:           1,
				SecondarySpIds:        []uint32{1},
				StoredSize:            0,
				VirtualPaymentAddress: common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				TotalDeposit:          nil,
				CreateAt:              0,
				CreateTxHash:          common.HexToHash("1"),
				CreateTime:            0,
				UpdateAt:              0,
				UpdateTxHash:          common.HexToHash("1"),
				UpdateTime:            0,
				Removed:               false,
			}, nil
		},
	).Times(1)
	objects, err := a.GfSpListObjectsInGVG(context.Background(), &types.GfSpListObjectsInGVGRequest{
		GvgId:      1,
		StartAfter: 0,
		Limit:      11111,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2222", objects.Objects[0].Object.ObjectInfo.ObjectName)
}

func TestMetadataModular_GfSpListObjectsInGVG_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsInGVG(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(uint32, common.Hash, int) ([]*bsdb.Object, *bsdb.Bucket, error) {
			return []*bsdb.Object{
					&bsdb.Object{
						ID:                  1,
						Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
						LocalVirtualGroupId: 0,
						BucketName:          "barry",
						ObjectName:          "2222",
						ObjectID:            common.HexToHash("2"),
						BucketID:            common.HexToHash("2"),
						PayloadSize:         0,
						Visibility:          "",
						ContentType:         "",
						CreateAt:            0,
						CreateTime:          0,
						ObjectStatus:        "",
						RedundancyType:      "",
						SourceType:          "",
						Checksums:           nil,
						LockedBalance:       common.HexToHash("2"),
						Removed:             false,
						UpdateTime:          0,
						UpdateAt:            0,
						DeleteAt:            0,
						DeleteReason:        "",
						CreateTxHash:        common.HexToHash("2"),
						UpdateTxHash:        common.HexToHash("2"),
						SealTxHash:          common.HexToHash("2"),
					},
				},
				&bsdb.Bucket{
					ID:                         848,
					Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					BucketName:                 "barry",
					Visibility:                 "VISIBILITY_TYPE_PRIVATE",
					BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					SourceType:                 "SOURCE_TYPE_ORIGIN",
					CreateAt:                   0,
					CreateTime:                 0,
					CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
					PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					GlobalVirtualGroupFamilyID: 1,
					ChargedReadQuota:           0,
					PaymentPriceTime:           0,
					Removed:                    false,
					Status:                     "",
					DeleteAt:                   0,
					DeleteReason:               "",
					UpdateAt:                   0,
					UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
					UpdateTime:                 0,
				}, nil
		},
	).Times(1)
	m.EXPECT().GetGvgByBucketAndLvgID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32) (*bsdb.GlobalVirtualGroup, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListObjectsInGVG(context.Background(), &types.GfSpListObjectsInGVGRequest{
		GvgId:      1,
		StartAfter: 0,
		Limit:      11111,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpListObjectsInGVG_Failed2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListObjectsInGVG(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(uint32, common.Hash, int) ([]*bsdb.Object, *bsdb.Bucket, error) {
			return nil, nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListObjectsInGVG(context.Background(), &types.GfSpListObjectsInGVGRequest{
		GvgId:      1,
		StartAfter: 0,
		Limit:      11111,
	})
	assert.NotNil(t, err)
}

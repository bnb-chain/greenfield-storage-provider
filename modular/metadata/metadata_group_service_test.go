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

func TestMetadataModular_GfSpGetGroupList_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListGroupsByNameAndSourceType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, int, int, bool) ([]*bsdb.Group, int64, error) {
			return []*bsdb.Group{&bsdb.Group{
				ID:             1,
				Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
				GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				GroupName:      "test",
				SourceType:     "SOURCE_TYPE_ORIGIN",
				Extra:          "",
				AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
				Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
				ExpirationTime: 0,
				CreateAt:       0,
				CreateTime:     0,
				UpdateAt:       0,
				UpdateTime:     0,
				Removed:        false,
			}}, 1, nil
		},
	).Times(1)
	m.EXPECT().GetGroupMembersCount(gomock.Any()).DoAndReturn(
		func([]common.Hash) ([]*bsdb.GroupCount, error) {
			return []*bsdb.GroupCount{&bsdb.GroupCount{
				GroupID: common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				Count:   1,
			}}, nil
		},
	).Times(1)
	groups, err := a.GfSpGetGroupList(context.Background(), &types.GfSpGetGroupListRequest{
		Name:           "e",
		Prefix:         "t",
		SourceType:     "SOURCE_TYPE_ORIGIN",
		Limit:          1,
		Offset:         0,
		IncludeRemoved: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), groups.Count)
}

func TestMetadataModular_GfSpGetGroupList_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListGroupsByNameAndSourceType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, int, int, bool) ([]*bsdb.Group, int64, error) {
			return nil, 0, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpGetGroupList(context.Background(), &types.GfSpGetGroupListRequest{
		Name:           "e",
		Prefix:         "t",
		SourceType:     "SOURCE_TYPE_ORIGIN",
		Limit:          1,
		Offset:         0,
		IncludeRemoved: true,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpGetGroupList_Failed2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListGroupsByNameAndSourceType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, int, int, bool) ([]*bsdb.Group, int64, error) {
			return []*bsdb.Group{&bsdb.Group{
				ID:             1,
				Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
				GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				GroupName:      "test",
				SourceType:     "SOURCE_TYPE_ORIGIN",
				Extra:          "",
				AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
				Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
				ExpirationTime: 0,
				CreateAt:       0,
				CreateTime:     0,
				UpdateAt:       0,
				UpdateTime:     0,
				Removed:        false,
			}}, 1, nil
		},
	).Times(1)
	m.EXPECT().GetGroupMembersCount(gomock.Any()).DoAndReturn(
		func([]common.Hash) ([]*bsdb.GroupCount, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpGetGroupList(context.Background(), &types.GfSpGetGroupListRequest{
		Name:           "e",
		Prefix:         "t",
		SourceType:     "SOURCE_TYPE_ORIGIN",
		Limit:          1,
		Offset:         0,
		IncludeRemoved: true,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpGetUserGroups_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetUserGroups(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Address, common.Hash, int) ([]*bsdb.Group, error) {
			return []*bsdb.Group{&bsdb.Group{
				ID:             1,
				Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
				GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				GroupName:      "test",
				SourceType:     "SOURCE_TYPE_ORIGIN",
				Extra:          "",
				AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
				Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
				ExpirationTime: 0,
				CreateAt:       0,
				CreateTime:     0,
				UpdateAt:       0,
				UpdateTime:     0,
				Removed:        false,
			}}, nil
		},
	).Times(1)
	groups, err := a.GfSpGetUserGroups(context.Background(), &types.GfSpGetUserGroupsRequest{
		AccountId:  "0x84A0D38D64498414B14CD979159D57557345CD8B",
		Limit:      0,
		StartAfter: 0,
	})
	assert.Nil(t, err)
	assert.Equal(t, "test", groups.Groups[0].Group.GroupName)
}

func TestMetadataModular_GfSpGetUserGroups_Success2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetUserGroups(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Address, common.Hash, int) ([]*bsdb.Group, error) {
			return []*bsdb.Group{&bsdb.Group{
				ID:             1,
				Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
				GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				GroupName:      "test",
				SourceType:     "SOURCE_TYPE_ORIGIN",
				Extra:          "",
				AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
				Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
				ExpirationTime: 0,
				CreateAt:       0,
				CreateTime:     0,
				UpdateAt:       0,
				UpdateTime:     0,
				Removed:        false,
			}}, nil
		},
	).Times(1)
	groups, err := a.GfSpGetUserGroups(context.Background(), &types.GfSpGetUserGroupsRequest{
		AccountId:  "0x84A0D38D64498414B14CD979159D57557345CD8B",
		Limit:      1111,
		StartAfter: 0,
	})
	assert.Nil(t, err)
	assert.Equal(t, "test", groups.Groups[0].Group.GroupName)
}

func TestMetadataModular_GfSpGetUserGroups_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetUserGroups(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Address, common.Hash, int) ([]*bsdb.Group, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpGetUserGroups(context.Background(), &types.GfSpGetUserGroupsRequest{
		AccountId:  "0x84A0D38D64498414B14CD979159D57557345CD8B",
		Limit:      1,
		StartAfter: 0,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpGetGroupMembers_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetGroupMembers(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, common.Address, int) ([]*bsdb.Group, error) {
			return []*bsdb.Group{&bsdb.Group{
				ID:             1,
				Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
				GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				GroupName:      "test",
				SourceType:     "SOURCE_TYPE_ORIGIN",
				Extra:          "",
				AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
				Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
				ExpirationTime: 0,
				CreateAt:       0,
				CreateTime:     0,
				UpdateAt:       0,
				UpdateTime:     0,
				Removed:        false,
			}}, nil
		},
	).Times(1)
	groups, err := a.GfSpGetGroupMembers(context.Background(), &types.GfSpGetGroupMembersRequest{
		GroupId:    1,
		Limit:      0,
		StartAfter: "",
	})
	assert.Nil(t, err)
	assert.Equal(t, "test", groups.Groups[0].Group.GroupName)
}

func TestMetadataModular_GfSpGetGroupMembers_Success2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetGroupMembers(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, common.Address, int) ([]*bsdb.Group, error) {
			return []*bsdb.Group{&bsdb.Group{
				ID:             1,
				Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
				GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				GroupName:      "test",
				SourceType:     "SOURCE_TYPE_ORIGIN",
				Extra:          "",
				AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
				Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
				ExpirationTime: 0,
				CreateAt:       0,
				CreateTime:     0,
				UpdateAt:       0,
				UpdateTime:     0,
				Removed:        false,
			}}, nil
		},
	).Times(1)
	groups, err := a.GfSpGetGroupMembers(context.Background(), &types.GfSpGetGroupMembersRequest{
		GroupId:    1,
		Limit:      11110,
		StartAfter: "",
	})
	assert.Nil(t, err)
	assert.Equal(t, "test", groups.Groups[0].Group.GroupName)
}

func TestMetadataModular_GfSpGetGroupMembers_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetGroupMembers(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, common.Address, int) ([]*bsdb.Group, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpGetGroupMembers(context.Background(), &types.GfSpGetGroupMembersRequest{
		GroupId:    11111,
		Limit:      0,
		StartAfter: "",
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpGetUserOwnedGroups_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetUserOwnedGroups(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Address, common.Hash, int) ([]*bsdb.Group, error) {
			return []*bsdb.Group{&bsdb.Group{
				ID:             1,
				Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
				GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				GroupName:      "test",
				SourceType:     "SOURCE_TYPE_ORIGIN",
				Extra:          "",
				AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
				Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
				ExpirationTime: 0,
				CreateAt:       0,
				CreateTime:     0,
				UpdateAt:       0,
				UpdateTime:     0,
				Removed:        false,
			}}, nil
		},
	).Times(1)
	groups, err := a.GfSpGetUserOwnedGroups(context.Background(), &types.GfSpGetUserOwnedGroupsRequest{
		AccountId:  "0x84A0D38D64498414B14CD979159D57557345CD8B",
		Limit:      0,
		StartAfter: 0,
	})
	assert.Nil(t, err)
	assert.Equal(t, "test", groups.Groups[0].Group.GroupName)
}

func TestMetadataModular_GfSpGetUserOwnedGroups_Success2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetUserOwnedGroups(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Address, common.Hash, int) ([]*bsdb.Group, error) {
			return []*bsdb.Group{&bsdb.Group{
				ID:             1,
				Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
				GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				GroupName:      "test",
				SourceType:     "SOURCE_TYPE_ORIGIN",
				Extra:          "",
				AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
				Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
				ExpirationTime: 0,
				CreateAt:       0,
				CreateTime:     0,
				UpdateAt:       0,
				UpdateTime:     0,
				Removed:        false,
			}}, nil
		},
	).Times(1)
	groups, err := a.GfSpGetUserOwnedGroups(context.Background(), &types.GfSpGetUserOwnedGroupsRequest{
		AccountId:  "0x84A0D38D64498414B14CD979159D57557345CD8B",
		Limit:      11111,
		StartAfter: 0,
	})
	assert.Nil(t, err)
	assert.Equal(t, "test", groups.Groups[0].Group.GroupName)
}

func TestMetadataModular_GfSpGetUserOwnedGroups_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetUserOwnedGroups(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Address, common.Hash, int) ([]*bsdb.Group, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpGetUserOwnedGroups(context.Background(), &types.GfSpGetUserOwnedGroupsRequest{
		AccountId:  "0x84A0D38D64498414B14CD979159D57557345CD8B",
		Limit:      11111,
		StartAfter: 0,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpListGroupsByIDs_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListGroupsByIDs(gomock.Any()).DoAndReturn(
		func([]common.Hash) ([]*bsdb.Group, error) {
			return []*bsdb.Group{&bsdb.Group{
				ID:             1,
				Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
				GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				GroupName:      "test",
				SourceType:     "SOURCE_TYPE_ORIGIN",
				Extra:          "",
				AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
				Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
				ExpirationTime: 0,
				CreateAt:       0,
				CreateTime:     0,
				UpdateAt:       0,
				UpdateTime:     0,
				Removed:        false,
			}}, nil
		},
	).Times(1)
	groups, err := a.GfSpListGroupsByIDs(context.Background(), &types.GfSpListGroupsByIDsRequest{GroupIds: []uint64{1}})
	assert.Nil(t, err)
	assert.Equal(t, "test", groups.Groups[1].Group.GroupName)
}

func TestMetadataModular_GfSpListGroupsByIDs_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListGroupsByIDs(gomock.Any()).DoAndReturn(
		func([]common.Hash) ([]*bsdb.Group, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListGroupsByIDs(context.Background(), &types.GfSpListGroupsByIDsRequest{GroupIds: []uint64{1}})
	assert.NotNil(t, err)
}

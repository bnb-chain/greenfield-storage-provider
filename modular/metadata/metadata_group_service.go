package metadata

import (
	"context"

	"cosmossdk.io/math"
	"github.com/forbole/juno/v4/common"

	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	model "github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	storage_types "github.com/bnb-chain/greenfield/x/storage/types"
)

// GfSpGetGroupList get group list by queryName/prefix/sourceType
func (r *MetadataModular) GfSpGetGroupList(ctx context.Context, req *types.GfSpGetGroupListRequest) (resp *types.GfSpGetGroupListResponse, err error) {
	var (
		groups        []*model.Group
		groupCounts   []*model.GroupCount
		groupIDs      []common.Hash
		groupCountMap map[common.Hash]int64
		res           []*types.Group
		count         int64
	)

	ctx = log.Context(ctx, req)
	groups, count, err = r.baseApp.GfBsDB().ListGroupsByNameAndSourceType(req.Name, req.Prefix, req.SourceType, int(req.Limit), int(req.Offset), req.IncludeRemoved)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get group list", "error", err)
		return nil, err
	}

	res = make([]*types.Group, len(groups))
	if len(groups) > 0 {
		// generate group IDS
		groupIDs = make([]common.Hash, len(groups))
		for i, group := range groups {
			groupIDs[i] = group.GroupID
		}

		// get group counts by ids
		groupCounts, err = r.baseApp.GfBsDB().GetGroupMembersCount(groupIDs)
		if err != nil {
			log.CtxErrorw(ctx, "failed to get group count", "error", err)
			return nil, err
		}
		groupCountMap = make(map[common.Hash]int64)
		for _, g := range groupCounts {
			groupCountMap[g.GroupID] = g.Count
		}

		for i, group := range groups {
			res[i] = &types.Group{
				Group: &storage_types.GroupInfo{
					Owner:      group.Owner.String(),
					GroupName:  group.GroupName,
					SourceType: storage_types.SourceType(storage_types.SourceType_value[group.SourceType]),
					Id:         math.NewUintFromBigInt(group.GroupID.Big()),
					Extra:      group.Extra,
					Tags:       group.GetResourceTags(),
				},
				Operator:        group.Operator.String(),
				CreateAt:        group.CreateAt,
				CreateTime:      group.CreateTime,
				UpdateAt:        group.UpdateAt,
				UpdateTime:      group.UpdateTime,
				NumberOfMembers: groupCountMap[group.GroupID],
				Removed:         group.Removed,
			}
		}
	}

	resp = &types.GfSpGetGroupListResponse{
		Groups: res,
		Count:  count,
	}
	log.CtxInfo(ctx, "succeed to get group list")
	return resp, nil
}

// GfSpGetUserGroups get groups info by a user address
func (r *MetadataModular) GfSpGetUserGroups(ctx context.Context, req *types.GfSpGetUserGroupsRequest) (resp *types.GfSpGetUserGroupsResponse, err error) {
	var (
		groups []*model.Group
		res    []*types.GroupMember
		limit  int
	)

	ctx = log.Context(ctx, req)
	limit = int(req.Limit)
	//if the user doesn't specify a limit, the default value is ListGroupsDefaultMaxKeys
	if req.Limit == 0 {
		limit = model.ListGroupsDefaultLimit
	}

	// If the user specifies a value exceeding ListGroupsLimitSize, the response will only return up to ListGroupsLimitSize groups
	if req.Limit > model.ListGroupsLimitSize {
		limit = model.ListGroupsLimitSize
	}

	groups, err = r.baseApp.GfBsDB().GetUserGroups(common.HexToAddress(req.AccountId), common.BigToHash(math.NewUint(req.StartAfter).BigInt()), limit)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get user groups", "error", err)
		return nil, err
	}

	res = make([]*types.GroupMember, len(groups))
	for i, group := range groups {
		res[i] = &types.GroupMember{
			Group: &storage_types.GroupInfo{
				Owner:      group.Owner.String(),
				GroupName:  group.GroupName,
				SourceType: storage_types.SourceType(storage_types.SourceType_value[group.SourceType]),
				Id:         math.NewUintFromBigInt(group.GroupID.Big()),
				Extra:      group.Extra,
				Tags:       group.GetResourceTags(),
			},
			AccountId:      group.AccountID.String(),
			Operator:       group.Operator.String(),
			CreateAt:       group.CreateAt,
			CreateTime:     group.CreateTime,
			UpdateAt:       group.UpdateAt,
			UpdateTime:     group.UpdateTime,
			ExpirationTime: group.ExpirationTime,
			Removed:        group.Removed,
		}
	}

	resp = &types.GfSpGetUserGroupsResponse{Groups: res}
	log.CtxInfo(ctx, "succeed to get user groups")
	return resp, nil
}

// GfSpGetGroupMembers get group members by group id
func (r *MetadataModular) GfSpGetGroupMembers(ctx context.Context, req *types.GfSpGetGroupMembersRequest) (resp *types.GfSpGetGroupMembersResponse, err error) {
	var (
		groups []*model.Group
		res    []*types.GroupMember
		limit  int
	)

	ctx = log.Context(ctx, req)
	limit = int(req.Limit)
	//if the user doesn't specify a limit, the default value is ListGroupsDefaultMaxKeys
	if req.Limit == 0 {
		limit = model.ListGroupsDefaultLimit
	}

	// If the user specifies a value exceeding ListGroupsLimitSize, the response will only return up to ListGroupsLimitSize groups
	if req.Limit > model.ListGroupsLimitSize {
		limit = model.ListGroupsLimitSize
	}

	// If the user doesn't specify a start after, the default value is 0
	if req.StartAfter == "" {
		req.StartAfter = "0"
	}

	groups, err = r.baseApp.GfBsDB().GetGroupMembers(common.BigToHash(math.NewUint(req.GroupId).BigInt()), common.HexToAddress(req.StartAfter), limit)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get group members by group id", "error", err)
		return nil, err
	}

	res = make([]*types.GroupMember, len(groups))
	for i, group := range groups {
		res[i] = &types.GroupMember{
			Group: &storage_types.GroupInfo{
				Owner:      group.Owner.String(),
				GroupName:  group.GroupName,
				SourceType: storage_types.SourceType(storage_types.SourceType_value[group.SourceType]),
				Id:         math.NewUintFromBigInt(group.GroupID.Big()),
				Extra:      group.Extra,
				Tags:       group.GetResourceTags(),
			},
			AccountId:      group.AccountID.String(),
			Operator:       group.Operator.String(),
			CreateAt:       group.CreateAt,
			CreateTime:     group.CreateTime,
			UpdateAt:       group.UpdateAt,
			UpdateTime:     group.UpdateTime,
			ExpirationTime: group.ExpirationTime,
			Removed:        group.Removed,
		}
	}

	resp = &types.GfSpGetGroupMembersResponse{Groups: res}
	log.CtxInfo(ctx, "succeed to get group members by group id")
	return resp, nil
}

// GfSpGetUserOwnedGroups retrieve groups where the user is the owner
func (r *MetadataModular) GfSpGetUserOwnedGroups(ctx context.Context, req *types.GfSpGetUserOwnedGroupsRequest) (resp *types.GfSpGetUserOwnedGroupsResponse, err error) {
	var (
		groups []*model.Group
		res    []*types.GroupMember
		limit  int
	)

	ctx = log.Context(ctx, req)
	limit = int(req.Limit)
	//if the user doesn't specify a limit, the default value is ListGroupsDefaultMaxKeys
	if req.Limit == 0 {
		limit = model.ListGroupsDefaultLimit
	}

	// If the user specifies a value exceeding ListGroupsLimitSize, the response will only return up to ListGroupsLimitSize groups
	if req.Limit > model.ListGroupsLimitSize {
		limit = model.ListGroupsLimitSize
	}

	groups, err = r.baseApp.GfBsDB().GetUserOwnedGroups(common.HexToAddress(req.AccountId), common.BigToHash(math.NewUint(req.StartAfter).BigInt()), limit)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get user groups", "error", err)
		return nil, err
	}

	res = make([]*types.GroupMember, len(groups))
	for i, group := range groups {
		res[i] = &types.GroupMember{
			Group: &storage_types.GroupInfo{
				Owner:      group.Owner.String(),
				GroupName:  group.GroupName,
				SourceType: storage_types.SourceType(storage_types.SourceType_value[group.SourceType]),
				Id:         math.NewUintFromBigInt(group.GroupID.Big()),
				Extra:      group.Extra,
				Tags:       group.GetResourceTags(),
			},
			AccountId:      group.AccountID.String(),
			Operator:       group.Operator.String(),
			CreateAt:       group.CreateAt,
			CreateTime:     group.CreateTime,
			UpdateAt:       group.UpdateAt,
			UpdateTime:     group.UpdateTime,
			ExpirationTime: group.ExpirationTime,
			Removed:        group.Removed,
		}
	}

	resp = &types.GfSpGetUserOwnedGroupsResponse{Groups: res}
	log.CtxInfo(ctx, "succeed to get user groups")
	return resp, nil
}

// GfSpListGroupsByIDs list groups by group ids
func (r *MetadataModular) GfSpListGroupsByIDs(ctx context.Context, req *types.GfSpListGroupsByIDsRequest) (resp *types.GfSpListGroupsByIDsResponse, err error) {
	var (
		groups    []*model.Group
		ids       []common.Hash
		groupsMap map[uint64]*types.Group
	)

	ids = make([]common.Hash, len(req.GroupIds))
	for i, id := range req.GroupIds {
		ids[i] = common.BigToHash(math.NewUint(id).BigInt())
	}

	groups, err = r.baseApp.GfBsDB().ListGroupsByIDs(ids)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list groups by group ids", "error", err)
		return nil, err
	}

	groupsMap = make(map[uint64]*types.Group)
	for _, id := range req.GroupIds {
		groupsMap[id] = nil
	}

	for _, group := range groups {
		groupsMap[group.GroupID.Big().Uint64()] = &types.Group{
			Group: &storage_types.GroupInfo{
				Owner:      group.Owner.String(),
				GroupName:  group.GroupName,
				SourceType: storage_types.SourceType(storage_types.SourceType_value[group.SourceType]),
				Id:         math.NewUintFromBigInt(group.GroupID.Big()),
				Extra:      group.Extra,
				Tags:       group.GetResourceTags(),
			},
			Operator:   group.Operator.String(),
			CreateAt:   group.CreateAt,
			CreateTime: group.CreateTime,
			UpdateAt:   group.UpdateAt,
			UpdateTime: group.UpdateTime,
			Removed:    group.Removed,
		}
	}
	resp = &types.GfSpListGroupsByIDsResponse{Groups: groupsMap}
	log.CtxInfo(ctx, "succeed to list groups by group ids")
	return resp, nil
}

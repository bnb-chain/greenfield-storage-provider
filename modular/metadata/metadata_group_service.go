package metadata

import (
	"context"

	"cosmossdk.io/math"
	storage_types "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	model "github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

// GfSpGetGroupList get group list by queryName/prefix/sourceType
func (r *MetadataModular) GfSpGetGroupList(ctx context.Context, req *types.GfSpGetGroupListRequest) (resp *types.GfSpGetGroupListResponse, err error) {
	var (
		groups []*model.Group
		res    []*types.Group
	)

	ctx = log.Context(ctx, req)
	groups, err = r.baseApp.GfBsDB().ListGroupsByNameAndSourceType(req.Name, req.Prefix, req.SourceType, int(req.Limit), int(req.Offset))
	if err != nil {
		log.CtxErrorw(ctx, "failed to get group list", "error", err)
		return nil, err
	}

	res = make([]*types.Group, len(groups))
	for i, group := range groups {
		res[i] = &types.Group{
			Group: &storage_types.GroupInfo{
				Owner:      group.Owner.String(),
				GroupName:  group.GroupName,
				SourceType: storage_types.SourceType(storage_types.SourceType_value[group.SourceType]),
				Id:         math.NewUintFromBigInt(group.GroupID.Big()),
				Extra:      group.Extra,
			},
			Operator:   group.Operator.String(),
			CreateAt:   group.CreateAt,
			CreateTime: group.CreateTime,
			UpdateAt:   group.UpdateAt,
			UpdateTime: group.UpdateTime,
			Removed:    group.Removed,
		}
	}

	resp = &types.GfSpGetGroupListResponse{Groups: res}
	log.CtxInfo(ctx, "succeed to get group list")
	return resp, nil
}

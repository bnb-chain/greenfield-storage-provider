package metadata

import (
	"context"

	"cosmossdk.io/math"
	"github.com/forbole/juno/v4/common"

	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	model "github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

// GfSpListVirtualGroupFamiliesBySpID list virtual group families by sp id
func (r *MetadataModular) GfSpListVirtualGroupFamiliesBySpID(ctx context.Context, req *types.GfSpListVirtualGroupFamiliesBySpIDRequest) (resp *types.GfSpListVirtualGroupFamiliesBySpIDResponse, err error) {
	var (
		families []*model.VirtualGroupFamily
		res      []*types.GlobalVirtualGroupFamily
	)

	ctx = log.Context(ctx, req)
	families, err = r.baseApp.GfBsDB().ListVirtualGroupFamiliesBySpID(req.SpId)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list virtual group families by sp id", "error", err)
		return nil, err
	}

	res = make([]*types.GlobalVirtualGroupFamily, len(families))
	for i, family := range families {
		res[i] = &types.GlobalVirtualGroupFamily{
			Id:                    family.GlobalVirtualGroupFamilyId,
			GlobalVirtualGroupIds: family.GlobalVirtualGroupIds,
			VirtualPaymentAddress: family.VirtualPaymentAddress.String(),
		}
	}

	resp = &types.GfSpListVirtualGroupFamiliesBySpIDResponse{GlobalVirtualGroupFamilies: res}
	log.CtxInfow(ctx, "succeed to list virtual group families by sp id")
	return resp, nil
}

// GfSpGetGlobalVirtualGroupByGvgID get global virtual group by gvg id
func (r *MetadataModular) GfSpGetGlobalVirtualGroupByGvgID(ctx context.Context, req *types.GfSpGetGlobalVirtualGroupByGvgIDRequest) (resp *types.GfSpGetGlobalVirtualGroupByGvgIDResponse, err error) {
	var (
		gvg *model.GlobalVirtualGroup
		res *types.GlobalVirtualGroup
	)

	ctx = log.Context(ctx, req)
	gvg, err = r.baseApp.GfBsDB().GetGlobalVirtualGroupByGvgID(req.GvgId)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get global virtual group by gvg id", "error", err)
		return nil, err
	}

	res = &types.GlobalVirtualGroup{
		Id:                    gvg.GlobalVirtualGroupId,
		FamilyId:              gvg.FamilyId,
		PrimarySpId:           gvg.PrimarySpId,
		SecondarySpIds:        gvg.SecondarySpIds,
		StoredSize:            gvg.StoredSize,
		VirtualPaymentAddress: gvg.VirtualPaymentAddress.String(),
		TotalDeposit:          math.NewIntFromBigInt(gvg.TotalDeposit.Raw()),
	}

	resp = &types.GfSpGetGlobalVirtualGroupByGvgIDResponse{GlobalVirtualGroup: res}
	log.CtxInfow(ctx, "succeed to get global virtual group by gvg id")
	return resp, nil
}

// GfSpGetVirtualGroupFamilyBindingOnBucket get virtual group family binding on bucket
func (r *MetadataModular) GfSpGetVirtualGroupFamilyBindingOnBucket(ctx context.Context, req *types.GfSpGetVirtualGroupFamilyBindingOnBucketRequest) (resp *types.GfSpGetVirtualGroupFamilyBindingOnBucketResponse, err error) {
	var (
		family *model.VirtualGroupFamily
		res    *types.GlobalVirtualGroupFamily
	)

	ctx = log.Context(ctx, req)
	family, err = r.baseApp.GfBsDB().GetVirtualGroupFamilyBindingOnBucket(common.BigToHash(math.NewUint(req.BucketId).BigInt()))
	if err != nil {
		log.CtxErrorw(ctx, "failed to get virtual group family binding on bucket", "error", err)
		return nil, err
	}

	res = &types.GlobalVirtualGroupFamily{
		Id:                    family.GlobalVirtualGroupFamilyId,
		GlobalVirtualGroupIds: family.GlobalVirtualGroupIds,
		VirtualPaymentAddress: family.VirtualPaymentAddress.String(),
	}

	resp = &types.GfSpGetVirtualGroupFamilyBindingOnBucketResponse{GlobalVirtualGroupFamily: res}
	log.CtxInfow(ctx, "succeed to get virtual group family binding on bucket")
	return resp, nil
}

// GfSpGetVirtualGroupFamily get virtual group families by vgf id
func (r *MetadataModular) GfSpGetVirtualGroupFamily(ctx context.Context, req *types.GfSpGetVirtualGroupFamilyRequest) (resp *types.GfSpGetVirtualGroupFamilyResponse, err error) {
	var (
		family *model.VirtualGroupFamily
		res    *types.GlobalVirtualGroupFamily
	)

	ctx = log.Context(ctx, req)
	family, err = r.baseApp.GfBsDB().GetVirtualGroupFamiliesByVgfID(req.VgfId)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get virtual group families by vgf id", "error", err)
		return nil, err
	}

	res = &types.GlobalVirtualGroupFamily{
		Id:                    family.GlobalVirtualGroupFamilyId,
		GlobalVirtualGroupIds: family.GlobalVirtualGroupIds,
		VirtualPaymentAddress: family.VirtualPaymentAddress.String(),
	}

	resp = &types.GfSpGetVirtualGroupFamilyResponse{Vgf: res}
	log.CtxInfow(ctx, "succeed to get virtual group families by vgf id")
	return resp, nil
}

// GfSpGetGlobalVirtualGroup get global virtual group by lvg id and bucket id
func (r *MetadataModular) GfSpGetGlobalVirtualGroup(ctx context.Context, req *types.GfSpGetGlobalVirtualGroupRequest) (resp *types.GfSpGetGlobalVirtualGroupResponse, err error) {
	var (
		gvg *model.GlobalVirtualGroup
		res *types.GlobalVirtualGroup
	)

	ctx = log.Context(ctx, req)
	gvg, err = r.baseApp.GfBsDB().GfSpGetGvgByBucketAndLvgID(common.BigToHash(math.NewUint(req.BucketId).BigInt()), req.LvgId)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get global virtual group by lvg id and bucket id", "error", err)
		return nil, err
	}

	res = &types.GlobalVirtualGroup{
		Id:                    gvg.GlobalVirtualGroupId,
		FamilyId:              gvg.FamilyId,
		PrimarySpId:           gvg.PrimarySpId,
		SecondarySpIds:        gvg.SecondarySpIds,
		StoredSize:            gvg.StoredSize,
		VirtualPaymentAddress: gvg.VirtualPaymentAddress.String(),
		TotalDeposit:          math.NewIntFromBigInt(gvg.TotalDeposit.Raw()),
	}

	resp = &types.GfSpGetGlobalVirtualGroupResponse{Gvg: res}
	log.CtxInfow(ctx, "succeed to get global virtual group by lvg id and bucket id")
	return resp, nil
}

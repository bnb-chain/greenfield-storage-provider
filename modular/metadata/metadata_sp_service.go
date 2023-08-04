package metadata

import (
	"context"
	"errors"

	"cosmossdk.io/math"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

// GfSpGetEndpointBySpId get endpoint by sp id
func (r *MetadataModular) GfSpGetEndpointBySpId(
	ctx context.Context,
	req *types.GfSpGetEndpointBySpIdRequest) (
	resp *types.GfSpGetEndpointBySpIdResponse, err error) {
	ctx = log.Context(ctx, req)

	sp, err := r.baseApp.GfSpDB().GetSpById(req.SpId)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get sp", "error", err)
		return
	}

	resp = &types.GfSpGetEndpointBySpIdResponse{Endpoint: sp.Endpoint}
	log.CtxInfow(ctx, "succeed to get endpoint by a sp id")
	return resp, nil
}

// GfSpGetSPInfo get sp info by operator address
func (r *MetadataModular) GfSpGetSPInfo(ctx context.Context, req *types.GfSpGetSPInfoRequest) (resp *types.GfSpGetSPInfoResponse, err error) {
	var (
		sp  *bsdb.StorageProvider
		res *sptypes.StorageProvider
	)

	ctx = log.Context(ctx, req)
	sp, err = r.baseApp.GfBsDB().GetSPByAddress(common.HexToAddress(req.OperatorAddress))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoSuchSP
		}
		log.CtxErrorw(ctx, "failed to get sp info by operator address", "error", err)
		return nil, err
	}

	res = &sptypes.StorageProvider{
		Id:              sp.SpId,
		OperatorAddress: sp.OperatorAddress.String(),
		FundingAddress:  sp.FundingAddress.String(),
		SealAddress:     sp.SealAddress.String(),
		ApprovalAddress: sp.ApprovalAddress.String(),
		GcAddress:       sp.GcAddress.String(),
		TotalDeposit:    math.NewIntFromBigInt(sp.TotalDeposit.Raw()),
		Status:          sptypes.Status(sptypes.Status_value[sp.Status]),
		Endpoint:        sp.Endpoint,
		Description: sptypes.Description{
			Moniker:         sp.Moniker,
			Identity:        sp.Identity,
			Website:         sp.Website,
			SecurityContact: sp.SecurityContact,
			Details:         sp.Details,
		},
		BlsKey: []byte(sp.BlsKey),
	}

	resp = &types.GfSpGetSPInfoResponse{StorageProvider: res}
	log.CtxInfow(ctx, "succeed to get sp info by operator address")
	return resp, nil
}

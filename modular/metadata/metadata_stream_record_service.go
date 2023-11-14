package metadata

import (
	"context"
	"sort"
	"time"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
)

func (r *MetadataModular) GfSpPrimarySpIncomeDetails(ctx context.Context, req *types.GfSpPrimarySpIncomeDetailsRequest) (resp *types.GfSpPrimarySpIncomeDetailsResponse, err error) {
	resp = &types.GfSpPrimarySpIncomeDetailsResponse{}
	ctx = log.Context(ctx, req)

	currentTimestampInSec := time.Now().Unix()
	primarySpIncomeMetaList, err := r.baseApp.GfBsDB().GetPrimarySPStreamRecordBySpID(req.GetSpId())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get primary Sp Income Details", "error", err)
		return nil, err
	}
	resp.CurrentTimestamp = currentTimestampInSec
	resp.PrimarySpIncomeDetails = make([]*types.PrimarySpIncomeDetail, len(primarySpIncomeMetaList))
	for i, primarySpIncomeMeta := range primarySpIncomeMetaList {
		streamRecord := &paymenttypes.StreamRecord{
			Account:           primarySpIncomeMeta.Account.String(),
			CrudTimestamp:     primarySpIncomeMeta.CrudTimestamp,
			NetflowRate:       math.NewIntFromBigInt(primarySpIncomeMeta.NetflowRate.Raw()),
			StaticBalance:     math.NewIntFromBigInt(primarySpIncomeMeta.StaticBalance.Raw()),
			BufferBalance:     math.NewIntFromBigInt(primarySpIncomeMeta.BufferBalance.Raw()),
			LockBalance:       math.NewIntFromBigInt(primarySpIncomeMeta.LockBalance.Raw()),
			Status:            paymenttypes.StreamAccountStatus(paymenttypes.StreamAccountStatus_value[primarySpIncomeMeta.Status]),
			SettleTimestamp:   primarySpIncomeMeta.SettleTimestamp,
			OutFlowCount:      primarySpIncomeMeta.OutFlowCount,
			FrozenNetflowRate: math.NewIntFromBigInt(primarySpIncomeMeta.FrozenNetflowRate.Raw()),
		}

		resp.PrimarySpIncomeDetails[i] = &types.PrimarySpIncomeDetail{
			VgfId:        primarySpIncomeMeta.GlobalVirtualGroupFamilyId,
			StreamRecord: streamRecord,
			Income:       streamRecord.NetflowRate.Mul(math.NewInt(currentTimestampInSec - streamRecord.CrudTimestamp)).Add(streamRecord.StaticBalance),
		}
	}

	log.CtxInfow(ctx, "succeed to get primary sp income details")
	return resp, nil
}
func (r *MetadataModular) GfSpSecondarySpIncomeDetails(ctx context.Context, req *types.GfSpSecondarySpIncomeDetailsRequest) (resp *types.GfSpSecondarySpIncomeDetailsResponse, err error) {
	resp = &types.GfSpSecondarySpIncomeDetailsResponse{}

	ctx = log.Context(ctx, req)

	currentTimestampInSec := time.Now().Unix()
	secondarySpIncomeMetaList, err := r.baseApp.GfBsDB().GetSecondarySPStreamRecordBySpID(req.GetSpId())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get secondary Sp Income Details", "error", err)
		return nil, err
	}
	resp.CurrentTimestamp = currentTimestampInSec
	secondarySpIncomeDetails := make([]*types.SecondarySpIncomeDetail, len(secondarySpIncomeMetaList))
	for i, secondarySpIncomeMeta := range secondarySpIncomeMetaList {
		streamRecord := &paymenttypes.StreamRecord{
			Account:           secondarySpIncomeMeta.Account.String(),
			CrudTimestamp:     secondarySpIncomeMeta.CrudTimestamp,
			NetflowRate:       math.NewIntFromBigInt(secondarySpIncomeMeta.NetflowRate.Raw()),
			StaticBalance:     math.NewIntFromBigInt(secondarySpIncomeMeta.StaticBalance.Raw()),
			BufferBalance:     math.NewIntFromBigInt(secondarySpIncomeMeta.BufferBalance.Raw()),
			LockBalance:       math.NewIntFromBigInt(secondarySpIncomeMeta.LockBalance.Raw()),
			Status:            paymenttypes.StreamAccountStatus(paymenttypes.StreamAccountStatus_value[secondarySpIncomeMeta.Status]),
			SettleTimestamp:   secondarySpIncomeMeta.SettleTimestamp,
			OutFlowCount:      secondarySpIncomeMeta.OutFlowCount,
			FrozenNetflowRate: math.NewIntFromBigInt(secondarySpIncomeMeta.FrozenNetflowRate.Raw()),
		}

		secondarySpIncomeDetails[i] = &types.SecondarySpIncomeDetail{
			GvgId:        secondarySpIncomeMeta.GlobalVirtualGroupId,
			StreamRecord: streamRecord,
			Income:       streamRecord.NetflowRate.Mul(math.NewInt(currentTimestampInSec - streamRecord.CrudTimestamp)).Add(streamRecord.StaticBalance).Quo(math.NewInt(int64(6))),
		}
	}
	// sort by Income desc
	sort.Slice(secondarySpIncomeDetails, func(i, j int) bool {
		return secondarySpIncomeDetails[i].Income.GT(secondarySpIncomeDetails[j].Income)
	})
	resp.SecondarySpIncomeDetails = secondarySpIncomeDetails

	log.CtxInfow(ctx, "succeed to get secondary sp income details")
	return resp, nil
}

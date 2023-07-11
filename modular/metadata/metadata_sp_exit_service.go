package metadata

import (
	"context"
	"fmt"
	"net/http"

	"cosmossdk.io/math"
	storage_types "github.com/bnb-chain/greenfield/x/storage/types"
	virtual_types "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"github.com/forbole/juno/v4/common"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	model "github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

var (
	ErrNoEvents = gfsperrors.Register(MetadataModularName, http.StatusNotFound, 90004, "not found events")
)

// GfSpListVirtualGroupFamiliesBySpID list virtual group families by sp id
func (r *MetadataModular) GfSpListVirtualGroupFamiliesBySpID(ctx context.Context, req *types.GfSpListVirtualGroupFamiliesBySpIDRequest) (resp *types.GfSpListVirtualGroupFamiliesBySpIDResponse, err error) {
	var (
		families []*model.VirtualGroupFamily
		res      []*virtual_types.GlobalVirtualGroupFamily
	)

	ctx = log.Context(ctx, req)
	families, err = r.baseApp.GfBsDB().ListVirtualGroupFamiliesBySpID(req.SpId)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list virtual group families by sp id", "error", err)
		return nil, err
	}

	res = make([]*virtual_types.GlobalVirtualGroupFamily, len(families))
	for i, family := range families {
		res[i] = &virtual_types.GlobalVirtualGroupFamily{
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
		res *virtual_types.GlobalVirtualGroup
	)

	ctx = log.Context(ctx, req)
	gvg, err = r.baseApp.GfBsDB().GetGlobalVirtualGroupByGvgID(req.GvgId)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get global virtual group by gvg id", "error", err)
		return nil, err
	}

	if gvg != nil {
		if err != nil {
			log.CtxErrorw(ctx, "failed to parse global virtual group ids", "error", err)
			return nil, err
		}
		res = &virtual_types.GlobalVirtualGroup{
			Id:                    gvg.GlobalVirtualGroupId,
			FamilyId:              gvg.FamilyId,
			PrimarySpId:           gvg.PrimarySpId,
			SecondarySpIds:        gvg.SecondarySpIds,
			StoredSize:            gvg.StoredSize,
			VirtualPaymentAddress: gvg.VirtualPaymentAddress.String(),
			TotalDeposit:          math.NewIntFromBigInt(gvg.TotalDeposit.Raw()),
		}
	}

	resp = &types.GfSpGetGlobalVirtualGroupByGvgIDResponse{GlobalVirtualGroup: res}
	log.CtxInfow(ctx, "succeed to get global virtual group by gvg id")
	return resp, nil
}

// GfSpGetVirtualGroupFamily get virtual group families by vgf id
func (r *MetadataModular) GfSpGetVirtualGroupFamily(ctx context.Context, req *types.GfSpGetVirtualGroupFamilyRequest) (resp *types.GfSpGetVirtualGroupFamilyResponse, err error) {
	var (
		family *model.VirtualGroupFamily
		res    *virtual_types.GlobalVirtualGroupFamily
	)

	ctx = log.Context(ctx, req)
	family, err = r.baseApp.GfBsDB().GetVirtualGroupFamiliesByVgfID(req.VgfId)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get virtual group families by vgf id", "error", err)
		return nil, err
	}

	if family != nil {
		res = &virtual_types.GlobalVirtualGroupFamily{
			Id:                    family.GlobalVirtualGroupFamilyId,
			GlobalVirtualGroupIds: family.GlobalVirtualGroupIds,
			VirtualPaymentAddress: family.VirtualPaymentAddress.String(),
		}
	}

	resp = &types.GfSpGetVirtualGroupFamilyResponse{Vgf: res}
	log.CtxInfow(ctx, "succeed to get virtual group families by vgf id")
	return resp, nil
}

// GfSpGetGlobalVirtualGroup get global virtual group by lvg id and bucket id
func (r *MetadataModular) GfSpGetGlobalVirtualGroup(ctx context.Context, req *types.GfSpGetGlobalVirtualGroupRequest) (resp *types.GfSpGetGlobalVirtualGroupResponse, err error) {
	var (
		gvg *model.GlobalVirtualGroup
		res *virtual_types.GlobalVirtualGroup
	)

	ctx = log.Context(ctx, req)
	gvg, err = r.baseApp.GfBsDB().GetGvgByBucketAndLvgID(common.BigToHash(math.NewUint(req.BucketId).BigInt()), req.LvgId)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get global virtual group by lvg id and bucket id", "error", err)
		return nil, err
	}

	if gvg != nil {
		res = &virtual_types.GlobalVirtualGroup{
			Id:                    gvg.GlobalVirtualGroupId,
			FamilyId:              gvg.FamilyId,
			PrimarySpId:           gvg.PrimarySpId,
			SecondarySpIds:        gvg.SecondarySpIds,
			StoredSize:            gvg.StoredSize,
			VirtualPaymentAddress: gvg.VirtualPaymentAddress.String(),
			TotalDeposit:          math.NewIntFromBigInt(gvg.TotalDeposit.Raw()),
		}
	}

	resp = &types.GfSpGetGlobalVirtualGroupResponse{Gvg: res}
	log.CtxInfow(ctx, "succeed to get global virtual group by lvg id and bucket id")
	return resp, nil
}

// GfSpListMigrateBucketEvents list migrate bucket events
func (r *MetadataModular) GfSpListMigrateBucketEvents(ctx context.Context, req *types.GfSpListMigrateBucketEventsRequest) (resp *types.GfSpListMigrateBucketEventsResponse, err error) {
	var (
		events           []*model.EventMigrationBucket
		completeEvents   []*model.EventCompleteMigrationBucket
		cancelEvents     []*model.EventCancelMigrationBucket
		spEvent          []*storage_types.EventMigrationBucket
		spCompleteEvents []*storage_types.EventCompleteMigrationBucket
		//spCancelEvents    []*storage_types.EventCancelMigrationBucket
		eventsMap         map[storage_types.Uint]*storage_types.EventMigrationBucket
		completeEventsMap map[storage_types.Uint]*storage_types.EventCompleteMigrationBucket
		cancelEventsMap   map[storage_types.Uint]*storage_types.EventCancelMigrationBucket
		res               []*types.ListMigrateBucketEvents
	)

	ctx = log.Context(ctx, req)
	events, completeEvents, cancelEvents, err = r.baseApp.GfBsDB().ListMigrateBucketEvents(req.BlockId, req.SpId)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list migrate bucket events", "error", err)
		return nil, err
	}

	if events == nil && completeEvents == nil && cancelEvents == nil {
		return nil, ErrNoEvents
	}

	eventsMap = make(map[storage_types.Uint]*storage_types.EventMigrationBucket)
	spEvent = make([]*storage_types.EventMigrationBucket, len(events))
	for i, event := range events {
		e := &storage_types.EventMigrationBucket{
			Operator:       event.Operator.String(),
			BucketName:     event.BucketName,
			BucketId:       math.NewUintFromBigInt(event.BucketID.Big()),
			DstPrimarySpId: event.DstPrimarySpId,
		}
		spEvent[i] = e
		eventsMap[e.BucketId] = e
	}

	completeEventsMap = make(map[storage_types.Uint]*storage_types.EventCompleteMigrationBucket)
	spCompleteEvents = make([]*storage_types.EventCompleteMigrationBucket, len(completeEvents))
	for i, event := range completeEvents {
		e := &storage_types.EventCompleteMigrationBucket{
			Operator:                   event.Operator.String(),
			BucketName:                 event.BucketName,
			BucketId:                   math.NewUintFromBigInt(event.BucketID.Big()),
			GlobalVirtualGroupFamilyId: event.GlobalVirtualGroupFamilyId,
			// TODO BARRY
			//GvgMappings:                event.GvgMappings,
		}
		spCompleteEvents[i] = e
		completeEventsMap[e.BucketId] = e
	}

	//cancelEventsMap = make(map[storage_types.Uint]*storage_types.EventCancelMigrationBucket)
	//spCancelEvents = make([]*storage_types.EventCancelMigrationBucket, len(cancelEvents))
	//for i, event := range cancelEvents {
	//	e := &storage_types.EventCancelMigrationBucket{
	//		Operator:   event.Operator.String(),
	//		BucketName: event.BucketName,
	//		BucketId:   math.NewUintFromBigInt(event.BucketID.Big()),
	//	}
	//	spCancelEvents[i] = e
	//	cancelEventsMap[e.BucketId] = e
	//}

	res = make([]*types.ListMigrateBucketEvents, 0)
	for _, event := range eventsMap {
		complete := completeEventsMap[event.BucketId]
		cancel := cancelEventsMap[event.BucketId]
		res = append(res, &types.ListMigrateBucketEvents{
			Events:         event,
			CompleteEvents: complete,
			CancelEvents:   cancel,
		})
	}

	resp = &types.GfSpListMigrateBucketEventsResponse{Events: res}
	log.CtxInfow(ctx, "succeed to list migrate bucket events")
	return resp, nil
}

// GfSpListSwapOutEvents list swap out events
func (r *MetadataModular) GfSpListSwapOutEvents(ctx context.Context, req *types.GfSpListSwapOutEventsRequest) (resp *types.GfSpListSwapOutEventsResponse, err error) {
	var (
		events            []*model.EventSwapOut
		completeEvents    []*model.EventCompleteSwapOut
		cancelEvents      []*model.EventCancelSwapOut
		spEvent           []*virtual_types.EventSwapOut
		spCompleteEvents  []*virtual_types.EventCompleteSwapOut
		spCancelEvents    []*virtual_types.EventCancelSwapOut
		eventsMap         map[string]*virtual_types.EventSwapOut
		completeEventsMap map[string]*virtual_types.EventCompleteSwapOut
		cancelEventsMap   map[string]*virtual_types.EventCancelSwapOut
		res               []*types.ListSwapOutEvents
	)

	ctx = log.Context(ctx, req)
	events, completeEvents, cancelEvents, err = r.baseApp.GfBsDB().ListSwapOutEvents(req.BlockId, req.SpId)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list migrate swap out events", "error", err)
		return nil, err
	}

	if events == nil && completeEvents == nil && cancelEvents == nil {
		return nil, ErrNoEvents
	}

	eventsMap = make(map[string]*virtual_types.EventSwapOut)
	spEvent = make([]*virtual_types.EventSwapOut, len(events))
	for i, event := range events {
		e := &virtual_types.EventSwapOut{
			StorageProviderId:          event.StorageProviderId,
			GlobalVirtualGroupFamilyId: event.GlobalVirtualGroupFamilyId,
			GlobalVirtualGroupIds:      event.GlobalVirtualGroupIds,
			SuccessorSpId:              event.SuccessorSpId,
		}
		spEvent[i] = e
		// StorageProviderId and GlobalVirtualGroupFamilyId can be identified continuous event
		idx := fmt.Sprintf("%d+%d", event.StorageProviderId, event.GlobalVirtualGroupFamilyId)
		eventsMap[idx] = e
	}

	completeEventsMap = make(map[string]*virtual_types.EventCompleteSwapOut)
	spCompleteEvents = make([]*virtual_types.EventCompleteSwapOut, len(completeEvents))
	for i, event := range completeEvents {
		e := &virtual_types.EventCompleteSwapOut{
			StorageProviderId:          event.StorageProviderId,
			GlobalVirtualGroupFamilyId: event.GlobalVirtualGroupFamilyId,
			GlobalVirtualGroupIds:      event.GlobalVirtualGroupIds,
		}
		spCompleteEvents[i] = e
		idx := fmt.Sprintf("%d+%d", event.StorageProviderId, event.GlobalVirtualGroupFamilyId)
		completeEventsMap[idx] = e
	}

	cancelEventsMap = make(map[string]*virtual_types.EventCancelSwapOut)
	spCancelEvents = make([]*virtual_types.EventCancelSwapOut, len(cancelEvents))
	for i, event := range cancelEvents {
		e := &virtual_types.EventCancelSwapOut{
			StorageProviderId:          event.StorageProviderId,
			GlobalVirtualGroupFamilyId: event.GlobalVirtualGroupFamilyId,
			GlobalVirtualGroupIds:      event.GlobalVirtualGroupIds,
		}
		spCancelEvents[i] = e
		idx := fmt.Sprintf("%d+%d", event.StorageProviderId, event.GlobalVirtualGroupFamilyId)
		cancelEventsMap[idx] = e
	}

	res = make([]*types.ListSwapOutEvents, 0)
	for _, event := range eventsMap {
		idx := fmt.Sprintf("%d+%d", event.StorageProviderId, event.GlobalVirtualGroupFamilyId)
		complete := completeEventsMap[idx]
		cancel := cancelEventsMap[idx]
		res = append(res, &types.ListSwapOutEvents{
			Events:         event,
			CompleteEvents: complete,
			CancelEvents:   cancel,
		})
	}

	resp = &types.GfSpListSwapOutEventsResponse{Events: res}
	log.CtxInfow(ctx, "succeed to list migrate swap out events")
	return resp, nil
}

// GfSpListGlobalVirtualGroupsBySecondarySP list global virtual group by secondary sp id
func (r *MetadataModular) GfSpListGlobalVirtualGroupsBySecondarySP(ctx context.Context, req *types.GfSpListGlobalVirtualGroupsBySecondarySPRequest) (resp *types.GfSpListGlobalVirtualGroupsBySecondarySPResponse, err error) {
	var (
		groups []*model.GlobalVirtualGroup
		res    []*virtual_types.GlobalVirtualGroup
	)

	ctx = log.Context(ctx, req)
	groups, err = r.baseApp.GfBsDB().ListGvgBySecondarySpID(req.SpId)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get global virtual group by lvg id and bucket id", "error", err)
		return nil, err
	}

	res = make([]*virtual_types.GlobalVirtualGroup, len(groups))
	for i, gvg := range groups {
		res[i] = &virtual_types.GlobalVirtualGroup{
			Id:                    gvg.GlobalVirtualGroupId,
			FamilyId:              gvg.FamilyId,
			PrimarySpId:           gvg.PrimarySpId,
			SecondarySpIds:        gvg.SecondarySpIds,
			StoredSize:            gvg.StoredSize,
			VirtualPaymentAddress: gvg.VirtualPaymentAddress.String(),
			TotalDeposit:          math.NewIntFromBigInt(gvg.TotalDeposit.Raw()),
		}
	}

	resp = &types.GfSpListGlobalVirtualGroupsBySecondarySPResponse{Groups: res}
	log.CtxInfow(ctx, "succeed to get global virtual group by secondary sp id")
	return resp, nil
}

// GfSpListGlobalVirtualGroupsByBucket list global virtual group by bucket id
func (r *MetadataModular) GfSpListGlobalVirtualGroupsByBucket(ctx context.Context, req *types.GfSpListGlobalVirtualGroupsByBucketRequest) (resp *types.GfSpListGlobalVirtualGroupsByBucketResponse, err error) {
	var (
		groups []*model.GlobalVirtualGroup
		res    []*virtual_types.GlobalVirtualGroup
	)

	ctx = log.Context(ctx, req)
	groups, err = r.baseApp.GfBsDB().ListGvgByBucketID(common.BigToHash(math.NewUint(req.BucketId).BigInt()))
	if err != nil {
		log.CtxErrorw(ctx, "failed to list global virtual group by bucket id", "error", err)
		return nil, err
	}

	res = make([]*virtual_types.GlobalVirtualGroup, len(groups))
	for i, gvg := range groups {
		res[i] = &virtual_types.GlobalVirtualGroup{
			Id:                    gvg.GlobalVirtualGroupId,
			FamilyId:              gvg.FamilyId,
			PrimarySpId:           gvg.PrimarySpId,
			SecondarySpIds:        gvg.SecondarySpIds,
			StoredSize:            gvg.StoredSize,
			VirtualPaymentAddress: gvg.VirtualPaymentAddress.String(),
			TotalDeposit:          math.NewIntFromBigInt(gvg.TotalDeposit.Raw()),
		}
	}

	resp = &types.GfSpListGlobalVirtualGroupsByBucketResponse{Groups: res}
	log.CtxInfow(ctx, "succeed to list global virtual group by bucket id")
	return resp, nil
}

// GfSpListSpExitEvents list migrate sp exit events
func (r *MetadataModular) GfSpListSpExitEvents(ctx context.Context, req *types.GfSpListSpExitEventsRequest) (resp *types.GfSpListSpExitEventsResponse, err error) {
	var (
		event           *model.EventStorageProviderExit
		completeEvent   *model.EventCompleteStorageProviderExit
		spEvent         *virtual_types.EventStorageProviderExit
		spCompleteEvent *virtual_types.EventCompleteStorageProviderExit
	)
	log.Debugw("GfSpListSpExitEvents", "block-id", req.BlockId, "operator-address", req.OperatorAddress)
	ctx = log.Context(ctx, req)
	event, completeEvent, err = r.baseApp.GfBsDB().ListSpExitEvents(req.BlockId, common.HexToAddress(req.OperatorAddress))
	if err != nil {
		log.CtxErrorw(ctx, "failed to list migrate swap out events", "error", err)
		return nil, err
	}

	if event != nil {
		spEvent = &virtual_types.EventStorageProviderExit{
			StorageProviderId: event.StorageProviderId,
			OperatorAddress:   event.OperatorAddress.String(),
		}
	}

	if completeEvent != nil {
		spCompleteEvent = &virtual_types.EventCompleteStorageProviderExit{
			StorageProviderId: completeEvent.StorageProviderId,
			OperatorAddress:   completeEvent.OperatorAddress.String(),
			TotalDeposit:      math.NewIntFromBigInt(completeEvent.TotalDeposit.Raw()),
		}
	}

	resp = &types.GfSpListSpExitEventsResponse{Events: &types.ListSpExitEvents{
		Event:         spEvent,
		CompleteEvent: spCompleteEvent,
	}}
	log.CtxInfow(ctx, "succeed to list migrate swap out events")
	return resp, nil
}

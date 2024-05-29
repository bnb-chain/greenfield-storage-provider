package metadata

import (
	"context"

	"cosmossdk.io/math"
	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	model "github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	storage_types "github.com/bnb-chain/greenfield/x/storage/types"
	virtual_types "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

// GfSpListVirtualGroupFamiliesBySpID list virtual group families by sp id
func (r *MetadataModular) GfSpListVirtualGroupFamiliesBySpID(ctx context.Context, req *types.GfSpListVirtualGroupFamiliesBySpIDRequest) (resp *types.GfSpListVirtualGroupFamiliesBySpIDResponse, err error) {
	var (
		families []*model.GlobalVirtualGroupFamily
		res      []*virtual_types.GlobalVirtualGroupFamily
	)

	ctx = log.Context(ctx, req)
	log.Debugw("GfSpListVirtualGroupFamiliesBySpID", "sp-id", req.SpId)
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
	log.CtxInfow(ctx, "succeed to list virtual group families by sp id", "request", req, "response", resp)
	return resp, nil
}

// GfSpGetGlobalVirtualGroupByGvgID get global virtual group by gvg id
func (r *MetadataModular) GfSpGetGlobalVirtualGroupByGvgID(ctx context.Context, req *types.GfSpGetGlobalVirtualGroupByGvgIDRequest) (resp *types.GfSpGetGlobalVirtualGroupByGvgIDResponse, err error) {
	var (
		gvg *model.GlobalVirtualGroup
		res *virtual_types.GlobalVirtualGroup
	)

	ctx = log.Context(ctx, req)
	log.Debugw("GfSpGetGlobalVirtualGroupByGvgID", "gvg-id", req.GvgId)
	gvg, err = r.baseApp.GfBsDB().GetGlobalVirtualGroupByGvgID(req.GvgId)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get global virtual group by gvg id", "error", err)
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

	resp = &types.GfSpGetGlobalVirtualGroupByGvgIDResponse{GlobalVirtualGroup: res}
	log.CtxInfow(ctx, "succeed to get global virtual group by gvg id", "request", req, "response", resp)
	return resp, nil
}

// GfSpGetVirtualGroupFamily get virtual group families by vgf id
func (r *MetadataModular) GfSpGetVirtualGroupFamily(ctx context.Context, req *types.GfSpGetVirtualGroupFamilyRequest) (resp *types.GfSpGetVirtualGroupFamilyResponse, err error) {
	var (
		family *model.GlobalVirtualGroupFamily
		res    *virtual_types.GlobalVirtualGroupFamily
	)

	ctx = log.Context(ctx, req)
	log.Debugw("GfSpGetVirtualGroupFamily", "vgf-id", req.VgfId)
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
	log.CtxInfow(ctx, "succeed to get virtual group families by vgf id", "request", req, "response", resp)
	return resp, nil
}

// GfSpGetGlobalVirtualGroup get global virtual group by lvg id and bucket id
func (r *MetadataModular) GfSpGetGlobalVirtualGroup(ctx context.Context, req *types.GfSpGetGlobalVirtualGroupRequest) (resp *types.GfSpGetGlobalVirtualGroupResponse, err error) {
	var (
		gvg *model.GlobalVirtualGroup
		res *virtual_types.GlobalVirtualGroup
	)

	ctx = log.Context(ctx, req)
	log.Debugw("GfSpGetGlobalVirtualGroup", "lvg-id", req.LvgId, "bucket-id", req.BucketId)
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
	log.CtxInfow(ctx, "succeed to get global virtual group by lvg id and bucket id", "request", req, "response", resp)
	return resp, nil
}

// GfSpListMigrateBucketEvents list migrate bucket events
func (r *MetadataModular) GfSpListMigrateBucketEvents(ctx context.Context, req *types.GfSpListMigrateBucketEventsRequest) (resp *types.GfSpListMigrateBucketEventsResponse, err error) {
	var (
		events            []*model.EventMigrationBucket
		completeEvents    []*model.EventCompleteMigrationBucket
		cancelEvents      []*model.EventCancelMigrationBucket
		rejectEvents      []*model.EventRejectMigrateBucket
		spEvent           *storage_types.EventMigrationBucket
		spCompleteEvent   *storage_types.EventCompleteMigrationBucket
		spCancelEvent     *storage_types.EventCancelMigrationBucket
		spRejectEvent     *storage_types.EventRejectMigrateBucket
		eventsMap         map[common.Hash]*model.EventMigrationBucket
		completeEventsMap map[common.Hash]*model.EventCompleteMigrationBucket
		cancelEventsMap   map[common.Hash]*model.EventCancelMigrationBucket
		rejectEventsMap   map[common.Hash]*model.EventRejectMigrateBucket
		res               []*types.ListMigrateBucketEvents
		filters           []func(*gorm.DB) *gorm.DB
		latestBlock       uint64
	)

	ctx = log.Context(ctx, req)
	latestBlock, err = r.baseApp.GfBsDB().GetLatestBlockNumber()
	if err != nil {
		log.CtxErrorw(ctx, "failed to list migrate bucket events", "error", err)
		return nil, err
	}
	if latestBlock < req.BlockId {
		log.CtxError(ctx, "failed to list migrate bucket events due to request block id exceed current block syncer block height")
		return nil, ErrExceedBlockHeight
	}
	log.Debugw("GfSpListMigrateBucketEvents", "sp-id", req.SpId, "block-id", req.BlockId)
	filters = append(filters, model.CreateAtFilter(int64(req.BlockId)))
	events, completeEvents, cancelEvents, rejectEvents, err = r.baseApp.GfBsDB().ListMigrateBucketEvents(req.SpId, filters...)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list migrate bucket events", "error", err)
		return nil, err
	}

	eventsMap = make(map[common.Hash]*model.EventMigrationBucket)
	for _, e := range events {
		eventsMap[e.BucketID] = e
	}

	completeEventsMap = make(map[common.Hash]*model.EventCompleteMigrationBucket)
	for _, e := range completeEvents {
		completeEventsMap[e.BucketID] = e
	}

	cancelEventsMap = make(map[common.Hash]*model.EventCancelMigrationBucket)
	for _, e := range cancelEvents {
		cancelEventsMap[e.BucketID] = e
	}

	rejectEventsMap = make(map[common.Hash]*model.EventRejectMigrateBucket)
	for _, e := range rejectEvents {
		rejectEventsMap[e.BucketID] = e
	}

	res = make([]*types.ListMigrateBucketEvents, 0)
	for _, event := range eventsMap {
		complete := completeEventsMap[event.BucketID]
		cancel := cancelEventsMap[event.BucketID]
		reject := rejectEventsMap[event.BucketID]
		spCompleteEvent = nil
		spCancelEvent = nil
		spRejectEvent = nil
		spEvent = &storage_types.EventMigrationBucket{
			Operator:       event.Operator.String(),
			BucketName:     event.BucketName,
			BucketId:       math.NewUintFromBigInt(event.BucketID.Big()),
			DstPrimarySpId: event.DstPrimarySpId,
		}
		if complete != nil && complete.CreateAt >= event.CreateAt {
			spCompleteEvent = &storage_types.EventCompleteMigrationBucket{
				Operator:                   complete.Operator.String(),
				BucketName:                 complete.BucketName,
				BucketId:                   math.NewUintFromBigInt(complete.BucketID.Big()),
				GlobalVirtualGroupFamilyId: complete.GlobalVirtualGroupFamilyId,
			}
		}
		if cancel != nil && cancel.CreateAt >= event.CreateAt && (complete == nil || cancel.CreateAt > complete.CreateAt) {
			spCancelEvent = &storage_types.EventCancelMigrationBucket{
				Operator:   cancel.Operator.String(),
				BucketName: cancel.BucketName,
				BucketId:   math.NewUintFromBigInt(cancel.BucketID.Big()),
			}
		}
		if reject != nil && reject.CreateAt >= event.CreateAt && (complete == nil || reject.CreateAt > complete.CreateAt) {
			spRejectEvent = &storage_types.EventRejectMigrateBucket{
				Operator:   reject.Operator.String(),
				BucketName: reject.BucketName,
				BucketId:   math.NewUintFromBigInt(reject.BucketID.Big()),
			}
		}
		if spCompleteEvent == nil {
			res = append(res, &types.ListMigrateBucketEvents{
				Event:         spEvent,
				RejectEvent:   spRejectEvent,
				CompleteEvent: spCompleteEvent,
				CancelEvent:   spCancelEvent,
			})
		}
	}

	resp = &types.GfSpListMigrateBucketEventsResponse{Events: res}
	log.CtxInfow(ctx, "succeed to list migrate bucket events", "request", req, "response", resp)
	return resp, nil
}

// GfSpListCompleteMigrationBucketEvents list migrate bucket events, sp_id should be src sp id
func (r *MetadataModular) GfSpListCompleteMigrationBucketEvents(ctx context.Context, req *types.GfSpListCompleteMigrationBucketEventsRequest) (resp *types.GfSpListCompleteMigrationBucketEventsResponse, err error) {
	var (
		completeEvents    []*model.EventCompleteMigrationBucket
		spCompleteEvent   *storage_types.EventCompleteMigrationBucket
		completeEventsMap map[common.Hash]*model.EventCompleteMigrationBucket
		res               []*storage_types.EventCompleteMigrationBucket
		filters           []func(*gorm.DB) *gorm.DB
		latestBlock       uint64
	)

	ctx = log.Context(ctx, req)
	if latestBlock, err = r.baseApp.GfBsDB().GetLatestBlockNumber(); err != nil {
		log.CtxErrorw(ctx, "failed to list migrate bucket events", "error", err)
		return nil, err
	}
	if latestBlock < req.BlockId {
		//log.CtxError(ctx, "failed to list migrate bucket events due to request block id exceed current block syncer block height")
		return nil, ErrExceedBlockHeight
	}
	log.Debugw("GfSpListCompleteMigrationBucketEvents", "src_sp_id", req.SrcSpId, "block_id", req.BlockId)
	filters = append(filters, model.CreateAtEqualFilter(int64(req.BlockId)))
	if completeEvents, err = r.baseApp.GfBsDB().ListCompleteMigrationBucket(req.SrcSpId, filters...); err != nil {
		log.CtxErrorw(ctx, "failed to list complete migrate bucket events", "error", err)
		return nil, err
	}

	completeEventsMap = make(map[common.Hash]*model.EventCompleteMigrationBucket)
	for _, e := range completeEvents {
		completeEventsMap[e.BucketID] = e
	}

	res = make([]*storage_types.EventCompleteMigrationBucket, 0)

	for _, e := range completeEvents {
		spCompleteEvent = &storage_types.EventCompleteMigrationBucket{
			Operator:                   e.Operator.String(),
			BucketName:                 e.BucketName,
			BucketId:                   math.NewUintFromBigInt(e.BucketID.Big()),
			GlobalVirtualGroupFamilyId: e.GlobalVirtualGroupFamilyId,
		}
		res = append(res, spCompleteEvent)
	}

	resp = &types.GfSpListCompleteMigrationBucketEventsResponse{CompleteEvents: res}
	log.CtxInfow(ctx, "succeed to list complete migrate bucket events", "request", req, "response", resp)
	return resp, nil
}

// GfSpListSwapOutEvents list swap out events
func (r *MetadataModular) GfSpListSwapOutEvents(ctx context.Context, req *types.GfSpListSwapOutEventsRequest) (resp *types.GfSpListSwapOutEventsResponse, err error) {
	var (
		events            []*model.EventSwapOut
		completeEvents    []*model.EventCompleteSwapOut
		cancelEvents      []*model.EventCancelSwapOut
		spEvent           *virtual_types.EventSwapOut
		spCompleteEvent   *virtual_types.EventCompleteSwapOut
		spCancelEvent     *virtual_types.EventCancelSwapOut
		eventsMap         map[string]*model.EventSwapOut
		completeEventsMap map[string]*model.EventCompleteSwapOut
		cancelEventsMap   map[string]*model.EventCancelSwapOut
		res               []*types.ListSwapOutEvents
		idx               string
		latestBlock       uint64
	)

	ctx = log.Context(ctx, req)
	latestBlock, err = r.baseApp.GfBsDB().GetLatestBlockNumber()
	if err != nil {
		log.CtxErrorw(ctx, "failed to list migrate swap out events", "error", err)
		return nil, err
	}
	if latestBlock < req.BlockId {
		log.CtxError(ctx, "failed to list migrate swap out events due to request block id exceed current block syncer block height")
		return nil, ErrExceedBlockHeight
	}
	log.Debugw("GfSpListSwapOutEvents", "sp-id", req.SpId, "block-id", req.BlockId)
	events, completeEvents, cancelEvents, err = r.baseApp.GfBsDB().ListSwapOutEvents(req.BlockId, req.SpId)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list migrate swap out events", "error", err)
		return nil, err
	}

	eventsMap = make(map[string]*model.EventSwapOut)
	for _, e := range events {
		idx = model.CreateSwapOutIdx(e.GlobalVirtualGroupFamilyId, e.StorageProviderId, e.GlobalVirtualGroupIds)
		eventsMap[idx] = e
	}

	completeEventsMap = make(map[string]*model.EventCompleteSwapOut)
	for _, e := range completeEvents {
		idx = model.CreateSwapOutIdx(e.GlobalVirtualGroupFamilyId, e.SrcStorageProviderId, e.GlobalVirtualGroupIds)
		completeEventsMap[idx] = e
	}

	cancelEventsMap = make(map[string]*model.EventCancelSwapOut)
	for _, e := range cancelEvents {
		idx = model.CreateSwapOutIdx(e.GlobalVirtualGroupFamilyId, e.StorageProviderId, e.GlobalVirtualGroupIds)
		cancelEventsMap[idx] = e
	}

	res = make([]*types.ListSwapOutEvents, 0)
	for _, event := range eventsMap {
		idx = model.CreateSwapOutIdx(event.GlobalVirtualGroupFamilyId, event.StorageProviderId, event.GlobalVirtualGroupIds)
		complete := completeEventsMap[idx]
		cancel := cancelEventsMap[idx]
		spCompleteEvent = nil
		spCancelEvent = nil
		spEvent = &virtual_types.EventSwapOut{
			StorageProviderId:          event.StorageProviderId,
			GlobalVirtualGroupFamilyId: event.GlobalVirtualGroupFamilyId,
			GlobalVirtualGroupIds:      event.GlobalVirtualGroupIds,
			SuccessorSpId:              event.SuccessorSpId,
		}
		if complete != nil && complete.CreateAt >= event.CreateAt {
			spCompleteEvent = &virtual_types.EventCompleteSwapOut{
				StorageProviderId:          complete.StorageProviderId,
				SrcStorageProviderId:       complete.SrcStorageProviderId,
				GlobalVirtualGroupFamilyId: complete.GlobalVirtualGroupFamilyId,
				GlobalVirtualGroupIds:      complete.GlobalVirtualGroupIds,
			}
		}
		if cancel != nil && cancel.CreateAt >= event.CreateAt && complete == nil {
			spCancelEvent = &virtual_types.EventCancelSwapOut{
				StorageProviderId:          cancel.StorageProviderId,
				GlobalVirtualGroupFamilyId: cancel.GlobalVirtualGroupFamilyId,
				GlobalVirtualGroupIds:      cancel.GlobalVirtualGroupIds,
				SuccessorSpId:              cancel.SuccessorSpId,
			}
		}
		res = append(res, &types.ListSwapOutEvents{
			Events:         spEvent,
			CompleteEvents: spCompleteEvent,
			CancelEvents:   spCancelEvent,
		})
	}

	resp = &types.GfSpListSwapOutEventsResponse{Events: res}
	log.CtxInfow(ctx, "succeed to list migrate swap out events", "request", req, "response", resp)
	return resp, nil
}

// GfSpListGlobalVirtualGroupsBySecondarySP list global virtual group by secondary sp id
func (r *MetadataModular) GfSpListGlobalVirtualGroupsBySecondarySP(ctx context.Context, req *types.GfSpListGlobalVirtualGroupsBySecondarySPRequest) (resp *types.GfSpListGlobalVirtualGroupsBySecondarySPResponse, err error) {
	var (
		groups []*model.GlobalVirtualGroup
		res    []*virtual_types.GlobalVirtualGroup
	)

	ctx = log.Context(ctx, req)
	log.Debugw("GfSpListGlobalVirtualGroupsBySecondarySP", "sp-id", req.SpId)
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
	log.CtxInfow(ctx, "succeed to get global virtual group by secondary sp id", "request", req, "response", resp)
	return resp, nil
}

// GfSpListGlobalVirtualGroupsByBucket list global virtual group by bucket id
func (r *MetadataModular) GfSpListGlobalVirtualGroupsByBucket(ctx context.Context, req *types.GfSpListGlobalVirtualGroupsByBucketRequest) (resp *types.GfSpListGlobalVirtualGroupsByBucketResponse, err error) {
	var (
		groups []*model.GlobalVirtualGroup
		res    []*virtual_types.GlobalVirtualGroup
	)

	ctx = log.Context(ctx, req)
	log.Debugw("GfSpListGlobalVirtualGroupsByBucket", "bucket-id", req.BucketId)
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
	log.CtxInfow(ctx, "succeed to list global virtual group by bucket id", "request", req, "response", resp)
	return resp, nil
}

// GfSpListSpExitEvents list migrate sp exit events
func (r *MetadataModular) GfSpListSpExitEvents(ctx context.Context, req *types.GfSpListSpExitEventsRequest) (resp *types.GfSpListSpExitEventsResponse, err error) {
	var (
		event           *model.EventStorageProviderExit
		completeEvent   *model.EventCompleteStorageProviderExit
		spEvent         *virtual_types.EventStorageProviderExit
		spCompleteEvent *virtual_types.EventCompleteStorageProviderExit
		latestBlock     uint64
	)
	ctx = log.Context(ctx, req)
	latestBlock, err = r.baseApp.GfBsDB().GetLatestBlockNumber()
	if err != nil {
		log.CtxErrorw(ctx, "failed to list sp exit events", "error", err)
		return nil, err
	}
	if latestBlock < req.BlockId {
		log.CtxError(ctx, "failed to list sp exit events due to request block id exceed current block syncer block height")
		return nil, ErrExceedBlockHeight
	}

	event, completeEvent, err = r.baseApp.GfBsDB().ListSpExitEvents(req.BlockId, req.SpId)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list sp exit events", "error", err)
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
	log.CtxInfow(ctx, "succeed to list sp exit events", "request", req, "response", resp)
	return resp, nil
}

// GfSpGetSPMigratingBucketNumber get the latest active migrating bucket by specific sp
func (r *MetadataModular) GfSpGetSPMigratingBucketNumber(ctx context.Context, req *types.GfSpGetSPMigratingBucketNumberRequest) (resp *types.GfSpGetSPMigratingBucketNumberResponse, err error) {
	var (
		events            []*model.EventMigrationBucket
		completeEvents    []*model.EventCompleteMigrationBucket
		cancelEvents      []*model.EventCancelMigrationBucket
		rejectEvents      []*model.EventRejectMigrateBucket
		eventsMap         map[common.Hash]*model.EventMigrationBucket
		completeEventsMap map[common.Hash]*model.EventCompleteMigrationBucket
		cancelEventsMap   map[common.Hash]*model.EventCancelMigrationBucket
		rejectEventsMap   map[common.Hash]*model.EventRejectMigrateBucket
		cancelFlag        bool
		completeFlag      bool
		rejectFlag        bool
		count             uint64
	)

	ctx = log.Context(ctx, req)

	events, completeEvents, cancelEvents, rejectEvents, err = r.baseApp.GfBsDB().ListMigrateBucketEvents(req.SpId)
	if err != nil {
		log.CtxErrorw(ctx, "failed to list migrate bucket events", "error", err)
		return nil, err
	}

	eventsMap = make(map[common.Hash]*model.EventMigrationBucket)
	for _, e := range events {
		eventsMap[e.BucketID] = e
	}

	completeEventsMap = make(map[common.Hash]*model.EventCompleteMigrationBucket)
	for _, e := range completeEvents {
		completeEventsMap[e.BucketID] = e
	}

	cancelEventsMap = make(map[common.Hash]*model.EventCancelMigrationBucket)
	for _, e := range cancelEvents {
		cancelEventsMap[e.BucketID] = e
	}

	rejectEventsMap = make(map[common.Hash]*model.EventRejectMigrateBucket)
	for _, e := range rejectEvents {
		rejectEventsMap[e.BucketID] = e
	}

	for _, event := range eventsMap {
		completeFlag = false
		cancelFlag = false
		rejectFlag = false
		complete := completeEventsMap[event.BucketID]
		cancel := cancelEventsMap[event.BucketID]
		reject := rejectEventsMap[event.BucketID]
		if complete != nil && complete.CreateAt >= event.CreateAt {
			completeFlag = true
		}
		if cancel != nil && cancel.CreateAt >= event.CreateAt && complete == nil {
			cancelFlag = true
		}
		if reject != nil && reject.CreateAt >= event.CreateAt && complete == nil {
			rejectFlag = true
		}
		if !completeFlag && !cancelFlag && !rejectFlag && event != nil {
			count++
		}
	}

	resp = &types.GfSpGetSPMigratingBucketNumberResponse{Count: count}
	log.CtxInfow(ctx, "succeed to get the latest active migrating bucket by specific sp", "request", req, "response", resp)
	return resp, nil
}

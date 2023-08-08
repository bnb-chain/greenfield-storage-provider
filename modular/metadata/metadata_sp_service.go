package metadata

import (
	"context"
	"errors"

	"cosmossdk.io/math"
	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
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

func (r *MetadataModular) GfSpGetStatus(ctx context.Context, req *types.GfSpGetStatusRequest) (resp *types.GfSpGetStatusResponse, err error) {
	var (
		res                     *types.Status
		blockSyncerVersion      *types.BlockSyncerInfo
		storageProviderInfo     *types.StorageProviderInfo
		chainInfo               *types.ChainInfo
		version                 string
		spVersion               *bsdb.SpVersion
		epoch                   *bsdb.Epoch
		bsUpdateTime            int64
		bsBlockHash             string
		bsBlockHeight           int64
		defaultCharacterSetName string
		defaultCollationName    string
	)

	ctx = log.Context(ctx, req)
	version, err = r.baseApp.GfBsDB().GetMysqlVersion()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get mysql version info", "error", err)
	}

	epoch, err = r.baseApp.GfBsDB().GetEpoch()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get epoch info", "error", err)
	}

	if epoch != nil {
		bsUpdateTime = epoch.UpdateTime
		bsBlockHash = epoch.BlockHash.String()
		bsBlockHeight = epoch.BlockHeight
	}

	defaultCharacterSetName, err = r.baseApp.GfBsDB().GetDefaultCharacterSet()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get default character set name", "error", err)
	}

	defaultCollationName, err = r.baseApp.GfBsDB().GetDefaultCollationName()
	if err != nil {
		log.CtxErrorw(ctx, "failed to get default collation name", "error", err)
	}

	spVersion = r.baseApp.GfBsDB().GetSpVersion()

	blockSyncerVersion = &types.BlockSyncerInfo{
		BsUpdateTime:              bsUpdateTime,
		BsBlockHash:               bsBlockHash,
		BsBlockHeight:             bsBlockHeight,
		BsMysqlVersion:            version,
		BsDefaultCharacterSetName: defaultCharacterSetName,
		BsDefaultCollationName:    defaultCollationName,
		BsModules:                 BsModules,
		BsEnableDualDb:            BsEnableDualDB,
		BsWorkers:                 uint32(BsWorkers),
	}

	storageProviderInfo = &types.StorageProviderInfo{
		SpOperatorAddress: SpOperatorAddress,
		SpDomainName:      GatewayDomainName,
		SpGoVersion:       spVersion.SpGoVersion,
		SpCodeCommit:      spVersion.SpCodeCommit,
		SpCodeVersion:     spVersion.SpCodeVersion,
		SpOperatingSystem: spVersion.SpOperatingSystem,
		SpArchitecture:    spVersion.SpArchitecture,
	}

	chainInfo = &types.ChainInfo{
		ChainId:      ChainID,
		ChainAddress: ChainAddress,
	}

	res = &types.Status{
		BlockSyncerInfo:     blockSyncerVersion,
		StorageProviderInfo: storageProviderInfo,
		ChainInfo:           chainInfo,
	}

	resp = &types.GfSpGetStatusResponse{Status: res}
	log.CtxInfow(ctx, "succeed to get status")
	return resp, nil
}

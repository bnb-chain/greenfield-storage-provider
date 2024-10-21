package consensus

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	paymenttypes "github.com/evmos/evmos/v12/x/payment/types"
	sptypes "github.com/evmos/evmos/v12/x/sp/types"
	storagetypes "github.com/evmos/evmos/v12/x/storage/types"
	virtualgrouptypes "github.com/evmos/evmos/v12/x/virtualgroup/types"
)

// Consensus is the interface to query mechain consensus data. the consensus
// data can come from validator, full-node, or other off-chain data service
//
//go:generate mockgen -source=./consensus.go -destination=./consensus_mock.go -package=consensus
type Consensus interface {
	// CurrentHeight returns the current mechain height - 1,
	CurrentHeight(ctx context.Context) (uint64, error)
	// HasAccount returns an indicator whether the account has been created.
	HasAccount(ctx context.Context, account string) (bool, error)
	// ListSPs returns all SP info.
	ListSPs(ctx context.Context) ([]*sptypes.StorageProvider, error)
	// QuerySP returns the sp info by operator address.
	QuerySP(context.Context, string) (*sptypes.StorageProvider, error)
	// QuerySPByID returns the sp info by sp id.
	QuerySPByID(context.Context, uint32) (*sptypes.StorageProvider, error)
	// QuerySPFreeQuota returns the sp free quota by operator address.
	QuerySPFreeQuota(context.Context, string) (uint64, error)
	// QuerySPPrice returns the sp price info
	QuerySPPrice(ctx context.Context, operatorAddress string) (sptypes.SpStoragePrice, error)
	// ListBondedValidators returns all bonded validators info.
	ListBondedValidators(ctx context.Context) ([]stakingtypes.Validator, error)
	// ListVirtualGroupFamilies return all virtual group family which primary sp is spID.
	ListVirtualGroupFamilies(ctx context.Context, spID uint32) ([]*virtualgrouptypes.GlobalVirtualGroupFamily, error)
	// QueryVirtualGroupFamily return the virtual group family info.
	QueryVirtualGroupFamily(ctx context.Context, vgfID uint32) (*virtualgrouptypes.GlobalVirtualGroupFamily, error)
	// QueryGlobalVirtualGroup returns the global virtual group info.
	QueryGlobalVirtualGroup(ctx context.Context, gvgID uint32) (*virtualgrouptypes.GlobalVirtualGroup, error)
	// ListGlobalVirtualGroupsByFamilyID returns gvg list by family.
	ListGlobalVirtualGroupsByFamilyID(ctx context.Context, vgfID uint32) ([]*virtualgrouptypes.GlobalVirtualGroup, error)
	// AvailableGlobalVirtualGroupFamilies submits a list global virtual group families Id to chain and return the filtered families which are able to serve create bucket request.
	AvailableGlobalVirtualGroupFamilies(ctx context.Context, globalVirtualGroupFamiliesIDs []uint32) ([]uint32, error)
	// QueryVirtualGroupParams returns the virtual group params.
	QueryVirtualGroupParams(ctx context.Context) (*virtualgrouptypes.Params, error)
	// QueryStorageParams returns the storage params.
	QueryStorageParams(ctx context.Context) (params *storagetypes.Params, err error)
	// QueryStorageParamsByTimestamp returns the storage params by block create time.
	QueryStorageParamsByTimestamp(ctx context.Context, timestamp int64) (params *storagetypes.Params, err error)
	// QueryBucketInfo returns the bucket info by bucket name.
	QueryBucketInfo(ctx context.Context, bucket string) (*storagetypes.BucketInfo, error)
	// QueryBucketExtraInfo returns the bucket extra info by bucket name.
	QueryBucketExtraInfo(ctx context.Context, bucket string) (bucketInfo *storagetypes.BucketExtraInfo, err error)
	// QueryBucketInfoById returns the bucket info by bucket id.
	QueryBucketInfoById(ctx context.Context, bucketId uint64) (bucketInfo *storagetypes.BucketInfo, err error)
	// QueryObjectInfo returns the object info by bucket and object name.
	QueryObjectInfo(ctx context.Context, bucket, object string) (*storagetypes.ObjectInfo, error)
	// QueryObjectInfoByID returns the object info by object ID.
	QueryObjectInfoByID(ctx context.Context, objectID string) (*storagetypes.ObjectInfo, error)
	// QueryBucketInfoAndObjectInfo returns the bucket and object info by bucket and object name.
	QueryBucketInfoAndObjectInfo(ctx context.Context, bucket, object string) (*storagetypes.BucketInfo, *storagetypes.ObjectInfo, error)
	// QueryPaymentStreamRecord returns the account payment status.
	QueryPaymentStreamRecord(ctx context.Context, account string) (*paymenttypes.StreamRecord, error)
	// VerifyGetObjectPermission returns an indicator whether the account has permission to get object.
	VerifyGetObjectPermission(ctx context.Context, account, bucket, object string) (bool, error)
	// VerifyPutObjectPermission returns an indicator whether the account has permission to put object.
	VerifyPutObjectPermission(ctx context.Context, account, bucket, object string) (bool, error)
	// ListenObjectSeal returns an indicator whether the object is successfully sealed before timeOutHeight.
	ListenObjectSeal(ctx context.Context, objectID uint64, timeOutHeight int) (bool, error)
	// ListenRejectUnSealObject returns an indication of the object is rejected.
	ListenRejectUnSealObject(ctx context.Context, objectID uint64, timeoutHeight int) (bool, error)
	// ConfirmTransaction is used to confirm whether the transaction is on the chain.
	ConfirmTransaction(ctx context.Context, txHash string) (*sdk.TxResponse, error)
	// WaitForNextBlock is used to chain generate a new block.
	WaitForNextBlock(ctx context.Context) error
	// QuerySwapInInfo is used to query the onchain swapIn info
	QuerySwapInInfo(ctx context.Context, familyID, gvgID uint32) (*virtualgrouptypes.SwapInInfo, error)
	// VerifyUpdateObjectPermission returns an indicator whether the account has permission to update object.
	VerifyUpdateObjectPermission(ctx context.Context, account, bucket, object string) (bool, error)
	// QueryShadowObjectInfo is used to query the onchain ShadowObjectInfo
	QueryShadowObjectInfo(ctx context.Context, bucket, object string) (*storagetypes.ShadowObjectInfo, error)
	// Close the Consensus interface.
	Close() error
}

var _ Consensus = (*NullConsensus)(nil)

type NullConsensus struct{}

func (*NullConsensus) CurrentHeight(context.Context) (uint64, error) { return 0, nil }
func (*NullConsensus) HasAccount(context.Context, string) (bool, error) {
	return false, nil
}
func (*NullConsensus) ListSPs(context.Context) ([]*sptypes.StorageProvider, error) {
	return nil, nil
}
func (*NullConsensus) QuerySP(context.Context, string) (*sptypes.StorageProvider, error) {
	return nil, nil
}
func (*NullConsensus) QuerySPFreeQuota(context.Context, string) (uint64, error) {
	return 0, nil
}
func (*NullConsensus) QuerySPPrice(context.Context, string) (sptypes.SpStoragePrice, error) {
	return sptypes.SpStoragePrice{}, nil
}
func (*NullConsensus) QuerySPByID(context.Context, uint32) (*sptypes.StorageProvider, error) {
	return nil, nil
}
func (*NullConsensus) ListBondedValidators(context.Context) ([]stakingtypes.Validator, error) {
	return nil, nil
}
func (*NullConsensus) ListVirtualGroupFamilies(context.Context, uint32) ([]*virtualgrouptypes.GlobalVirtualGroupFamily, error) {
	return nil, nil
}
func (*NullConsensus) ListGlobalVirtualGroupsByFamilyID(context.Context, uint32) ([]*virtualgrouptypes.GlobalVirtualGroup, error) {
	return nil, nil
}
func (*NullConsensus) AvailableGlobalVirtualGroupFamilies(context.Context, []uint32) ([]uint32, error) {
	return nil, nil
}
func (*NullConsensus) QueryVirtualGroupFamily(context.Context, uint32) (*virtualgrouptypes.GlobalVirtualGroupFamily, error) {
	return nil, nil
}
func (*NullConsensus) QueryGlobalVirtualGroup(context.Context, uint32) (*virtualgrouptypes.GlobalVirtualGroup, error) {
	return nil, nil
}
func (*NullConsensus) QueryVirtualGroupParams(context.Context) (*virtualgrouptypes.Params, error) {
	return nil, nil
}
func (*NullConsensus) QueryStorageParams(context.Context) (*storagetypes.Params, error) {
	return nil, nil
}
func (*NullConsensus) QueryStorageParamsByTimestamp(context.Context, int64) (*storagetypes.Params, error) {
	return nil, nil
}
func (*NullConsensus) QueryBucketInfo(context.Context, string) (*storagetypes.BucketInfo, error) {
	return nil, nil
}
func (*NullConsensus) QueryBucketExtraInfo(context.Context, string) (*storagetypes.BucketExtraInfo, error) {
	return nil, nil
}
func (*NullConsensus) QueryBucketInfoById(context.Context, uint64) (*storagetypes.BucketInfo, error) {
	return nil, nil
}
func (*NullConsensus) QueryObjectInfo(context.Context, string, string) (*storagetypes.ObjectInfo, error) {
	return nil, nil
}
func (*NullConsensus) QueryObjectInfoByID(context.Context, string) (*storagetypes.ObjectInfo, error) {
	return nil, nil
}
func (*NullConsensus) QueryBucketInfoAndObjectInfo(context.Context, string, string) (*storagetypes.BucketInfo, *storagetypes.ObjectInfo, error) {
	return nil, nil, nil
}
func (*NullConsensus) QueryPaymentStreamRecord(context.Context, string) (*paymenttypes.StreamRecord, error) {
	return nil, nil
}
func (*NullConsensus) VerifyGetObjectPermission(context.Context, string, string, string) (bool, error) {
	return false, nil
}
func (*NullConsensus) VerifyPutObjectPermission(context.Context, string, string, string) (bool, error) {
	return false, nil
}
func (*NullConsensus) ListenObjectSeal(context.Context, uint64, int) (bool, error) {
	return false, nil
}
func (*NullConsensus) ListenRejectUnSealObject(context.Context, uint64, int) (bool, error) {
	return false, nil
}
func (*NullConsensus) ConfirmTransaction(context.Context, string) (*sdk.TxResponse, error) {
	return nil, nil
}
func (*NullConsensus) WaitForNextBlock(context.Context) error {
	return nil
}
func (c *NullConsensus) QuerySwapInInfo(ctx context.Context, familyID, gvgID uint32) (*virtualgrouptypes.SwapInInfo, error) {
	return nil, nil
}
func (c *NullConsensus) VerifyUpdateObjectPermission(ctx context.Context, account, bucket, object string) (bool, error) {
	return false, nil
}
func (c *NullConsensus) QueryShadowObjectInfo(ctx context.Context, bucket, object string) (*storagetypes.ShadowObjectInfo, error) {
	return nil, nil
}
func (*NullConsensus) Close() error { return nil }

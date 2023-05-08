package consensus

import (
	"context"

	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

type Consensus interface {
	CurrentHeight(ctx context.Context) (uint64, error)
	HasAccount(ctx context.Context, account string) (bool, error)
	QuerySPInfo(ctx context.Context) ([]*sptypes.StorageProvider, error)
	QueryStorageParams(ctx context.Context) (params *storagetypes.Params, err error)
	QueryBucketInfo(ctx context.Context, bucket string) (*storagetypes.BucketInfo, error)
	QueryObjectInfo(ctx context.Context, bucket, object string) (*storagetypes.ObjectInfo, error)
	QueryObjectInfoByID(ctx context.Context, objectID string) (*storagetypes.ObjectInfo, error)
	QueryBucketInfoAndObjectInfo(ctx context.Context, bucket, object string) (*storagetypes.BucketInfo, *storagetypes.ObjectInfo, error)
	QueryPaymentStreamRecord(ctx context.Context, account string) (*paymenttypes.StreamRecord, error)
	VerifyGetObjectPermission(ctx context.Context, account, bucket, object string) (bool, error)
	VerifyPutObjectPermission(ctx context.Context, account, bucket, object string) (bool, error)
	ListenObjectSeal(ctx context.Context, objectID uint64, timeOutHeight int) (bool, error)
	Close() error
}

var _ Consensus = (*NullConsensus)(nil)

type NullConsensus struct{}

func (*NullConsensus) CurrentHeight(context.Context) (uint64, error) { return 0, nil }
func (*NullConsensus) HasAccount(context.Context, string) (bool, error) {
	return false, nil
}
func (*NullConsensus) QuerySPInfo(context.Context) ([]*sptypes.StorageProvider, error) {
	return nil, nil
}
func (*NullConsensus) QueryStorageParams(context.Context) (*storagetypes.Params, error) {
	return nil, nil
}
func (*NullConsensus) QueryBucketInfo(context.Context, string) (*storagetypes.BucketInfo, error) {
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
func (*NullConsensus) Close() error { return nil }

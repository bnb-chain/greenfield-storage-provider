package task

import (
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var _ Task = (*NullTask)(nil)
var _ ApprovalTask = (*NullTask)(nil)
var _ ApprovalCreateBucketTask = (*NullTask)(nil)
var _ ApprovalCreateObjectTask = (*NullTask)(nil)
var _ ApprovalReplicatePieceTask = (*NullTask)(nil)
var _ ObjectTask = (*NullTask)(nil)
var _ UploadObjectTask = (*NullTask)(nil)
var _ ReplicatePieceTask = (*NullTask)(nil)
var _ SealObjectTask = (*NullTask)(nil)
var _ ReceivePieceTask = (*NullTask)(nil)
var _ DownloadObjectTask = (*NullTask)(nil)
var _ ChallengePieceTask = (*NullTask)(nil)
var _ GCTask = (*NullTask)(nil)
var _ GCZombiePieceTask = (*NullTask)(nil)
var _ GCMetaTask = (*NullTask)(nil)

type NullTask struct{}

func (*NullTask) Key() TKey                                { return "" }
func (*NullTask) Type() TType                              { return 0 }
func (*NullTask) GetAddress() string                       { return "" }
func (*NullTask) SetAddress(string)                        { return }
func (*NullTask) GetCreateTime() int64                     { return 0 }
func (*NullTask) SetCreateTime(int64)                      { return }
func (*NullTask) GetUpdateTime() int64                     { return 0 }
func (*NullTask) SetUpdateTime(int64)                      { return }
func (*NullTask) GetTimeout() int64                        { return 0 }
func (*NullTask) SetTimeout(int64)                         { return }
func (*NullTask) Expired() bool                            { return false }
func (*NullTask) ExceedTimeout() bool                      { return false }
func (*NullTask) GetPriority() TPriority                   { return 0 }
func (*NullTask) SetPriority(TPriority)                    { return }
func (*NullTask) SetRetry(int)                             { return }
func (*NullTask) IncRetry()                                { return }
func (*NullTask) ExceedRetry() bool                        { return false }
func (*NullTask) GetRetry() int64                          { return 0 }
func (*NullTask) GetMaxRetry() int64                       { return 0 }
func (*NullTask) SetMaxRetry(int64)                        { return }
func (*NullTask) EstimateLimit() rcmgr.Limit               { return nil }
func (*NullTask) Error() error                             { return nil }
func (*NullTask) SetError(error)                           { return }
func (*NullTask) GetExpiredHeight() uint64                 { return 0 }
func (*NullTask) SetExpiredHeight(uint64)                  { return }
func (*NullTask) GetObjectInfo() *storagetypes.ObjectInfo  { return nil }
func (*NullTask) SetObjectInfo(*storagetypes.ObjectInfo)   { return }
func (*NullTask) GetStorageParams() *storagetypes.Params   { return nil }
func (*NullTask) SetStorageParams(*storagetypes.Params)    { return }
func (*NullTask) GetGCZombiePieceStatus() (uint64, uint64) { return 0, 0 }
func (*NullTask) SetGCZombiePieceStatus(uint64, uint64)    { return }
func (*NullTask) GetGCMetaStatus() (uint64, uint64)        { return 0, 0 }
func (*NullTask) SetGCMetaStatus(uint64, uint64)           { return }
func (*NullTask) InitApprovalCreateBucketTask(*storagetypes.MsgCreateBucket, TPriority) {
	return
}
func (*NullTask) GetCreateBucketInfo() *storagetypes.MsgCreateBucket { return nil }
func (*NullTask) SetCreateBucketInfo(*storagetypes.MsgCreateBucket)  { return }
func (*NullTask) InitApprovalCreateObjectTask(*storagetypes.MsgCreateObject, TPriority) {
	return
}
func (*NullTask) GetCreateObjectInfo() *storagetypes.MsgCreateObject { return nil }
func (*NullTask) SetCreateObjectInfo(*storagetypes.MsgCreateObject)  { return }
func (*NullTask) InitApprovalReplicatePieceTask(*storagetypes.ObjectInfo, *storagetypes.Params, TPriority, string) {
	return
}
func (*NullTask) GetAskSpOperatorAddress() string                                     { return "" }
func (*NullTask) SetAskSpOperatorAddress(string)                                      { return }
func (*NullTask) GetAskSignature() []byte                                             { return nil }
func (*NullTask) SetAskSignature([]byte)                                              { return }
func (*NullTask) GetApprovedSpOperatorAddress() string                                { return "" }
func (*NullTask) SetApprovedSpOperatorAddress(string)                                 { return }
func (*NullTask) GetApprovedSignature() []byte                                        { return nil }
func (*NullTask) SetApprovedSignature([]byte)                                         { return }
func (*NullTask) GetApprovedSpEndpoint() string                                       { return "" }
func (*NullTask) SetApprovedSpEndpoint(string)                                        { return }
func (*NullTask) GetApprovedSpApprovalAddress() string                                { return "" }
func (*NullTask) SetApprovedSpApprovalAddress(string)                                 { return }
func (*NullTask) InitUploadObjectTask(*storagetypes.ObjectInfo, *storagetypes.Params) { return }
func (*NullTask) InitReplicatePieceTask(*storagetypes.ObjectInfo, *storagetypes.Params, TPriority, int64, int64) {
	return
}
func (*NullTask) GetSealed() bool                 { return false }
func (*NullTask) SetSealed(bool)                  { return }
func (*NullTask) GetSecondarySignature() [][]byte { return nil }
func (*NullTask) SetSecondarySignature([][]byte)  { return }
func (*NullTask) InitSealObjectTask(*storagetypes.ObjectInfo, *storagetypes.Params, TPriority, [][]byte, int64, int64) {
	return
}
func (*NullTask) InitReceivePieceTask(*storagetypes.ObjectInfo, *storagetypes.Params, TPriority, uint32, int32, int64) {
	return
}
func (*NullTask) GetReplicateIdx() uint32  { return 0 }
func (*NullTask) SetReplicateIdx(uint32)   { return }
func (*NullTask) GetPieceIdx() int32       { return 0 }
func (*NullTask) SetPieceIdx(int32)        { return }
func (*NullTask) GetPieceSize() int64      { return 0 }
func (*NullTask) SetPieceSize(int64)       { return }
func (*NullTask) SetPieceChecksum([]byte)  { return }
func (*NullTask) GetPieceChecksum() []byte { return nil }
func (*NullTask) GetSignature() []byte     { return nil }
func (*NullTask) SetSignature([]byte)      { return }
func (*NullTask) InitDownloadObjectTask(*storagetypes.ObjectInfo, *storagetypes.Params, TPriority, int64, int64, int64, int64) {
	return
}
func (*NullTask) GetBucketInfo() *storagetypes.BucketInfo { return nil }
func (*NullTask) GetUserAddress() string                  { return "" }
func (*NullTask) GetSize() int64                          { return 0 }
func (*NullTask) GetLow() int64                           { return 0 }
func (*NullTask) GetHigh() int64                          { return 0 }
func (*NullTask) InitChallengePieceTask(*storagetypes.ObjectInfo, *storagetypes.BucketInfo, TPriority, int32, uint32, int64, int64) {
	return
}
func (*NullTask) SetBucketInfo(*storagetypes.BucketInfo) { return }
func (*NullTask) SetUserAddress(string)                  { return }
func (*NullTask) GetSegmentIdx() uint32                  { return 0 }
func (*NullTask) SetSegmentIdx(uint32)                   { return }
func (*NullTask) GetRedundancyIdx() int32                { return 0 }
func (*NullTask) SetRedundancyIdx(idx int32)             { return }
func (*NullTask) GetIntegrityHash() []byte               { return nil }
func (*NullTask) SetIntegrityHash([]byte)                { return }
func (*NullTask) GetPieceHash() [][]byte                 { return nil }
func (*NullTask) SetPieceHash([][]byte)                  { return }
func (*NullTask) GetPieceDataSize() int64                { return 0 }
func (*NullTask) SetPieceDataSize(int64)                 { return }
func (*NullTask) GetSignBytes() []byte                   { return nil }

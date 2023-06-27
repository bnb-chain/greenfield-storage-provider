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
var _ MigratePieceTask = (*NullTask)(nil)
var _ RecoveryPieceTask = (*NullTask)(nil)

type NullTask struct{}

func (*NullTask) Key() TKey                                                             { return "" }
func (*NullTask) Type() TType                                                           { return 0 }
func (*NullTask) Info() string                                                          { return "" }
func (*NullTask) GetAddress() string                                                    { return "" }
func (*NullTask) SetAddress(string)                                                     {}
func (*NullTask) GetCreateTime() int64                                                  { return 0 }
func (*NullTask) SetCreateTime(int64)                                                   {}
func (*NullTask) GetUpdateTime() int64                                                  { return 0 }
func (*NullTask) SetUpdateTime(int64)                                                   {}
func (*NullTask) GetTimeout() int64                                                     { return 0 }
func (*NullTask) SetTimeout(int64)                                                      {}
func (*NullTask) Expired() bool                                                         { return false }
func (*NullTask) ExceedTimeout() bool                                                   { return false }
func (*NullTask) GetPriority() TPriority                                                { return 0 }
func (*NullTask) SetPriority(TPriority)                                                 {}
func (*NullTask) SetRetry(int)                                                          {}
func (*NullTask) IncRetry()                                                             {}
func (*NullTask) ExceedRetry() bool                                                     { return false }
func (*NullTask) GetRetry() int64                                                       { return 0 }
func (*NullTask) GetMaxRetry() int64                                                    { return 0 }
func (*NullTask) SetMaxRetry(int64)                                                     {}
func (*NullTask) EstimateLimit() rcmgr.Limit                                            { return nil }
func (*NullTask) Error() error                                                          { return nil }
func (*NullTask) SetError(error)                                                        {}
func (*NullTask) GetExpiredHeight() uint64                                              { return 0 }
func (*NullTask) SetExpiredHeight(uint64)                                               {}
func (*NullTask) GetObjectInfo() *storagetypes.ObjectInfo                               { return nil }
func (*NullTask) SetObjectInfo(*storagetypes.ObjectInfo)                                {}
func (*NullTask) GetStorageParams() *storagetypes.Params                                { return nil }
func (*NullTask) SetStorageParams(*storagetypes.Params)                                 {}
func (*NullTask) GetGCZombiePieceStatus() (uint64, uint64)                              { return 0, 0 }
func (*NullTask) SetGCZombiePieceStatus(uint64, uint64)                                 {}
func (*NullTask) GetGCMetaStatus() (uint64, uint64)                                     { return 0, 0 }
func (*NullTask) SetGCMetaStatus(uint64, uint64)                                        {}
func (*NullTask) InitApprovalCreateBucketTask(*storagetypes.MsgCreateBucket, TPriority) {}
func (*NullTask) GetCreateBucketInfo() *storagetypes.MsgCreateBucket                    { return nil }
func (*NullTask) SetCreateBucketInfo(*storagetypes.MsgCreateBucket)                     {}
func (*NullTask) InitApprovalCreateObjectTask(*storagetypes.MsgCreateObject, TPriority) {}
func (*NullTask) GetCreateObjectInfo() *storagetypes.MsgCreateObject                    { return nil }
func (*NullTask) SetCreateObjectInfo(*storagetypes.MsgCreateObject)                     {}
func (*NullTask) InitApprovalReplicatePieceTask(*storagetypes.ObjectInfo, *storagetypes.Params, TPriority, string) {
}
func (*NullTask) GetAskSpOperatorAddress() string      { return "" }
func (*NullTask) SetAskSpOperatorAddress(string)       {}
func (*NullTask) GetAskSignature() []byte              { return nil }
func (*NullTask) SetAskSignature([]byte)               {}
func (*NullTask) GetApprovedSpOperatorAddress() string { return "" }
func (*NullTask) SetApprovedSpOperatorAddress(string)  {}
func (*NullTask) GetApprovedSignature() []byte         { return nil }
func (*NullTask) SetApprovedSignature([]byte)          {}
func (*NullTask) GetApprovedSpEndpoint() string        { return "" }
func (*NullTask) SetApprovedSpEndpoint(string)         {}
func (*NullTask) GetApprovedSpApprovalAddress() string { return "" }
func (*NullTask) SetApprovedSpApprovalAddress(string)  {}
func (*NullTask) InitUploadObjectTask(uint32, *storagetypes.ObjectInfo, *storagetypes.Params, int64) {
}
func (*NullTask) GetVirtualGroupFamilyId() uint32 {
	return 0
}

func (*NullTask) GetGlobalVirtualGroupId() uint32 {
	return 0
}

func (*NullTask) InitReplicatePieceTask(*storagetypes.ObjectInfo, *storagetypes.Params, TPriority, int64, int64) {
}
func (*NullTask) InitRecoverPieceTask(*storagetypes.ObjectInfo, *storagetypes.Params, TPriority, uint32, int32, uint64, int64, int64) {
}
func (*NullTask) GetSealed() bool                  { return false }
func (*NullTask) SetSealed(bool)                   {}
func (*NullTask) GetSecondaryAddresses() []string  { return nil }
func (*NullTask) GetSecondarySignatures() [][]byte { return nil }
func (*NullTask) SetSecondarySignatures([][]byte)  {}
func (*NullTask) SetSecondaryAddresses([]string)   {}
func (*NullTask) GetSecondaryEndpoints() []string  { return nil }
func (*NullTask) InitSealObjectTask(uint32, *storagetypes.ObjectInfo, *storagetypes.Params, TPriority, []string, [][]byte, int64, int64) {
}
func (*NullTask) InitReceivePieceTask(uint32, *storagetypes.ObjectInfo, *storagetypes.Params, TPriority, uint32, int32, int64) {
}
func (*NullTask) GetReplicateIdx() uint32  { return 0 }
func (*NullTask) SetReplicateIdx(uint32)   {}
func (*NullTask) GetPieceIdx() int32       { return 0 }
func (*NullTask) SetPieceIdx(int32)        {}
func (*NullTask) GetPieceSize() int64      { return 0 }
func (*NullTask) SetPieceSize(int64)       {}
func (*NullTask) SetPieceChecksum([]byte)  {}
func (*NullTask) GetPieceChecksum() []byte { return nil }
func (*NullTask) GetSignature() []byte     { return nil }
func (*NullTask) SetSignature([]byte)      {}
func (*NullTask) InitDownloadObjectTask(*storagetypes.ObjectInfo, *storagetypes.BucketInfo, *storagetypes.Params, TPriority, string, int64, int64, int64, int64) {
}
func (*NullTask) GetBucketInfo() *storagetypes.BucketInfo { return nil }
func (*NullTask) GetUserAddress() string                  { return "" }
func (*NullTask) GetSize() int64                          { return 0 }
func (*NullTask) GetLow() int64                           { return 0 }
func (*NullTask) GetHigh() int64                          { return 0 }
func (*NullTask) InitChallengePieceTask(*storagetypes.ObjectInfo, *storagetypes.BucketInfo, *storagetypes.Params, TPriority, string, int32, uint32, int64, int64) {
}
func (*NullTask) SetBucketInfo(*storagetypes.BucketInfo) {}
func (*NullTask) SetUserAddress(string)                  {}
func (*NullTask) GetSegmentIdx() uint32                  { return 0 }
func (*NullTask) GetEcIdx() int32                        { return 0 }
func (*NullTask) SetSegmentIdx(uint32)                   {}
func (*NullTask) GetRecovered() bool                     { return false }
func (*NullTask) SetRecoverDone()                        {}
func (*NullTask) GetRedundancyIdx() int32                { return 0 }
func (*NullTask) SetRedundancyIdx(idx int32)             {}
func (*NullTask) GetIntegrityHash() []byte               { return nil }
func (*NullTask) SetIntegrityHash([]byte)                {}
func (*NullTask) GetPieceHash() [][]byte                 { return nil }
func (*NullTask) SetPieceHash([][]byte)                  {}
func (*NullTask) GetPieceDataSize() int64                { return 0 }
func (*NullTask) SetPieceDataSize(int64)                 {}
func (*NullTask) GetSignBytes() []byte                   { return nil }
func (*NullTask) InitMigratePieceTask(object *storagetypes.ObjectInfo, params *storagetypes.Params, priority TPriority, segmentIdx uint32, ecIdx int32, pieceSize uint64, timeout, retry int64) {
}

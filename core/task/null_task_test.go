package task

import "testing"

func TestNullTask(t *testing.T) {
	n := &NullTask{}
	n.Key()
	n.Type()
	n.Info()
	n.GetAddress()
	n.SetAddress("")
	n.GetCreateTime()
	n.SetCreateTime(0)
	n.GetUpdateTime()
	n.SetUpdateTime(0)
	n.GetTimeout()
	n.SetTimeout(0)
	n.Expired()
	n.ExceedTimeout()
	n.GetPriority()
	n.SetPriority(0)
	n.SetRetry(0)
	n.IncRetry()
	n.ExceedRetry()
	n.GetRetry()
	n.GetMaxRetry()
	n.SetMaxRetry(0)
	n.EstimateLimit()
	_ = n.Error()
	n.SetLogs("")
	_ = n.GetLogs()
	n.AppendLog("")
	n.SetError(nil)
	n.GetExpiredHeight()
	n.SetExpiredHeight(0)
	n.GetObjectInfo()
	n.SetObjectInfo(nil)
	n.GetStorageParams()
	n.SetStorageParams(nil)
	n.GetGCMetaStatus()
	n.SetGCMetaStatus(0, 0)
	n.InitApprovalCreateBucketTask("", nil, nil, 0)
	n.GetCreateBucketInfo()
	n.SetCreateBucketInfo(nil)
	n.InitApprovalCreateObjectTask("", nil, nil, 0)
	n.GetCreateObjectInfo()
	n.SetCreateObjectInfo(nil)
	n.InitApprovalReplicatePieceTask(nil, nil, 0, "")
	n.GetAskSpOperatorAddress()
	n.SetAskSpOperatorAddress("")
	n.GetAskSignature()
	n.SetAskSignature(nil)
	n.GetApprovedSpOperatorAddress()
	n.SetApprovedSpOperatorAddress("")
	n.GetApprovedSignature()
	n.SetApprovedSignature(nil)
	n.GetApprovedSpEndpoint()
	n.SetApprovedSpEndpoint("")
	n.GetApprovedSpApprovalAddress()
	n.SetApprovedSpApprovalAddress("")
	n.InitUploadObjectTask(0, nil, nil, 0, false)
	n.GetVirtualGroupFamilyId()
	n.GetGlobalVirtualGroupId()
	n.SetGlobalVirtualGroupID(0)
	n.GetBucketMigration()
	n.SetBucketMigration(true)
	n.InitReplicatePieceTask(nil, nil, 0, 0, 0, false)
	n.InitRecoverPieceTask(nil, nil, 0, 0, 0, 0, 0, 0)
	n.GetSealed()
	n.SetSealed(true)
	n.GetSecondaryAddresses()
	n.GetSecondarySignatures()
	n.SetSecondarySignatures(nil)
	n.SetSecondaryAddresses(nil)
	n.GetSecondaryEndpoints()
	n.InitSealObjectTask(0, nil, nil, 0, nil, nil, 0, 0)
	n.InitReceivePieceTask(0, nil, nil, 0, 0, 0, 0)
	n.GetReplicateIdx()
	n.SetReplicateIdx(0)
	n.GetPieceIdx()
	n.SetPieceIdx(0)
	n.GetPieceSize()
	n.SetPieceSize(0)
	n.SetPieceChecksum(nil)
	n.GetPieceChecksum()
	n.GetSignature()
	n.SetSignature(nil)
	n.InitDownloadObjectTask(nil, nil, nil, 0, "", 0, 0, 0, 0)
	n.GetBucketInfo()
	n.GetUserAddress()
	n.GetSize()
	n.GetLow()
	n.GetHigh()
	n.InitChallengePieceTask(nil, nil, nil, 0, "", 0, 0, 0, 0)
	n.SetBucketInfo(nil)
	n.SetUserAddress("")
	n.GetSegmentIdx()
	n.GetEcIdx()
	n.SetSegmentIdx(0)
	n.GetRecovered()
	n.SetRecoverDone()
	n.GetRedundancyIdx()
	n.SetRedundancyIdx(0)
	n.GetIntegrityHash()
	n.SetIntegrityHash(nil)
	n.GetPieceHash()
	n.SetPieceHash(nil)
	n.GetPieceDataSize()
	n.SetPieceDataSize(0)
	n.GetSignBytes()
	n.InitMigrateGVGTask(0, 0, nil, 0, nil, 0, 0)
	n.GetSrcGvg()
	n.SetSrcGvg(nil)
	n.GetDestGvg()
	n.SetDestGvg(nil)
	n.GetSrcSp()
	n.SetSrcSp(nil)
	n.GetDestSp()
	n.SetDestSp(nil)
	n.GetBucketID()
	n.SetBucketID(0)
	n.GetLastMigratedObjectID()
	n.SetLastMigratedObjectID(0)
	n.GetFinished()
	n.SetFinished(true)
	n.GetNotAvailableSpIdx()
	n.SetNotAvailableSpIdx(0)
}

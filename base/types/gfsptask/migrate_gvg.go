package gfsptask

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var _ coretask.MigrateGVGTask = &GfSpMigrateGVGTask{}

func (m *GfSpMigrateGVGTask) InitMigrateGVGTask(priority coretask.TPriority, bucketID uint64, srcGvg *virtualgrouptypes.GlobalVirtualGroup,
	redundancyIndex int32, srcSP *sptypes.StorageProvider, timeout, retry int64) {
	m.Reset()
	m.Task = &GfSpTask{}
	m.SetPriority(priority)
	m.BucketId = bucketID
	m.SrcGvg = srcGvg
	m.RedundancyIdx = redundancyIndex
	m.SrcSp = srcSP
	m.SetTimeout(timeout)
	m.SetMaxRetry(retry)
	m.SetRetry(0)
}

func (m *GfSpMigrateGVGTask) Key() coretask.TKey {
	return GfSpMigrateGVGTaskKey(m.GetSrcGvg().GetId(), m.GetBucketId(), m.GetRedundancyIdx())
}

func (m *GfSpMigrateGVGTask) Type() coretask.TType {
	return coretask.TypeTaskMigrateGVG
}

func (m *GfSpMigrateGVGTask) Info() string {
	return fmt.Sprintf(
		"key[%s], type[%s], priority[%d], limit[%s], src_gvg_id[%d], bucket_id[%d], redundancy_index[%d], last_migrated_object_id[%d], finished[%t], %s",
		m.Key(), coretask.TaskTypeName(m.Type()), m.GetPriority(), m.EstimateLimit().String(),
		m.GetSrcGvg().GetId(), m.GetBucketId(), m.GetRedundancyIdx(),
		m.GetLastMigratedObjectId(), m.GetFinished(), m.GetTask().Info())
}

func (m *GfSpMigrateGVGTask) GetAddress() string {
	return m.GetTask().GetAddress()
}

func (m *GfSpMigrateGVGTask) SetAddress(address string) {
	m.GetTask().SetAddress(address)
}

func (m *GfSpMigrateGVGTask) GetCreateTime() int64 {
	return m.GetTask().GetCreateTime()
}

func (m *GfSpMigrateGVGTask) SetCreateTime(time int64) {
	m.GetTask().SetCreateTime(time)
}

func (m *GfSpMigrateGVGTask) GetUpdateTime() int64 {
	return m.GetTask().GetUpdateTime()
}

func (m *GfSpMigrateGVGTask) SetUpdateTime(time int64) {
	m.GetTask().SetUpdateTime(time)
}

func (m *GfSpMigrateGVGTask) GetTimeout() int64 {
	return m.GetTask().GetTimeout()
}

func (m *GfSpMigrateGVGTask) SetTimeout(time int64) {
	m.GetTask().SetTimeout(time)
}

func (m *GfSpMigrateGVGTask) ExceedTimeout() bool {
	return m.GetTask().ExceedTimeout()
}

func (m *GfSpMigrateGVGTask) GetRetry() int64 {
	return m.GetTask().GetRetry()
}

func (m *GfSpMigrateGVGTask) IncRetry() {
	m.GetTask().IncRetry()
}

func (m *GfSpMigrateGVGTask) SetRetry(retry int) {
	m.GetTask().SetRetry(retry)
}

func (m *GfSpMigrateGVGTask) GetMaxRetry() int64 {
	return m.GetTask().GetMaxRetry()
}

func (m *GfSpMigrateGVGTask) SetMaxRetry(limit int64) {
	m.GetTask().SetMaxRetry(limit)
}

func (m *GfSpMigrateGVGTask) ExceedRetry() bool {
	return m.GetTask().ExceedRetry()
}

func (m *GfSpMigrateGVGTask) Expired() bool {
	return m.GetTask().Expired()
}

func (m *GfSpMigrateGVGTask) GetPriority() coretask.TPriority {
	return m.GetTask().GetPriority()
}

func (m *GfSpMigrateGVGTask) SetPriority(priority coretask.TPriority) {
	m.GetTask().SetPriority(priority)
}

func (m *GfSpMigrateGVGTask) EstimateLimit() corercmgr.Limit {
	return LimitEstimateByPriority(m.GetPriority())
}

func (m *GfSpMigrateGVGTask) SetLogs(logs string) {
	m.GetTask().SetLogs(logs)
}

func (m *GfSpMigrateGVGTask) GetLogs() string {
	return m.GetTask().GetLogs()
}

func (m *GfSpMigrateGVGTask) AppendLog(log string) {
	m.GetTask().AppendLog(log)
}

func (m *GfSpMigrateGVGTask) GetUserAddress() string {
	return m.GetTask().GetUserAddress()
}

func (m *GfSpMigrateGVGTask) SetUserAddress(s string) {
	m.GetTask().SetUserAddress(s)
}

func (m *GfSpMigrateGVGTask) Error() error {
	return m.GetTask().Error()
}

func (m *GfSpMigrateGVGTask) SetError(err error) {
	m.GetTask().SetError(err)
}

func (m *GfSpMigrateGVGTask) SetSrcGvg(gvg *virtualgrouptypes.GlobalVirtualGroup) {
	m.SrcGvg = gvg
}

func (m *GfSpMigrateGVGTask) SetDestGvg(gvg *virtualgrouptypes.GlobalVirtualGroup) {
	m.DestGvg = gvg
}

func (m *GfSpMigrateGVGTask) SetSrcSp(srcSP *sptypes.StorageProvider) {
	m.SrcSp = srcSP
}

func (m *GfSpMigrateGVGTask) GetBucketID() uint64 {
	return m.GetBucketId()
}

func (m *GfSpMigrateGVGTask) SetBucketID(bucketID uint64) {
	m.BucketId = bucketID
}

func (m *GfSpMigrateGVGTask) SetRedundancyIdx(rIdx int32) {
	m.RedundancyIdx = rIdx
}

func (m *GfSpMigrateGVGTask) GetLastMigratedObjectID() uint64 {
	return m.GetLastMigratedObjectId()
}

func (m *GfSpMigrateGVGTask) SetLastMigratedObjectID(lastMigratedObjectID uint64) {
	m.LastMigratedObjectId = lastMigratedObjectID
}

func (m *GfSpMigrateGVGTask) SetMigratedBytesSize(migratedBytesSize uint64) {
	m.MigratedBytesSize = migratedBytesSize
}

func (m *GfSpMigrateGVGTask) SetFinished(finished bool) {
	m.Finished = finished
}

func (m *GfSpMigrateGVGTask) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&GfSpMigrateGVGTask{
		BucketId:      m.GetBucketID(),
		SrcGvg:        m.GetSrcGvg(),
		RedundancyIdx: m.GetRedundancyIdx(),
		ExpireTime:    m.GetExpireTime(),
	}))
}

func (m *GfSpMigrateGVGTask) SetSignature(signature []byte) {
	m.Signature = signature
}

// ======================= MigratePieceTask =====================================

func (g *GfSpMigratePieceTask) Key() coretask.TKey {
	return GfSpMigratePieceTaskKey(g.GetObjectInfo().GetObjectName(), g.GetObjectInfo().Id.String(),
		g.GetSegmentIdx(), g.GetRedundancyIdx())
}

// ======================= GfSpBucketMigrationInfo =====================================

func (m *GfSpBucketMigrationInfo) Key() coretask.TKey {
	return GfSpMigrateBucketTaskKey(m.GetBucketId())
}

func (m *GfSpBucketMigrationInfo) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&GfSpBucketMigrationInfo{
		BucketId:          m.GetBucketId(),
		Finished:          m.GetFinished(),
		MigratedBytesSize: m.GetMigratedBytesSize(),
		ExpireTime:        m.GetExpireTime(),
	}))
}

func (m *GfSpBucketMigrationInfo) SetSignature(signature []byte) {
	m.Signature = signature
}

func (m *GfSpBucketMigrationInfo) Info() string {
	return fmt.Sprintf(
		"key[%s],   bucket_id[%d], migrated_bytes_size[%d], finished[%t].",
		m.Key(), m.GetBucketId(), m.GetMigratedBytesSize(), m.GetFinished())
}

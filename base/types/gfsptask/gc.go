package gfsptask

import (
	"fmt"
	"time"

	corercmgr "github.com/zkMeLabs/mechain-storage-provider/core/rcmgr"
	coretask "github.com/zkMeLabs/mechain-storage-provider/core/task"
)

var (
	_ coretask.GCObjectTask      = &GfSpGCObjectTask{}
	_ coretask.GCZombiePieceTask = &GfSpGCZombiePieceTask{}
	_ coretask.GCMetaTask        = &GfSpGCMetaTask{}
)

func (m *GfSpGCObjectTask) InitGCObjectTask(priority coretask.TPriority, start, end uint64, timeout int64) {
	m.Reset()
	m.Task = &GfSpTask{}
	m.SetStartBlockNumber(start)
	m.SetEndBlockNumber(end)
	m.SetPriority(priority)
	m.SetCreateTime(time.Now().Unix())
	m.SetUpdateTime(time.Now().Unix())
	m.SetTimeout(timeout)
}

func (m *GfSpGCObjectTask) Key() coretask.TKey {
	return GfSpGCObjectTaskKey(
		m.GetStartBlockNumber(),
		m.GetEndBlockNumber(),
		m.GetCreateTime())
}

func (m *GfSpGCObjectTask) Type() coretask.TType {
	return coretask.TypeTaskGCObject
}

func (m *GfSpGCObjectTask) Info() string {
	return fmt.Sprintf(
		"key[%s], type[%s], priority[%d], limit[%s], start_block[%d], end_block[%d], current_block[%d], last_deleted_object_id[%d], %s",
		m.Key(), coretask.TaskTypeName(m.Type()), m.GetPriority(), m.EstimateLimit().String(),
		m.GetStartBlockNumber(), m.GetEndBlockNumber(), m.GetCurrentBlockNumber(),
		m.GetLastDeletedObjectId(), m.GetTask().Info())
}

func (m *GfSpGCObjectTask) GetAddress() string {
	return m.GetTask().GetAddress()
}

func (m *GfSpGCObjectTask) SetAddress(address string) {
	m.GetTask().SetAddress(address)
}

func (m *GfSpGCObjectTask) GetCreateTime() int64 {
	return m.GetTask().GetCreateTime()
}

func (m *GfSpGCObjectTask) SetCreateTime(time int64) {
	m.GetTask().SetCreateTime(time)
}

func (m *GfSpGCObjectTask) GetUpdateTime() int64 {
	return m.GetTask().GetUpdateTime()
}

func (m *GfSpGCObjectTask) SetUpdateTime(time int64) {
	m.GetTask().SetUpdateTime(time)
}

func (m *GfSpGCObjectTask) GetTimeout() int64 {
	return m.GetTask().GetTimeout()
}

func (m *GfSpGCObjectTask) SetTimeout(time int64) {
	m.GetTask().SetTimeout(time)
}

func (m *GfSpGCObjectTask) ExceedTimeout() bool {
	return m.GetTask().ExceedTimeout()
}

func (m *GfSpGCObjectTask) GetRetry() int64 {
	return m.GetTask().GetRetry()
}

func (m *GfSpGCObjectTask) IncRetry() {
	m.GetTask().IncRetry()
}

func (m *GfSpGCObjectTask) SetRetry(retry int64) {
	m.GetTask().SetRetry(retry)
}

func (m *GfSpGCObjectTask) GetMaxRetry() int64 {
	return m.GetTask().GetMaxRetry()
}

func (m *GfSpGCObjectTask) SetMaxRetry(limit int64) {
	m.GetTask().SetMaxRetry(limit)
}

func (m *GfSpGCObjectTask) ExceedRetry() bool {
	return m.GetTask().ExceedRetry()
}

func (m *GfSpGCObjectTask) Expired() bool {
	return m.GetTask().Expired()
}

func (m *GfSpGCObjectTask) GetPriority() coretask.TPriority {
	return m.GetTask().GetPriority()
}

func (m *GfSpGCObjectTask) SetPriority(priority coretask.TPriority) {
	m.GetTask().SetPriority(priority)
}

func (m *GfSpGCObjectTask) EstimateLimit() corercmgr.Limit {
	return LimitEstimateByPriority(m.GetPriority())
}

func (m *GfSpGCObjectTask) GetUserAddress() string {
	return m.GetTask().GetUserAddress()
}

func (m *GfSpGCObjectTask) SetUserAddress(address string) {
	m.GetTask().SetUserAddress(address)
}

func (m *GfSpGCObjectTask) SetLogs(logs string) {
	m.GetTask().SetLogs(logs)
}

func (m *GfSpGCObjectTask) GetLogs() string {
	return m.GetTask().GetLogs()
}

func (m *GfSpGCObjectTask) AppendLog(log string) {
	m.GetTask().AppendLog(log)
}

func (m *GfSpGCObjectTask) Error() error {
	return m.GetTask().Error()
}

func (m *GfSpGCObjectTask) SetError(err error) {
	m.GetTask().SetError(err)
}

func (m *GfSpGCObjectTask) SetStartBlockNumber(block uint64) {
	m.StartBlockNumber = block
}

func (m *GfSpGCObjectTask) SetEndBlockNumber(block uint64) {
	m.EndBlockNumber = block
}

func (m *GfSpGCObjectTask) GetGCObjectProgress() (uint64, uint64) {
	return m.GetCurrentBlockNumber(), m.GetLastDeletedObjectId()
}

func (m *GfSpGCObjectTask) SetGCObjectProgress(block uint64, object uint64) {
	m.CurrentBlockNumber = block
	m.LastDeletedObjectId = object
}

func (m *GfSpGCObjectTask) SetCurrentBlockNumber(block uint64) {
	m.CurrentBlockNumber = block
}

func (m *GfSpGCObjectTask) SetLastDeletedObjectId(object uint64) {
	m.LastDeletedObjectId = object
}

func (m *GfSpGCZombiePieceTask) InitGCZombiePieceTask(priority coretask.TPriority, start, end uint64, timeout int64) {
	m.Reset()
	m.Task = &GfSpTask{}
	m.SetStartObjectID(start)
	m.SetEndObjectID(end)
	m.SetPriority(priority)
	m.SetCreateTime(time.Now().Unix())
	m.SetUpdateTime(time.Now().Unix())
	m.SetTimeout(timeout)
}

func (m *GfSpGCZombiePieceTask) Key() coretask.TKey {
	return GfSpGCZombiePieceTaskKey(m.GetStartObjectId(), m.GetEndObjectId(), m.GetCreateTime())
}

func (m *GfSpGCZombiePieceTask) Type() coretask.TType {
	return coretask.TypeTaskGCZombiePiece
}

func (m *GfSpGCZombiePieceTask) Info() string {
	return fmt.Sprintf("key[%s], type[%s], priority[%d], limit[%s], start_object_id[%d], end_object_id[%d],%s",
		m.Key(), coretask.TaskTypeName(m.Type()), m.GetPriority(),
		m.EstimateLimit().String(), m.GetStartObjectId(), m.GetEndObjectId(), m.GetTask().Info())
}

func (m *GfSpGCZombiePieceTask) GetAddress() string {
	return m.GetTask().GetAddress()
}

func (m *GfSpGCZombiePieceTask) SetAddress(address string) {
	m.GetTask().SetAddress(address)
}

func (m *GfSpGCZombiePieceTask) GetCreateTime() int64 {
	return m.GetTask().GetCreateTime()
}

func (m *GfSpGCZombiePieceTask) SetCreateTime(time int64) {
	m.GetTask().SetCreateTime(time)
}

func (m *GfSpGCZombiePieceTask) GetUpdateTime() int64 {
	return m.GetTask().GetUpdateTime()
}

func (m *GfSpGCZombiePieceTask) SetUpdateTime(time int64) {
	m.GetTask().SetUpdateTime(time)
}

func (m *GfSpGCZombiePieceTask) GetTimeout() int64 {
	return m.GetTask().GetTimeout()
}

func (m *GfSpGCZombiePieceTask) SetTimeout(time int64) {
	m.GetTask().SetTimeout(time)
}

func (m *GfSpGCZombiePieceTask) ExceedTimeout() bool {
	return m.GetTask().ExceedTimeout()
}

func (m *GfSpGCZombiePieceTask) GetRetry() int64 {
	return m.GetTask().GetRetry()
}

func (m *GfSpGCZombiePieceTask) IncRetry() {
	m.GetTask().IncRetry()
}

func (m *GfSpGCZombiePieceTask) SetRetry(retry int64) {
	m.GetTask().SetRetry(retry)
}

func (m *GfSpGCZombiePieceTask) GetMaxRetry() int64 {
	return m.GetTask().GetMaxRetry()
}

func (m *GfSpGCZombiePieceTask) SetMaxRetry(limit int64) {
	m.GetTask().SetMaxRetry(limit)
}

func (m *GfSpGCZombiePieceTask) ExceedRetry() bool {
	return m.GetTask().ExceedRetry()
}

func (m *GfSpGCZombiePieceTask) Expired() bool {
	return m.GetTask().Expired()
}

func (m *GfSpGCZombiePieceTask) GetPriority() coretask.TPriority {
	return m.GetTask().GetPriority()
}

func (m *GfSpGCZombiePieceTask) SetPriority(priority coretask.TPriority) {
	m.GetTask().SetPriority(priority)
}

func (m *GfSpGCZombiePieceTask) EstimateLimit() corercmgr.Limit {
	return LimitEstimateByPriority(m.GetPriority())
}

func (m *GfSpGCZombiePieceTask) GetUserAddress() string {
	return m.GetTask().GetUserAddress()
}

func (m *GfSpGCZombiePieceTask) SetUserAddress(address string) {
	m.GetTask().SetUserAddress(address)
}

func (m *GfSpGCZombiePieceTask) SetLogs(logs string) {
	m.GetTask().SetLogs(logs)
}

func (m *GfSpGCZombiePieceTask) GetLogs() string {
	return m.GetTask().GetLogs()
}

func (m *GfSpGCZombiePieceTask) AppendLog(log string) {
	m.GetTask().AppendLog(log)
}

func (m *GfSpGCZombiePieceTask) Error() error {
	return m.GetTask().Error()
}

func (m *GfSpGCZombiePieceTask) SetError(err error) {
	m.GetTask().SetError(err)
}

func (m *GfSpGCZombiePieceTask) SetStartObjectID(id uint64) {
	m.StartObjectId = id
}

func (m *GfSpGCZombiePieceTask) SetEndObjectID(id uint64) {
	m.EndObjectId = id
}

func (m *GfSpGCStaleVersionObjectTask) InitGCStaleVersionObjectTask(priority coretask.TPriority,
	objectID uint64,
	redundancyIndex int32,
	integrityChecksum []byte,
	pieceChecksumList [][]byte,
	version int64,
	objectSize uint64,
	timeout int64,
) {
	m.Reset()
	m.Task = &GfSpTask{}
	m.SetObjectID(objectID)
	m.SetRedundancyIndex(redundancyIndex)
	m.SetVersion(version)
	m.SetObjectSize(objectSize)
	m.SetIntegrityChecksum(integrityChecksum)
	m.SetPieceChecksumList(pieceChecksumList)
	m.SetPriority(priority)
	m.SetCreateTime(time.Now().Unix())
	m.SetUpdateTime(time.Now().Unix())
	m.SetTimeout(timeout)
}

func (m *GfSpGCStaleVersionObjectTask) Key() coretask.TKey {
	return GfSpGCStaleVersionObjectTaskKey(m.GetObjectId(), m.GetRedundancyIndex(), m.GetVersion(), m.GetCreateTime())
}

func (m *GfSpGCStaleVersionObjectTask) Type() coretask.TType {
	return coretask.TypeTaskGCStaleVersionObject
}

func (m *GfSpGCStaleVersionObjectTask) Info() string {
	return fmt.Sprintf("key[%s], type[%s], priority[%d], limit[%s], object_id[%d], version[%d], %s",
		m.Key(), coretask.TaskTypeName(m.Type()), m.GetPriority(),
		m.EstimateLimit().String(), m.GetObjectId(), m.GetVersion(),
		m.GetTask().Info())
}

func (m *GfSpGCStaleVersionObjectTask) GetAddress() string {
	return m.GetTask().GetAddress()
}

func (m *GfSpGCStaleVersionObjectTask) SetAddress(address string) {
	m.GetTask().SetAddress(address)
}

func (m *GfSpGCStaleVersionObjectTask) GetCreateTime() int64 {
	return m.GetTask().GetCreateTime()
}

func (m *GfSpGCStaleVersionObjectTask) SetCreateTime(time int64) {
	m.GetTask().SetCreateTime(time)
}

func (m *GfSpGCStaleVersionObjectTask) GetUpdateTime() int64 {
	return m.GetTask().GetUpdateTime()
}

func (m *GfSpGCStaleVersionObjectTask) SetUpdateTime(time int64) {
	m.GetTask().SetUpdateTime(time)
}

func (m *GfSpGCStaleVersionObjectTask) GetTimeout() int64 {
	return m.GetTask().GetTimeout()
}

func (m *GfSpGCStaleVersionObjectTask) SetTimeout(time int64) {
	m.GetTask().SetTimeout(time)
}

func (m *GfSpGCStaleVersionObjectTask) ExceedTimeout() bool {
	return m.GetTask().ExceedTimeout()
}

func (m *GfSpGCStaleVersionObjectTask) GetRetry() int64 {
	return m.GetTask().GetRetry()
}

func (m *GfSpGCStaleVersionObjectTask) IncRetry() {
	m.GetTask().IncRetry()
}

func (m *GfSpGCStaleVersionObjectTask) SetRetry(retry int64) {
	m.GetTask().SetRetry(retry)
}

func (m *GfSpGCStaleVersionObjectTask) GetMaxRetry() int64 {
	return m.GetTask().GetMaxRetry()
}

func (m *GfSpGCStaleVersionObjectTask) SetMaxRetry(limit int64) {
	m.GetTask().SetMaxRetry(limit)
}

func (m *GfSpGCStaleVersionObjectTask) ExceedRetry() bool {
	return m.GetTask().ExceedRetry()
}

func (m *GfSpGCStaleVersionObjectTask) Expired() bool {
	return m.GetTask().Expired()
}

func (m *GfSpGCStaleVersionObjectTask) GetPriority() coretask.TPriority {
	return m.GetTask().GetPriority()
}

func (m *GfSpGCStaleVersionObjectTask) SetPriority(priority coretask.TPriority) {
	m.GetTask().SetPriority(priority)
}

func (m *GfSpGCStaleVersionObjectTask) EstimateLimit() corercmgr.Limit {
	return LimitEstimateByPriority(m.GetPriority())
}

func (m *GfSpGCStaleVersionObjectTask) GetUserAddress() string {
	return m.GetTask().GetUserAddress()
}

func (m *GfSpGCStaleVersionObjectTask) SetUserAddress(address string) {
	m.GetTask().SetUserAddress(address)
}

func (m *GfSpGCStaleVersionObjectTask) SetLogs(logs string) {
	m.GetTask().SetLogs(logs)
}

func (m *GfSpGCStaleVersionObjectTask) GetLogs() string {
	return m.GetTask().GetLogs()
}

func (m *GfSpGCStaleVersionObjectTask) AppendLog(log string) {
	m.GetTask().AppendLog(log)
}

func (m *GfSpGCStaleVersionObjectTask) Error() error {
	return m.GetTask().Error()
}

func (m *GfSpGCStaleVersionObjectTask) SetError(err error) {
	m.GetTask().SetError(err)
}

func (m *GfSpGCStaleVersionObjectTask) SetObjectID(id uint64) {
	m.ObjectId = id
}

func (m *GfSpGCStaleVersionObjectTask) SetVersion(version int64) {
	m.Version = version
}

func (m *GfSpGCStaleVersionObjectTask) SetObjectSize(objectSize uint64) {
	m.ObjectSize = objectSize
}

func (m *GfSpGCStaleVersionObjectTask) SetRedundancyIndex(redundancyIndex int32) {
	m.RedundancyIndex = redundancyIndex
}

func (m *GfSpGCStaleVersionObjectTask) SetIntegrityChecksum(integrityChecksum []byte) {
	m.IntegrityChecksum = integrityChecksum
}

func (m *GfSpGCStaleVersionObjectTask) SetPieceChecksumList(pieceChecksumList [][]byte) {
	m.PieceChecksumList = pieceChecksumList
}

func (m *GfSpGCMetaTask) InitGCMetaTask(priority coretask.TPriority, timeout int64) {
	m.Reset()
	m.Task = &GfSpTask{}
	m.SetPriority(priority)
	m.SetCreateTime(time.Now().Unix())
	m.SetUpdateTime(time.Now().Unix())
	m.SetTimeout(timeout)
}

func (m *GfSpGCMetaTask) Key() coretask.TKey {
	return GfSpGfSpGCMetaTaskKey(m.GetCreateTime())
}

func (m *GfSpGCMetaTask) Type() coretask.TType {
	return coretask.TypeTaskGCMeta
}

func (m *GfSpGCMetaTask) Info() string {
	return fmt.Sprintf("key[%s], type[%s], priority[%d], limit[%s], %s",
		m.Key(), coretask.TaskTypeName(m.Type()), m.GetPriority(),
		m.EstimateLimit().String(), m.GetTask().Info())
}

func (m *GfSpGCMetaTask) GetAddress() string {
	return m.GetTask().GetAddress()
}

func (m *GfSpGCMetaTask) SetAddress(address string) {
	m.GetTask().SetAddress(address)
}

func (m *GfSpGCMetaTask) GetCreateTime() int64 {
	return m.GetTask().GetCreateTime()
}

func (m *GfSpGCMetaTask) SetCreateTime(time int64) {
	m.GetTask().SetCreateTime(time)
}

func (m *GfSpGCMetaTask) GetUpdateTime() int64 {
	return m.GetTask().GetUpdateTime()
}

func (m *GfSpGCMetaTask) SetUpdateTime(time int64) {
	m.GetTask().SetUpdateTime(time)
}

func (m *GfSpGCMetaTask) GetTimeout() int64 {
	return m.GetTask().GetTimeout()
}

func (m *GfSpGCMetaTask) SetTimeout(time int64) {
	m.GetTask().SetTimeout(time)
}

func (m *GfSpGCMetaTask) ExceedTimeout() bool {
	return m.GetTask().ExceedTimeout()
}

func (m *GfSpGCMetaTask) GetRetry() int64 {
	return m.GetTask().GetRetry()
}

func (m *GfSpGCMetaTask) IncRetry() {
	m.GetTask().IncRetry()
}

func (m *GfSpGCMetaTask) SetRetry(retry int64) {
	m.GetTask().SetRetry(retry)
}

func (m *GfSpGCMetaTask) GetMaxRetry() int64 {
	return m.GetTask().GetMaxRetry()
}

func (m *GfSpGCMetaTask) SetMaxRetry(limit int64) {
	m.GetTask().SetMaxRetry(limit)
}

func (m *GfSpGCMetaTask) ExceedRetry() bool {
	return m.GetTask().ExceedRetry()
}

func (m *GfSpGCMetaTask) Expired() bool {
	return m.GetTask().Expired()
}

func (m *GfSpGCMetaTask) GetPriority() coretask.TPriority {
	return m.GetTask().GetPriority()
}

func (m *GfSpGCMetaTask) SetPriority(priority coretask.TPriority) {
	m.GetTask().SetPriority(priority)
}

func (m *GfSpGCMetaTask) EstimateLimit() corercmgr.Limit {
	return LimitEstimateByPriority(m.GetPriority())
}

func (m *GfSpGCMetaTask) GetUserAddress() string {
	return m.GetTask().GetUserAddress()
}

func (m *GfSpGCMetaTask) SetUserAddress(address string) {
	m.GetTask().SetUserAddress(address)
}

func (m *GfSpGCMetaTask) SetLogs(logs string) {
	m.GetTask().SetLogs(logs)
}

func (m *GfSpGCMetaTask) GetLogs() string {
	return m.GetTask().GetLogs()
}

func (m *GfSpGCMetaTask) AppendLog(log string) {
	m.GetTask().AppendLog(log)
}

func (m *GfSpGCMetaTask) Error() error {
	return m.GetTask().Error()
}

func (m *GfSpGCMetaTask) SetError(err error) {
	m.GetTask().SetError(err)
}

func (m *GfSpGCMetaTask) GetGCMetaStatus() (uint64, uint64) {
	return m.GetCurrentIdx(), m.GetDeleteCount()
}

func (m *GfSpGCMetaTask) SetGCMetaStatus(current uint64, delete uint64) {
	m.CurrentIdx = current
	m.DeleteCount = delete
}

package gfsptask

import (
	"fmt"
	"time"

	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
)

var _ coretask.GCObjectTask = &GfSpGCObjectTask{}
var _ coretask.GCZombiePieceTask = &GfSpGCZombiePieceTask{}
var _ coretask.GCMetaTask = &GfSpGCMetaTask{}

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

func (m *GfSpGCObjectTask) SetRetry(retry int) {
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

func (m *GfSpGCZombiePieceTask) Key() coretask.TKey {
	return GfSpGCZombiePieceTaskKey(m.GetCreateTime())
}

func (m *GfSpGCZombiePieceTask) Type() coretask.TType {
	return coretask.TypeTaskGCZombiePiece
}

func (m *GfSpGCZombiePieceTask) Info() string {
	return fmt.Sprintf("key[%s], type[%s], priority[%d], limit[%s], %s",
		m.Key(), coretask.TaskTypeName(m.Type()), m.GetPriority(),
		m.EstimateLimit().String(), m.GetTask().Info())
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

func (m *GfSpGCZombiePieceTask) SetRetry(retry int) {
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

func (m *GfSpGCZombiePieceTask) Error() error {
	return m.GetTask().Error()
}

func (m *GfSpGCZombiePieceTask) SetError(err error) {
	m.GetTask().SetError(err)
}

func (m *GfSpGCZombiePieceTask) GetGCZombiePieceStatus() (uint64, uint64) {
	return m.GetObjectId(), m.GetDeleteCount()
}

func (m *GfSpGCZombiePieceTask) SetGCZombiePieceStatus(object uint64, delete uint64) {
	m.ObjectId = object
	m.DeleteCount = delete
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

func (m *GfSpGCMetaTask) SetRetry(retry int) {
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

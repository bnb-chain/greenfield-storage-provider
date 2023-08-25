package gfsptask

import (
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsplimit"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var _ coretask.ApprovalCreateBucketTask = &GfSpCreateBucketApprovalTask{}
var _ coretask.ApprovalMigrateBucketTask = &GfSpMigrateBucketApprovalTask{}
var _ coretask.ApprovalCreateObjectTask = &GfSpCreateObjectApprovalTask{}
var _ coretask.ApprovalReplicatePieceTask = &GfSpReplicatePieceApprovalTask{}

func (m *GfSpCreateBucketApprovalTask) InitApprovalCreateBucketTask(
	account string, bucket *storagetypes.MsgCreateBucket, fingerprint []byte, priority coretask.TPriority) {
	m.Reset()
	m.Task = &GfSpTask{}
	m.Fingerprint = fingerprint
	m.GetTask().SetCreateTime(time.Now().Unix())
	m.GetTask().SetUpdateTime(time.Now().Unix())
	m.SetUserAddress(account)
	m.SetCreateBucketInfo(bucket)
	m.SetPriority(priority)
}

func (m *GfSpCreateBucketApprovalTask) Key() coretask.TKey {
	return GfSpCreateBucketApprovalTaskKey(
		m.GetCreateBucketInfo().GetBucketName(),
		m.GetUserAddress(),
		m.Fingerprint)
}

func (m *GfSpCreateBucketApprovalTask) Type() coretask.TType {
	return coretask.TypeTaskCreateBucketApproval
}

func (m *GfSpCreateBucketApprovalTask) Info() string {
	return fmt.Sprintf("key[%s], type[%s], priority[%d], limit[%s], expiredHeight[%d], %s",
		m.Key(), coretask.TaskTypeName(m.Type()), m.GetPriority(), m.EstimateLimit().String(),
		m.GetExpiredHeight(), m.GetTask().Info())
}

func (m *GfSpCreateBucketApprovalTask) GetAddress() string {
	return m.GetTask().GetAddress()
}

func (m *GfSpCreateBucketApprovalTask) SetAddress(address string) {
	m.GetTask().SetAddress(address)
}

func (m *GfSpCreateBucketApprovalTask) GetCreateTime() int64 {
	return m.GetTask().GetCreateTime()
}

func (m *GfSpCreateBucketApprovalTask) SetCreateTime(time int64) {
	m.GetTask().SetCreateTime(time)
}

func (m *GfSpCreateBucketApprovalTask) GetUpdateTime() int64 {
	return m.GetTask().GetUpdateTime()
}

func (m *GfSpCreateBucketApprovalTask) SetUpdateTime(time int64) {
	m.GetTask().SetUpdateTime(time)
}

func (m *GfSpCreateBucketApprovalTask) GetTimeout() int64 {
	return m.GetTask().GetTimeout()
}

func (m *GfSpCreateBucketApprovalTask) SetTimeout(time int64) {
	m.GetTask().SetTimeout(time)
}

func (m *GfSpCreateBucketApprovalTask) ExceedTimeout() bool {
	return m.GetTask().ExceedTimeout()
}

func (m *GfSpCreateBucketApprovalTask) GetRetry() int64 {
	return m.GetTask().GetRetry()
}

func (m *GfSpCreateBucketApprovalTask) IncRetry() {
	m.GetTask().IncRetry()
}

func (m *GfSpCreateBucketApprovalTask) SetRetry(retry int) {
	m.GetTask().SetRetry(retry)
}

func (m *GfSpCreateBucketApprovalTask) GetMaxRetry() int64 {
	return m.GetTask().GetMaxRetry()
}

func (m *GfSpCreateBucketApprovalTask) SetMaxRetry(limit int64) {
	m.GetTask().SetMaxRetry(limit)
}

func (m *GfSpCreateBucketApprovalTask) ExceedRetry() bool {
	return m.GetTask().ExceedRetry()
}

func (m *GfSpCreateBucketApprovalTask) Expired() bool {
	return m.GetTask().Expired()
}

func (m *GfSpCreateBucketApprovalTask) GetPriority() coretask.TPriority {
	return m.GetTask().GetPriority()
}

func (m *GfSpCreateBucketApprovalTask) SetPriority(priority coretask.TPriority) {
	m.GetTask().SetPriority(priority)
}

func (m *GfSpCreateBucketApprovalTask) EstimateLimit() corercmgr.Limit {
	l := &gfsplimit.GfSpLimit{}
	l.Add(LimitEstimateByPriority(m.GetPriority()))
	return l
}

func (m *GfSpCreateBucketApprovalTask) Error() error {
	return m.GetTask().Error()
}

func (m *GfSpCreateBucketApprovalTask) GetUserAddress() string {
	return m.GetTask().GetUserAddress()
}

func (m *GfSpCreateBucketApprovalTask) SetUserAddress(address string) {
	m.GetTask().SetUserAddress(address)
}

func (m *GfSpCreateBucketApprovalTask) SetLogs(logs string) {
	m.GetTask().SetLogs(logs)
}

func (m *GfSpCreateBucketApprovalTask) GetLogs() string {
	return m.GetTask().GetLogs()
}

func (m *GfSpCreateBucketApprovalTask) AppendLog(log string) {
	m.GetTask().AppendLog(log)
}

func (m *GfSpCreateBucketApprovalTask) SetError(err error) {
	m.GetTask().SetError(err)
}

func (m *GfSpCreateBucketApprovalTask) SetExpiredHeight(height uint64) {
	m.GetCreateBucketInfo().GetPrimarySpApproval().ExpiredHeight = height
}

func (m *GfSpCreateBucketApprovalTask) GetExpiredHeight() uint64 {
	return m.GetCreateBucketInfo().GetPrimarySpApproval().GetExpiredHeight()
}

func (m *GfSpCreateBucketApprovalTask) SetCreateBucketInfo(bucket *storagetypes.MsgCreateBucket) {
	m.CreateBucketInfo = bucket
}

func (m *GfSpMigrateBucketApprovalTask) InitApprovalMigrateBucketTask(bucket *storagetypes.MsgMigrateBucket, priority coretask.TPriority) {
	m.Reset()
	m.Task = &GfSpTask{}
	m.GetTask().SetCreateTime(time.Now().Unix())
	m.GetTask().SetUpdateTime(time.Now().Unix())
	m.SetMigrateBucketInfo(bucket)
	m.SetPriority(priority)
}

func (m *GfSpMigrateBucketApprovalTask) Key() coretask.TKey {
	return GfSpMigrateBucketApprovalTaskKey(m.GetMigrateBucketInfo().GetBucketName(),
		hex.EncodeToString(m.GetMigrateBucketInfo().GetSignBytes()))
}

func (m *GfSpMigrateBucketApprovalTask) Type() coretask.TType {
	return coretask.TypeTaskMigrateBucketApproval
}

func (m *GfSpMigrateBucketApprovalTask) Info() string {
	return fmt.Sprintf("key[%s], type[%s], priority[%d], limit[%s], expiredHeight[%d], %s",
		m.Key(), coretask.TaskTypeName(m.Type()), m.GetPriority(), m.EstimateLimit().String(),
		m.GetExpiredHeight(), m.GetTask().Info())
}

func (m *GfSpMigrateBucketApprovalTask) GetAddress() string {
	return m.GetTask().GetAddress()
}

func (m *GfSpMigrateBucketApprovalTask) SetAddress(address string) {
	m.GetTask().SetAddress(address)
}

func (m *GfSpMigrateBucketApprovalTask) GetCreateTime() int64 {
	return m.GetTask().GetCreateTime()
}

func (m *GfSpMigrateBucketApprovalTask) SetCreateTime(time int64) {
	m.GetTask().SetCreateTime(time)
}

func (m *GfSpMigrateBucketApprovalTask) GetUpdateTime() int64 {
	return m.GetTask().GetUpdateTime()
}

func (m *GfSpMigrateBucketApprovalTask) SetUpdateTime(time int64) {
	m.GetTask().SetUpdateTime(time)
}

func (m *GfSpMigrateBucketApprovalTask) GetTimeout() int64 {
	return m.GetTask().GetTimeout()
}

func (m *GfSpMigrateBucketApprovalTask) SetTimeout(time int64) {
	m.GetTask().SetTimeout(time)
}

func (m *GfSpMigrateBucketApprovalTask) ExceedTimeout() bool {
	return m.GetTask().ExceedTimeout()
}

func (m *GfSpMigrateBucketApprovalTask) GetRetry() int64 {
	return m.GetTask().GetRetry()
}

func (m *GfSpMigrateBucketApprovalTask) IncRetry() {
	m.GetTask().IncRetry()
}

func (m *GfSpMigrateBucketApprovalTask) SetRetry(retry int) {
	m.GetTask().SetRetry(retry)
}

func (m *GfSpMigrateBucketApprovalTask) GetMaxRetry() int64 {
	return m.GetTask().GetMaxRetry()
}

func (m *GfSpMigrateBucketApprovalTask) SetMaxRetry(limit int64) {
	m.GetTask().SetMaxRetry(limit)
}

func (m *GfSpMigrateBucketApprovalTask) ExceedRetry() bool {
	return m.GetTask().ExceedRetry()
}

func (m *GfSpMigrateBucketApprovalTask) Expired() bool {
	return m.GetTask().Expired()
}

func (m *GfSpMigrateBucketApprovalTask) GetPriority() coretask.TPriority {
	return m.GetTask().GetPriority()
}

func (m *GfSpMigrateBucketApprovalTask) SetPriority(priority coretask.TPriority) {
	m.GetTask().SetPriority(priority)
}

func (m *GfSpMigrateBucketApprovalTask) EstimateLimit() corercmgr.Limit {
	l := &gfsplimit.GfSpLimit{}
	l.Add(LimitEstimateByPriority(m.GetPriority()))
	return l
}

func (m *GfSpMigrateBucketApprovalTask) Error() error {
	return m.GetTask().Error()
}

func (m *GfSpMigrateBucketApprovalTask) SetError(err error) {
	m.GetTask().SetError(err)
}

func (m *GfSpMigrateBucketApprovalTask) SetExpiredHeight(height uint64) {
	m.GetMigrateBucketInfo().GetDstPrimarySpApproval().ExpiredHeight = height
}

func (m *GfSpMigrateBucketApprovalTask) GetExpiredHeight() uint64 {
	return m.GetMigrateBucketInfo().GetDstPrimarySpApproval().GetExpiredHeight()
}

func (m *GfSpMigrateBucketApprovalTask) SetMigrateBucketInfo(bucket *storagetypes.MsgMigrateBucket) {
	m.MigrateBucketInfo = bucket
}

func (m *GfSpMigrateBucketApprovalTask) SetLogs(logs string) {
	m.GetTask().SetLogs(logs)
}

func (m *GfSpMigrateBucketApprovalTask) GetLogs() string {
	return m.GetTask().GetLogs()
}

func (m *GfSpMigrateBucketApprovalTask) AppendLog(log string) {
	m.GetTask().AppendLog(log)
}

func (m *GfSpMigrateBucketApprovalTask) GetUserAddress() string {
	return m.GetTask().GetUserAddress()
}

func (m *GfSpMigrateBucketApprovalTask) SetUserAddress(address string) {
	m.GetTask().SetUserAddress(address)
}

func (m *GfSpCreateObjectApprovalTask) InitApprovalCreateObjectTask(account string, object *storagetypes.MsgCreateObject,
	fingerprint []byte, priority coretask.TPriority) {
	m.Reset()
	m.Task = &GfSpTask{}
	m.Fingerprint = fingerprint
	m.GetTask().SetCreateTime(time.Now().Unix())
	m.GetTask().SetUpdateTime(time.Now().Unix())
	m.SetUserAddress(account)
	m.SetCreateObjectInfo(object)
	m.SetPriority(priority)
}

func (m *GfSpCreateObjectApprovalTask) Key() coretask.TKey {
	return GfSpCreateObjectApprovalTaskKey(
		m.GetCreateObjectInfo().GetBucketName(),
		m.GetCreateObjectInfo().GetObjectName(),
		m.GetUserAddress(),
		m.Fingerprint)
}

func (m *GfSpCreateObjectApprovalTask) Type() coretask.TType {
	return coretask.TypeTaskCreateObjectApproval
}

func (m *GfSpCreateObjectApprovalTask) Info() string {
	return fmt.Sprintf("key[%s], type[%s], priority[%d], limit[%s], expiedHeight[%d], %s",
		m.Key(), coretask.TaskTypeName(m.Type()), m.GetPriority(), m.EstimateLimit().String(),
		m.GetExpiredHeight(), m.GetTask().Info())
}

func (m *GfSpCreateObjectApprovalTask) GetAddress() string {
	return m.GetTask().GetAddress()
}

func (m *GfSpCreateObjectApprovalTask) SetAddress(address string) {
	m.GetTask().SetAddress(address)
}

func (m *GfSpCreateObjectApprovalTask) GetCreateTime() int64 {
	return m.GetTask().GetCreateTime()
}

func (m *GfSpCreateObjectApprovalTask) SetCreateTime(time int64) {
	m.GetTask().SetCreateTime(time)
}

func (m *GfSpCreateObjectApprovalTask) GetUpdateTime() int64 {
	return m.GetTask().GetUpdateTime()
}

func (m *GfSpCreateObjectApprovalTask) SetUpdateTime(time int64) {
	m.GetTask().SetUpdateTime(time)
}

func (m *GfSpCreateObjectApprovalTask) GetTimeout() int64 {
	return m.GetTask().GetTimeout()
}

func (m *GfSpCreateObjectApprovalTask) SetTimeout(time int64) {
	m.GetTask().SetTimeout(time)
}

func (m *GfSpCreateObjectApprovalTask) ExceedTimeout() bool {
	return m.GetTask().ExceedTimeout()
}

func (m *GfSpCreateObjectApprovalTask) GetRetry() int64 {
	return m.GetTask().GetRetry()
}

func (m *GfSpCreateObjectApprovalTask) IncRetry() {
	m.GetTask().IncRetry()
}

func (m *GfSpCreateObjectApprovalTask) SetRetry(retry int) {
	m.GetTask().SetRetry(retry)
}

func (m *GfSpCreateObjectApprovalTask) GetMaxRetry() int64 {
	return m.GetTask().GetMaxRetry()
}

func (m *GfSpCreateObjectApprovalTask) SetMaxRetry(limit int64) {
	m.GetTask().SetMaxRetry(limit)
}

func (m *GfSpCreateObjectApprovalTask) ExceedRetry() bool {
	return m.GetTask().ExceedRetry()
}

func (m *GfSpCreateObjectApprovalTask) Expired() bool {
	return m.GetTask().Expired()
}

func (m *GfSpCreateObjectApprovalTask) GetPriority() coretask.TPriority {
	return m.GetTask().GetPriority()
}

func (m *GfSpCreateObjectApprovalTask) SetPriority(priority coretask.TPriority) {
	m.GetTask().SetPriority(priority)
}

func (m *GfSpCreateObjectApprovalTask) EstimateLimit() corercmgr.Limit {
	l := &gfsplimit.GfSpLimit{}
	l.Add(LimitEstimateByPriority(m.GetPriority()))
	return l
}

func (m *GfSpCreateObjectApprovalTask) GetUserAddress() string {
	return m.GetTask().GetUserAddress()
}

func (m *GfSpCreateObjectApprovalTask) SetUserAddress(address string) {
	m.GetTask().SetUserAddress(address)
}

func (m *GfSpCreateObjectApprovalTask) SetLogs(logs string) {
	m.GetTask().SetLogs(logs)
}

func (m *GfSpCreateObjectApprovalTask) GetLogs() string {
	return m.GetTask().GetLogs()
}

func (m *GfSpCreateObjectApprovalTask) AppendLog(log string) {
	m.GetTask().AppendLog(log)
}

func (m *GfSpCreateObjectApprovalTask) Error() error {
	return m.GetTask().Error()
}

func (m *GfSpCreateObjectApprovalTask) SetError(err error) {
	m.GetTask().SetError(err)
}

func (m *GfSpCreateObjectApprovalTask) SetExpiredHeight(height uint64) {
	m.GetCreateObjectInfo().GetPrimarySpApproval().ExpiredHeight = height
}

func (m *GfSpCreateObjectApprovalTask) GetExpiredHeight() uint64 {
	return m.GetCreateObjectInfo().GetPrimarySpApproval().GetExpiredHeight()
}

func (m *GfSpCreateObjectApprovalTask) SetCreateObjectInfo(object *storagetypes.MsgCreateObject) {
	m.CreateObjectInfo = object
}

func (m *GfSpReplicatePieceApprovalTask) InitApprovalReplicatePieceTask(object *storagetypes.ObjectInfo,
	params *storagetypes.Params, priority coretask.TPriority, askOpAddress string) {
	m.Reset()
	m.Task = &GfSpTask{}
	m.SetStorageParams(params)
	m.GetTask().SetCreateTime(time.Now().Unix())
	m.GetTask().SetUpdateTime(time.Now().Unix())
	m.SetObjectInfo(object)
	m.SetPriority(priority)
	m.SetAskSpOperatorAddress(askOpAddress)
}

func (m *GfSpReplicatePieceApprovalTask) GetSignBytes() []byte {
	fakeMsg := &GfSpReplicatePieceApprovalTask{
		ObjectInfo:    m.GetObjectInfo(),
		StorageParams: m.GetStorageParams(),
		Task:          &GfSpTask{CreateTime: m.GetCreateTime()},
		ExpiredHeight: m.GetExpiredHeight(),
	}
	bz := ModuleCdc.MustMarshalJSON(fakeMsg)
	return sdk.MustSortJSON(bz)
}

func (m *GfSpReplicatePieceApprovalTask) Key() coretask.TKey {
	return GfSpReplicatePieceApprovalTaskKey(
		m.GetObjectInfo().GetBucketName(),
		m.GetObjectInfo().GetObjectName(),
		m.GetObjectInfo().Id.String())
}

func (m *GfSpReplicatePieceApprovalTask) Type() coretask.TType {
	return coretask.TypeTaskReplicatePieceApproval
}

func (m *GfSpReplicatePieceApprovalTask) Info() string {
	return fmt.Sprintf("key[%s], type[%s], priority[%d], limit[%s], expiedHeight[%d], %s",
		m.Key(), coretask.TaskTypeName(m.Type()), m.GetPriority(), m.EstimateLimit().String(),
		m.GetExpiredHeight(), m.GetTask().Info())
}

func (m *GfSpReplicatePieceApprovalTask) GetAddress() string {
	return m.GetTask().GetAddress()
}

func (m *GfSpReplicatePieceApprovalTask) SetAddress(address string) {
	m.GetTask().SetAddress(address)
}

func (m *GfSpReplicatePieceApprovalTask) GetCreateTime() int64 {
	return m.GetTask().GetCreateTime()
}

func (m *GfSpReplicatePieceApprovalTask) SetCreateTime(time int64) {
	m.GetTask().SetCreateTime(time)
}

func (m *GfSpReplicatePieceApprovalTask) GetUpdateTime() int64 {
	return m.GetTask().GetUpdateTime()
}

func (m *GfSpReplicatePieceApprovalTask) SetUpdateTime(time int64) {
	m.GetTask().SetUpdateTime(time)
}

func (m *GfSpReplicatePieceApprovalTask) GetTimeout() int64 {
	return m.GetTask().GetTimeout()
}

func (m *GfSpReplicatePieceApprovalTask) SetTimeout(time int64) {
	m.GetTask().SetTimeout(time)
}

func (m *GfSpReplicatePieceApprovalTask) ExceedTimeout() bool {
	return m.GetTask().ExceedTimeout()
}

func (m *GfSpReplicatePieceApprovalTask) GetRetry() int64 {
	return m.GetTask().GetRetry()
}

func (m *GfSpReplicatePieceApprovalTask) IncRetry() {
	m.GetTask().IncRetry()
}

func (m *GfSpReplicatePieceApprovalTask) SetRetry(retry int) {
	m.GetTask().SetRetry(retry)
}

func (m *GfSpReplicatePieceApprovalTask) GetMaxRetry() int64 {
	return m.GetTask().GetMaxRetry()
}

func (m *GfSpReplicatePieceApprovalTask) SetMaxRetry(limit int64) {
	m.GetTask().SetMaxRetry(limit)
}

func (m *GfSpReplicatePieceApprovalTask) ExceedRetry() bool {
	return m.GetTask().ExceedRetry()
}

func (m *GfSpReplicatePieceApprovalTask) Expired() bool {
	return m.GetTask().Expired()
}

func (m *GfSpReplicatePieceApprovalTask) GetPriority() coretask.TPriority {
	return m.GetTask().GetPriority()
}

func (m *GfSpReplicatePieceApprovalTask) SetPriority(priority coretask.TPriority) {
	m.GetTask().SetPriority(priority)
}

func (m *GfSpReplicatePieceApprovalTask) EstimateLimit() corercmgr.Limit {
	l := &gfsplimit.GfSpLimit{}
	l.Add(LimitEstimateByPriority(m.GetPriority()))
	return l
}

func (m *GfSpReplicatePieceApprovalTask) GetUserAddress() string {
	return m.GetTask().GetUserAddress()
}

func (m *GfSpReplicatePieceApprovalTask) SetUserAddress(address string) {
	m.GetTask().SetUserAddress(address)
}

func (m *GfSpReplicatePieceApprovalTask) SetLogs(logs string) {
	m.GetTask().SetLogs(logs)
}

func (m *GfSpReplicatePieceApprovalTask) GetLogs() string {
	return m.GetTask().GetLogs()
}

func (m *GfSpReplicatePieceApprovalTask) AppendLog(log string) {
	m.GetTask().AppendLog(log)
}

func (m *GfSpReplicatePieceApprovalTask) Error() error {
	return m.GetTask().Error()
}

func (m *GfSpReplicatePieceApprovalTask) SetError(err error) {
	m.GetTask().SetError(err)
}

func (m *GfSpReplicatePieceApprovalTask) SetExpiredHeight(height uint64) {
	m.ExpiredHeight = height
}

func (m *GfSpReplicatePieceApprovalTask) SetObjectInfo(object *storagetypes.ObjectInfo) {
	m.ObjectInfo = object
}

func (m *GfSpReplicatePieceApprovalTask) SetStorageParams(params *storagetypes.Params) {
	m.StorageParams = params
}

func (m *GfSpReplicatePieceApprovalTask) SetAskSpOperatorAddress(address string) {
	m.AskSpOperatorAddress = address
}

func (m *GfSpReplicatePieceApprovalTask) SetAskSignature(signature []byte) {
	m.AskSignature = signature
}

func (m *GfSpReplicatePieceApprovalTask) SetApprovedSpOperatorAddress(address string) {
	m.ApprovedSpOperatorAddress = address
}

func (m *GfSpReplicatePieceApprovalTask) SetApprovedSignature(signature []byte) {
	m.ApprovedSignature = signature
}

func (m *GfSpReplicatePieceApprovalTask) SetApprovedSpEndpoint(endpoint string) {
	m.ApprovedSpEndpoint = endpoint
}

func (m *GfSpReplicatePieceApprovalTask) SetApprovedSpApprovalAddress(address string) {
	m.ApprovedSpApprovalAddress = address
}

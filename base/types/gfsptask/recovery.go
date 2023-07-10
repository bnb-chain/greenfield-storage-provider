package gfsptask

import (
	"fmt"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsplimit"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (m *GfSpRecoverPieceTask) InitRecoverPieceTask(object *storagetypes.ObjectInfo, params *storagetypes.Params,
	priority coretask.TPriority, segmentIdx uint32, ecIdx int32, pieceSize uint64, timeout int64, retry int64) {
	m.Reset()
	m.Task = &GfSpTask{}
	m.SetCreateTime(time.Now().Unix())
	m.SetUpdateTime(time.Now().Unix())
	m.SetObjectInfo(object)
	m.SetStorageParams(params)
	m.SetPriority(priority)
	m.SetTimeout(timeout)
	m.SetMaxRetry(retry)
	m.SetSegmentIndex(segmentIdx)
	m.SetECIndex(ecIdx)
	m.SetPieceSize(pieceSize)
}

func (m *GfSpRecoverPieceTask) Key() coretask.TKey {
	return GfSpRecoverPieceTaskKey(
		m.GetObjectInfo().GetBucketName(),
		m.GetObjectInfo().GetObjectName(),
		m.GetObjectInfo().Id.String(),
		m.GetSegmentIdx(),
		m.GetEcIdx(),
		m.GetCreateTime(),
	)
}

func (m *GfSpRecoverPieceTask) Type() coretask.TType {
	return coretask.TypeTaskRecoverPiece
}

func (m *GfSpRecoverPieceTask) Info() string {
	return fmt.Sprintf("key[%s], type[%s], priority[%d], piece index[%d], %s",
		m.Key(), coretask.TaskTypeName(m.Type()), m.GetPriority(),
		m.GetSegmentIdx(), m.GetTask().Info())
}

func (m *GfSpRecoverPieceTask) GetAddress() string {
	return m.GetTask().GetAddress()
}

func (m *GfSpRecoverPieceTask) SetAddress(address string) {
	m.GetTask().SetAddress(address)
}

func (m *GfSpRecoverPieceTask) GetCreateTime() int64 {
	return m.GetTask().GetCreateTime()
}

func (m *GfSpRecoverPieceTask) SetCreateTime(time int64) {
	m.GetTask().SetCreateTime(time)
}

func (m *GfSpRecoverPieceTask) GetUpdateTime() int64 {
	return m.GetTask().GetUpdateTime()
}

func (m *GfSpRecoverPieceTask) SetUpdateTime(time int64) {
	m.GetTask().SetUpdateTime(time)
}

func (m *GfSpRecoverPieceTask) GetTimeout() int64 {
	return m.GetTask().GetTimeout()
}

func (m *GfSpRecoverPieceTask) SetTimeout(time int64) {
	m.GetTask().SetTimeout(time)
}

func (m *GfSpRecoverPieceTask) ExceedTimeout() bool {
	return m.GetTask().ExceedTimeout()
}

func (m *GfSpRecoverPieceTask) GetRetry() int64 {
	return m.GetTask().GetRetry()
}

func (m *GfSpRecoverPieceTask) IncRetry() {
	m.GetTask().IncRetry()
}

func (m *GfSpRecoverPieceTask) SetRetry(retry int) {
	m.GetTask().SetRetry(retry)
}

func (m *GfSpRecoverPieceTask) GetMaxRetry() int64 {
	return m.GetTask().GetMaxRetry()
}

func (m *GfSpRecoverPieceTask) SetMaxRetry(limit int64) {
	m.GetTask().SetMaxRetry(limit)
}

func (m *GfSpRecoverPieceTask) SetSegmentIndex(index uint32) {
	m.SegmentIdx = index
}

func (m *GfSpRecoverPieceTask) SetECIndex(index int32) {
	m.EcIdx = index
}

func (m *GfSpRecoverPieceTask) ExceedRetry() bool {
	return m.GetTask().ExceedRetry()
}

func (m *GfSpRecoverPieceTask) Expired() bool {
	return m.GetTask().Expired()
}

func (m *GfSpRecoverPieceTask) GetPriority() coretask.TPriority {
	return m.GetTask().GetPriority()
}

func (m *GfSpRecoverPieceTask) SetPriority(priority coretask.TPriority) {
	m.GetTask().SetPriority(priority)
}

func (m *GfSpRecoverPieceTask) SetObjectInfo(object *storagetypes.ObjectInfo) {
	m.ObjectInfo = object
}

func (m *GfSpRecoverPieceTask) SetStorageParams(param *storagetypes.Params) {
	m.StorageParams = param
}

func (m *GfSpRecoverPieceTask) GetUserAddress() string {
	return m.GetTask().GetUserAddress()
}

func (m *GfSpRecoverPieceTask) SetUserAddress(address string) {
	m.GetTask().SetUserAddress(address)
}

func (m *GfSpRecoverPieceTask) SetLogs(logs string) {
	m.GetTask().SetLogs(logs)
}

func (m *GfSpRecoverPieceTask) GetLogs() string {
	return m.GetTask().GetLogs()
}

func (m *GfSpRecoverPieceTask) AppendLog(log string) {
	m.GetTask().AppendLog(log)
}

func (m *GfSpRecoverPieceTask) Error() error {
	return m.GetTask().Error()
}

func (m *GfSpRecoverPieceTask) SetError(err error) {
	m.GetTask().SetError(err)
}

func (m *GfSpRecoverPieceTask) EstimateLimit() corercmgr.Limit {
	l := &gfsplimit.GfSpLimit{Memory: 2 * int64(m.GetObjectInfo().PayloadSize)}
	l.Add(LimitEstimateByPriority(m.GetPriority()))
	return l
}

func (m *GfSpRecoverPieceTask) SetSignature(signature []byte) {
	m.Signature = signature
}

func (m *GfSpRecoverPieceTask) SetPieceSize(size uint64) {
	m.PieceSize = size
}

func (m *GfSpRecoverPieceTask) SetRecoverDone() {
	m.Recovered = true
}

func (m *GfSpRecoverPieceTask) GetSignBytes() []byte {
	fakeMsg := &GfSpRecoverPieceTask{
		ObjectInfo:    m.GetObjectInfo(),
		StorageParams: m.GetStorageParams(),
		Task:          &GfSpTask{CreateTime: m.GetCreateTime()},
		PieceSize:     m.GetPieceSize(),
		SegmentIdx:    m.GetSegmentIdx(),
		// TODO rename EcIdx to ReplicateIdx
		EcIdx: m.GetEcIdx(),
	}
	bz := ModuleCdc.MustMarshalJSON(fakeMsg)
	return sdk.MustSortJSON(bz)
}

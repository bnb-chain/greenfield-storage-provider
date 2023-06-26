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

func (m *GfSpRecoveryPieceTask) InitRecoveryPieceTask(object *storagetypes.ObjectInfo, params *storagetypes.Params,
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

func (m *GfSpRecoveryPieceTask) Key() coretask.TKey {
	return GfSpRecoveryPieceTaskKey(
		m.GetObjectInfo().GetBucketName(),
		m.GetObjectInfo().GetObjectName(),
		m.GetObjectInfo().Id.String(),
		m.GetSegmentIdx(),
		m.GetEcIdx(),
		m.GetCreateTime(),
	)
}

func (m *GfSpRecoveryPieceTask) Type() coretask.TType {
	return coretask.TypeTaskRecoverPiece
}

func (m *GfSpRecoveryPieceTask) Info() string {
	return fmt.Sprintf("key[%s], type[%s], priority[%d], piece index[%d], %s",
		m.Key(), coretask.TaskTypeName(m.Type()), m.GetPriority(),
		m.GetSegmentIdx(), m.GetTask().Info())
}

func (m *GfSpRecoveryPieceTask) GetAddress() string {
	return m.GetTask().GetAddress()
}

func (m *GfSpRecoveryPieceTask) SetAddress(address string) {
	m.GetTask().SetAddress(address)
}

func (m *GfSpRecoveryPieceTask) GetCreateTime() int64 {
	return m.GetTask().GetCreateTime()
}

func (m *GfSpRecoveryPieceTask) SetCreateTime(time int64) {
	m.GetTask().SetCreateTime(time)
}

func (m *GfSpRecoveryPieceTask) GetUpdateTime() int64 {
	return m.GetTask().GetUpdateTime()
}

func (m *GfSpRecoveryPieceTask) SetUpdateTime(time int64) {
	m.GetTask().SetUpdateTime(time)
}

func (m *GfSpRecoveryPieceTask) GetTimeout() int64 {
	return m.GetTask().GetTimeout()
}

func (m *GfSpRecoveryPieceTask) SetTimeout(time int64) {
	m.GetTask().SetTimeout(time)
}

func (m *GfSpRecoveryPieceTask) ExceedTimeout() bool {
	return m.GetTask().ExceedTimeout()
}

func (m *GfSpRecoveryPieceTask) GetRetry() int64 {
	return m.GetTask().GetRetry()
}

func (m *GfSpRecoveryPieceTask) IncRetry() {
	m.GetTask().IncRetry()
}

func (m *GfSpRecoveryPieceTask) SetRetry(retry int) {
	m.GetTask().SetRetry(retry)
}

func (m *GfSpRecoveryPieceTask) GetMaxRetry() int64 {
	return m.GetTask().GetMaxRetry()
}

func (m *GfSpRecoveryPieceTask) SetMaxRetry(limit int64) {
	m.GetTask().SetMaxRetry(limit)
}

func (m *GfSpRecoveryPieceTask) SetSegmentIndex(index uint32) {
	m.SegmentIdx = index
}

func (m *GfSpRecoveryPieceTask) SetECIndex(index int32) {
	m.EcIdx = index
}

func (m *GfSpRecoveryPieceTask) ExceedRetry() bool {
	return m.GetTask().ExceedRetry()
}

func (m *GfSpRecoveryPieceTask) Expired() bool {
	return m.GetTask().Expired()
}

func (m *GfSpRecoveryPieceTask) GetPriority() coretask.TPriority {
	return m.GetTask().GetPriority()
}

func (m *GfSpRecoveryPieceTask) SetPriority(priority coretask.TPriority) {
	m.GetTask().SetPriority(priority)
}

func (m *GfSpRecoveryPieceTask) SetObjectInfo(object *storagetypes.ObjectInfo) {
	m.ObjectInfo = object
}

func (m *GfSpRecoveryPieceTask) SetStorageParams(param *storagetypes.Params) {
	m.StorageParams = param
}

func (m *GfSpRecoveryPieceTask) Error() error {
	return m.GetTask().Error()
}

func (m *GfSpRecoveryPieceTask) SetError(err error) {
	m.GetTask().SetError(err)
}

func (m *GfSpRecoveryPieceTask) EstimateLimit() corercmgr.Limit {
	l := &gfsplimit.GfSpLimit{Memory: 2 * int64(m.GetObjectInfo().PayloadSize)}
	l.Add(LimitEstimateByPriority(m.GetPriority()))
	return l
}

func (m *GfSpRecoveryPieceTask) SetSignature(signature []byte) {
	m.Signature = signature
}

func (m *GfSpRecoveryPieceTask) SetPieceSize(size uint64) {
	m.PieceSize = size
}

func (m *GfSpRecoveryPieceTask) SetRecoverDone() {
	m.Recovered = true
}

func (m *GfSpRecoveryPieceTask) GetSignBytes() []byte {
	fakeMsg := &GfSpRecoveryPieceTask{
		ObjectInfo:    m.GetObjectInfo(),
		StorageParams: m.GetStorageParams(),
		Task:          &GfSpTask{CreateTime: m.GetCreateTime()},
		PieceSize:     m.GetPieceSize(),
		SegmentIdx:    m.GetSegmentIdx(),
		EcIdx:         m.GetEcIdx(),
	}
	bz := ModuleCdc.MustMarshalJSON(fakeMsg)
	return sdk.MustSortJSON(bz)
}

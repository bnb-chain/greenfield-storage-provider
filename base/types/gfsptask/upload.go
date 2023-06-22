package gfsptask

import (
	"encoding/hex"
	"fmt"
	"math"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsplimit"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var _ coretask.UploadObjectTask = &GfSpUploadObjectTask{}
var _ coretask.ReplicatePieceTask = &GfSpReplicatePieceTask{}
var _ coretask.SealObjectTask = &GfSpSealObjectTask{}
var _ coretask.ReceivePieceTask = &GfSpReceivePieceTask{}

func (m *GfSpUploadObjectTask) InitUploadObjectTask(vgfID uint32, object *storagetypes.ObjectInfo, params *storagetypes.Params, timeout int64) {
	m.Reset()
	m.Task = &GfSpTask{}
	m.VirtualGroupFamilyId = vgfID
	m.SetCreateTime(time.Now().Unix())
	m.SetUpdateTime(time.Now().Unix())
	m.SetTimeout(timeout)
	m.SetObjectInfo(object)
	m.SetStorageParams(params)
}

func (m *GfSpUploadObjectTask) Key() coretask.TKey {
	return GfSpUploadObjectTaskKey(
		m.GetObjectInfo().GetBucketName(),
		m.GetObjectInfo().GetObjectName(),
		m.GetObjectInfo().Id.String())
}

func (m *GfSpUploadObjectTask) Type() coretask.TType {
	return coretask.TypeTaskUpload
}

func (m *GfSpUploadObjectTask) Info() string {
	return fmt.Sprintf("key[%s], type[%s], priority[%d], limit[%s], %s",
		m.Key(), coretask.TaskTypeName(m.Type()), m.GetPriority(),
		m.EstimateLimit().String(), m.GetTask().Info())
}

func (m *GfSpUploadObjectTask) GetAddress() string {
	return m.GetTask().GetAddress()
}

func (m *GfSpUploadObjectTask) SetAddress(address string) {
	m.GetTask().SetAddress(address)
}

func (m *GfSpUploadObjectTask) GetCreateTime() int64 {
	return m.GetTask().GetCreateTime()
}

func (m *GfSpUploadObjectTask) SetCreateTime(time int64) {
	m.GetTask().SetCreateTime(time)
}

func (m *GfSpUploadObjectTask) GetUpdateTime() int64 {
	return m.GetTask().GetUpdateTime()
}

func (m *GfSpUploadObjectTask) SetUpdateTime(time int64) {
	m.GetTask().SetUpdateTime(time)
}

func (m *GfSpUploadObjectTask) GetTimeout() int64 {
	return m.GetTask().GetTimeout()
}

func (m *GfSpUploadObjectTask) SetTimeout(time int64) {
	m.GetTask().SetTimeout(time)
}

func (m *GfSpUploadObjectTask) ExceedTimeout() bool {
	return m.GetTask().ExceedTimeout()
}

func (m *GfSpUploadObjectTask) GetRetry() int64 {
	return m.GetTask().GetRetry()
}

func (m *GfSpUploadObjectTask) IncRetry() {
	m.GetTask().IncRetry()
}

func (m *GfSpUploadObjectTask) SetRetry(retry int) {
	m.GetTask().SetRetry(retry)
}

func (m *GfSpUploadObjectTask) GetMaxRetry() int64 {
	return m.GetTask().GetMaxRetry()
}

func (m *GfSpUploadObjectTask) SetMaxRetry(limit int64) {
	m.GetTask().SetMaxRetry(limit)
}

func (m *GfSpUploadObjectTask) ExceedRetry() bool {
	return m.GetTask().ExceedRetry()
}

func (m *GfSpUploadObjectTask) Expired() bool {
	return m.GetTask().Expired()
}

func (m *GfSpUploadObjectTask) GetPriority() coretask.TPriority {
	return m.GetTask().GetPriority()
}

func (m *GfSpUploadObjectTask) SetPriority(priority coretask.TPriority) {
	m.GetTask().SetPriority(priority)
}

func (m *GfSpUploadObjectTask) EstimateLimit() corercmgr.Limit {
	l := &gfsplimit.GfSpLimit{}
	if m.GetObjectInfo().GetPayloadSize() >= m.GetStorageParams().VersionedParams.GetMaxSegmentSize() {
		l.Memory = int64(m.GetStorageParams().VersionedParams.GetMaxSegmentSize()) * 2
	} else {
		l.Memory = int64(m.GetObjectInfo().GetPayloadSize()) * 2
	}
	l.Add(LimitEstimateByPriority(m.GetPriority()))
	return l
}

func (m *GfSpUploadObjectTask) Error() error {
	return m.GetTask().Error()
}

func (m *GfSpUploadObjectTask) SetError(err error) {
	m.GetTask().SetError(err)
}

func (m *GfSpUploadObjectTask) SetObjectInfo(object *storagetypes.ObjectInfo) {
	m.ObjectInfo = object
}

func (m *GfSpUploadObjectTask) SetStorageParams(param *storagetypes.Params) {
	m.StorageParams = param
}

func (m *GfSpReplicatePieceTask) InitReplicatePieceTask(object *storagetypes.ObjectInfo, params *storagetypes.Params,
	priority coretask.TPriority, timeout int64, retry int64) {
	m.Reset()
	m.Task = &GfSpTask{}
	m.SetCreateTime(time.Now().Unix())
	m.SetUpdateTime(time.Now().Unix())
	m.SetObjectInfo(object)
	m.SetStorageParams(params)
	m.SetPriority(priority)
	m.SetTimeout(timeout)
	m.SetMaxRetry(retry)
}

func (m *GfSpReplicatePieceTask) Key() coretask.TKey {
	return GfSpReplicatePieceTaskKey(
		m.GetObjectInfo().GetBucketName(),
		m.GetObjectInfo().GetObjectName(),
		m.GetObjectInfo().Id.String())
}

func (m *GfSpReplicatePieceTask) Type() coretask.TType {
	return coretask.TypeTaskReplicatePiece
}

func (m *GfSpReplicatePieceTask) Info() string {
	return fmt.Sprintf("key[%s], type[%s], priority[%d], limit[%s], %s",
		m.Key(), coretask.TaskTypeName(m.Type()), m.GetPriority(),
		m.EstimateLimit().String(), m.GetTask().Info())
}

func (m *GfSpReplicatePieceTask) GetAddress() string {
	return m.GetTask().GetAddress()
}

func (m *GfSpReplicatePieceTask) SetAddress(address string) {
	m.GetTask().SetAddress(address)
}

func (m *GfSpReplicatePieceTask) GetCreateTime() int64 {
	return m.GetTask().GetCreateTime()
}

func (m *GfSpReplicatePieceTask) SetCreateTime(time int64) {
	m.GetTask().SetCreateTime(time)
}

func (m *GfSpReplicatePieceTask) GetUpdateTime() int64 {
	return m.GetTask().GetUpdateTime()
}

func (m *GfSpReplicatePieceTask) SetUpdateTime(time int64) {
	m.GetTask().SetUpdateTime(time)
}

func (m *GfSpReplicatePieceTask) GetTimeout() int64 {
	return m.GetTask().GetTimeout()
}

func (m *GfSpReplicatePieceTask) SetTimeout(time int64) {
	m.GetTask().SetTimeout(time)
}

func (m *GfSpReplicatePieceTask) ExceedTimeout() bool {
	return m.GetTask().ExceedTimeout()
}

func (m *GfSpReplicatePieceTask) GetRetry() int64 {
	return m.GetTask().GetRetry()
}

func (m *GfSpReplicatePieceTask) IncRetry() {
	m.GetTask().IncRetry()
}

func (m *GfSpReplicatePieceTask) SetRetry(retry int) {
	m.GetTask().SetRetry(retry)
}

func (m *GfSpReplicatePieceTask) GetMaxRetry() int64 {
	return m.GetTask().GetMaxRetry()
}

func (m *GfSpReplicatePieceTask) SetMaxRetry(limit int64) {
	m.GetTask().SetMaxRetry(limit)
}

func (m *GfSpReplicatePieceTask) ExceedRetry() bool {
	return m.GetTask().ExceedRetry()
}

func (m *GfSpReplicatePieceTask) Expired() bool {
	return m.GetTask().Expired()
}

func (m *GfSpReplicatePieceTask) GetPriority() coretask.TPriority {
	return m.GetTask().GetPriority()
}

func (m *GfSpReplicatePieceTask) SetPriority(priority coretask.TPriority) {
	m.GetTask().SetPriority(priority)
}

func (m *GfSpReplicatePieceTask) EstimateLimit() corercmgr.Limit {
	l := &gfsplimit.GfSpLimit{}
	if m.GetObjectInfo().GetRedundancyType() == storagetypes.REDUNDANCY_REPLICA_TYPE {
		if m.GetObjectInfo().GetPayloadSize() < m.GetStorageParams().GetMaxPayloadSize() {
			l.Memory = int64(m.GetObjectInfo().GetPayloadSize())
		} else {
			l.Memory = int64(m.GetStorageParams().VersionedParams.GetMaxSegmentSize())
		}
	} else {
		if m.GetObjectInfo().GetPayloadSize() < m.GetStorageParams().GetMaxPayloadSize() {
			size := float64(m.GetStorageParams().VersionedParams.GetMaxSegmentSize()) *
				(float64(m.GetStorageParams().VersionedParams.GetRedundantDataChunkNum()) +
					float64(m.GetStorageParams().VersionedParams.GetRedundantParityChunkNum())) /
				float64(m.GetStorageParams().VersionedParams.GetRedundantDataChunkNum())
			l.Memory = int64(math.Ceil(size))
		} else {
			// it is an estimation method, within a few bytes of error
			size := float64(m.GetObjectInfo().GetPayloadSize()) *
				(float64(m.GetStorageParams().VersionedParams.GetRedundantDataChunkNum()) +
					float64(m.GetStorageParams().VersionedParams.GetRedundantParityChunkNum())) /
				float64(m.GetStorageParams().VersionedParams.GetRedundantDataChunkNum())
			l.Memory = int64(math.Ceil(size))
		}
	}
	l.Add(LimitEstimateByPriority(m.GetPriority()))
	return l
}

func (m *GfSpReplicatePieceTask) SetSealed(sealed bool) {
	m.Sealed = sealed
}

func (m *GfSpReplicatePieceTask) SetSecondarySignatures(signatures [][]byte) {
	m.SecondarySignatures = signatures
}

func (m *GfSpReplicatePieceTask) SetSecondaryAddresses(addresses []string) {
	m.SecondaryAddresses = addresses
}

func (m *GfSpReplicatePieceTask) SetObjectInfo(object *storagetypes.ObjectInfo) {
	m.ObjectInfo = object
}

func (m *GfSpReplicatePieceTask) SetStorageParams(param *storagetypes.Params) {
	m.StorageParams = param
}

func (m *GfSpReplicatePieceTask) Error() error {
	return m.GetTask().Error()
}

func (m *GfSpReplicatePieceTask) SetError(err error) {
	m.GetTask().SetError(err)
}

func (m *GfSpSealObjectTask) InitSealObjectTask(vgfID uint32, object *storagetypes.ObjectInfo, params *storagetypes.Params, priority coretask.TPriority,
	endpoints []string, signatures [][]byte, timeout int64, retry int64) {
	m.Reset()
	m.Task = &GfSpTask{}
	m.GlobalVirtualGroupId = vgfID
	m.SecondaryEndpoints = endpoints
	m.SetCreateTime(time.Now().Unix())
	m.SetUpdateTime(time.Now().Unix())
	m.SetObjectInfo(object)
	m.SetStorageParams(params)
	m.SetPriority(priority)
	m.SetTimeout(timeout)
	m.SetMaxRetry(retry)
	m.SetSecondarySignatures(signatures)
}

func (m *GfSpSealObjectTask) Key() coretask.TKey {
	return GfSpSealObjectTaskKey(
		m.GetObjectInfo().GetBucketName(),
		m.GetObjectInfo().GetObjectName(),
		m.GetObjectInfo().Id.String())
}

func (m *GfSpSealObjectTask) Type() coretask.TType {
	return coretask.TypeTaskSealObject
}

func (m *GfSpSealObjectTask) Info() string {
	return fmt.Sprintf("key[%s], type[%s], priority[%d], limit[%s], %s",
		m.Key(), coretask.TaskTypeName(m.Type()), m.GetPriority(),
		m.EstimateLimit().String(), m.GetTask().Info())
}

func (m *GfSpSealObjectTask) GetAddress() string {
	return m.GetTask().GetAddress()
}

func (m *GfSpSealObjectTask) SetAddress(address string) {
	m.GetTask().SetAddress(address)
}

func (m *GfSpSealObjectTask) GetCreateTime() int64 {
	return m.GetTask().GetCreateTime()
}

func (m *GfSpSealObjectTask) SetCreateTime(time int64) {
	m.GetTask().SetCreateTime(time)
}

func (m *GfSpSealObjectTask) GetUpdateTime() int64 {
	return m.GetTask().GetUpdateTime()
}

func (m *GfSpSealObjectTask) SetUpdateTime(time int64) {
	m.GetTask().SetUpdateTime(time)
}

func (m *GfSpSealObjectTask) GetTimeout() int64 {
	return m.GetTask().GetTimeout()
}

func (m *GfSpSealObjectTask) SetTimeout(time int64) {
	m.GetTask().SetTimeout(time)
}

func (m *GfSpSealObjectTask) ExceedTimeout() bool {
	return m.GetTask().ExceedTimeout()
}

func (m *GfSpSealObjectTask) GetRetry() int64 {
	return m.GetTask().GetRetry()
}

func (m *GfSpSealObjectTask) IncRetry() {
	m.GetTask().IncRetry()
}

func (m *GfSpSealObjectTask) SetRetry(retry int) {
	m.GetTask().SetRetry(retry)
}

func (m *GfSpSealObjectTask) GetMaxRetry() int64 {
	return m.GetTask().GetMaxRetry()
}

func (m *GfSpSealObjectTask) SetMaxRetry(limit int64) {
	m.GetTask().SetMaxRetry(limit)
}

func (m *GfSpSealObjectTask) ExceedRetry() bool {
	return m.GetTask().ExceedRetry()
}

func (m *GfSpSealObjectTask) Expired() bool {
	return m.GetTask().Expired()
}

func (m *GfSpSealObjectTask) GetPriority() coretask.TPriority {
	return m.GetTask().GetPriority()
}

func (m *GfSpSealObjectTask) SetPriority(priority coretask.TPriority) {
	m.GetTask().SetPriority(priority)
}

func (m *GfSpSealObjectTask) EstimateLimit() corercmgr.Limit {
	l := &gfsplimit.GfSpLimit{}
	l.Add(LimitEstimateByPriority(m.GetPriority()))
	return l
}

func (m *GfSpSealObjectTask) SetSecondaryAddresses(addresses []string) {
	m.SecondaryAddresses = addresses
}

func (m *GfSpSealObjectTask) SetSecondarySignatures(signatures [][]byte) {
	m.SecondarySignatures = signatures
}

func (m *GfSpSealObjectTask) SetObjectInfo(object *storagetypes.ObjectInfo) {
	m.ObjectInfo = object
}

func (m *GfSpSealObjectTask) SetStorageParams(param *storagetypes.Params) {
	m.StorageParams = param
}

func (m *GfSpSealObjectTask) Error() error {
	return m.GetTask().Error()
}

func (m *GfSpSealObjectTask) SetError(err error) {
	m.GetTask().SetError(err)
}

func (m *GfSpReceivePieceTask) InitReceivePieceTask(gvgID uint32, object *storagetypes.ObjectInfo, params *storagetypes.Params,
	priority coretask.TPriority, replicateIdx uint32, pieceIdx int32, pieceSize int64) {
	m.Reset()
	m.Task = &GfSpTask{}
	m.GlobalVirtualGroupId = gvgID
	m.SetCreateTime(time.Now().Unix())
	m.SetUpdateTime(time.Now().Unix())
	m.SetObjectInfo(object)
	m.SetStorageParams(params)
	m.SetPriority(priority)
	m.SetReplicateIdx(replicateIdx)
	m.SetPieceIdx(pieceIdx)
	m.SetPieceSize(pieceSize)
}

func (m *GfSpReceivePieceTask) GetSignBytes() []byte {
	fakeMsg := &GfSpReceivePieceTask{
		ObjectInfo:    m.GetObjectInfo(),
		StorageParams: m.GetStorageParams(),
		Task:          &GfSpTask{CreateTime: m.GetCreateTime()},
		ReplicateIdx:  m.GetReplicateIdx(),
		PieceIdx:      m.GetPieceIdx(),
		PieceSize:     m.GetPieceSize(),
		PieceChecksum: m.GetPieceChecksum(),
	}
	bz := ModuleCdc.MustMarshalJSON(fakeMsg)
	return sdk.MustSortJSON(bz)
}

func (m *GfSpReceivePieceTask) Key() coretask.TKey {
	return GfSpReceivePieceTaskKey(
		m.GetObjectInfo().GetBucketName(),
		m.GetObjectInfo().GetObjectName(),
		m.GetObjectInfo().Id.String(),
		m.GetReplicateIdx(),
		m.GetPieceIdx())
}

func (m *GfSpReceivePieceTask) Type() coretask.TType {
	return coretask.TypeTaskReceivePiece
}

func (m *GfSpReceivePieceTask) Info() string {
	return fmt.Sprintf("key[%s], type[%s], priority[%d], limit[%s] rIdx[%d], pIdx[%d], size[%d], checksum[%s], %s",
		m.Key(), coretask.TaskTypeName(m.Type()), m.GetPriority(), m.EstimateLimit().String(), m.GetReplicateIdx(),
		m.GetPieceIdx(), m.GetPieceSize(), hex.EncodeToString(m.GetPieceChecksum()), m.GetTask().Info())
}

func (m *GfSpReceivePieceTask) GetAddress() string {
	return m.GetTask().GetAddress()
}

func (m *GfSpReceivePieceTask) SetAddress(address string) {
	m.GetTask().SetAddress(address)
}

func (m *GfSpReceivePieceTask) GetCreateTime() int64 {
	return m.GetTask().GetCreateTime()
}

func (m *GfSpReceivePieceTask) SetCreateTime(time int64) {
	m.GetTask().SetCreateTime(time)
}

func (m *GfSpReceivePieceTask) GetUpdateTime() int64 {
	return m.GetTask().GetUpdateTime()
}

func (m *GfSpReceivePieceTask) SetUpdateTime(time int64) {
	m.GetTask().SetUpdateTime(time)
}

func (m *GfSpReceivePieceTask) GetTimeout() int64 {
	return m.GetTask().GetTimeout()
}

func (m *GfSpReceivePieceTask) SetTimeout(time int64) {
	m.GetTask().SetTimeout(time)
}

func (m *GfSpReceivePieceTask) ExceedTimeout() bool {
	return m.GetTask().ExceedTimeout()
}

func (m *GfSpReceivePieceTask) GetRetry() int64 {
	return m.GetTask().GetRetry()
}

func (m *GfSpReceivePieceTask) IncRetry() {
	m.GetTask().IncRetry()
}

func (m *GfSpReceivePieceTask) SetRetry(retry int) {
	m.GetTask().SetRetry(retry)
}

func (m *GfSpReceivePieceTask) GetMaxRetry() int64 {
	return m.GetTask().GetMaxRetry()
}

func (m *GfSpReceivePieceTask) SetMaxRetry(limit int64) {
	m.GetTask().SetMaxRetry(limit)
}

func (m *GfSpReceivePieceTask) ExceedRetry() bool {
	return m.GetTask().ExceedRetry()
}

func (m *GfSpReceivePieceTask) Expired() bool {
	return m.GetTask().Expired()
}

func (m *GfSpReceivePieceTask) GetPriority() coretask.TPriority {
	return m.GetTask().GetPriority()
}

func (m *GfSpReceivePieceTask) SetPriority(priority coretask.TPriority) {
	m.GetTask().SetPriority(priority)
}

func (m *GfSpReceivePieceTask) EstimateLimit() corercmgr.Limit {
	l := &gfsplimit.GfSpLimit{Memory: m.GetPieceSize()}
	l.Add(LimitEstimateByPriority(m.GetPriority()))
	return l
}

func (m *GfSpReceivePieceTask) SetObjectInfo(object *storagetypes.ObjectInfo) {
	m.ObjectInfo = object
}

func (m *GfSpReceivePieceTask) SetStorageParams(param *storagetypes.Params) {
	m.StorageParams = param
}

func (m *GfSpReceivePieceTask) Error() error {
	return m.GetTask().Error()
}

func (m *GfSpReceivePieceTask) SetError(err error) {
	m.GetTask().SetError(err)
}

func (m *GfSpReceivePieceTask) SetReplicateIdx(idx uint32) {
	m.ReplicateIdx = idx
}

func (m *GfSpReceivePieceTask) SetPieceIdx(idx int32) {
	m.PieceIdx = idx
}

func (m *GfSpReceivePieceTask) SetPieceSize(size int64) {
	m.PieceSize = size
}

func (m *GfSpReceivePieceTask) SetPieceChecksum(checksum []byte) {
	m.PieceChecksum = checksum
}

func (m *GfSpReceivePieceTask) SetSealed(seal bool) {
	m.Sealed = seal
}

func (m *GfSpReceivePieceTask) SetSignature(signature []byte) {
	m.Signature = signature
}

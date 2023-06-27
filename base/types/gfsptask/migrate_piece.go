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

func (g *GfSpMigratePieceTask) InitMigratePieceTask(object *storagetypes.ObjectInfo, params *storagetypes.Params, priority coretask.TPriority,
	segmentIdx uint32, ecIdx int32, pieceSize uint64, timeout, retry int64) {
	g.Reset()
	g.Task = &GfSpTask{}
	g.SetCreateTime(time.Now().Unix())
	g.SetUpdateTime(time.Now().Unix())
	g.SetObjectInfo(object)
	g.SetStorageParams(params)
	g.SetPriority(priority)
	g.SetTimeout(timeout)
	g.SetMaxRetry(retry)
	g.SetSegmentIndex(segmentIdx)
	g.SetECIndex(ecIdx)
	g.SetPieceSize(pieceSize)
}

func (g *GfSpMigratePieceTask) Key() coretask.TKey {
	return GfSpMigratePieceTaskKey(
		g.GetObjectInfo().GetBucketName(),
		g.GetObjectInfo().GetObjectName(),
		g.GetObjectInfo().Id.String(),
		g.GetSegmentIdx(),
		g.GetEcIdx(),
		g.GetCreateTime(),
	)
}

func (g *GfSpMigratePieceTask) Type() coretask.TType {
	return coretask.TypeTaskMigratePiece
}

func (g *GfSpMigratePieceTask) Info() string {
	return fmt.Sprintf("key[%s], type[%s], priority[%d], segment index[%d], %s",
		g.Key(), coretask.TaskTypeName(g.Type()), g.GetPriority(),
		g.GetSegmentIdx(), g.GetTask().Info())
}

func (g *GfSpMigratePieceTask) GetAddress() string {
	return g.GetTask().GetAddress()
}

func (g *GfSpMigratePieceTask) SetAddress(address string) {
	g.GetTask().SetAddress(address)
}

func (g *GfSpMigratePieceTask) GetCreateTime() int64 {
	return g.GetTask().GetCreateTime()
}

func (g *GfSpMigratePieceTask) SetCreateTime(time int64) {
	g.GetTask().SetCreateTime(time)
}

func (g *GfSpMigratePieceTask) GetUpdateTime() int64 {
	return g.GetTask().GetUpdateTime()
}

func (g *GfSpMigratePieceTask) SetUpdateTime(time int64) {
	g.GetTask().SetUpdateTime(time)
}

func (g *GfSpMigratePieceTask) GetTimeout() int64 {
	return g.GetTask().GetTimeout()
}

func (g *GfSpMigratePieceTask) SetTimeout(time int64) {
	g.GetTask().SetTimeout(time)
}

func (g *GfSpMigratePieceTask) ExceedTimeout() bool {
	return g.GetTask().ExceedTimeout()
}

func (g *GfSpMigratePieceTask) GetRetry() int64 {
	return g.GetTask().GetRetry()
}

func (g *GfSpMigratePieceTask) IncRetry() {
	g.GetTask().IncRetry()
}

func (g *GfSpMigratePieceTask) SetRetry(retry int) {
	g.GetTask().SetRetry(retry)
}

func (g *GfSpMigratePieceTask) GetMaxRetry() int64 {
	return g.GetTask().GetMaxRetry()
}

func (g *GfSpMigratePieceTask) SetMaxRetry(limit int64) {
	g.GetTask().SetMaxRetry(limit)
}

func (g *GfSpMigratePieceTask) SetSegmentIndex(index uint32) {
	g.SegmentIdx = index
}

func (g *GfSpMigratePieceTask) SetECIndex(index int32) {
	g.EcIdx = index
}

func (g *GfSpMigratePieceTask) ExceedRetry() bool {
	return g.GetTask().ExceedRetry()
}

func (g *GfSpMigratePieceTask) Expired() bool {
	return g.GetTask().Expired()
}

func (g *GfSpMigratePieceTask) GetPriority() coretask.TPriority {
	return g.GetTask().GetPriority()
}

func (g *GfSpMigratePieceTask) SetPriority(priority coretask.TPriority) {
	g.GetTask().SetPriority(priority)
}

func (g *GfSpMigratePieceTask) SetObjectInfo(object *storagetypes.ObjectInfo) {
	g.ObjectInfo = object
}

func (g *GfSpMigratePieceTask) SetStorageParams(param *storagetypes.Params) {
	g.StorageParams = param
}

func (g *GfSpMigratePieceTask) Error() error {
	return g.GetTask().Error()
}

func (g *GfSpMigratePieceTask) SetError(err error) {
	g.GetTask().SetError(err)
}

func (g *GfSpMigratePieceTask) EstimateLimit() corercmgr.Limit {
	l := &gfsplimit.GfSpLimit{Memory: 2 * int64(g.GetObjectInfo().PayloadSize)}
	l.Add(LimitEstimateByPriority(g.GetPriority()))
	return l
}

func (g *GfSpMigratePieceTask) SetSignature(signature []byte) {
	g.Signature = signature
}

func (g *GfSpMigratePieceTask) SetPieceSize(size uint64) {
	g.PieceSize = size
}

func (g *GfSpMigratePieceTask) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&GfSpMigratePieceTask{
		Task:          &GfSpTask{CreateTime: g.GetCreateTime()},
		ObjectInfo:    g.GetObjectInfo(),
		StorageParams: g.GetStorageParams(),
		SegmentIdx:    g.GetSegmentIdx(),
		EcIdx:         g.GetEcIdx(),
		PieceSize:     g.GetPieceSize(),
	}))
}

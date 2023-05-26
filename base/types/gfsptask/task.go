package gfsptask

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsplimit"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
)

var _ coretask.Task = &GfSpTask{}

func (m *GfSpTask) Key() coretask.TKey {
	return ""
}

func (m *GfSpTask) Type() coretask.TType {
	return coretask.TypeTaskUnknown
}

func (m *GfSpTask) Info() string {
	return fmt.Sprintf("create[%d], updata[%d], timeout[%d], retry[%d], max_retry[%d], runner[%s], error[%v]",
		m.GetCreateTime(), m.GetUpdateTime(), m.GetTimeout(), m.GetRetry(), m.GetMaxRetry(), m.GetAddress(), m.GetErr())
}
func (m *GfSpTask) SetAddress(address string) {
	m.Address = address
}

func (m *GfSpTask) SetCreateTime(time int64) {
	m.CreateTime = time
}

func (m *GfSpTask) SetUpdateTime(time int64) {
	m.UpdateTime = time
}

func (m *GfSpTask) SetTimeout(timeout int64) {
	m.Timeout = timeout
}

func (m *GfSpTask) ExceedTimeout() bool {
	if m.Retry == 0 {
		return false
	}
	return m.GetUpdateTime()+m.GetTimeout() < time.Now().Unix()
}

func (m *GfSpTask) IncRetry() {
	m.Retry++
}

func (m *GfSpTask) SetRetry(retry int) {
	m.Retry = int64(retry)
}

func (m *GfSpTask) ExceedRetry() bool {
	return m.GetRetry() >= m.GetMaxRetry()
}

func (m *GfSpTask) Expired() bool {
	if m.ExceedRetry() && m.ExceedTimeout() {
		return true
	}
	return false
}

func (m *GfSpTask) GetPriority() coretask.TPriority {
	return coretask.TPriority(m.TaskPriority)
}

func (m *GfSpTask) SetPriority(priority coretask.TPriority) {
	m.TaskPriority = int32(priority)
}

func (m *GfSpTask) SetMaxRetry(limit int64) {
	m.MaxRetry = limit
}

func (m *GfSpTask) EstimateLimit() rcmgr.Limit {
	return nil
}

func (m *GfSpTask) Error() error {
	if m.GetErr() == nil {
		return nil
	} else if m.GetErr().GetInnerCode() == 0 {
		return nil
	}
	return m.GetErr()
}

func (m *GfSpTask) SetError(err error) {
	m.Err = gfsperrors.MakeGfSpError(err)
}

func LimitEstimateByPriority(priority coretask.TPriority) rcmgr.Limit {
	if priority < coretask.DefaultSmallerPriority {
		return &gfsplimit.GfSpLimit{Tasks: 1, TasksLowPriority: 1}
	} else if priority > coretask.DefaultLargerTaskPriority {
		return &gfsplimit.GfSpLimit{Tasks: 1, TasksHighPriority: 1}
	}
	return &gfsplimit.GfSpLimit{Tasks: 1, TasksMediumPriority: 1}
}

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&GfSpReplicatePieceApprovalTask{}, "p2p/ReplicatePieceApprovalTask", nil)
	cdc.RegisterConcrete(&GfSpReceivePieceTask{}, "secondary/ReceivePieceTask", nil)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(Amino)
)

func init() {
	RegisterCodec(Amino)
}

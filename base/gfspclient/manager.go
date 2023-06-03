package gfspclient

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsplimit"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

func (s *GfSpClient) CreateUploadObject(ctx context.Context, task coretask.UploadObjectTask) error {
	conn, connErr := s.ManagerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect manager", "error", connErr)
		return ErrRpcUnknown
	}
	req := &gfspserver.GfSpBeginTaskRequest{
		Request: &gfspserver.GfSpBeginTaskRequest_UploadObjectTask{
			UploadObjectTask: task.(*gfsptask.GfSpUploadObjectTask),
		},
	}
	resp, err := gfspserver.NewGfSpManageServiceClient(conn).GfSpBeginTask(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to create upload object task", "error", err)
		return ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		resp.GetErr()
	}
	return nil
}

func (s *GfSpClient) AskTask(ctx context.Context, limit corercmgr.Limit) (coretask.Task, error) {
	conn, connErr := s.ManagerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect manager", "error", connErr)
		return nil, ErrRpcUnknown
	}
	req := &gfspserver.GfSpAskTaskRequest{
		NodeLimit: limit.(*gfsplimit.GfSpLimit),
	}
	resp, err := gfspserver.NewGfSpManageServiceClient(conn).GfSpAskTask(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to ask task", "error", err)
		return nil, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	switch t := resp.GetResponse().(type) {
	case *gfspserver.GfSpAskTaskResponse_ReplicatePieceTask:
		return t.ReplicatePieceTask, nil
	case *gfspserver.GfSpAskTaskResponse_SealObjectTask:
		return t.SealObjectTask, nil
	case *gfspserver.GfSpAskTaskResponse_ReceivePieceTask:
		return t.ReceivePieceTask, nil
	case *gfspserver.GfSpAskTaskResponse_GcObjectTask:
		return t.GcObjectTask, nil
	case *gfspserver.GfSpAskTaskResponse_GcZombiePieceTask:
		return t.GcZombiePieceTask, nil
	case *gfspserver.GfSpAskTaskResponse_GcMetaTask:
		return t.GcMetaTask, nil
	default:
		return nil, ErrTypeMismatch
	}
}

func (s *GfSpClient) ReportTask(ctx context.Context, report coretask.Task) error {
	conn, connErr := s.ManagerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect manager", "error", connErr)
		return ErrRpcUnknown
	}
	req := &gfspserver.GfSpReportTaskRequest{}
	switch t := report.(type) {
	case *gfsptask.GfSpUploadObjectTask:
		req.Request = &gfspserver.GfSpReportTaskRequest_UploadObjectTask{
			UploadObjectTask: t,
		}
	case *gfsptask.GfSpReplicatePieceTask:
		req.Request = &gfspserver.GfSpReportTaskRequest_ReplicatePieceTask{
			ReplicatePieceTask: t,
		}
	case *gfsptask.GfSpReceivePieceTask:
		req.Request = &gfspserver.GfSpReportTaskRequest_ReceivePieceTask{
			ReceivePieceTask: t,
		}
	case *gfsptask.GfSpSealObjectTask:
		req.Request = &gfspserver.GfSpReportTaskRequest_SealObjectTask{
			SealObjectTask: t,
		}
	case *gfsptask.GfSpGCObjectTask:
		req.Request = &gfspserver.GfSpReportTaskRequest_GcObjectTask{
			GcObjectTask: t,
		}
	case *gfsptask.GfSpGCZombiePieceTask:
		req.Request = &gfspserver.GfSpReportTaskRequest_GcZombiePieceTask{
			GcZombiePieceTask: t,
		}
	case *gfsptask.GfSpGCMetaTask:
		req.Request = &gfspserver.GfSpReportTaskRequest_GcMetaTask{
			GcMetaTask: t,
		}
	case *gfsptask.GfSpDownloadObjectTask:
		req.Request = &gfspserver.GfSpReportTaskRequest_DownloadObjectTask{
			DownloadObjectTask: t,
		}
	case *gfsptask.GfSpChallengePieceTask:
		req.Request = &gfspserver.GfSpReportTaskRequest_ChallengePieceTask{
			ChallengePieceTask: t,
		}
	}
	resp, err := gfspserver.NewGfSpManageServiceClient(conn).GfSpReportTask(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to report task", "error", err)
		return ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return resp.GetErr()
	}
	return nil
}

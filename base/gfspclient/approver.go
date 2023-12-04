package gfspclient

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

func (s *GfSpClient) AskCreateBucketApproval(ctx context.Context, task coretask.ApprovalCreateBucketTask) (
	bool, coretask.ApprovalCreateBucketTask, error) {
	conn, connErr := s.ApproverConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect approver", "error", connErr)
		return false, nil, ErrRPCUnknownWithDetail("client failed to connect approver, error: ", connErr)
	}
	req := &gfspserver.GfSpAskApprovalRequest{
		Request: &gfspserver.GfSpAskApprovalRequest_CreateBucketApprovalTask{
			CreateBucketApprovalTask: task.(*gfsptask.GfSpCreateBucketApprovalTask),
		}}
	resp, err := gfspserver.NewGfSpApprovalServiceClient(conn).GfSpAskApproval(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to ask create bucket approval", "error", err)
		return false, nil, ErrRPCUnknownWithDetail("client failed to ask create bucket approval, error: ", err)
	}
	if resp.GetErr() != nil {
		return false, nil, resp.GetErr()
	}
	switch t := resp.Response.(type) {
	case *gfspserver.GfSpAskApprovalResponse_CreateBucketApprovalTask:
		task = t.CreateBucketApprovalTask
	default:
		return false, nil, ErrTypeMismatch
	}
	return resp.GetAllowed(), task, nil
}

func (s *GfSpClient) AskMigrateBucketApproval(ctx context.Context, task coretask.ApprovalMigrateBucketTask) (
	bool, coretask.ApprovalMigrateBucketTask, error) {
	conn, connErr := s.ApproverConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect approver", "error", connErr)
		return false, nil, ErrRPCUnknownWithDetail("client failed to connect approver, error: ", connErr)
	}
	req := &gfspserver.GfSpAskApprovalRequest{
		Request: &gfspserver.GfSpAskApprovalRequest_MigrateBucketApprovalTask{
			MigrateBucketApprovalTask: task.(*gfsptask.GfSpMigrateBucketApprovalTask),
		}}
	resp, err := gfspserver.NewGfSpApprovalServiceClient(conn).GfSpAskApproval(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to ask create bucket approval", "error", err)
		return false, nil, ErrRPCUnknownWithDetail("client failed to ask create bucket approval, error: ", err)
	}
	if resp.GetErr() != nil {
		return false, nil, resp.GetErr()
	}
	switch t := resp.Response.(type) {
	case *gfspserver.GfSpAskApprovalResponse_MigrateBucketApprovalTask:
		task = t.MigrateBucketApprovalTask
	default:
		return false, nil, ErrTypeMismatch
	}
	return resp.GetAllowed(), task, nil
}

func (s *GfSpClient) AskCreateObjectApproval(ctx context.Context, task coretask.ApprovalCreateObjectTask) (
	bool, coretask.ApprovalCreateObjectTask, error) {
	conn, connErr := s.ApproverConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect approver", "error", connErr)
		return false, nil, ErrRPCUnknownWithDetail("client failed to connect approver, error: ", connErr)
	}
	req := &gfspserver.GfSpAskApprovalRequest{
		Request: &gfspserver.GfSpAskApprovalRequest_CreateObjectApprovalTask{
			CreateObjectApprovalTask: task.(*gfsptask.GfSpCreateObjectApprovalTask),
		}}
	resp, err := gfspserver.NewGfSpApprovalServiceClient(conn).GfSpAskApproval(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to ask create object approval", "error", err)
		return false, nil, ErrRPCUnknownWithDetail("client failed to ask create object approval, error: ", err)
	}
	if resp.GetErr() != nil {
		return false, nil, resp.GetErr()
	}
	switch t := resp.Response.(type) {
	case *gfspserver.GfSpAskApprovalResponse_CreateObjectApprovalTask:
		task = t.CreateObjectApprovalTask
	default:
		return false, nil, ErrTypeMismatch
	}
	return resp.GetAllowed(), task, nil
}

package gfspclient

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
)

func (s *GfSpClient) AskCreateBucketApproval(
	ctx context.Context,
	task coretask.ApprovalCreateBucketTask) (
	bool, coretask.ApprovalCreateBucketTask, error) {
	conn, err := s.ApproverConn(ctx)
	if err != nil {
		return false, nil, err
	}
	req := &gfspserver.GfSpAskApprovalRequest{
		Request: &gfspserver.GfSpAskApprovalRequest_CreateBucketApprovalTask{
			CreateBucketApprovalTask: task.(*gfsptask.GfSpCreateBucketApprovalTask),
		}}
	resp, err := gfspserver.NewGfSpApprovalServiceClient(conn).GfSpAskApproval(ctx, req)
	if err != nil {
		return false, nil, ErrRpcUnknown
	}
	switch t := resp.Response.(type) {
	case *gfspserver.GfSpAskApprovalResponse_CreateBucketApprovalTask:
		task = t.CreateBucketApprovalTask
	default:
		return false, nil, ErrTypeMismatch
	}
	return resp.GetAllowed(), task, resp.GetErr()
}

func (s *GfSpClient) AskCreateObjectApproval(
	ctx context.Context,
	task coretask.ApprovalCreateObjectTask) (
	bool, coretask.ApprovalCreateObjectTask, error) {
	conn, err := s.ApproverConn(ctx)
	if err != nil {
		return false, nil, err
	}
	req := &gfspserver.GfSpAskApprovalRequest{
		Request: &gfspserver.GfSpAskApprovalRequest_CreateObjectApprovalTask{
			CreateObjectApprovalTask: task.(*gfsptask.GfSpCreateObjectApprovalTask),
		}}
	resp, err := gfspserver.NewGfSpApprovalServiceClient(conn).GfSpAskApproval(ctx, req)
	if err != nil {
		return false, nil, ErrRpcUnknown
	}
	switch t := resp.Response.(type) {
	case *gfspserver.GfSpAskApprovalResponse_CreateObjectApprovalTask:
		task = t.CreateObjectApprovalTask
	default:
		return false, nil, ErrTypeMismatch
	}
	return resp.GetAllowed(), task, resp.GetErr()
}

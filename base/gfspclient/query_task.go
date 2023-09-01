package gfspclient

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"google.golang.org/grpc"
)

func (s *GfSpClient) QueryTasks(ctx context.Context, endpoint string, subKey string, opts ...grpc.DialOption) ([]string, error) {
	conn, connErr := s.Connection(ctx, endpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect gfsp server", "error", connErr)
		return nil, ErrRPCUnknownWithDetail("client failed to connect gfsp server, error: " + connErr.Error())
	}
	defer conn.Close()
	req := &gfspserver.GfSpQueryTasksRequest{
		TaskSubKey: subKey,
	}
	resp, err := gfspserver.NewGfSpQueryTaskServiceClient(conn).GfSpQueryTasks(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to query tasks", "error", err)
		return nil, ErrRPCUnknownWithDetail("client failed to query tasks, error: " + err.Error())
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetTaskInfo(), nil
}

func (s *GfSpClient) QueryBucketMigrate(ctx context.Context, endpoint string, opts ...grpc.DialOption) (string, error) {
	conn, connErr := s.Connection(ctx, endpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect gfsp server", "error", connErr)
		return "", ErrRPCUnknownWithDetail("client failed to connect gfsp server, error: " + connErr.Error())
	}
	defer conn.Close()
	req := &gfspserver.GfSpQueryBucketMigrateRequest{}
	resp, err := gfspserver.NewGfSpQueryTaskServiceClient(conn).GfSpQueryBucketMigrate(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to query tasks", "error", err)
		return "", ErrRPCUnknownWithDetail("client failed to query tasks, error: " + err.Error())
	}
	if resp.GetErr() != nil {
		return "", resp.GetErr()
	}
	jsonData, err := json.Marshal(resp)
	if err != nil {
		return "", errors.New("error converting response to JSON")
	}
	return string(jsonData), nil
}

func (s *GfSpClient) QuerySPExit(ctx context.Context, endpoint string, opts ...grpc.DialOption) (string, error) {
	conn, connErr := s.Connection(ctx, endpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect gfsp server", "error", connErr)
		return "", ErrRPCUnknownWithDetail("client failed to connect gfsp server, error: " + connErr.Error())
	}
	defer conn.Close()
	req := &gfspserver.GfSpQuerySpExitRequest{}
	resp, err := gfspserver.NewGfSpQueryTaskServiceClient(conn).GfSpQuerySpExit(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to query tasks", "error", err)
		return "", ErrRPCUnknownWithDetail("client failed to query tasks, error: " + err.Error())
	}
	if resp.GetErr() != nil {
		return "", resp.GetErr()
	}
	jsonData, err := json.Marshal(resp)
	if err != nil {
		return "", errors.New("error converting response to JSON")
	}
	return string(jsonData), nil
}

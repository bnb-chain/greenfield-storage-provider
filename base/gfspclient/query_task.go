package gfspclient

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// QueryAPI for mock use
type QueryAPI interface {
	QueryTasks(ctx context.Context, endpoint string, subKey string) ([]string, error)
	QueryBucketMigrate(ctx context.Context, endpoint string) (string, error)
	QuerySPExit(ctx context.Context, endpoint string) (string, error)
}

func (s *GfSpClient) QueryTasks(ctx context.Context, endpoint string, subKey string) ([]string, error) {
	conn, connErr := s.Connection(ctx, endpoint)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect gfsp server", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &gfspserver.GfSpQueryTasksRequest{
		TaskSubKey: subKey,
	}
	resp, err := gfspserver.NewGfSpQueryTaskServiceClient(conn).GfSpQueryTasks(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to query tasks", "error", err)
		return nil, ErrRPCUnknown
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetTaskInfo(), nil
}

func (s *GfSpClient) QueryBucketMigrate(ctx context.Context, endpoint string) (string, error) {
	conn, connErr := s.Connection(ctx, endpoint)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect gfsp server", "error", connErr)
		return "", ErrRPCUnknown
	}
	defer conn.Close()
	req := &gfspserver.GfSpQueryBucketMigrateRequest{}
	resp, err := gfspserver.NewGfSpQueryTaskServiceClient(conn).GfSpQueryBucketMigrate(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to query tasks", "error", err)
		return "", ErrRPCUnknown
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

func (s *GfSpClient) QuerySPExit(ctx context.Context, endpoint string) (string, error) {
	conn, connErr := s.Connection(ctx, endpoint)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect gfsp server", "error", connErr)
		return "", ErrRPCUnknown
	}
	defer conn.Close()
	req := &gfspserver.GfSpQuerySpExitRequest{}
	resp, err := gfspserver.NewGfSpQueryTaskServiceClient(conn).GfSpQuerySpExit(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to query tasks", "error", err)
		return "", ErrRPCUnknown
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

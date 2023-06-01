package gfspapp

import (
	"context"
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
)

var _ gfspserver.GfSpQueryTaskServiceServer = &GfSpBaseApp{}

func (g *GfSpBaseApp) GfSpQueryTasks(ctx context.Context, req *gfspserver.GfSpQueryTasksRequest) (
	*gfspserver.GfSpQueryTasksResponse, error) {
	subKey := req.GetTaskSubKey()
	if len(subKey) == 0 {
		return &gfspserver.GfSpQueryTasksResponse{
			Err: gfsperrors.MakeGfSpError(fmt.Errorf("invalid query key"))}, nil
	}
	approverTasks, _ := g.approver.QueryTasks(ctx, coretask.TKey(subKey))
	downloaderTasks, _ := g.downloader.QueryTasks(ctx, coretask.TKey(subKey))
	managerTasks, _ := g.manager.QueryTasks(ctx, coretask.TKey(subKey))
	p2pTasks, _ := g.p2p.QueryTasks(ctx, coretask.TKey(subKey))
	receiverTasks, _ := g.receiver.QueryTasks(ctx, coretask.TKey(subKey))
	uploaderTasks, _ := g.uploader.QueryTasks(ctx, coretask.TKey(subKey))

	var taskInfo []string
	for _, task := range approverTasks {
		taskInfo = append(taskInfo, task.Info())
	}
	for _, task := range downloaderTasks {
		taskInfo = append(taskInfo, task.Info())
	}
	for _, task := range managerTasks {
		taskInfo = append(taskInfo, task.Info())
	}
	for _, task := range p2pTasks {
		taskInfo = append(taskInfo, task.Info())
	}
	for _, task := range receiverTasks {
		taskInfo = append(taskInfo, task.Info())
	}
	for _, task := range uploaderTasks {
		taskInfo = append(taskInfo, task.Info())
	}
	if len(taskInfo) == 0 {
		return &gfspserver.GfSpQueryTasksResponse{
			Err: gfsperrors.MakeGfSpError(fmt.Errorf("no match tasks"))}, nil
	}
	return &gfspserver.GfSpQueryTasksResponse{TaskInfo: taskInfo}, nil
}

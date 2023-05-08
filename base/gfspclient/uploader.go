package gfspclient

import (
	"context"
	"io"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

func (s *GfSpClient) UploadObject(
	ctx context.Context,
	task coretask.UploadObjectTask,
	stream io.Reader) error {
	conn, err := s.Connection(ctx, s.uploaderEndpoint)
	if err != nil {
		return err
	}
	defer conn.Close()
	client, err := gfspserver.NewGfSpUploadServiceClient(conn).GfSpUploadObject(ctx)
	if err != nil {
		log.CtxErrorw(ctx, "failed to new uploader stream client", "error", err)
		return ErrRpcUnknown
	}
	var (
		buf = make([]byte, model.DefaultStreamBufSize)
	)
	for {
		n, err := stream.Read(buf)
		if err == io.EOF {
			if n != 0 {
				req := &gfspserver.GfSpUploadObjectRequest{
					UploadObjectTask: task.(*gfsptask.GfSpUploadObjectTask),
					Payload:          buf[0:n],
				}
				err = client.Send(req)
				if err != nil {
					log.CtxErrorw(ctx, "failed to send the upload stream data", "error", err)
					return ErrRpcUnknown
				}
			}
			err = client.CloseSend()
			if err != nil {
				log.CtxErrorw(ctx, "failed to close upload stream", "error", err)
				return ErrRpcUnknown
			}
		}
		if err != nil {
			log.CtxErrorw(ctx, "failed to close upload stream", "error", err)
			return ErrExceptionsStream
		}
		req := &gfspserver.GfSpUploadObjectRequest{
			UploadObjectTask: task.(*gfsptask.GfSpUploadObjectTask),
			Payload:          buf[0:n],
		}
		err = client.Send(req)
		if err != nil {
			log.CtxErrorw(ctx, "failed to send the upload stream data", "error", err)
			return ErrRpcUnknown
		}
	}
}

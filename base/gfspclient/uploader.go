package gfspclient

import (
	"context"
	"io"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

func (s *GfSpClient) UploadObject(ctx context.Context, task coretask.UploadObjectTask, stream io.Reader) error {
	conn, connErr := s.Connection(ctx, s.uploaderEndpoint)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect uploader", "error", connErr)
		return ErrRpcUnknown
	}
	var sendSize = 0
	defer func() {
		conn.Close()
		if task != nil {
			log.CtxDebugw(ctx, "succeed to send payload data", "info", task.Info(),
				"send_size", sendSize)
		} else {
			log.CtxDebugw(ctx, "finished to send payload data", "send_size", sendSize)
		}
	}()
	client, err := gfspserver.NewGfSpUploadServiceClient(conn).GfSpUploadObject(ctx)
	if err != nil {
		log.CtxErrorw(ctx, "failed to new uploader stream client", "error", err)
		return ErrRpcUnknown
	}
	var (
		buf = make([]byte, DefaultStreamBufSize)
	)
	for {
		n, streamErr := stream.Read(buf)
		sendSize += n
		if streamErr == io.EOF {
			if n != 0 {
				req := &gfspserver.GfSpUploadObjectRequest{
					UploadObjectTask: task.(*gfsptask.GfSpUploadObjectTask),
					Payload:          buf[0:n],
				}
				err = client.Send(req)
				if err != nil {
					log.CtxErrorw(ctx, "failed to send the last upload stream data", "error", err)
					return ErrRpcUnknown
				}
			}
			resp, closeErr := client.CloseAndRecv()
			if closeErr != nil {
				log.CtxErrorw(ctx, "failed to close upload stream", "error", closeErr)
				return ErrRpcUnknown
			}
			if resp.GetErr() != nil {
				return resp.GetErr()
			}
			return nil
		}
		if streamErr != nil {
			log.CtxErrorw(ctx, "failed to read upload data stream", "error", streamErr)
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

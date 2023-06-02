package gfspclient

import (
	"context"
	"io"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

func (s *GfSpClient) UploadObject(ctx context.Context, task coretask.UploadObjectTask, stream io.Reader) error {
	startConnectUploader := time.Now()
	conn, connErr := s.Connection(ctx, s.uploaderEndpoint)
	metrics.PerfUploadTimeHistogram.WithLabelValues("connect_to_uploader").Observe(time.Since(startConnectUploader).Seconds())
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
		metrics.PerfUploadTimeHistogram.WithLabelValues("client_total_time").Observe(time.Since(startConnectUploader).Seconds())
	}()
	startGetUploaderClient := time.Now()
	client, err := gfspserver.NewGfSpUploadServiceClient(conn).GfSpUploadObject(ctx)
	metrics.PerfUploadTimeHistogram.WithLabelValues("get_uploader_client").Observe(time.Since(startGetUploaderClient).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to new uploader stream client", "error", err)
		return ErrRpcUnknown
	}
	var (
		buf = make([]byte, DefaultStreamBufSize)
	)
	for {
		startReadFromSDK := time.Now()
		n, streamErr := stream.Read(buf)
		metrics.PerfUploadTimeHistogram.WithLabelValues("read_from_sdk").Observe(time.Since(startReadFromSDK).Seconds())
		sendSize += n
		if streamErr == io.EOF {
			if n != 0 {
				req := &gfspserver.GfSpUploadObjectRequest{
					UploadObjectTask: task.(*gfsptask.GfSpUploadObjectTask),
					Payload:          buf[0:n],
				}
				startSendUploader := time.Now()
				err = client.Send(req)
				metrics.PerfUploadTimeHistogram.WithLabelValues("send_to_uploader").Observe(time.Since(startSendUploader).Seconds())
				if err != nil {
					log.CtxErrorw(ctx, "failed to send the last upload stream data", "error", err)
					return ErrRpcUnknown
				}
			}
			startCloseClient := time.Now()
			resp, closeErr := client.CloseAndRecv()
			metrics.PerfUploadTimeHistogram.WithLabelValues("close_client").Observe(time.Since(startCloseClient).Seconds())
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
		startSendUploader := time.Now()
		err = client.Send(req)
		metrics.PerfUploadTimeHistogram.WithLabelValues("send_to_uploader").Observe(time.Since(startSendUploader).Seconds())
		if err != nil {
			log.CtxErrorw(ctx, "failed to send the upload stream data", "error", err)
			return ErrRpcUnknown
		}
	}
}

func (s *GfSpClient) ResumableUploadObject(
	ctx context.Context,
	task coretask.ResumableUploadObjectTask,
	stream io.Reader) {
}

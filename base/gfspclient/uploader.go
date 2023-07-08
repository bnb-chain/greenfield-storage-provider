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
	startTime := time.Now()
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
	buf := make([]byte, DefaultStreamBufSize)
	metrics.PerfPutObjectTime.WithLabelValues("client_put_object_prepare_cost").Observe(time.Since(startTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("client_put_object_prepare_end").Observe(time.Now().Sub(time.UnixMilli(task.GetCreateTime())).Seconds())
	for {
		startReadFromSDK := time.Now()
		n, streamErr := stream.Read(buf)
		metrics.PerfPutObjectTime.WithLabelValues("client_put_object_read_data_cost").Observe(time.Since(startReadFromSDK).Seconds())
		metrics.PerfPutObjectTime.WithLabelValues("client_put_object_read_data_end").Observe(time.Now().Sub(time.UnixMilli(task.GetCreateTime())).Seconds())
		sendSize += n
		if streamErr == io.EOF {
			if n != 0 {
				req := &gfspserver.GfSpUploadObjectRequest{
					UploadObjectTask: task.(*gfsptask.GfSpUploadObjectTask),
					Payload:          buf[0:n],
				}
				startSendUploader := time.Now()
				err = client.Send(req)
				metrics.PerfPutObjectTime.WithLabelValues("client_put_object_send_cost").Observe(time.Since(startSendUploader).Seconds())
				metrics.PerfPutObjectTime.WithLabelValues("client_put_object_send_end").Observe(time.Now().Sub(time.UnixMilli(task.GetCreateTime())).Seconds())
				if err != nil {
					log.CtxErrorw(ctx, "failed to send the last upload stream data", "error", err)
					return ErrRpcUnknown
				}
			}
			startCloseClient := time.Now()
			resp, closeErr := client.CloseAndRecv()
			metrics.PerfPutObjectTime.WithLabelValues("client_put_object_send_last_cost").Observe(time.Since(startCloseClient).Seconds())
			metrics.PerfPutObjectTime.WithLabelValues("client_put_object_send_last_end").Observe(time.Now().Sub(time.UnixMilli(task.GetCreateTime())).Seconds())
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
		metrics.PerfPutObjectTime.WithLabelValues("client_put_object_send_cost").Observe(time.Since(startSendUploader).Seconds())
		metrics.PerfPutObjectTime.WithLabelValues("client_put_object_send_end").Observe(time.Now().Sub(time.UnixMilli(task.GetCreateTime())).Seconds())
		if err != nil {
			log.CtxErrorw(ctx, "failed to send the upload stream data", "error", err)
			return ErrRpcUnknown
		}
	}
}

func (s *GfSpClient) ResumableUploadObject(ctx context.Context, task coretask.ResumableUploadObjectTask, stream io.Reader) error {
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
			log.CtxDebugw(ctx, "failed to send payload data", "send_size", sendSize)
		}
	}()
	client, err := gfspserver.NewGfSpUploadServiceClient(conn).GfSpResumableUploadObject(ctx)

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
				req := &gfspserver.GfSpResumableUploadObjectRequest{
					ResumableUploadObjectTask: task.(*gfsptask.GfSpResumableUploadObjectTask),
					Payload:                   buf[0:n],
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
		req := &gfspserver.GfSpResumableUploadObjectRequest{
			ResumableUploadObjectTask: task.(*gfsptask.GfSpResumableUploadObjectTask),
			Payload:                   buf[0:n],
		}
		err = client.Send(req)
		if err != nil {
			log.CtxErrorw(ctx, "failed to send the upload stream data", "error", err)
			return ErrRpcUnknown
		}
	}
}

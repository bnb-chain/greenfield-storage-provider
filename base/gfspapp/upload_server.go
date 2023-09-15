package gfspapp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

var (
	ErrUploadObjectDangling  = gfsperrors.Register(BaseCodeSpace, http.StatusBadRequest, 991101, "OoooH... request lost")
	ErrUploadExhaustResource = gfsperrors.Register(BaseCodeSpace, http.StatusBadRequest, 991102, "server overload, try again later")
	ErrExceptionsStream      = gfsperrors.Register(BaseCodeSpace, http.StatusBadRequest, 991103, "stream closed abnormally")
)

var _ gfspserver.GfSpUploadServiceServer = &GfSpBaseApp{}

func (g *GfSpBaseApp) GfSpUploadObject(stream gfspserver.GfSpUploadService_GfSpUploadObjectServer) error {
	var (
		span          rcmgr.ResourceScopeSpan
		task          *gfsptask.GfSpUploadObjectTask
		req           *gfspserver.GfSpUploadObjectRequest
		resp          = &gfspserver.GfSpUploadObjectResponse{}
		pRead, pWrite = io.Pipe()
		initChan      = make(chan struct{})
		errChan       = make(chan error)
		ctx, cancel   = context.WithCancel(context.Background())
		err           error
		receiveSize   int
	)
	startTime := time.Now()
	defer func() {
		defer cancel()
		if span != nil {
			span.Done()
		}
		if task != nil {
			g.uploader.PostUploadObject(ctx, task)
			log.CtxDebugw(ctx, "finished to receive object stream data", "info", task.Info(),
				"receive_size", receiveSize, "error", err)
		} else {
			log.CtxDebugw(ctx, "finished to receive object stream data", "receive_size", receiveSize,
				"error", err)
		}
		if err != nil {
			resp.Err = gfsperrors.MakeGfSpError(err)
			metrics.ReqCounter.WithLabelValues(UploaderFailurePutObject).Inc()
			metrics.ReqTime.WithLabelValues(UploaderFailurePutObject).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(UploaderSuccessPutObject).Inc()
			metrics.ReqTime.WithLabelValues(UploaderSuccessPutObject).Observe(time.Since(startTime).Seconds())
		}

		err = stream.SendAndClose(resp)
		if err != nil {
			log.CtxErrorw(ctx, "failed to close upload object stream", "error", err)
		}
	}()

	go func() {
		init := false
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			req, err = stream.Recv()
			if err == io.EOF {
				if len(req.GetPayload()) != 0 {
					_, _ = pWrite.Write(req.GetPayload())
				}
				log.CtxDebug(ctx, "received last upload stream data")
				err = nil
				_ = pWrite.Close()
				return
			}
			if err != nil {
				log.CtxErrorw(ctx, "failed to receive object ", "error", err)
				err = ErrExceptionsStream
				_ = pWrite.CloseWithError(err)
				errChan <- err
				return
			}
			if !init {
				init = true
				task = req.GetUploadObjectTask()
				if task == nil {
					log.CtxError(ctx, "[BUG] failed to receive object, upload object task pointer dangling!!!")
					err = ErrUploadObjectDangling
					_ = pWrite.CloseWithError(err)
					errChan <- err
					return
				}
				ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
				span, err = g.uploader.ReserveResource(ctx, task.EstimateLimit().ScopeStat())
				if err != nil {
					log.CtxErrorw(ctx, "failed to reserve resource", "error", err)
					err = ErrUploadExhaustResource
					_ = pWrite.CloseWithError(err)
					errChan <- err
					return
				}
				err = g.uploader.PreUploadObject(ctx, task)
				task.AppendLog(fmt.Sprintf("uploader-prepare-upload-task-cost:%d", time.Now().UnixMilli()-startTime.UnixMilli()))
				if err != nil {
					log.CtxErrorw(ctx, "failed to pre upload object", "error", err)
					_ = pWrite.CloseWithError(err)
					errChan <- err
					return
				}
				close(initChan)
			}
			receiveSize += len(req.GetPayload())
			_, _ = pWrite.Write(req.GetPayload())
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case <-initChan:
		log.CtxDebug(ctx, "received first upload stream data")
	case err1 := <-errChan:
		log.Errorw("failed to upload object", "error", err1)
		return err1
	}
	err = g.uploader.HandleUploadObjectTask(ctx, task, pRead)
	if err != nil {
		log.CtxErrorw(ctx, "failed to upload object data", "error", err)
		_ = pWrite.CloseWithError(err)
		return err
	}
	log.CtxDebug(ctx, "succeed to upload object")
	return nil
}

func (g *GfSpBaseApp) GfSpResumableUploadObject(stream gfspserver.GfSpUploadService_GfSpResumableUploadObjectServer) error {
	var (
		span          rcmgr.ResourceScopeSpan
		task          *gfsptask.GfSpResumableUploadObjectTask
		req           *gfspserver.GfSpResumableUploadObjectRequest
		resp          = &gfspserver.GfSpResumableUploadObjectResponse{}
		pRead, pWrite = io.Pipe()
		initCh        = make(chan struct{})
		errChan       = make(chan error)
		ctx, cancel   = context.WithCancel(context.Background())
		err           error
		receiveSize   int
	)
	defer func() {
		defer cancel()
		if span != nil {
			span.Done()
		}
		if task != nil {
			g.uploader.PostResumableUploadObject(ctx, task)
			log.CtxDebugw(ctx, "finished to receive object stream data", "info", task.Info(),
				"receive_size", receiveSize, "error", err)
		} else {
			log.CtxDebugw(ctx, "finished to receive object stream data", "receive_size", receiveSize,
				"error", err)
		}
		if err != nil {
			resp.Err = gfsperrors.MakeGfSpError(err)
		}

		err = stream.SendAndClose(resp)
		if err != nil {
			log.CtxErrorw(ctx, "failed to close upload object stream", "error", err)
		}
	}()

	go func() {
		init := false
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			req, err = stream.Recv()
			if err == io.EOF {
				if len(req.GetPayload()) != 0 {
					_, _ = pWrite.Write(req.GetPayload())
				}
				log.CtxDebug(ctx, "received last resumable upload stream data")
				err = nil
				_ = pWrite.Close()
				return
			}
			if err != nil {
				log.CtxErrorw(ctx, "failed to receive object ", "error", err)
				err = ErrExceptionsStream
				_ = pWrite.CloseWithError(err)
				errChan <- err
				return
			}
			if !init {
				init = true
				task = req.GetResumableUploadObjectTask()
				if task == nil {
					log.CtxError(ctx, "[BUG] failed to receive resumable object, resumable upload object task pointer dangling!!!")
					err = ErrUploadObjectDangling
					_ = pWrite.CloseWithError(err)
					errChan <- err
					return
				}
				ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
				span, err = g.uploader.ReserveResource(ctx, task.EstimateLimit().ScopeStat())
				if err != nil {
					log.CtxErrorw(ctx, "failed to reserve resource", "error", err)
					err = ErrUploadExhaustResource
					_ = pWrite.CloseWithError(err)
					errChan <- err
					return
				}
				err = g.uploader.PreResumableUploadObject(ctx, task)
				if err != nil {
					log.CtxErrorw(ctx, "failed to pre resumable upload object", "error", err)
					_ = pWrite.CloseWithError(err)
					errChan <- err
					return
				}
				close(initCh)
			}
			receiveSize += len(req.GetPayload())
			_, _ = pWrite.Write(req.GetPayload())
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case <-initCh:
		log.CtxDebug(ctx, "received first resumable upload stream data")
	case err1 := <-errChan:
		log.Errorw("failed to resumable upload object", "error", err1)
		return err1
	}
	err = g.uploader.HandleResumableUploadObjectTask(ctx, task, pRead)
	if err != nil {
		log.CtxErrorw(ctx, "failed to resumable upload object data", "error", err)
		_ = pWrite.CloseWithError(err)
		return err
	}
	log.CtxDebug(ctx, "succeed to resumable upload object")
	return nil
}

// gRPCUploadStream for mock use
// Note: gRPCUploadStream interface is forbidden to be used in non-UT code
//
// nolint:unused
//
//go:generate mockgen -source=./upload_server.go -destination=./upload_server_mock.go -package=gfspapp
type gRPCUploadStream interface {
	gfspserver.GfSpUploadService_GfSpUploadObjectServer
}

// gRPCResumableUploadStream for mock use
// Note: gRPCResumableUploadStream interface is forbidden to be used in non-UT code
//
// nolint:unused
type gRPCResumableUploadStream interface {
	gfspserver.GfSpUploadService_GfSpResumableUploadObjectServer
}

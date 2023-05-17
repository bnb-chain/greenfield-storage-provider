package gfspapp

import (
	"context"
	"io"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

var (
	ErrUploadObjectDangling  = gfsperrors.Register(BaseCodeSpace, http.StatusInternalServerError, 991101, "OoooH... request lost")
	ErrUploadExhaustResource = gfsperrors.Register(BaseCodeSpace, http.StatusServiceUnavailable, 991102, "server overload, try again later")
)

var _ gfspserver.GfSpUploadServiceServer = &GfSpBaseApp{}

func (g *GfSpBaseApp) GfSpUploadObject(stream gfspserver.GfSpUploadService_GfSpUploadObjectServer) error {
	var (
		span          rcmgr.ResourceScopeSpan
		task          *gfsptask.GfSpUploadObjectTask
		req           *gfspserver.GfSpUploadObjectRequest
		resp          = &gfspserver.GfSpUploadObjectResponse{}
		pRead, pWrite = io.Pipe()
		ctx           = context.Background()
		errCh         = make(chan error)
		err           error
		receiveSize   int
	)
	defer func() {
		if span != nil {
			span.Done()
		}
		if task != nil {
			metrics.UploadObjectSizeHistogram.WithLabelValues(g.uploader.Name()).Observe(
				float64(task.GetObjectInfo().GetPayloadSize()))
			g.uploader.PostUploadObject(ctx, task)
			log.CtxDebugw(ctx, "finish to receive object stream data", "receive_size", receiveSize,
				"object_size", task.GetObjectInfo().GetPayloadSize(), "error", err)
		} else {
			log.CtxDebugw(ctx, "finish to receive object stream data",
				"receive_size", receiveSize, "error", err)
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
					pWrite.Write(req.GetPayload())
				}
				log.CtxDebugw(ctx, "received last upload stream data")
				err = nil
				pWrite.Close()
				return
			}
			if err != nil {
				log.CtxErrorw(ctx, "failed to receive object ", "error", err)
				pWrite.CloseWithError(err)
				errCh <- err
				return
			}
			if !init {
				init = true
				task = req.GetUploadObjectTask()
				if task == nil {
					log.CtxErrorw(ctx, "[BUG] failed to receive object, upload object task pointer dangling !!!")
					err = ErrUploadObjectDangling
					errCh <- err
					pWrite.CloseWithError(err)
					return
				}
				ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
				span, err = g.uploader.ReserveResource(ctx, task.EstimateLimit().ScopeStat())
				if err != nil {
					log.CtxErrorw(ctx, "failed to reserve resource", "error", err)
					err = ErrUploadExhaustResource
					errCh <- err
					pWrite.CloseWithError(err)
					return
				}
				err = g.uploader.PreUploadObject(ctx, task)
				if err != nil {
					log.CtxErrorw(ctx, "failed to pre upload object", "error", err)
					errCh <- err
					pWrite.CloseWithError(err)
					return
				}
				errCh <- nil
			}
			receiveSize += len(req.GetPayload())
			pWrite.Write(req.GetPayload())
		}
	}()

	if err = <-errCh; err != nil {
		log.CtxErrorw(ctx, "failed to begin upload object data", "error", err)
		return nil
	}
	go func() {
		errCh <- g.uploader.HandleUploadObjectTask(ctx, task, pRead)
	}()
	if err = <-errCh; err != nil {
		log.CtxErrorw(ctx, "failed to upload object data", "error", err)
		pWrite.CloseWithError(err)
		return err
	}
	log.CtxDebugw(ctx, "succeed to upload object")
	return nil
}

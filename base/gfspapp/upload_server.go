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
	ErrUploadObjectDangling  = gfsperrors.Register(BaseCodeSpace, http.StatusInternalServerError, 99111, "OoooH... request lost")
	ErrUploadExhaustResource = gfsperrors.Register(BaseCodeSpace, http.StatusServiceUnavailable, 99112, "server overload, try again later")
)

var _ gfspserver.GfSpUploadServiceServer = &GfSpBaseApp{}

func (g *GfSpBaseApp) GfSpUploadObject(stream gfspserver.GfSpUploadService_GfSpUploadObjectServer) error {
	var (
		span          rcmgr.ResourceScopeSpan
		task          *gfsptask.GfSpUploadObjectTask
		req           *gfspserver.GfSpUploadObjectRequest
		resp          = &gfspserver.GfSpUploadObjectResponse{}
		initCh        = make(chan struct{})
		pRead, pWrite = io.Pipe()
		ctx, cancel   = context.WithCancel(context.Background())
		err           error
	)
	defer func() {
		defer cancel()
		if span != nil {
			span.Done()
		}
		if task != nil {
			metrics.UploadObjectSizeHistogram.WithLabelValues(g.uploader.Name()).Observe(
				float64(task.GetObjectInfo().GetPayloadSize()))
			g.uploader.PostUploadObject(ctx, task)
		}
		if err != nil {
			resp.Err = gfsperrors.MakeGfSpError(err)
		}
		log.CtxDebugw(ctx, "finished to receive object stream data", "error", err)
		err = stream.SendAndClose(resp)
		if err != nil {
			log.CtxErrorw(ctx, "failed to close upload object stream", "error", err)
		}
	}()

	span, err = g.uploader.ReserveResource(ctx, task.EstimateLimit().ScopeStat())
	if err != nil {
		log.CtxErrorw(ctx, "failed to reserve resource", "error", err)
		err = ErrUploadExhaustResource
		return nil
	}

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
				log.CtxDebugw(ctx, "received last upload stream data")
				pWrite.Close()
				return
			}
			if err != nil {
				log.CtxErrorw(ctx, "failed to receive object ", "error", err)
				pWrite.CloseWithError(err)
				return
			}
			if !init {
				init = true
				task = req.GetUploadObjectTask()
				if task == nil {
					log.CtxErrorw(ctx, "[BUG] failed to receive object, upload object task pointer dangling !!!")
					err = ErrUploadObjectDangling
					pWrite.CloseWithError(err)
					return
				}
				ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
				err = g.uploader.PreUploadObject(ctx, task)
				if err != nil {
					log.CtxErrorw(ctx, "failed to pre upload object", "error", err)
					pWrite.CloseWithError(err)
					return
				}
				initCh <- struct{}{}
			}
			pWrite.Write(req.GetPayload())
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case <-initCh:
		log.CtxDebugw(ctx, "received first upload stream data")
	}
	err = g.uploader.HandleUploadObjectTask(ctx, task, pRead)
	if err != nil {
		log.CtxErrorw(ctx, "failed to upload object data", "error", err)
		pWrite.CloseWithError(err)
		return err
	}
	log.CtxErrorw(ctx, "success to upload object")
	pWrite.Close()
	return nil
}

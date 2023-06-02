package gfspapp

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
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
		initCh        = make(chan struct{})
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
			g.GfSpDB().InsertUploadEvent(task.GetObjectInfo().Id.Uint64(), corespdb.UploaderEndReceiveData, task.Key().String())
			metrics.UploadObjectSizeHistogram.WithLabelValues(g.uploader.Name()).Observe(
				float64(task.GetObjectInfo().GetPayloadSize()))
			g.uploader.PostUploadObject(ctx, task)
			log.CtxDebugw(ctx, "finish to receive object stream data", "info", task.Info(),
				"receive_size", receiveSize, "error", err)
		} else {
			log.CtxDebugw(ctx, "finish to receive object stream data",
				"receive_size", receiveSize, "error", err)
		}
		if err != nil {
			resp.Err = gfsperrors.MakeGfSpError(err)
		}

		closeTime := time.Now()
		err = stream.SendAndClose(resp)
		metrics.PerfUploadTimeHistogram.WithLabelValues("server_send_and_close_time").Observe(time.Since(closeTime).Seconds())
		if err != nil {
			log.CtxErrorw(ctx, "failed to close upload object stream", "error", err)
		}
	}()

	go func() {
		serverStartTime := time.Now()
		defer metrics.PerfUploadTimeHistogram.WithLabelValues("server_total_time").Observe(time.Since(serverStartTime).Seconds())
		init := false
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			startReadFromGateway := time.Now()
			req, err = stream.Recv()
			metrics.PerfUploadTimeHistogram.WithLabelValues("producer_read_from_gateway").Observe(time.Since(startReadFromGateway).Seconds())
			if err == io.EOF {
				if len(req.GetPayload()) != 0 {
					startForwardToConsumer := time.Now()
					pWrite.Write(req.GetPayload())
					metrics.PerfUploadTimeHistogram.WithLabelValues("producer_to_consumer").Observe(time.Since(startForwardToConsumer).Seconds())
				}
				log.CtxDebugw(ctx, "received last upload stream data")
				err = nil
				pWrite.Close()
				return
			}
			if err != nil {
				log.CtxErrorw(ctx, "failed to receive object ", "error", err)
				err = ErrExceptionsStream
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
				g.GfSpDB().InsertUploadEvent(task.GetObjectInfo().Id.Uint64(), corespdb.UploaderBeginReceiveData, task.Key().String())
				ctx = log.WithValue(ctx, log.CtxKeyTask, task.Key().String())
				span, err = g.uploader.ReserveResource(ctx, task.EstimateLimit().ScopeStat())
				if err != nil {
					log.CtxErrorw(ctx, "failed to reserve resource", "error", err)
					err = ErrUploadExhaustResource
					pWrite.CloseWithError(err)
					return
				}
				err = g.uploader.PreUploadObject(ctx, task)
				if err != nil {
					log.CtxErrorw(ctx, "failed to pre upload object", "error", err)
					pWrite.CloseWithError(err)
					return
				}
				initCh <- struct{}{}
			}
			receiveSize += len(req.GetPayload())
			startForwardToConsumer := time.Now()
			pWrite.Write(req.GetPayload())
			metrics.PerfUploadTimeHistogram.WithLabelValues("producer_to_consumer").Observe(time.Since(startForwardToConsumer).Seconds())
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
		return nil
	}
	log.CtxDebugw(ctx, "succeed to upload object")
	return nil
}

func (g *GfSpBaseApp) GfSpResumableUploadObject(stream gfspserver.GfSpUploadService_GfSpResumableUploadObjectServer) error {
	return nil
}

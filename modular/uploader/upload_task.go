package uploader

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	ErrDanglingDownloadTask = gfsperrors.Register(module.UploadModularName, http.StatusBadRequest, 110001, "OoooH... request lost, try again later")
	ErrNotCreatedState      = gfsperrors.Register(module.UploadModularName, http.StatusForbidden, 110002, "object not created state")
	ErrRepeatedTask         = gfsperrors.Register(module.UploadModularName, http.StatusNotAcceptable, 110003, "put object request repeated")
	ErrInvalidIntegrity     = gfsperrors.Register(module.UploadModularName, http.StatusNotAcceptable, 110004, "invalid payload data integrity hash")
	ErrClosedStream         = gfsperrors.Register(module.UploadModularName, http.StatusBadRequest, 110005, "upload payload data stream exception")
	ErrPieceStore           = gfsperrors.Register(module.UploadModularName, http.StatusInternalServerError, 115101, "server slipped away, try again later")
	ErrGfSpDB               = gfsperrors.Register(module.UploadModularName, http.StatusInternalServerError, 115001, "server slipped away, try again later")
)

func (u *UploadModular) PreUploadObject(ctx context.Context, uploadObjectTask coretask.UploadObjectTask) error {
	if uploadObjectTask == nil || uploadObjectTask.GetObjectInfo() == nil || uploadObjectTask.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed to pre upload object, task pointer dangling")
		return ErrDanglingDownloadTask
	}
	if uploadObjectTask.GetObjectInfo().GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED {
		log.CtxErrorw(ctx, "failed to pre upload object, object not create")
		return ErrNotCreatedState
	}
	if u.uploadQueue.Has(uploadObjectTask.Key()) {
		log.CtxErrorw(ctx, "failed to pre upload object, task repeated")
		return ErrRepeatedTask
	}
	if err := u.baseApp.GfSpClient().CreateUploadObject(ctx, uploadObjectTask); err != nil {
		log.CtxErrorw(ctx, "failed to begin upload object task")
		return err
	}
	return nil
}

func (u *UploadModular) HandleUploadObjectTask(ctx context.Context, uploadObjectTask coretask.UploadObjectTask, stream io.Reader) error {
	if err := u.uploadQueue.Push(uploadObjectTask); err != nil {
		log.CtxErrorw(ctx, "failed to push upload queue", "error", err)
		return err
	}
	defer u.uploadQueue.PopByKey(uploadObjectTask.Key())

	segmentSize := u.baseApp.PieceOp().MaxSegmentPieceSize(
		uploadObjectTask.GetObjectInfo().GetPayloadSize(),
		uploadObjectTask.GetStorageParams().GetMaxSegmentSize())
	var (
		err       error
		segIdx    uint32 = 0
		pieceKey  string
		signature []byte
		integrity []byte
		checksums [][]byte
		readN     int
		readSize  int
		data      = make([]byte, segmentSize)
	)
	defer func() {
		if err != nil {
			uploadObjectTask.SetError(err)
		}
		log.CtxDebugw(ctx, "finish to read data from stream", "info", uploadObjectTask.Info(),
			"read_size", readSize, "error", err)
		startReportManager := time.Now()
		err = u.baseApp.GfSpClient().ReportTask(ctx, uploadObjectTask)
		metrics.PerfUploadTimeHistogram.WithLabelValues("report_to_manager").Observe(time.Since(startReportManager).Seconds())
	}()

	for {
		if err != nil {
			return err
		}
		data = data[0:segmentSize]
		startReadFromGateway := time.Now()
		readN, err = StreamReadAt(stream, data)
		metrics.PerfUploadTimeHistogram.WithLabelValues("consumer_read_from_gateway").Observe(time.Since(startReadFromGateway).Seconds())
		readSize += readN
		data = data[0:readN]

		if err == io.EOF {
			err = nil
			if readN != 0 { // the last segment piece
				pieceKey = u.baseApp.PieceOp().SegmentPieceKey(uploadObjectTask.GetObjectInfo().Id.Uint64(), segIdx)
				checksums = append(checksums, hash.GenerateChecksum(data))
				startPutPiece := time.Now()
				if err = u.baseApp.PieceStore().PutPiece(ctx, pieceKey, data); err != nil {
					metrics.PerfUploadTimeHistogram.WithLabelValues("put_to_piecestore").Observe(time.Since(startPutPiece).Seconds())
					log.CtxErrorw(ctx, "failed to put segment piece to piece store",
						"piece_key", pieceKey, "error", err)
					return ErrPieceStore
				}
				metrics.PerfUploadTimeHistogram.WithLabelValues("put_to_piecestore").Observe(time.Since(startPutPiece).Seconds())
			}
			startSignSignature := time.Now()
			if signature, integrity, err = u.baseApp.GfSpClient().SignIntegrityHash(ctx,
				uploadObjectTask.GetObjectInfo().Id.Uint64(), checksums); err != nil {
				metrics.PerfUploadTimeHistogram.WithLabelValues("sign_from_signer").Observe(time.Since(startSignSignature).Seconds())
				log.CtxErrorw(ctx, "failed to sign the integrity hash", "error", err)
				return err
			}
			metrics.PerfUploadTimeHistogram.WithLabelValues("sign_from_signer").Observe(time.Since(startSignSignature).Seconds())
			if !bytes.Equal(integrity, uploadObjectTask.GetObjectInfo().GetChecksums()[0]) {
				log.CtxErrorw(ctx, "failed to put object due to check integrity hash not consistent",
					"actual_integrity", hex.EncodeToString(integrity),
					"expected_integrity", hex.EncodeToString(uploadObjectTask.GetObjectInfo().GetChecksums()[0]))
				err = ErrInvalidIntegrity
				return ErrInvalidIntegrity
			}
			integrityMeta := &corespdb.IntegrityMeta{
				ObjectID:          uploadObjectTask.GetObjectInfo().Id.Uint64(),
				PieceChecksumList: checksums,
				IntegrityChecksum: integrity,
				Signature:         signature,
			}
			startUpdateSignature := time.Now()
			if err = u.baseApp.GfSpDB().SetObjectIntegrity(integrityMeta); err != nil {
				metrics.PerfUploadTimeHistogram.WithLabelValues("update_to_sqldb").Observe(time.Since(startUpdateSignature).Seconds())
				log.CtxErrorw(ctx, "failed to write integrity hash to db", "error", err)
				return ErrGfSpDB
			}
			metrics.PerfUploadTimeHistogram.WithLabelValues("update_to_sqldb").Observe(time.Since(startUpdateSignature).Seconds())
			log.CtxDebugw(ctx, "succeed to upload payload to piece store")
			return nil
		}
		if err != nil {
			log.CtxErrorw(ctx, "stream closed abnormally", "piece_key", pieceKey, "error", err)
			return ErrClosedStream
		}
		pieceKey = u.baseApp.PieceOp().SegmentPieceKey(uploadObjectTask.GetObjectInfo().Id.Uint64(), segIdx)
		checksums = append(checksums, hash.GenerateChecksum(data))
		startPutPiece := time.Now()
		err = u.baseApp.PieceStore().PutPiece(ctx, pieceKey, data)
		if err != nil {
			metrics.PerfUploadTimeHistogram.WithLabelValues("put_to_piecestore").Observe(time.Since(startPutPiece).Seconds())
			log.CtxErrorw(ctx, "failed to put segment piece to piece store", "error", err)
			return ErrPieceStore
		}
		metrics.PerfUploadTimeHistogram.WithLabelValues("put_to_piecestore").Observe(time.Since(startPutPiece).Seconds())
		segIdx++
	}
}

func StreamReadAt(stream io.Reader, b []byte) (int, error) {
	if len(b) == 0 {
		return 0, fmt.Errorf("failed to read due to invalid args")
	}

	var (
		totalReadLen int
		curReadLen   int
		err          error
	)

	for {
		curReadLen, err = stream.Read(b[totalReadLen:])
		totalReadLen += curReadLen
		if err != nil || totalReadLen == len(b) {
			return totalReadLen, err
		}
	}
}

func (u *UploadModular) PostUploadObject(ctx context.Context, uploadObjectTask coretask.UploadObjectTask) {
}

func (u *UploadModular) QueryTasks(ctx context.Context, subKey coretask.TKey) (
	[]coretask.Task, error) {
	uploadTasks, _ := taskqueue.ScanTQueueBySubKey(u.uploadQueue, subKey)
	return uploadTasks, nil
}

func (u *UploadModular) PreResumableUploadObject(
	ctx context.Context,
	task coretask.ResumableUploadObjectTask) error {
	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed to pre upload object, task pointer dangling")
		return ErrDanglingDownloadTask
	}
	if task.GetObjectInfo().GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED {
		log.CtxErrorw(ctx, "failed to pre upload object, object not create")
		return ErrNotCreatedState
	}
	if u.uploadQueue.Has(task.Key()) {
		log.CtxErrorw(ctx, "failed to pre download object, task repeated")
		return ErrRepeatedTask
	}
	if err := u.baseApp.GfSpClient().CreateResumableUploadObject(ctx, task); err != nil {
		log.CtxErrorw(ctx, "failed to begin upload object task")
		return err
	}
	return nil
}

func (u *UploadModular) HandleResumableUploadObjectTask(
	ctx context.Context,
	task coretask.ResumableUploadObjectTask,
	stream io.Reader) error {
	if err := u.resumeableUploadQueue.Push(task); err != nil {
		log.CtxErrorw(ctx, "failed to push upload queue", "error", err)
		return ErrExceedTask
	}
	segmentSize := u.baseApp.PieceOp().MaxSegmentPieceSize(
		task.GetObjectInfo().GetPayloadSize(),
		task.GetStorageParams().GetMaxSegmentSize())
	var (
		err       error
		segIdx    uint32 = 0
		pieceKey  string
		signature []byte
		integrity []byte
		checksums [][]byte
		readN     int
		readSize  int
		data      = make([]byte, segmentSize)
	)
	defer func() {
		defer u.resumeableUploadQueue.PopByKey(task.Key())
		if err != nil {
			task.SetError(err)
		}
		log.CtxDebugw(ctx, "finish to read data from stream", "info", task.Info(),
			"read_size", readSize, "error", err)
		err = u.baseApp.GfSpClient().ReportTask(ctx, task)
	}()

	for {
		if err != nil {
			return err
		}
		data = data[0:segmentSize]
		readN, err = StreamReadAt(stream, data)
		readSize += readN
		data = data[0:readN]

		if err == io.EOF {
			err = nil
			if readN != 0 {
				pieceKey = u.baseApp.PieceOp().SegmentPieceKey(task.GetObjectInfo().Id.Uint64(), segIdx)
				checksums = append(checksums, hash.GenerateChecksum(data))
				err = u.baseApp.PieceStore().PutPiece(ctx, pieceKey, data)
				if err != nil {
					log.CtxErrorw(ctx, "put segment piece to piece store",
						"piece_key", pieceKey, "error", err)
					return ErrPieceStore
				}
			}
			if task.GetCompleted() {
				signature, integrity, err = u.baseApp.GfSpClient().SignIntegrityHash(ctx,
					task.GetObjectInfo().Id.Uint64(), checksums)
				if err != nil {
					log.CtxErrorw(ctx, "failed to sign the integrity hash", "error", err)
					return err
				}
				if !bytes.Equal(integrity, task.GetObjectInfo().GetChecksums()[0]) {
					log.CtxErrorw(ctx, "invalid integrity hash",
						"integrity", hex.EncodeToString(integrity),
						"expect", hex.EncodeToString(task.GetObjectInfo().GetChecksums()[0]))
					err = ErrInvalidIntegrity
					return ErrInvalidIntegrity
				}
				integrityMeta := &corespdb.IntegrityMeta{
					ObjectID:          task.GetObjectInfo().Id.Uint64(),
					PieceChecksumList: checksums,
					IntegrityChecksum: integrity,
					Signature:         signature,
				}
				err = u.baseApp.GfSpDB().SetObjectIntegrity(integrityMeta)
				if err != nil {
					log.CtxErrorw(ctx, "failed to write integrity hash to db", "error", err)
					return ErrGfSpDB
				}
			}

			log.CtxDebugw(ctx, "succeed to upload payload to piece store")
			return nil
		}
		if err != nil {
			log.CtxErrorw(ctx, "stream closed abnormally", "piece_key", pieceKey, "error", err)
			return ErrClosedStream
		}
		pieceKey = u.baseApp.PieceOp().SegmentPieceKey(task.GetObjectInfo().Id.Uint64(), segIdx)
		checksums = append(checksums, hash.GenerateChecksum(data))
		err = u.baseApp.PieceStore().PutPiece(ctx, pieceKey, data)
		if err != nil {
			log.CtxErrorw(ctx, "put segment piece to piece store", "error", err)
			return ErrPieceStore
		}
		segIdx++
	}

}

func (*UploadModular) PostResumableUploadObject(ctx context.Context, task coretask.ResumableUploadObjectTask) {
}

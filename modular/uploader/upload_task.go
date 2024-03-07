package uploader

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/store/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	rtyAttNum = uint(3)
	rtyAttem  = retry.Attempts(rtyAttNum)
	rtyDelay  = retry.Delay(time.Millisecond * 500)
	rtyErr    = retry.LastErrorOnly(true)
)

var (
	ErrDanglingUploadTask   = gfsperrors.Register(module.UploadModularName, http.StatusBadRequest, 110001, "OoooH... request lost, try again later")
	ErrNotCreatedState      = gfsperrors.Register(module.UploadModularName, http.StatusForbidden, 110002, "object not created state")
	ErrRepeatedTask         = gfsperrors.Register(module.UploadModularName, http.StatusNotAcceptable, 110003, "put object request repeated")
	ErrInvalidIntegrity     = gfsperrors.Register(module.UploadModularName, http.StatusNotAcceptable, 110004, "invalid payload data integrity hash")
	ErrClosedStream         = gfsperrors.Register(module.UploadModularName, http.StatusBadRequest, 110005, "upload payload data stream exception")
	ErrInvalidUploadRequest = gfsperrors.Register(module.UploadModularName, http.StatusConflict, 110006, "the object had already been fully uploaded and any further uploading attempt is not allowed")
	ErrGetObjectUploadState = gfsperrors.Register(module.UploadModularName, http.StatusInternalServerError, 110007, "failed to get upload object state")
)

func ErrPieceStoreWithDetail(detail string) *gfsperrors.GfSpError {
	return gfsperrors.Register(module.UploadModularName, http.StatusInternalServerError, 115101, detail)
}

func ErrGfSpDBWithDetail(detail string) *gfsperrors.GfSpError {
	return gfsperrors.Register(module.UploadModularName, http.StatusInternalServerError, 115001, detail)
}

func (u *UploadModular) PreUploadObject(ctx context.Context, uploadObjectTask coretask.UploadObjectTask) error {
	if uploadObjectTask == nil || uploadObjectTask.GetObjectInfo() == nil || uploadObjectTask.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed to pre upload object, task pointer dangling")
		return ErrDanglingUploadTask
	}
	if uploadObjectTask.GetObjectInfo().GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED && !uploadObjectTask.GetObjectInfo().GetIsUpdating() {
		log.CtxErrorw(ctx, "failed to pre upload object, object not created nor being updated")
		return ErrNotCreatedState
	}

	startTime := time.Now()
	has := u.uploadQueue.Has(uploadObjectTask.Key())
	metrics.PerfPutObjectTime.WithLabelValues("uploader_put_object_check_repeat_cost").Observe(time.Since(startTime).Seconds())
	if has {
		log.CtxErrorw(ctx, "failed to pre upload object, task repeated")
		return ErrRepeatedTask
	}
	createUploadTime := time.Now()
	err := u.baseApp.GfSpClient().CreateUploadObject(ctx, uploadObjectTask)
	metrics.PerfPutObjectTime.WithLabelValues("uploader_put_object_report_upload_manager_cost").Observe(time.Since(createUploadTime).Seconds())
	metrics.PerfPutObjectTime.WithLabelValues("uploader_put_object_report_upload_manager_end").Observe(time.Since(time.Unix(uploadObjectTask.GetCreateTime(), 0)).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to begin upload object task")
		return err
	}

	taskState, _, errGetUploadObjectState := u.baseApp.GfSpClient().GetUploadObjectState(ctx, uploadObjectTask.GetObjectInfo().Id.Uint64())
	if errGetUploadObjectState != nil {
		log.CtxErrorw(ctx, "failed to get upload object state")
		return ErrGetObjectUploadState
	}
	if !types.CheckAllowUploadStatus(types.TaskState(taskState)) {
		// It is not allowed to upload piece or object for an object id which had already been fully uploaded.
		log.CtxErrorw(ctx, "failed to put object as the target object had already fully uploaded")
		return ErrInvalidUploadRequest
	}

	return nil
}

func (u *UploadModular) HandleUploadObjectTask(ctx context.Context, uploadObjectTask coretask.UploadObjectTask, stream io.Reader) error {
	if err := u.uploadQueue.Push(uploadObjectTask); err != nil {
		log.CtxErrorw(ctx, "failed to push upload queue", "error", err)
		return err
	}
	defer u.uploadQueue.PopByKey(uploadObjectTask.Key())
	segmentSize := u.baseApp.PieceOp().MaxSegmentPieceSize(uploadObjectTask.GetObjectInfo().GetPayloadSize(),
		uploadObjectTask.GetStorageParams().GetMaxSegmentSize())
	var (
		err       error
		segIdx    uint32 = 0
		pieceKey  string
		integrity []byte
		checksums [][]byte
		readN     int
		readSize  int
		data      = make([]byte, segmentSize)
	)
	metrics.PerfPutObjectTime.WithLabelValues("uploader_put_object_begin_from_task_create").Observe(time.Since(time.Unix(uploadObjectTask.GetCreateTime(), 0)).Seconds())
	defer func() {
		if err != nil {
			uploadObjectTask.SetError(err)
		}

		log.CtxDebugw(ctx, "finished to read data from stream", "info", uploadObjectTask.Info(),
			"read_size", readSize, "error", err)
		uploadObjectTask.AppendLog("uploader-report-upload-task")
		metrics.PerfPutObjectTime.WithLabelValues("uploader_put_object_end_from_task_create").Observe(time.Since(time.Unix(uploadObjectTask.GetCreateTime(), 0)).Seconds())
		go func() {
			metrics.PerfPutObjectTime.WithLabelValues("uploader_put_object_before_report_manager_end").Observe(time.Since(time.Unix(uploadObjectTask.GetCreateTime(), 0)).Seconds())
			if err = retry.Do(func() error {
				return u.baseApp.GfSpClient().ReportTask(context.Background(), uploadObjectTask)
			}, rtyAttem,
				rtyDelay,
				rtyErr,
				retry.OnRetry(func(n uint, err error) {
					log.CtxErrorw(ctx, "failed to report upload object task", "error", err, "attempt", n, "max_attempts", rtyAttNum)
				})); err != nil {
				log.CtxErrorw(ctx, "failed to report upload object task", "error", err)
			}
			metrics.PerfPutObjectTime.WithLabelValues("uploader_put_object_after_report_manager_end").Observe(time.Since(time.Unix(uploadObjectTask.GetCreateTime(), 0)).Seconds())
		}()
	}()
	startTime := time.Now()
	for {
		if err != nil {
			return err
		}
		data = data[0:segmentSize]
		startReadFromGateway := time.Now()
		readN, err = StreamReadAt(stream, data)
		metrics.PerfPutObjectTime.WithLabelValues("uploader_put_object_server_read_data_cost").Observe(time.Since(startReadFromGateway).Seconds())
		metrics.PerfPutObjectTime.WithLabelValues("uploader_put_object_server_read_data_end").Observe(time.Since(time.Unix(uploadObjectTask.GetCreateTime(), 0)).Seconds())
		readSize += readN
		data = data[0:readN]

		if err == io.EOF {
			err = nil
			if readN != 0 { // the last segment piece
				pieceKey = u.baseApp.PieceOp().SegmentPieceKey(uploadObjectTask.GetObjectInfo().Id.Uint64(), segIdx, uploadObjectTask.GetObjectInfo().Version)
				checksums = append(checksums, hash.GenerateChecksum(data))
				startPutPiece := time.Now()
				err = u.baseApp.PieceStore().PutPiece(ctx, pieceKey, data)
				metrics.PerfPutObjectTime.WithLabelValues("uploader_put_object_server_put_piece_cost").Observe(time.Since(startPutPiece).Seconds())
				metrics.PerfPutObjectTime.WithLabelValues("uploader_put_object_server_put_piece_end").Observe(time.Since(startTime).Seconds())
				if err != nil {
					log.CtxErrorw(ctx, "failed to put segment piece to piece store", "piece_key", pieceKey, "error", err)
					return ErrPieceStoreWithDetail("failed to put segment piece to piece store, piece_key: " + pieceKey + ", error: " + err.Error())
				}
			}
			integrity = hash.GenerateIntegrityHash(checksums)
			expectedChecksum := uploadObjectTask.GetObjectInfo().GetChecksums()[0]
			if !bytes.Equal(integrity, expectedChecksum) {
				log.CtxErrorw(ctx, "failed to put object due to check integrity hash not consistent",
					"object_info", uploadObjectTask.GetObjectInfo(), "actual_integrity", hex.EncodeToString(integrity),
					"expected_integrity", hex.EncodeToString(expectedChecksum))
				err = ErrInvalidIntegrity
				return ErrInvalidIntegrity
			}
			startUpdateSignature := time.Now()
			if uploadObjectTask.GetObjectInfo().GetIsUpdating() {
				integrityMeta := &corespdb.ShadowIntegrityMeta{
					ObjectID:          uploadObjectTask.GetObjectInfo().Id.Uint64(),
					RedundancyIndex:   piecestore.PrimarySPRedundancyIndex,
					PieceChecksumList: checksums,
					IntegrityChecksum: integrity,
					Version:           uploadObjectTask.GetObjectInfo().GetVersion(),
				}
				err = u.baseApp.GfSpDB().SetShadowObjectIntegrity(integrityMeta)
			} else {
				integrityMeta := &corespdb.IntegrityMeta{
					ObjectID:          uploadObjectTask.GetObjectInfo().Id.Uint64(),
					RedundancyIndex:   piecestore.PrimarySPRedundancyIndex,
					PieceChecksumList: checksums,
					IntegrityChecksum: integrity,
				}
				err = u.baseApp.GfSpDB().SetObjectIntegrity(integrityMeta)
			}
			metrics.PerfPutObjectTime.WithLabelValues("uploader_put_object_set_integrity_cost").Observe(time.Since(startUpdateSignature).Seconds())
			metrics.PerfPutObjectTime.WithLabelValues("uploader_put_object_set_integrity_end").Observe(time.Since(time.Unix(uploadObjectTask.GetCreateTime(), 0)).Seconds())
			if err != nil {
				log.CtxErrorw(ctx, "failed to write integrity hash to db", "error", err)
				return ErrGfSpDBWithDetail("failed to write integrity hash to db, error: " + err.Error())
			}
			err = u.baseApp.GfSpDB().UpdateUploadProgress(&corespdb.UploadObjectMeta{
				ObjectID:  uploadObjectTask.GetObjectInfo().Id.Uint64(),
				TaskState: types.TaskState_TASK_STATE_UPLOAD_OBJECT_DONE,
			})
			if err != nil {
				log.CtxErrorw(ctx, "failed to update upload progress", "error", err)
				return ErrGfSpDBWithDetail("failed to update upload progress, error: " + err.Error())
			}
			log.CtxDebugw(ctx, "succeed to upload payload to piece store")
			return nil
		}
		if err != nil {
			log.CtxErrorw(ctx, "stream closed abnormally", "piece_key", pieceKey, "error", err)
			return ErrClosedStream
		}
		checksums = append(checksums, hash.GenerateChecksum(data))
		startPutPiece := time.Now()
		pieceKey = u.baseApp.PieceOp().SegmentPieceKey(uploadObjectTask.GetObjectInfo().Id.Uint64(), segIdx, uploadObjectTask.GetObjectInfo().Version)
		err = u.baseApp.PieceStore().PutPiece(ctx, pieceKey, data)
		metrics.PerfPutObjectTime.WithLabelValues("uploader_put_object_server_put_piece_cost").Observe(time.Since(startPutPiece).Seconds())
		metrics.PerfPutObjectTime.WithLabelValues("uploader_put_object_server_put_piece_end").Observe(time.Since(time.Unix(uploadObjectTask.GetCreateTime(), 0)).Seconds())
		if err != nil {
			log.CtxErrorw(ctx, "failed to put segment piece to piece store", "error", err)
			return ErrPieceStoreWithDetail("failed to put segment piece to piece store, error: " + err.Error())
		}
		segIdx++
	}
}

func (u *UploadModular) PostUploadObject(ctx context.Context, uploadObjectTask coretask.UploadObjectTask) {
}

func (u *UploadModular) QueryTasks(ctx context.Context, subKey coretask.TKey) ([]coretask.Task, error) {
	uploadTasks, _ := taskqueue.ScanTQueueBySubKey(u.uploadQueue, subKey)
	return uploadTasks, nil
}

func (u *UploadModular) PreResumableUploadObject(ctx context.Context, task coretask.ResumableUploadObjectTask) error {
	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		log.CtxErrorw(ctx, "failed to pre upload object, task pointer dangling")
		return ErrDanglingUploadTask
	}
	if task.GetObjectInfo().GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED && !task.GetObjectInfo().GetIsUpdating() {
		log.CtxErrorw(ctx, "failed to pre upload object, object not created or under updating")
		return ErrNotCreatedState
	}
	if u.resumeableUploadQueue.Has(task.Key()) {
		log.CtxErrorw(ctx, "failed to pre upload object, task repeated")
		return ErrRepeatedTask
	}
	if err := u.baseApp.GfSpClient().CreateResumableUploadObject(ctx, task); err != nil {
		log.CtxErrorw(ctx, "failed to begin upload object task")
		return err
	}

	taskState, _, errGetUploadObjectState := u.baseApp.GfSpClient().GetUploadObjectState(ctx, task.GetObjectInfo().Id.Uint64())
	if errGetUploadObjectState != nil {
		log.CtxErrorw(ctx, "failed to get upload object state")
		return ErrGetObjectUploadState
	}
	log.CtxInfo(ctx, "taskState is ", taskState, "task is ", task)
	if !types.CheckAllowUploadStatus(types.TaskState(taskState)) {
		// It is not allowed to upload piece or object for an object id which had already been fully uploaded.
		log.CtxErrorw(ctx, "failed to put object as the target object had already fully uploaded")
		return ErrInvalidUploadRequest
	}
	return nil
}

func (u *UploadModular) HandleResumableUploadObjectTask(ctx context.Context, task coretask.ResumableUploadObjectTask,
	stream io.Reader) error {
	if err := u.resumeableUploadQueue.Push(task); err != nil {
		log.CtxErrorw(ctx, "failed to push upload queue", "error", err)
		return err
	}
	defer u.resumeableUploadQueue.PopByKey(task.Key())

	segmentSize := u.baseApp.PieceOp().MaxSegmentPieceSize(task.GetObjectInfo().GetPayloadSize(), task.GetStorageParams().GetMaxSegmentSize())
	offset := task.GetResumeOffset()
	isUpdate := task.GetObjectInfo().GetIsUpdating()

	var (
		err      error
		segIdx   = uint32(int64(offset) / segmentSize)
		pieceKey string
		readN    int
		readSize int
		data     = make([]byte, segmentSize)
	)
	defer func() {
		if err != nil {
			task.SetError(err)
		}
		log.CtxDebugw(ctx, "finished to read data from stream", "info", task.Info(),
			"read_size", readSize, "error", err)
		go func() {
			metrics.PerfPutObjectTime.WithLabelValues("uploader_put_object_before_report_manager_end").Observe(time.Since(time.Unix(task.GetCreateTime(), 0)).Seconds())
			if err = retry.Do(func() error {
				return u.baseApp.GfSpClient().ReportTask(context.Background(), task)
			}, rtyAttem,
				rtyDelay,
				rtyErr,
				retry.OnRetry(func(n uint, err error) {
					log.CtxErrorw(ctx, "failed to report upload object task", "error", err, "attempt", n, "max_attempts", rtyAttNum)
				})); err != nil {
				log.CtxErrorw(ctx, "failed to report upload object task", "error", err)
			}
			metrics.PerfPutObjectTime.WithLabelValues("uploader_put_object_after_report_manager_end").Observe(time.Since(time.Unix(task.GetCreateTime(), 0)).Seconds())
		}()
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
				pieceKey = u.baseApp.PieceOp().SegmentPieceKey(task.GetObjectInfo().Id.Uint64(), segIdx, task.GetObjectInfo().Version)
				err = u.baseApp.PieceStore().PutPiece(ctx, pieceKey, data)
				if err != nil {
					log.CtxErrorw(ctx, "failed to put segment piece to piece store", "piece_key", pieceKey, "error", err)
					return ErrPieceStoreWithDetail(fmt.Sprintf("failed to put segment piece to piece store, piece_key: %s, error: %s",
						pieceKey, err.Error()))
				}
				err = u.updatePieceCheckSum(task, data, isUpdate)
				if err != nil {
					log.CtxErrorw(ctx, "failed to append integrity checksum to db", "error", err)
					return ErrGfSpDBWithDetail("failed to append integrity checksum to db, error: " + err.Error())
				}
			}
			if task.GetCompleted() {
				var pieceChecksumList [][]byte
				pieceChecksumList, err = u.getPieceCheckSumList(task, isUpdate)
				if err != nil {
					log.CtxErrorw(ctx, "failed to get object integrity hash", "error", err)
					return err
				}
				integrityHash := hash.GenerateIntegrityHash(pieceChecksumList)
				if !bytes.Equal(integrityHash, task.GetObjectInfo().GetChecksums()[0]) {
					log.CtxErrorw(ctx, "invalid integrity hash", "object_info", task.GetObjectInfo(),
						"actual", hex.EncodeToString(integrityHash), "expected", hex.EncodeToString(task.GetObjectInfo().GetChecksums()[0]))
					err = ErrInvalidIntegrity
					return ErrInvalidIntegrity
				}
				err = u.updateIntegrityChecksum(task.GetObjectInfo().Id.Uint64(), integrityHash, isUpdate)
				if err != nil {
					log.CtxErrorw(ctx, "failed to write integrity hash to db", "error", err)
					return ErrGfSpDBWithDetail("failed to write integrity hash to db, error: " + err.Error())
				}
				err = u.baseApp.GfSpDB().UpdateUploadProgress(&corespdb.UploadObjectMeta{
					ObjectID:  task.GetObjectInfo().Id.Uint64(),
					TaskState: types.TaskState_TASK_STATE_UPLOAD_OBJECT_DONE,
				})
				if err != nil {
					log.CtxErrorw(ctx, "failed to update upload progress", "error", err)
					return ErrGfSpDBWithDetail("failed to update upload progress, error: " + err.Error())
				}
			}

			log.CtxDebug(ctx, "succeed to upload payload to piece store")
			return nil
		}
		if err != nil {
			log.CtxErrorw(ctx, "stream closed abnormally", "piece_key", pieceKey, "error", err)
			return ErrClosedStream
		}
		pieceKey = u.baseApp.PieceOp().SegmentPieceKey(task.GetObjectInfo().Id.Uint64(), segIdx, task.GetObjectInfo().GetVersion())
		err = u.baseApp.PieceStore().PutPiece(ctx, pieceKey, data)
		if err != nil {
			log.CtxErrorw(ctx, "failed to put segment piece to piece store", "error", err)
			return ErrPieceStoreWithDetail("failed to put segment piece to piece store, error: " + err.Error())
		}
		err = u.updatePieceCheckSum(task, data, isUpdate)
		if err != nil {
			log.CtxErrorw(ctx, "failed to append integrity checksum to db", "error", err)
			return ErrGfSpDBWithDetail("failed to append integrity checksum to db, error: " + err.Error())
		}
		segIdx++
	}
}

func (*UploadModular) PostResumableUploadObject(ctx context.Context, task coretask.ResumableUploadObjectTask) {
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

func (u *UploadModular) updatePieceCheckSum(task coretask.ResumableUploadObjectTask, data []byte, isUpdate bool) error {
	var err error
	if isUpdate {
		err = u.baseApp.GfSpDB().UpdateShadowPieceChecksum(task.GetObjectInfo().Id.Uint64(), piecestore.PrimarySPRedundancyIndex, hash.GenerateChecksum(data), task.GetObjectInfo().GetVersion())
	} else {
		err = u.baseApp.GfSpDB().UpdatePieceChecksum(task.GetObjectInfo().Id.Uint64(), piecestore.PrimarySPRedundancyIndex, hash.GenerateChecksum(data))
	}
	return err
}

func (u *UploadModular) getPieceCheckSumList(task coretask.ResumableUploadObjectTask, isUpdate bool) ([][]byte, error) {
	if isUpdate {
		shadowIntegrityMeta, err := u.baseApp.GfSpDB().GetShadowObjectIntegrity(task.GetObjectInfo().Id.Uint64(), piecestore.PrimarySPRedundancyIndex)
		if err != nil {
			return nil, err
		}
		return shadowIntegrityMeta.PieceChecksumList, nil
	}
	integrityMeta, err := u.baseApp.GfSpDB().GetObjectIntegrity(task.GetObjectInfo().Id.Uint64(), piecestore.PrimarySPRedundancyIndex)
	if err != nil {
		return nil, err
	}
	return integrityMeta.PieceChecksumList, nil
}

func (u *UploadModular) updateIntegrityChecksum(objectID uint64, integrityHash []byte, isUpdate bool) error {
	var err error
	if isUpdate {
		shadowIntegrityMeta := &corespdb.ShadowIntegrityMeta{
			ObjectID:          objectID,
			RedundancyIndex:   piecestore.PrimarySPRedundancyIndex,
			IntegrityChecksum: integrityHash,
		}
		err = u.baseApp.GfSpDB().UpdateShadowIntegrityChecksum(shadowIntegrityMeta)
	} else {
		integrityMeta := &corespdb.IntegrityMeta{
			ObjectID:          objectID,
			RedundancyIndex:   piecestore.PrimarySPRedundancyIndex,
			IntegrityChecksum: integrityHash,
		}
		err = u.baseApp.GfSpDB().UpdateIntegrityChecksum(integrityMeta)
	}
	return err
}

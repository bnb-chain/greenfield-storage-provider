package uploader

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	ErrDanglingDownloadTask = gfsperrors.Register(module.UploadModularName, http.StatusBadRequest, 110001, "OoooH... request lost, try again later")
	ErrNotCreatedState      = gfsperrors.Register(module.UploadModularName, http.StatusForbidden, 110002, "object not created state")
	ErrRepeatedTask         = gfsperrors.Register(module.UploadModularName, http.StatusNotAcceptable, 110003, "put object request repeated")
	ErrExceedTask           = gfsperrors.Register(module.UploadModularName, http.StatusNotAcceptable, 110004, "OoooH... request exceed, try again later")
	ErrInvalidIntegrity     = gfsperrors.Register(module.UploadModularName, http.StatusNotAcceptable, 110005, "invalid payload data integrity hash")
	ErrClosedStream         = gfsperrors.Register(module.UploadModularName, http.StatusBadRequest, 110006, "upload payload data stream exception")
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
		return ErrExceedTask
	}
	defer u.uploadQueue.PopByKey(uploadObjectTask.Key())

	segmentSize := u.baseApp.PieceOp().MaxSegmentPieceSize(
		uploadObjectTask.GetObjectInfo().GetPayloadSize(),
		uploadObjectTask.GetStorageParams().VersionedParams.GetMaxSegmentSize())
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
		err = u.baseApp.GfSpClient().ReportTask(ctx, uploadObjectTask)
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
			if readN != 0 { // the last segment piece
				pieceKey = u.baseApp.PieceOp().SegmentPieceKey(uploadObjectTask.GetObjectInfo().Id.Uint64(), segIdx)
				checksums = append(checksums, hash.GenerateChecksum(data))
				if err = u.baseApp.PieceStore().PutPiece(ctx, pieceKey, data); err != nil {
					log.CtxErrorw(ctx, "failed to put segment piece to piece store",
						"piece_key", pieceKey, "error", err)
					return ErrPieceStore
				}
			}
			if signature, integrity, err = u.baseApp.GfSpClient().SignIntegrityHash(ctx,
				uploadObjectTask.GetObjectInfo().Id.Uint64(), checksums); err != nil {
				log.CtxErrorw(ctx, "failed to sign the integrity hash", "error", err)
				return err
			}
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
			if err = u.baseApp.GfSpDB().SetObjectIntegrity(integrityMeta); err != nil {
				log.CtxErrorw(ctx, "failed to write integrity hash to db", "error", err)
				return ErrGfSpDB
			}
			log.CtxDebugw(ctx, "succeed to upload payload to piece store")
			return nil
		}
		if err != nil {
			log.CtxErrorw(ctx, "stream closed abnormally", "piece_key", pieceKey, "error", err)
			return ErrClosedStream
		}
		pieceKey = u.baseApp.PieceOp().SegmentPieceKey(uploadObjectTask.GetObjectInfo().Id.Uint64(), segIdx)
		checksums = append(checksums, hash.GenerateChecksum(data))
		err = u.baseApp.PieceStore().PutPiece(ctx, pieceKey, data)
		if err != nil {
			log.CtxErrorw(ctx, "failed to put segment piece to piece store", "error", err)
			return ErrPieceStore
		}
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

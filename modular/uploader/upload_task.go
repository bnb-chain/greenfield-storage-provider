package uploader

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	ErrDanglingDownloadTask = gfsperrors.Register(module.UploadModularName, http.StatusInternalServerError, 110001, "OoooH... request lost, try again later")
	ErrNotCreatedState      = gfsperrors.Register(module.UploadModularName, http.StatusForbidden, 110002, "object not created state")
	ErrRepeatedTask         = gfsperrors.Register(module.UploadModularName, http.StatusBadRequest, 110003, "put object request repeated")
	ErrExceedTask           = gfsperrors.Register(module.UploadModularName, http.StatusServiceUnavailable, 110004, "OoooH... request exceed, try again later")
	ErrInvalidIntegrity     = gfsperrors.Register(module.UploadModularName, http.StatusNotAcceptable, 110005, "invalid payload data integrity hash")
	ErrClosedStream         = gfsperrors.Register(module.UploadModularName, http.StatusInternalServerError, 110006, "upload payload data stream exception")
	ErrPieceStore           = gfsperrors.Register(module.UploadModularName, http.StatusInternalServerError, 115101, "server slipped away, try again later")
	ErrGfSpDB               = gfsperrors.Register(module.UploadModularName, http.StatusInternalServerError, 115001, "server slipped away, try again later")
)

func (u *UploadModular) PreUploadObject(
	ctx context.Context,
	task coretask.UploadObjectTask) error {
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
	if err := u.baseApp.GfSpClient().CreateUploadObject(ctx, task); err != nil {
		log.CtxErrorw(ctx, "failed to begin upload object task")
		return err
	}
	return nil
}

func (u *UploadModular) HandleUploadObjectTask(
	ctx context.Context,
	task coretask.UploadObjectTask,
	stream io.Reader) error {
	if err := u.uploadQueue.Push(task); err != nil {
		log.CtxErrorw(ctx, "failed to push challenge piece queue", "error", err)
		return ErrExceedTask
	}
	segmentSize := u.baseApp.PieceOp().MaxSegmentSize(
		task.GetObjectInfo().GetPayloadSize(),
		task.GetStorageParams().VersionedParams.GetMaxSegmentSize())
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
		wg        sync.WaitGroup
	)
	defer func() {
		wg.Wait()
		defer u.uploadQueue.PopByKey(task.Key())
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
			log.CtxDebugw(ctx, "succeed to upload payload to piece store")
			return nil
		}
		if err != nil {
			log.CtxErrorw(ctx, "stream closed abnormally", "piece_key", pieceKey, "error", err)
			return ErrClosedStream
		}
		pieceKey = u.baseApp.PieceOp().SegmentPieceKey(task.GetObjectInfo().Id.Uint64(), segIdx)
		checksums = append(checksums, hash.GenerateChecksum(data))
		pieceData := make([]byte, len(data))
		copy(pieceData, data)
		//wg.Add(1)
		//go func(key string, piece []byte) {
		//	defer wg.Done()
		pieceErr := u.baseApp.PieceStore().PutPiece(ctx, pieceKey, data)
		if pieceErr != nil {
			log.CtxErrorw(ctx, "put segment piece to piece store", "error", pieceErr)
			err = ErrPieceStore
			task.SetError(pieceErr)
		}
		//}(pieceKey, pieceData)
		//segIdx++
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

func (u *UploadModular) PostUploadObject(
	ctx context.Context,
	task coretask.UploadObjectTask) {
	return
}

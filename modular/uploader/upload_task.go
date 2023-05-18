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
		segIdx                uint32 = 0
		pieceKey              string
		checksums             [][]byte
		readSize              int
		space                 = make([]byte, segmentSize)
		storeCtx, storeCancel = context.WithCancel(ctx)
		wg                    sync.WaitGroup
	)
	defer func() {
		wg.Wait()
		defer u.uploadQueue.PopByKey(task.Key())
		err := u.baseApp.GfSpClient().ReportTask(ctx, task)
		if err != nil {
			log.CtxDebugw(ctx, "failed to read data from stream", "read_size", readSize,
				"object_size", task.GetObjectInfo().GetPayloadSize(), "error", err)
		}
		log.CtxDebugw(ctx, "succeed to read data from stream", "read_size", readSize,
			"object_size", task.GetObjectInfo().GetPayloadSize())
	}()
	for {
		select {
		case <-storeCtx.Done():
			return ErrPieceStore
		default:
		}
		pieceKey = u.baseApp.PieceOp().SegmentPieceKey(task.GetObjectInfo().Id.Uint64(), segIdx)
		segIdx++
		n, err := StreamReadAt(stream, space)
		data := space[0:n]
		readSize += n
		if len(data) != 0 {
			checksums = append(checksums, hash.GenerateChecksum(data))
		}
		if err == io.EOF {
			if n != 0 {
				err = u.baseApp.PieceStore().PutPiece(ctx, pieceKey, data)
				if err != nil {
					log.CtxErrorw(ctx, "put segment piece to piece store", "piece_key", pieceKey, "error", err)
					task.SetError(ErrPieceStore)
					return ErrPieceStore
				}
			}
			signature, integrity, err := u.baseApp.GfSpClient().SignIntegrityHash(ctx,
				task.GetObjectInfo().Id.Uint64(), checksums)
			if err != nil {
				log.CtxErrorw(ctx, "failed to sign the integrity hash", "error", err)
				return err
			}
			if !bytes.Equal(integrity, task.GetObjectInfo().GetChecksums()[0]) {
				log.CtxErrorw(ctx, "invalid integrity hash", "integrity", hex.EncodeToString(integrity),
					"expect", hex.EncodeToString(task.GetObjectInfo().GetChecksums()[0]))
				task.SetError(ErrInvalidIntegrity)
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
				task.SetError(ErrGfSpDB)
				return ErrGfSpDB
			}
			log.CtxDebugw(ctx, "succeed to upload payload to piece store")
			return nil
		}
		if err != nil {
			log.CtxErrorw(ctx, "stream closed abnormally", "piece_key", pieceKey, "error", err)
			return ErrClosedStream
		}
		wg.Add(1)
		go func(key string, pieceData []byte) {
			defer wg.Done()
			select {
			case <-storeCtx.Done():
				return
			default:
			}
			err = u.baseApp.PieceStore().PutPiece(ctx, key, pieceData)
			if err != nil {
				log.CtxErrorw(ctx, "put segment piece to piece store", "error", err)
				task.SetError(ErrPieceStore)
				storeCancel()
			}
		}(pieceKey, data)
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

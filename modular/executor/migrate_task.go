package executor

import (
	"bytes"
	"context"
	"sync/atomic"

	"github.com/bnb-chain/greenfield-common/go/hash"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// HandleMigratePieceTask handle the migrate piece task, it will send requests to the exiting SP to get piece data
// integrity hash and piece checksum to migrate the receiving SP. PieceData would be written to PieceStore, integrity hash
// and piece checksum will be written sql db.
// currently get and handle data one by one; in the future, use concurrency
// storagetypes.ObjectInfo struct contains LocalVirtualGroupId field which we can use it to get a GVG consisting of
// one PrimarySP and six ordered secondarySP(the order cannot be changed). Therefore, we can know what kinds of object
// we want to migrate: primary or secondary. Now we cannot use objectInfo operator address or secondaryAddress straightly.
// We should encapsulate a new method to get.
// objectInfo->lvg->gvg->(1 primarySP, 6 secondarySPs)
func (e *ExecuteModular) HandleMigratePieceTask(ctx context.Context, task coretask.MigratePieceTask, srcSPEndpoint string) error {
	// send requests to srcSPEndpoint to get data; migrate object to own Uploader or receiver
	var (
		dataShards   = task.GetStorageParams().VersionedParams.GetRedundantDataChunkNum()
		parityShards = task.GetStorageParams().VersionedParams.GetRedundantParityChunkNum()
		ecPieceCount = dataShards + parityShards
		err          error
	)
	defer func() {
		if err != nil {
			task.SetError(err)
		}
		if task.Error() != nil {
			log.CtxErrorw(ctx, "failed to migrate piece", "error", err)
		}
	}()

	if task == nil || task.GetObjectInfo() == nil || task.GetStorageParams() == nil {
		err = ErrDanglingPointer
		return err
	}

	// index indicates the index of secondarySP, if index == 0 represents it's primarySP
	index := checkIsPrimary(task.GetObjectInfo(), srcSPEndpoint)
	log.Infow("secondarySP index", "index", index)
	if index == 0 { // get all segment pieces of an object
		segmentCount := e.baseApp.PieceOp().SegmentPieceCount(task.GetObjectInfo().GetPayloadSize(),
			task.GetStorageParams().VersionedParams.GetMaxSegmentSize())
		if err = e.migrateSegmentPieces(ctx, task, srcSPEndpoint, int(segmentCount)); err != nil {
			log.CtxErrorw(ctx, "failed to migrate segment pieces", "error", err)
			return err
		}
	} else {
		if err = e.migrateECPieces(ctx, task, srcSPEndpoint, int(ecPieceCount), index); err != nil {
			log.CtxErrorw(ctx, "failed to migrate ec pieces", "error", err)
			return err
		}
	}
	return nil
}

func checkIsPrimary(objectInfo *storagetypes.ObjectInfo, endpoint string) int {
	// TODO: check this objectID belongs to PrimarySP or SecondarySP
	// PrimarySP: return true and compute segmentCount
	// SecondarySP: return false and to get which segment ec piece
	return 0
}

func (e *ExecuteModular) migrateSegmentPieces(ctx context.Context, mTask coretask.MigratePieceTask, endpoint string,
	segmentCount int) error {
	var (
		segDataList     = make([][]byte, segmentCount)
		quitCh          = make(chan bool)
		totalSegmentNum = int32(0)
	)
	for i := 0; i < segmentCount; i++ {
		segDataList[i] = nil
		go func(segIdx int) {
			// 1. get segment data
			pieceData, err := e.doMigratingPiece(ctx, mTask, endpoint, true)
			if err != nil {
				log.CtxErrorw(ctx, "failed to migrate segment piece data", "error", err, "segment number", segIdx)
				return
			}
			segDataList[segIdx] = pieceData
			if atomic.AddInt32(&totalSegmentNum, 1) == int32(segmentCount) {
				quitCh <- true
			}
		}(i)
	}

	select {
	case <-quitCh: // all the task finished
		log.CtxInfo(ctx, "migrate all segment pieces successfully")
	}

	if err := e.setMigratePiecesMetadata(mTask, segDataList, 0); err != nil {
		log.CtxErrorw(ctx, "failed to set segment piece checksum and integrity hash into spdb", "error", err)
		return err
	}
	for i, j := range segDataList {
		pieceKey := e.baseApp.PieceOp().SegmentPieceKey(mTask.GetObjectInfo().Id.Uint64(), uint32(i))
		err := e.baseApp.PieceStore().PutPiece(ctx, pieceKey, j)
		if err != nil {
			log.CtxErrorw(ctx, "failed to put segment piece to primarySP", "pieceKey", pieceKey, "error", err)
		}
	}
	return nil
}

func (e *ExecuteModular) migrateECPieces(ctx context.Context, mTask coretask.MigratePieceTask, endpoint string,
	ecPieceCount, segIndex int) error {
	var (
		ecDataList = make([][]byte, ecPieceCount)
		quitCh     = make(chan bool)
		ecNum      = int32(0)
	)
	for i := 0; i < ecPieceCount; i++ {
		ecDataList[i] = nil
		go func(ecIdx int) {
			ecPieceData, err := e.doMigratingPiece(ctx, mTask, endpoint, false)
			if err != nil {
				log.CtxErrorw(ctx, "failed to migrate ec piece data", "error", err)
				return
			}
			ecDataList[ecIdx] = ecPieceData
			if atomic.AddInt32(&ecNum, 1) == int32(ecPieceCount) {
				quitCh <- true
			}
		}(i)
	}

	select {
	case <-quitCh:
		log.CtxInfo(ctx, "migrate all ec pieces successfully")
	}

	if err := e.setMigratePiecesMetadata(mTask, ecDataList, segIndex); err != nil {
		log.CtxErrorw(ctx, "failed to set segment piece checksum and integrity hash into spdb", "error", err)
		return err
	}
	for i, j := range ecDataList {
		pieceKey := e.baseApp.PieceOp().ECPieceKey(mTask.GetObjectInfo().Id.Uint64(), uint32(segIndex), uint32(i))
		err := e.baseApp.PieceStore().PutPiece(ctx, pieceKey, j)
		if err != nil {
			log.CtxErrorw(ctx, "failed to put ec piece to primarySP", "pieceKey", pieceKey, "error", err)
		}
	}
	return nil
}

func (e *ExecuteModular) doMigratingPiece(ctx context.Context, mTask coretask.MigratePieceTask, endpoint string,
	isPrimary bool) ([]byte, error) {
	var (
		sig       []byte
		pieceData []byte
		err       error
	)
	sig, err = e.baseApp.GfSpClient().SignMigratePiece(ctx, mTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign migrate piece task", "object_name", mTask.GetObjectInfo().GetObjectName(), "error", err)
		return nil, err
	}
	mTask.SetSignature(sig)
	// migrate primarySP or secondarySP piece data
	pieceData, err = e.baseApp.GfSpClient().MigratePieceBetweenSPs(ctx, mTask, endpoint, isPrimary)
	if err != nil {
		log.CtxErrorw(ctx, "failed to migrate piece data", "isPrimary", isPrimary, "error", err)
		return nil, err
	}
	log.CtxDebugw(ctx, "succeed to migrate piece data")
	return pieceData, nil
}

// setMigratePiecesMetadata writes piece checksum and integrity hash into db
// 1. generate piece checksum list and integrity hash
// 2. compare generated integrity hash to chain integrity hash, if they are equal write to db
func (e *ExecuteModular) setMigratePiecesMetadata(task coretask.MigratePieceTask, pieceData [][]byte, index int) error {
	pieceChecksumList := make([][]byte, len(pieceData))
	for i, v := range pieceData {
		pieceChecksumList[i] = nil
		checksum := hash.GenerateChecksum(v)
		pieceChecksumList = append(pieceChecksumList, checksum)
	}
	migratedIntegrityHash := hash.GenerateIntegrityHash(pieceChecksumList)
	var chainIntegrityHash []byte
	if index == 0 { // primarySP
		chainIntegrityHash = task.GetObjectInfo().GetChecksums()[index]
	} else { // secondarySP
		chainIntegrityHash = task.GetObjectInfo().GetChecksums()[index+1]
	}
	if !bytes.Equal(migratedIntegrityHash, chainIntegrityHash) {
		log.Errorw("migrated pieces integrity is different from integrity hash on chain", "object_name",
			task.GetObjectInfo().GetObjectName(), "object_id", task.GetObjectInfo().Id.String())
		return ErrMigratedPieceChecksum
	}

	if err := e.baseApp.GfSpDB().SetObjectIntegrity(&corespdb.IntegrityMeta{
		ObjectID:          task.GetObjectInfo().Id.Uint64(),
		IntegrityChecksum: migratedIntegrityHash,
		PieceChecksumList: pieceChecksumList,
	}); err != nil {
		log.Errorw("failed to set object integrity into spdb", "error", err)
		// TODO: define an sql db error
		return err
	}
	return nil
}

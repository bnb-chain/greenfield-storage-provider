package executor

import (
	"bytes"
	"context"
	"sync/atomic"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
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

const getPrimarySPEcIdx = -1

func (e *ExecuteModular) HandleMigratePieceTask(ctx context.Context, objectInfo *storagetypes.ObjectInfo, params *storagetypes.Params,
	srcSPEndpoint string) error {
	// send requests to srcSPEndpoint to get data; migrate object to own Uploader or receiver
	var (
		dataShards   = params.VersionedParams.GetRedundantDataChunkNum()
		parityShards = params.VersionedParams.GetRedundantParityChunkNum()
		ecPieceCount = dataShards + parityShards
		err          error
	)

	if objectInfo == nil || params == nil || srcSPEndpoint == "" {
		err = ErrDanglingPointer
		return err
	}

	index := checkIsPrimary(objectInfo, srcSPEndpoint)
	log.Infow("secondarySP index", "index", index)
	migratePiece := &gfspserver.GfSpMigratePiece{
		ObjectInfo:    objectInfo,
		StorageParams: params,
		ReplicateIdx:  0,
		EcIdx:         0,
	}
	if index < 0 { // get all segment pieces of an object
		migratePiece.EcIdx = getPrimarySPEcIdx
		migratePiece.ReplicateIdx = e.baseApp.PieceOp().SegmentPieceCount(objectInfo.GetPayloadSize(),
			params.VersionedParams.GetMaxSegmentSize())
		if err = e.migrateSegmentPieces(ctx, migratePiece, srcSPEndpoint); err != nil {
			log.CtxErrorw(ctx, "failed to migrate segment pieces", "error", err)
			return err
		}
	} else {
		migratePiece.EcIdx = int32(ecPieceCount)
		if err = e.migrateECPieces(ctx, migratePiece, srcSPEndpoint); err != nil {
			log.CtxErrorw(ctx, "failed to migrate ec pieces", "error", err)
			return err
		}
	}
	return nil
}

// TODO: to be completed
// index indicates it's primarySP or secondarySP. If index < 0 represents it's primarySP, otherwise, it's a secondarySP
func checkIsPrimary(objectInfo *storagetypes.ObjectInfo, endpoint string) int {
	// TODO: check this objectID belongs to PrimarySP or SecondarySP
	// PrimarySP: return true and compute segmentCount
	// SecondarySP: return false and to get which segment ec piece
	return 0
}

func (e *ExecuteModular) migrateSegmentPieces(ctx context.Context, migratePiece *gfspserver.GfSpMigratePiece, endpoint string) error {
	var (
		segmentCount    = migratePiece.GetReplicateIdx()
		segDataList     = make([][]byte, segmentCount)
		quitCh          = make(chan bool)
		totalSegmentNum = int32(0)
	)
	for i := 0; i < int(migratePiece.GetReplicateIdx()); i++ {
		segDataList[i] = nil
		go func(segIdx int) {
			// 1. get segment data
			pieceData, err := e.doMigratingPiece(ctx, migratePiece, endpoint)
			if err != nil {
				log.CtxErrorw(ctx, "failed to migrate segment piece data", "objectID",
					migratePiece.GetObjectInfo().Id.Uint64(), "object_name", migratePiece.GetObjectInfo().GetObjectName(),
					"segment index", segIdx, "error", err)
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

	if err := e.setMigratePiecesMetadata(migratePiece.GetObjectInfo(), segDataList, 0); err != nil {
		log.CtxErrorw(ctx, "failed to set segment piece checksum and integrity hash into spdb",
			"objectID", migratePiece.GetObjectInfo().Id.Uint64(), "object_name",
			migratePiece.GetObjectInfo().GetObjectName(), "error", err)
		return err
	}
	for i, j := range segDataList {
		pieceKey := e.baseApp.PieceOp().SegmentPieceKey(migratePiece.GetObjectInfo().Id.Uint64(), uint32(i))
		err := e.baseApp.PieceStore().PutPiece(ctx, pieceKey, j)
		if err != nil {
			log.CtxErrorw(ctx, "failed to put segment piece to primarySP", "pieceKey", pieceKey, "error", err)
		}
	}
	return nil
}

func (e *ExecuteModular) migrateECPieces(ctx context.Context, migratePiece *gfspserver.GfSpMigratePiece, endpoint string) error {
	var (
		replicateIdx = migratePiece.GetReplicateIdx()
		ecPieceCount = migratePiece.GetEcIdx()
		ecDataList   = make([][]byte, ecPieceCount)
		quitCh       = make(chan bool)
		ecNum        = int32(0)
	)
	for i := 0; i < int(ecPieceCount); i++ {
		ecDataList[i] = nil
		go func(ecIdx int) {
			ecPieceData, err := e.doMigratingPiece(ctx, migratePiece, endpoint)
			if err != nil {
				log.CtxErrorw(ctx, "failed to migrate ec piece data", "objectID",
					migratePiece.GetObjectInfo().Id.Uint64(), "object_name", migratePiece.GetObjectInfo().GetObjectName(),
					"segment index", replicateIdx, "ec index", ecIdx, "error", err)
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

	if err := e.setMigratePiecesMetadata(migratePiece.GetObjectInfo(), ecDataList, int(replicateIdx)); err != nil {
		log.CtxErrorw(ctx, "failed to set segment piece checksum and integrity hash into spdb", "error", err)
		return err
	}
	for i, j := range ecDataList {
		pieceKey := e.baseApp.PieceOp().ECPieceKey(migratePiece.GetObjectInfo().Id.Uint64(), replicateIdx, uint32(i))
		err := e.baseApp.PieceStore().PutPiece(ctx, pieceKey, j)
		if err != nil {
			log.CtxErrorw(ctx, "failed to put ec piece to primarySP", "pieceKey", pieceKey, "error", err)
		}
	}
	return nil
}

func (e *ExecuteModular) doMigratingPiece(ctx context.Context, migratePiece *gfspserver.GfSpMigratePiece,
	endpoint string) ([]byte, error) {
	var (
		sig       []byte
		pieceData []byte
		err       error
	)
	sig, err = e.baseApp.GfSpClient().SignMigratePiece(ctx, migratePiece)
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign migrate piece task", "objectID", migratePiece.GetObjectInfo().Id.Uint64(),
			"object_name", migratePiece.GetObjectInfo().GetObjectName(), "error", err)
		return nil, err
	}
	migratePiece.Signature = sig
	// migrate primarySP or secondarySP piece data
	pieceData, err = e.baseApp.GfSpClient().MigratePieceBetweenSPs(ctx, migratePiece, endpoint)
	if err != nil {
		log.CtxErrorw(ctx, "failed to migrate piece data", "objectID",
			migratePiece.GetObjectInfo().Id.Uint64(), "object_name", migratePiece.GetObjectInfo().GetObjectName(), "error", err)
		return nil, err
	}
	log.CtxDebugw(ctx, "succeed to migrate piece data", "objectID", migratePiece.GetObjectInfo().Id.Uint64(),
		"object_name", migratePiece.GetObjectInfo().GetObjectName())
	return pieceData, nil
}

// setMigratePiecesMetadata writes piece checksum and integrity hash into db
// 1. generate piece checksum list and integrity hash
// 2. compare generated integrity hash to chain integrity hash, if they are equal write to db
func (e *ExecuteModular) setMigratePiecesMetadata(objectInfo *storagetypes.ObjectInfo, pieceData [][]byte, index int) error {
	pieceChecksumList := make([][]byte, len(pieceData))
	for i, v := range pieceData {
		pieceChecksumList[i] = nil
		checksum := hash.GenerateChecksum(v)
		pieceChecksumList = append(pieceChecksumList, checksum)
	}
	migratedIntegrityHash := hash.GenerateIntegrityHash(pieceChecksumList)
	var chainIntegrityHash []byte
	if index == 0 { // primarySP
		chainIntegrityHash = objectInfo.GetChecksums()[index]
	} else { // secondarySP
		chainIntegrityHash = objectInfo.GetChecksums()[index+1]
	}
	if !bytes.Equal(migratedIntegrityHash, chainIntegrityHash) {
		log.Errorw("migrated pieces integrity is different from integrity hash on chain", "object_name",
			objectInfo.GetObjectName(), "object_id", objectInfo.Id.String())
		return ErrMigratedPieceChecksum
	}

	if err := e.baseApp.GfSpDB().SetObjectIntegrity(&corespdb.IntegrityMeta{
		ObjectID:          objectInfo.Id.Uint64(),
		IntegrityChecksum: migratedIntegrityHash,
		PieceChecksumList: pieceChecksumList,
	}); err != nil {
		log.Errorw("failed to set object integrity into spdb", "object_id", objectInfo.Id.String(),
			"object_name", objectInfo.GetObjectName(), "error", err)
		// TODO: define an sql db error
		return err
	}
	return nil
}

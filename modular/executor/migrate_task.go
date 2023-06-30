package executor

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

const primarySPECIdx = -1

// HandleMigratePieceTask handle the migrate piece task, it will send requests to the exiting SP to get piece data
// integrity hash and piece checksum to migrate the receiving SP. PieceData would be written to PieceStore, integrity hash
// and piece checksum will be written sql db.
// currently get and handle data one by one; in the future, use concurrency
// storagetypes.ObjectInfo struct contains LocalVirtualGroupId field which we can use it to get a GVG consisting of
// one PrimarySP and six ordered secondarySP(the order cannot be changed). Therefore, we can know what kinds of object
// we want to migrate: primary or secondary. Now we cannot use objectInfo operator address or secondaryAddress straightly.
// We should encapsulate a new method to get.
// objectInfo->lvg->gvg->(1 primarySP, 6 secondarySPs)
// HandleMigratePieceTask(ctx context.Context, objectInfo *storagetypes.ObjectInfo, params *storagetypes.Params, srcSPEndpoint string)
func (e *ExecuteModular) HandleMigratePieceTask(ctx context.Context, objectInfo *storagetypes.ObjectInfo, params *storagetypes.Params,
	srcSPEndpoint string) error {
	var (
		// dataShards   = params.VersionedParams.GetRedundantDataChunkNum()
		// parityShards = params.VersionedParams.GetRedundantParityChunkNum()
		// ecPieceCount = dataShards + parityShards
		err error
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
	}
	if index < 0 { // get all segment pieces of an object
		// migratePiece.ReplicateIdx = e.baseApp.PieceOp().SegmentPieceCount(objectInfo.GetPayloadSize(),
		// 	params.VersionedParams.GetMaxSegmentSize())
		migratePiece.EcIdx = primarySPECIdx
		if err = e.migrateSegmentPieces(ctx, migratePiece, srcSPEndpoint, primarySPECIdx); err != nil {
			log.CtxErrorw(ctx, "failed to migrate segment pieces", "error", err)
			return err
		}
	} else {
		// migratePiece.ReplicateIdx = e.baseApp.PieceOp().SegmentPieceCount(objectInfo.GetPayloadSize(),
		// 	params.VersionedParams.GetMaxSegmentSize())
		migratePiece.EcIdx = int32(index)
		if err = e.migrateECPieces(ctx, migratePiece, srcSPEndpoint, index); err != nil {
			log.CtxErrorw(ctx, "failed to migrate ec pieces", "error", err)
			return err
		}
	}
	return nil
}

func (e *ExecuteModular) mockHandleMigratePiece(ctx context.Context) {
	log.Info("enter mockHandleMigratePiece function")
	time.Sleep(15 * time.Second)
	log.Info("let's do!!!")
	objectOne, err := e.baseApp.Consensus().QueryObjectInfo(ctx, "migratepiece", "random_file")
	if err != nil {
		log.Errorw("failed to query objectOne", "error", err)
	}
	log.Infow("print object one detailed messages", "objectOne", objectOne)
	objectTwo, err := e.baseApp.Consensus().QueryObjectInfo(ctx, "mybu", "random")
	if err != nil {
		log.Errorw("failed to query objectTwo", "error", err)
	}
	log.Infow("print object two detailed messages", "objectTwo", objectTwo)
	params := &storagetypes.Params{
		VersionedParams: storagetypes.VersionedParams{
			MaxSegmentSize:          16777216,
			RedundantDataChunkNum:   4,
			RedundantParityChunkNum: 2,
			MinChargeSize:           1048576,
		},
		MaxPayloadSize:                   2147483648,
		MirrorBucketRelayerFee:           "250000000000000",
		MirrorBucketAckRelayerFee:        "250000000000000",
		MirrorObjectRelayerFee:           "250000000000000",
		MirrorObjectAckRelayerFee:        "250000000000000",
		MirrorGroupRelayerFee:            "250000000000000",
		MirrorGroupAckRelayerFee:         "250000000000000",
		MaxBucketsPerAccount:             100,
		DiscontinueCountingWindow:        10000,
		DiscontinueObjectMax:             18446744073709551615,
		DiscontinueBucketMax:             18446744073709551615,
		DiscontinueConfirmPeriod:         5,
		DiscontinueDeletionMax:           100,
		StalePolicyCleanupMax:            200,
		MinQuotaUpdateInterval:           2592000,
		MaxLocalVirtualGroupNumPerBucket: 10,
	}

	log.Infow("start executing migrate piece: PrimarySP", "objectInfoOne", objectOne.Id.String())
	ecPieceCount := params.VersionedParams.GetRedundantDataChunkNum() + params.VersionedParams.GetRedundantParityChunkNum()
	migratePieceOne := &gfspserver.GfSpMigratePiece{
		ObjectInfo:    objectOne,
		StorageParams: params,
	}
	// migratePieceOne.ReplicateIdx = e.baseApp.PieceOp().SegmentPieceCount(objectOne.GetPayloadSize(),
	// 	params.VersionedParams.GetMaxSegmentSize())
	if err := e.migrateSegmentPieces(ctx, migratePieceOne, "http://127.0.0.1:9033", primarySPECIdx); err != nil {
		log.CtxErrorw(ctx, "failed to migrate segment pieces", "error", err)
	}

	log.Infow("start executing migrate piece: SecondarySP", "objectInfoTwo", objectTwo.Id.String())
	migratePieceTwo := &gfspserver.GfSpMigratePiece{
		ObjectInfo:    objectTwo,
		StorageParams: params,
	}
	// migratePieceTwo.ReplicateIdx = e.baseApp.PieceOp().SegmentPieceCount(objectTwo.GetPayloadSize(),
	// 	params.VersionedParams.GetMaxSegmentSize())
	migratePieceTwo.EcIdx = int32(ecPieceCount)
	if err := e.migrateECPieces(ctx, migratePieceTwo, "http://127.0.0.1:9033", 0); err != nil {
		log.CtxErrorw(ctx, "failed to migrate ec pieces", "error", err)
	}
}

// TODO: to be completed
// index indicates it's primarySP or secondarySP. If index < 0 represents it's primarySP, otherwise, it's a secondarySP
func checkIsPrimary(objectInfo *storagetypes.ObjectInfo, endpoint string) int {
	// TODO: check this objectID belongs to PrimarySP or SecondarySP
	// PrimarySP: return true and compute segmentCount
	// SecondarySP: return false and to get which segment ec piece
	return 0
}

func (e *ExecuteModular) migrateSegmentPieces(ctx context.Context, migratePiece *gfspserver.GfSpMigratePiece, endpoint string,
	index int) error {
	var (
		segmentCount = e.baseApp.PieceOp().SegmentPieceCount(migratePiece.GetObjectInfo().GetPayloadSize(),
			migratePiece.GetStorageParams().VersionedParams.GetMaxSegmentSize())
		segDataList     = make([][]byte, segmentCount)
		quitCh          = make(chan bool)
		totalSegmentNum = int32(0)
	)
	for i := 0; i < int(segmentCount); i++ {
		segDataList[i] = nil
		go func(segIdx int) {
			// 1. get segment data
			migratePiece.ReplicateIdx = uint32(segIdx)
			migratePiece.EcIdx = primarySPECIdx
			pieceData, err := e.doPieceMigration(ctx, migratePiece, endpoint)
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

	ok := <-quitCh
	log.CtxInfow(ctx, "migrate all ec pieces successfully", "ok", ok)

	if err := e.setMigratePiecesMetadata(migratePiece.GetObjectInfo(), segDataList, index); err != nil {
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
			return err
		}
	}
	return nil
}

func (e *ExecuteModular) migrateECPieces(ctx context.Context, migratePiece *gfspserver.GfSpMigratePiece, endpoint string,
	index int) error {
	var (
		segmentCount = e.baseApp.PieceOp().SegmentPieceCount(migratePiece.GetObjectInfo().GetPayloadSize(),
			migratePiece.GetStorageParams().VersionedParams.GetMaxSegmentSize())
		ecDataList = make([][]byte, segmentCount)
		quitCh     = make(chan bool)
		ecNum      = int32(0)
	)
	for i := 0; i < int(segmentCount); i++ {
		ecDataList[i] = nil
		go func(segIdx int) {
			migratePiece.ReplicateIdx = uint32(segIdx)
			// the value of ecIdx comes from the index of secondarySP
			migratePiece.EcIdx = 0
			ecPieceData, err := e.doPieceMigration(ctx, migratePiece, endpoint)
			if err != nil {
				log.CtxErrorw(ctx, "failed to migrate ec piece data", "objectID",
					migratePiece.GetObjectInfo().Id.Uint64(), "object_name", migratePiece.GetObjectInfo().GetObjectName(),
					"segment index", segIdx, "ec index", migratePiece.GetEcIdx(), "error", err)
				return
			}
			ecDataList[segIdx] = ecPieceData
			if atomic.AddInt32(&ecNum, 1) == int32(segmentCount) {
				quitCh <- true
			}
		}(i)
	}

	ok := <-quitCh
	log.CtxInfo(ctx, "migrate all ec pieces successfully", "ok", ok)

	if err := e.setMigratePiecesMetadata(migratePiece.GetObjectInfo(), ecDataList, index); err != nil {
		log.CtxErrorw(ctx, "failed to set segment piece checksum and integrity hash into spdb", "error", err)
		return err
	}
	for i, j := range ecDataList {
		pieceKey := e.baseApp.PieceOp().ECPieceKey(migratePiece.GetObjectInfo().Id.Uint64(), uint32(i), uint32(index))
		err := e.baseApp.PieceStore().PutPiece(ctx, pieceKey, j)
		if err != nil {
			log.CtxErrorw(ctx, "failed to put ec piece to primarySP", "pieceKey", pieceKey, "error", err)
			return err
		}
	}
	return nil
}

func (e *ExecuteModular) doPieceMigration(ctx context.Context, migratePiece *gfspserver.GfSpMigratePiece,
	endpoint string) ([]byte, error) {
	var (
		// sig       []byte
		pieceData []byte
		err       error
	)
	// sig, err = e.baseApp.GfSpClient().SignMigratePiece(ctx, migratePiece)
	// if err != nil {
	// 	log.CtxErrorw(ctx, "failed to sign migrate piece task", "objectID", migratePiece.GetObjectInfo().Id.Uint64(),
	// 		"object_name", migratePiece.GetObjectInfo().GetObjectName(), "error", err)
	// 	return nil, err
	// }
	// migratePiece.Signature = sig
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
	if index < -1 || index > 5 {
		return fmt.Errorf("invalid index")
	}
	if index == -1 { // primarySP
		chainIntegrityHash = objectInfo.GetChecksums()[0]
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

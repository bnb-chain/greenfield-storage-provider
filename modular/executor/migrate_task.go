package executor

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"

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
	var err error
	if objectInfo == nil || params == nil || srcSPEndpoint == "" {
		err = ErrDanglingPointer
		return err
	}

	index := checkIsPrimary(objectInfo, srcSPEndpoint)
	log.Infow("secondarySP index", "index", index)
	migratePiece := gfspserver.GfSpMigratePiece{
		ObjectInfo:    objectInfo,
		StorageParams: params,
	}
	if index < 0 { // get all segment pieces of an object
		migratePiece.EcIdx = primarySPECIdx
		if err = e.doPieceMigration(ctx, migratePiece, srcSPEndpoint, primarySPECIdx); err != nil {
			log.CtxErrorw(ctx, "failed to migrate segment pieces", "error", err)
			return err
		}
	} else {
		migratePiece.EcIdx = int32(index)
		if err = e.doPieceMigration(ctx, migratePiece, srcSPEndpoint, index); err != nil {
			log.CtxErrorw(ctx, "failed to migrate ec pieces", "error", err)
			return err
		}
	}
	return nil
}

// TODO: mock function, this will be deleted in the future
// func (e *ExecuteModular) mockHandleMigratePiece(ctx context.Context) {
// 	log.Info("enter mockHandleMigratePiece function")
// 	time.Sleep(15 * time.Second)
// 	log.Info("let's do!!!")
// 	objectOne, err := e.baseApp.Consensus().QueryObjectInfo(ctx, "migratepiece", "random_file")
// 	if err != nil {
// 		log.Errorw("failed to query objectOne", "error", err)
// 	}
// 	log.Infow("print object one detailed messages", "objectOne", objectOne)
// 	objectTwo, err := e.baseApp.Consensus().QueryObjectInfo(ctx, "mybu", "random")
// 	if err != nil {
// 		log.Errorw("failed to query objectTwo", "error", err)
// 	}
// 	log.Infow("print object two detailed messages", "objectTwo", objectTwo)
// 	params := &storagetypes.Params{
// 		VersionedParams: storagetypes.VersionedParams{
// 			MaxSegmentSize:          16777216,
// 			RedundantDataChunkNum:   4,
// 			RedundantParityChunkNum: 2,
// 			MinChargeSize:           1048576,
// 		},
// 		MaxPayloadSize:                   2147483648,
// 		MirrorBucketRelayerFee:           "250000000000000",
// 		MirrorBucketAckRelayerFee:        "250000000000000",
// 		MirrorObjectRelayerFee:           "250000000000000",
// 		MirrorObjectAckRelayerFee:        "250000000000000",
// 		MirrorGroupRelayerFee:            "250000000000000",
// 		MirrorGroupAckRelayerFee:         "250000000000000",
// 		MaxBucketsPerAccount:             100,
// 		DiscontinueCountingWindow:        10000,
// 		DiscontinueObjectMax:             18446744073709551615,
// 		DiscontinueBucketMax:             18446744073709551615,
// 		DiscontinueConfirmPeriod:         5,
// 		DiscontinueDeletionMax:           100,
// 		StalePolicyCleanupMax:            200,
// 		MinQuotaUpdateInterval:           2592000,
// 		MaxLocalVirtualGroupNumPerBucket: 10,
// 	}
//
// 	log.Infow("start executing migrate piece: PrimarySP", "objectInfoOne", objectOne.Id.String())
// 	ecPieceCount := params.VersionedParams.GetRedundantDataChunkNum() + params.VersionedParams.GetRedundantParityChunkNum()
// 	migratePieceOne := gfspserver.GfSpMigratePiece{
// 		ObjectInfo:    objectOne,
// 		StorageParams: params,
// 	}
// 	if err := e.doPieceMigration(ctx, migratePieceOne, "http://127.0.0.1:9033", primarySPECIdx); err != nil {
// 		log.CtxErrorw(ctx, "failed to migrate segment pieces", "error", err)
// 	}
//
// 	log.Infow("start executing migrate piece: SecondarySP", "objectInfoTwo", objectTwo.Id.String())
// 	migratePieceTwo := gfspserver.GfSpMigratePiece{
// 		ObjectInfo:    objectTwo,
// 		StorageParams: params,
// 	}
// 	migratePieceTwo.EcIdx = int32(ecPieceCount)
// 	if err := e.doPieceMigration(ctx, migratePieceTwo, "http://127.0.0.1:9033", 0); err != nil {
// 		log.CtxErrorw(ctx, "failed to migrate ec pieces", "error", err)
// 	}
// }

// TODO: to be completed
// index indicates it's primarySP or secondarySP. If index < 0 represents it's primarySP, otherwise, it's a secondarySP
func checkIsPrimary(objectInfo *storagetypes.ObjectInfo, endpoint string) int {
	// TODO: check this objectID belongs to PrimarySP or SecondarySP
	// PrimarySP: return true and compute segmentCount
	// SecondarySP: return false and to get which segment ec piece
	return 0
}

// TODO: complete log info
func (e *ExecuteModular) doPieceMigration(ctx context.Context, migratePiece gfspserver.GfSpMigratePiece, endpoint string,
	index int) error {
	var (
		segmentCount = e.baseApp.PieceOp().SegmentPieceCount(migratePiece.GetObjectInfo().GetPayloadSize(),
			migratePiece.GetStorageParams().VersionedParams.GetMaxSegmentSize())
		pieceDataList = make([][]byte, segmentCount)
		quitCh        = make(chan bool)
		pieceNum      = int32(0)
	)
	for i := 0; i < int(segmentCount); i++ {
		go func(mp gfspserver.GfSpMigratePiece, segIdx int) {
			pieceDataList[segIdx] = nil
			// check migrating segment pieces or ec pieces
			mp.ReplicateIdx = uint32(segIdx)
			if index == primarySPECIdx {
				mp.EcIdx = primarySPECIdx
			} else {
				mp.EcIdx = int32(index)
			}
			pieceData, err := e.sendRequest(ctx, mp, endpoint)
			if err != nil {
				log.CtxErrorw(ctx, "failed to migrate piece data", "objectID", mp.GetObjectInfo().Id.Uint64(),
					"object_name", mp.GetObjectInfo().GetObjectName(), "segment index", segIdx, "ecIdx", index,
					"SP endpoint", endpoint, "error", err)
				return
			}
			pieceDataList[segIdx] = pieceData
			if atomic.AddInt32(&pieceNum, 1) == int32(segmentCount) {
				quitCh <- true
			}
		}(migratePiece, i)
	}

	ok := <-quitCh
	if !ok {
		log.CtxErrorw(ctx, "failed to get pieces from another SP", "SP endpoint", endpoint)
	}

	if err := e.setMigratePiecesMetadata(migratePiece.GetObjectInfo(), pieceDataList, index); err != nil {
		log.CtxErrorw(ctx, "failed to set piece checksum and integrity hash into spdb", "objectID",
			migratePiece.GetObjectInfo().Id.Uint64(), "object_name", migratePiece.GetObjectInfo().GetObjectName(), "error", err)
		return err
	}
	for i, j := range pieceDataList {
		var pieceKey string
		if index == primarySPECIdx {
			pieceKey = e.baseApp.PieceOp().SegmentPieceKey(migratePiece.GetObjectInfo().Id.Uint64(), uint32(i))
		} else {
			pieceKey = e.baseApp.PieceOp().ECPieceKey(migratePiece.GetObjectInfo().Id.Uint64(), uint32(i), uint32(index))
		}
		if err := e.baseApp.PieceStore().PutPiece(ctx, pieceKey, j); err != nil {
			log.CtxErrorw(ctx, "failed to put piece data to primarySP", "pieceKey", pieceKey, "error", err)
			return err
		}
	}
	log.Infow("migrate all pieces successfully", "objectID", migratePiece.GetObjectInfo().Id.Uint64(),
		"object_name", migratePiece.GetObjectInfo().GetObjectName(), "SP endpoint", endpoint)
	return nil
}

func (e *ExecuteModular) sendRequest(ctx context.Context, migratePiece gfspserver.GfSpMigratePiece,
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
	log.CtxInfow(ctx, "succeed to get piece from another sp", "objectID", migratePiece.GetObjectInfo().Id.Uint64(),
		"object_name", migratePiece.GetObjectInfo().GetObjectName(), "segIdx", migratePiece.GetReplicateIdx(),
		"ecIdx", migratePiece.GetEcIdx())
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
		pieceChecksumList[i] = checksum
	}
	migratedIntegrityHash := hash.GenerateIntegrityHash(pieceChecksumList)
	var chainIntegrityHash []byte
	if index < -1 || index > 5 {
		// TODO: define an error
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
	log.Infow("succeed to compute and set object integrity", "object_id", objectInfo.Id.String(),
		"object_name", objectInfo.GetObjectName())
	return nil
}

package manager

import (
	"context"
	"strings"
	"time"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	metatypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
)

const (
	defaultGCBlockSpanPerLoop           = 100
	defaultGCBlockSpanBeforeLatestBlock = 60
)

// GCWorker is responsible for releasing the space occupied by the deleted object in the piece-store.
// TODO: Will be refactored into task-node in the future.
type GCWorker struct {
	manager        *Manager
	currentGCBlock uint64 // TODO: load gc point from db
}

// Start is a non-blocking function that starts a goroutine execution logic internally.
func (w *GCWorker) Start() {
	go w.startGC()
	log.Info("start gc worker")

}

// Stop is responsible for stop gc.
func (w *GCWorker) Stop() {
	log.Info("stop gc worker")
}

// startGC starts an execution logic internally.
func (w *GCWorker) startGC() {
	var (
		gcLoopNumber          uint64
		gcObjectNumberOneLoop uint64
		height                uint64
		err                   error
		startBlock            uint64
		endBlock              uint64
		currentLatestBlock    uint64
		storageParams         *storagetypes.Params
		response              *metatypes.ListDeletedObjectsByBlockNumberRangeResponse
	)

	for {
		if gcLoopNumber%100 == 0 {
			height, err = w.manager.chain.GetCurrentHeight(context.Background())
			if err != nil {
				log.Errorw("failed to query current chain height", "error", err)
				time.Sleep(1 * time.Second)
				continue
			}
			currentLatestBlock = height
			log.Infow("succeed to fetch block height", "height", height)

			storageParams, err = w.manager.chain.QueryStorageParams(context.Background())
			if err != nil {
				log.Errorw("failed to query storage params", "error", err)
				time.Sleep(1 * time.Second)
				continue
			}
			log.Infow("succeed to fetch storage params", "storage_params", storageParams)
		}
		gcLoopNumber++
		gcObjectNumberOneLoop = 0

		startBlock = w.currentGCBlock
		endBlock = w.currentGCBlock + defaultGCBlockSpanPerLoop
		if startBlock+defaultGCBlockSpanBeforeLatestBlock > currentLatestBlock {
			log.Infow("skip gc and try again later",
				"start_block", startBlock, "latest_block", currentLatestBlock)
			time.Sleep(10 * time.Second)
			continue
		}
		if endBlock+defaultGCBlockSpanBeforeLatestBlock > currentLatestBlock {
			endBlock = currentLatestBlock - defaultGCBlockSpanBeforeLatestBlock
		}

		response, err = w.manager.metadata.ListDeletedObjectsByBlockNumberRange(context.Background(),
			&metatypes.ListDeletedObjectsByBlockNumberRangeRequest{
				StartBlockNumber: int64(startBlock),
				EndBlockNumber:   int64(endBlock),
				IsFullList:       true,
			})
		if err != nil {
			log.Warnw("failed to query deleted objects",
				"start_block", startBlock, "end_block", endBlock, "error", err)
			time.Sleep(1 * time.Second)
			continue
		}
		for _, object := range response.GetObjects() {
			gcObjectNumberOneLoop++
			// TODO: refine gc workflow by enrich metadata index.
			w.gcSegmentPiece(object.GetObjectInfo(), storageParams)
			w.gcECPiece(object.GetObjectInfo(), storageParams)
			log.Infow("succeed to gc object piece store", "object_info", object.GetObjectInfo())
		}

		log.Infow("succeed to gc one loop",
			"start_block", startBlock, "end_block", endBlock,
			"gc_object_number", gcObjectNumberOneLoop, "loop_number", gcLoopNumber)
		w.currentGCBlock = uint64(response.EndBlockNumber) + 1
	}
}

// gcSegmentPiece is used to gc segment piece.
func (w *GCWorker) gcSegmentPiece(objectInfo *storagetypes.ObjectInfo, storageParams *storagetypes.Params) {
	keyList := piecestore.GenerateObjectSegmentKeyList(objectInfo.Id.Uint64(),
		objectInfo.GetPayloadSize(), storageParams.VersionedParams.GetMaxSegmentSize())
	for _, key := range keyList {
		w.manager.pieceStore.DeletePiece(key)
	}
}

// gcECPiece is used to gc ec piece.
func (w *GCWorker) gcECPiece(objectInfo *storagetypes.ObjectInfo, storageParams *storagetypes.Params) {
	if objectInfo.GetRedundancyType() != storagetypes.REDUNDANCY_REPLICA_TYPE {
		return
	}
	for redundancyIndex, address := range objectInfo.GetSecondarySpAddresses() {
		if strings.Compare(w.manager.config.SpOperatorAddress, address) == 0 {
			keyList := piecestore.GenerateObjectECKeyList(
				objectInfo.Id.Uint64(), objectInfo.GetPayloadSize(),
				storageParams.VersionedParams.GetMaxSegmentSize(), uint64(redundancyIndex))
			for _, key := range keyList {
				w.manager.pieceStore.DeletePiece(key)
			}
		}
	}
}

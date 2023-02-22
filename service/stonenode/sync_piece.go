package stonenode

import (
	"bytes"
	"context"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// doSyncToSecondarySP send piece data to the secondary
func (node *StoneNodeService) doSyncToSecondarySP(ctx context.Context, resp *stypes.StoneHubServiceAllocStoneJobResponse,
	pieceDataBySecondary [][][]byte, urlList []string, secondarySPs []string) error {
	var (
		objectID       = resp.GetPieceJob().GetObjectId()
		payloadSize    = resp.GetPieceJob().GetPayloadSize()
		redundancyType = resp.GetPieceJob().GetRedundancyType()
	)
	// index represents which number ec in ec type, corresponding pieceData contains ec data
	// index represents which number sp in replica or inline type, it stores same two-dimensional slice
	// the length of pieceData represents the number of segments, SyncPiece is client stream interface, it sends
	// []byte data one by one, so it can be used to compute syncer server receives correct piece data
	for index, pieceData := range pieceDataBySecondary {
		go func(index int, pieceData [][]byte) {
			errMsg := &stypes.ErrMessage{}
			pieceJob := &stypes.PieceJob{
				ObjectId:       objectID,
				PayloadSize:    payloadSize,
				RedundancyType: redundancyType,
			}

			defer func() {
				// notify stone hub when an ec segment is done
				req := &stypes.StoneHubServiceDoneSecondaryPieceJobRequest{
					TraceId:    resp.GetTraceId(),
					PieceJob:   pieceJob,
					ErrMessage: errMsg,
				}
				// TBD:: according to the secondary to pick up the address
				if _, err := node.stoneHub.DoneSecondaryPieceJob(ctx, req); err != nil {
					log.CtxErrorw(ctx, "done secondary piece job to stone hub failed", "error", err)
					return
				}
			}()

			spInfo, err := sendRequest(pieceData, urlList[index], &stypes.SyncerInfo{
				ObjectId:          objectID,
				StorageProviderId: secondarySPs[index],
				PieceIndex:        uint32(index),
				PieceCount:        uint32(len(pieceData)),
				RedundancyType:    redundancyType,
			}, resp.GetTraceId())
			if err != nil {
				log.CtxErrorw(ctx, "send request to gateway failed", "error", err)
				errMsg.ErrCode = stypes.ErrCode_ERR_CODE_ERROR
				errMsg.ErrMsg = err.Error()
				return
			}

			if ok := verifyIntegrityHash(pieceData, spInfo); !ok {
				errMsg.ErrCode = stypes.ErrCode_ERR_CODE_ERROR
				errMsg.ErrMsg = merrors.ErrIntegrityHash.Error()
				return
			}
			log.Debug("verify secondary integrity hash successfully")

			pieceJob.StorageProviderSealInfo = spInfo
			log.CtxDebugw(ctx, "sync piece data to secondary", "secondary_provider", secondarySPs[index])
		}(index, pieceData)
	}
	log.Info("secondary piece job done")
	return nil
}

// verifyIntegrityHash verify secondary integrity hash is equal to local's
func verifyIntegrityHash(pieceData [][]byte, spInfo *stypes.StorageProviderSealInfo) bool {
	pieceHash := make([][]byte, 0)
	for _, value := range pieceData {
		pieceHash = append(pieceHash, hash.GenerateChecksum(value))
	}
	integrityHash := hash.GenerateIntegrityHash(pieceHash)
	if spInfo == nil || spInfo.GetIntegrityHash() == nil || !bytes.Equal(integrityHash, spInfo.GetIntegrityHash()) {
		log.Error("wrong secondary integrity hash")
		return false
	}
	log.Debugw("verify integrity hash", "local_integrity_hash", integrityHash,
		"remote_integrity_hash", spInfo.GetIntegrityHash())
	return true
}

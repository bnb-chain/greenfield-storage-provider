package stonenode

import (
	"bytes"
	"context"
	"errors"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// doSyncToSecondarySP send piece data to the secondary.
func (node *StoneNodeService) doSyncToSecondarySP(ctx context.Context, resp *stypes.StoneHubServiceAllocStoneJobResponse,
	pieceDataBySecondary [][][]byte, secondarySPs []string) error {
	var (
		objectID       = resp.GetPieceJob().GetObjectId()
		payloadSize    = resp.GetPieceJob().GetPayloadSize()
		redundancyType = resp.GetPieceJob().GetRedundancyType()
	)
	log.Infow("doSyncToSecondarySP", "pieceDataBySecondary", len(pieceDataBySecondary))
	// index在ec类型中代表第几个ec；在replica类型中代表第几sp，存储的二维数组是完全相同的
	// pieceData的长度代表有几个segment，在stream要一个接一个的发送pieceData中的[]byte，可以用来计算syncer server收到的数目是否正确
	for index, pieceData := range pieceDataBySecondary {
		log.Infow("yyyyyy", "piecedata", len(pieceData), "index", index)
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

			log.Infow("hhuuu", "piecedata", len(pieceData))
			syncResp, err := node.syncPiece(ctx, &stypes.SyncerInfo{
				ObjectId:          objectID,
				StorageProviderId: secondarySPs[index],
				PayloadSize:       payloadSize,
				PieceIndex:        uint32(index),
				PieceCount:        uint32(len(pieceData)),
				RedundancyType:    redundancyType,
			}, pieceData, resp.GetTraceId())
			// TBD:: retry alloc secondary sp and rat again.
			if err != nil {
				log.CtxErrorw(ctx, "sync to secondary piece job failed", "error", err)
				errMsg.ErrCode = stypes.ErrCode_ERR_CODE_ERROR
				errMsg.ErrMsg = err.Error()
				return
			}

			spInfo := syncResp.GetSecondarySpInfo()
			if ok := verifyIntegrityHash(pieceData, spInfo); !ok {
				log.CtxErrorw(ctx, "wrong secondary integrity hash", "error", err)
				errMsg.ErrCode = stypes.ErrCode_ERR_CODE_ERROR
				errMsg.ErrMsg = merrors.ErrIntegrityHash.Error()
				return
			}
			log.Info("verify secondary integrity hash successfully")

			pieceJob.StorageProviderSealInfo = spInfo
			log.CtxInfow(ctx, "sync piece data to secondary", "secondary_provider", secondarySPs[index])
		}(index, pieceData)
	}
	return nil
}

// verifyIntegrityHash check integrity is right
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
	log.Infow("verify integrity hash", "local_integrity_hash", integrityHash,
		"remote_integrity_hash", spInfo.GetIntegrityHash())
	return true
}

// syncPiece send rpc request to secondary storage provider to sync the piece data.
func (node *StoneNodeService) syncPiece(ctx context.Context, syncerInfo *stypes.SyncerInfo,
	pieceData [][]byte, traceID string) (*stypes.SyncerServiceSyncPieceResponse, error) {
	stream, err := node.syncer.SyncPiece(ctx)
	if err != nil {
		log.Errorw("sync secondary piece job error", "err", err)
		return nil, err
	}
	log.Infow("sync piece", "piece data", len(pieceData))

	// send data one by one to avoid exceeding rpc max msg size
	for _, value := range pieceData {
		if err := stream.Send(&stypes.SyncerServiceSyncPieceRequest{
			TraceId:    traceID,
			SyncerInfo: syncerInfo,
			PieceData:  value,
		}); err != nil {
			log.Errorw("client send request error", "error", err)
			return nil, err
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Errorw("client close error", "error", err, "traceID", resp.GetTraceId())
		return nil, err
	}
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.Errorw("sync piece sends to stone node response code is not success", "error", err, "traceID", resp.GetTraceId())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

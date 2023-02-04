package stonenode

import (
	"bytes"
	"context"
	"errors"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	ptypesv1pb "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypesv1pb "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// doSyncToSecondarySP send piece data to the secondary.
func (node *StoneNodeService) doSyncToSecondarySP(ctx context.Context, resp *stypesv1pb.StoneHubServiceAllocStoneJobResponse,
	pieceDataBySecondary map[string]map[string][]byte) error {
	var (
		objectID       = resp.GetPieceJob().GetObjectId()
		payloadSize    = resp.GetPieceJob().GetPayloadSize()
		redundancyType = resp.GetPieceJob().GetRedundancyType()
		txHash         = resp.GetTxHash()
	)
	for secondary, pieceData := range pieceDataBySecondary {
		go func(secondary string, pieceData map[string][]byte) {
			errMsg := &stypesv1pb.ErrMessage{}
			pieceJob := &stypesv1pb.PieceJob{
				TxHash:         txHash,
				ObjectId:       objectID,
				PayloadSize:    payloadSize,
				RedundancyType: redundancyType,
			}

			defer func() {
				// notify stone hub when an ec segment is done
				req := &stypesv1pb.StoneHubServiceDoneSecondaryPieceJobRequest{
					TraceId:    resp.GetTraceId(),
					TxHash:     pieceJob.GetTxHash(),
					PieceJob:   pieceJob,
					ErrMessage: errMsg,
				}
				// TBD:: according to the secondary to pick up the address
				if _, err := node.stoneHub.DoneSecondaryPieceJob(ctx, req); err != nil {
					log.CtxErrorw(ctx, "done secondary piece job to stone hub failed", "error", err)
					return
				}
			}()

			syncResp, err := node.syncPiece(ctx, &stypesv1pb.SyncerInfo{
				ObjectId:          objectID,
				TxHash:            txHash,
				StorageProviderId: secondary,
				PieceCount:        uint32(len(pieceData)),
				RedundancyType:    redundancyType,
			}, pieceData, resp.GetTraceId())
			// TBD:: retry alloc secondary sp and rat again.
			if err != nil {
				log.CtxErrorw(ctx, "sync to secondary piece job failed", "error", err)
				errMsg.ErrCode = stypesv1pb.ErrCode_ERR_CODE_ERROR
				errMsg.ErrMsg = err.Error()
				return
			}

			spInfo := syncResp.GetSecondarySpInfo()
			if ok := checkIntegrityHash(pieceData, spInfo); !ok {
				log.CtxErrorw(ctx, "wrong integrity hash", "error", err)
				errMsg.ErrCode = stypesv1pb.ErrCode_ERR_CODE_ERROR
				errMsg.ErrMsg = merrors.ErrIntegrityHash.Error()
				return
			}
			if redundancyType == ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE {
				log.Info("test kkkkkkkkkkkk", "secondary", secondary)
				pieceIndex, err := decodeSPKey(secondary)
				if err != nil {
					log.Errorw("decode sp key error", "error", err)
				}
				log.Infow("kankan pieceIndex", "pieceIndex", pieceIndex)
				pieceJob.StorageProviderSealInfo.PieceIdx = pieceIndex
			}
			pieceJob.StorageProviderSealInfo = spInfo
			log.CtxDebugw(ctx, "sync piece data to secondary", "secondary_provider", secondary)
			return
		}(secondary, pieceData)
	}
	return nil
}

// checkIntegrityHash check integrity is right
func checkIntegrityHash(pieceData map[string][]byte, spInfo *stypesv1pb.StorageProviderSealInfo) bool {
	var pieceHash [][]byte
	keys := util.GenericSortedKeys(pieceData)
	for _, key := range keys {
		pieceHash = append(pieceHash, hash.GenerateChecksum(pieceData[key]))
	}
	integrityHash := hash.GenerateIntegrityHash(pieceHash)
	if spInfo == nil || spInfo.GetIntegrityHash() == nil || !bytes.Equal(integrityHash, spInfo.GetIntegrityHash()) {
		log.Error("wrong secondary integrity hash")
		return false
	}
	log.Debugw("check integrity hash", "local_integrity_hash", integrityHash,
		"remote_integrity_hash", spInfo.GetIntegrityHash())
	return true
}

// syncPiece send rpc request to secondary storage provider to sync the piece data.
func (node *StoneNodeService) syncPiece(ctx context.Context, syncerInfo *stypesv1pb.SyncerInfo,
	pieceData map[string][]byte, traceID string) (*stypesv1pb.SyncerServiceSyncPieceResponse, error) {
	stream, err := node.syncer.SyncPiece(ctx)
	if err != nil {
		log.Errorw("sync secondary piece job error", "err", err)
		return nil, err
	}

	// send data one by one to avoid exceeding rpc max msg size
	for key, value := range pieceData {
		innerMap := make(map[string][]byte)
		innerMap[key] = value
		if err := stream.Send(&stypesv1pb.SyncerServiceSyncPieceRequest{
			TraceId:    traceID,
			SyncerInfo: syncerInfo,
			PieceData:  innerMap,
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
	if resp.GetErrMessage() != nil && resp.GetErrMessage().GetErrCode() != stypesv1pb.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
		log.Errorw("sync piece sends to stone node response code is not success", "error", err, "traceID", resp.GetTraceId())
		return nil, errors.New(resp.GetErrMessage().GetErrMsg())
	}
	return resp, nil
}

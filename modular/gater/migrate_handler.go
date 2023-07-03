package gater

import (
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const (
	primarySPECIdx = -1
)

// dest sp receive migrate gvg notify from src sp.
func (g *GateModular) notifyMigrateGVGHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		reqCtx        *RequestContext
		migrateGVGMsg []byte
	)
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(err)
			reqCtx.SetHttpCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHttpCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, _ = NewRequestContext(r, g)
	migrateGVGHeader := r.Header.Get(GnfdMigrateGVGMsgHeader)
	migrateGVGMsg, err = hex.DecodeString(migrateGVGHeader)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse migrate gvg header", "header", migrateGVGHeader)
		err = ErrDecodeMsg
		return
	}
	migrateGVG := gfsptask.GfSpMigrateGVGTask{}
	err = json.Unmarshal(migrateGVGMsg, &migrateGVG)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal migrate gvg msg", "header", migrateGVGHeader)
		err = ErrDecodeMsg
		return
	}
	// TODO: check approval.
	err = g.baseApp.GfSpClient().NotifyMigrateGVG(reqCtx.Context(), &migrateGVG)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to notify migrate gvg", "gvg", migrateGVG, "error", err)
	}
}

// migratePieceHandler handles the migrate piece request between SPs which is used in SP exiting case.
// First, gateway should verify Authorization header to ensure the requests are from correct SPs.
// Second, retrieve and get data from downloader module including: PrimarySP and SecondarySP pieces
// Third, transfer data to client which is a selected SP.
func (g *GateModular) migratePieceHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		reqCtx     *RequestContext
		migrateMsg []byte
		pieceData  []byte
	)
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(err)
			reqCtx.SetHttpCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHttpCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, _ = NewRequestContext(r, g)

	migrateHeader := r.Header.Get(GnfdMigratePieceMsgHeader)
	migrateMsg, err = hex.DecodeString(migrateHeader)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse migrate piece header", "header", migrateHeader)
		err = ErrDecodeMsg
		return
	}

	migratePiece := gfsptask.GfSpMigratePieceTask{}
	err = json.Unmarshal(migrateMsg, &migratePiece)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal migrate piece msg", "header", migrateHeader)
		err = ErrDecodeMsg
		return
	}

	// migratePieceSig := migratePiece.GetSignature()
	// _, pk, err := RecoverAddr(crypto.Keccak256(migratePiece.GetSignBytes()), migratePieceSig)
	// if err != nil {
	// 	log.CtxErrorw(reqCtx.Context(), "failed to get migrate piece address", "error", err)
	// 	err = ErrSignature
	// 	return
	// }
	//
	// if !secp256k1.VerifySignature(pk.Bytes(), crypto.Keccak256(migratePiece.GetSignBytes()), migratePieceSig[:len(migratePieceSig)-1]) {
	// 	log.CtxError(reqCtx.Context(), "failed to verify migrate piece signature")
	// 	err = ErrSignature
	// 	return
	// }

	objectInfo := migratePiece.GetObjectInfo()
	if objectInfo == nil {
		log.CtxError(reqCtx.Context(), "failed to get migrate piece object info")
		err = ErrInvalidHeader
		return
	}

	// TODO: Does this need to verify migratePiece.ObjectInfo to objectInfo on chain?
	chainObjectInfo, bucketInfo, params, err := getObjectChainMeta(reqCtx, g.baseApp, objectInfo.GetObjectName(), objectInfo.GetBucketName())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object on chain meta", "error", err)
		err = ErrInvalidHeader
		return
	}

	maxECIdx := int32(migratePiece.GetStorageParams().GetRedundantDataChunkNum()+migratePiece.GetStorageParams().GetRedundantParityChunkNum()) - 1
	objectID := migratePiece.GetObjectInfo().Id.Uint64()
	replicateIdx := migratePiece.GetReplicateIdx()
	ecIdx := migratePiece.GetEcIdx()
	if ecIdx < primarySPECIdx || ecIdx > maxECIdx {
		// TODO: customize an error to return to sp
		return
	}
	pieceTask := &gfsptask.GfSpDownloadPieceTask{}
	// if ecIdx is equal to -1, we should migrate pieces from primary SP
	if ecIdx == primarySPECIdx {
		segmentPieceKey := g.baseApp.PieceOp().SegmentPieceKey(objectID, replicateIdx)
		segmentPieceSize := g.baseApp.PieceOp().SegmentPieceSize(migratePiece.ObjectInfo.GetPayloadSize(),
			replicateIdx, migratePiece.GetStorageParams().GetMaxSegmentSize())
		log.Infow("migrate primary sp", "segmentPieceKey", segmentPieceKey, "segmentPieceSize", segmentPieceSize)
		pieceTask.InitDownloadPieceTask(chainObjectInfo, bucketInfo, params, coretask.DefaultSmallerPriority, false, "",
			uint64(segmentPieceSize), segmentPieceKey, 0, uint64(segmentPieceSize),
			g.baseApp.TaskTimeout(pieceTask, objectInfo.GetPayloadSize()), g.baseApp.TaskMaxRetry(pieceTask))
		pieceData, err = g.baseApp.GfSpClient().GetPiece(reqCtx.Context(), pieceTask)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to download segment piece", "error", err)
			return
		}
	} else { // in this case, we should migrate pieces from secondary SP
		ecPieceKey := g.baseApp.PieceOp().ECPieceKey(objectID, replicateIdx, uint32(ecIdx))
		ecPieceSize := g.baseApp.PieceOp().ECPieceSize(migratePiece.ObjectInfo.GetPayloadSize(), replicateIdx,
			migratePiece.GetStorageParams().GetMaxSegmentSize(), migratePiece.GetStorageParams().GetRedundantDataChunkNum())
		log.Infow("migrate secondary sp", "ecPieceKey", ecPieceKey, "ecPieceSize", ecPieceSize)
		pieceTask.InitDownloadPieceTask(chainObjectInfo, bucketInfo, params, coretask.DefaultSmallerPriority, false, "",
			uint64(ecPieceSize), ecPieceKey, 0, uint64(ecPieceSize),
			g.baseApp.TaskTimeout(pieceTask, objectInfo.GetPayloadSize()), g.baseApp.TaskMaxRetry(pieceTask))
		pieceData, err = g.baseApp.GfSpClient().GetPiece(reqCtx.Context(), pieceTask)
		if err != nil {
			return
		}
	}

	w.Write(pieceData)
	log.CtxInfow(reqCtx.Context(), "succeed to migrate one piece", "objectID", objectID, "replicateIdx",
		replicateIdx, "ecIdx", ecIdx)
}

package gfspclient

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"github.com/cosmos/gogoproto/proto"
)

// spilt server and client const definition avoids circular references
// TODO:: extract the common parts of http to the gfsp app layer
const (
	// ReplicateObjectPiecePath defines replicate-object path style
	ReplicateObjectPiecePath = "/greenfield/receiver/v1/replicate-piece"
	// GnfdReplicatePieceApprovalHeader defines secondary approved msg for replicating piece
	GnfdReplicatePieceApprovalHeader = "X-Gnfd-Replicate-Piece-Approval-Msg"
	// GnfdReceiveMsgHeader defines receive piece data meta
	GnfdReceiveMsgHeader = "X-Gnfd-Receive-Msg"
	// GnfdIntegrityHashHeader defines integrity hash, which is used by challenge and receiver
	GnfdIntegrityHashHeader = "X-Gnfd-Integrity-Hash"
	// GnfdIntegrityHashSignatureHeader defines integrity hash signature, which is used by receiver
	GnfdIntegrityHashSignatureHeader = "X-Gnfd-Integrity-Hash-Signature"
	// RecoveryObjectPiecePath defines recovery-object path style
	RecoveryObjectPiecePath = "/greenfield/recovery/v1/get-piece"
	// GnfdRecoveryMsgHeader defines receive piece data meta
	GnfdRecoveryMsgHeader = "X-Gnfd-Recovery-Msg"

	// MigratePiecePath defines migrate piece path which is used in SP exiting case
	MigratePiecePath = "/greenfield/migrate/v1/migrate-piece"
	// GnfdMigratePieceMsgHeader defines migrate piece msg header
	GnfdMigratePieceMsgHeader = "X-Gnfd-Migrate-Piece-Msg"
	// GnfdMigrateGVGMsgHeader defines migrate gvg msg header
	GnfdMigrateGVGMsgHeader = "X-Gnfd-Migrate-GVG-Msg"
	// NotifyMigrateSwapOutTaskPath defines dispatch migrate gvg task from src sp to dest sp.
	NotifyMigrateSwapOutTaskPath = "/greenfield/migrate/v1/notify-migrate-swap-out-task"
	// GnfdMigrateSwapOutMsgHeader defines migrate swap out msg header
	GnfdMigrateSwapOutMsgHeader = "X-Gnfd-Migrate-Swap-Out-Msg"
	// SecondarySPMigrationBucketApprovalPath defines secondary sp sign migration bucket approval
	SecondarySPMigrationBucketApprovalPath = "/greenfield/migrate/v1/migration-bucket-approval"
	// SwapOutApprovalPath defines get swap out approval path
	SwapOutApprovalPath = "/greenfield/migrate/v1/get-swap-out-approval"
	// GnfdSecondarySPMigrationBucketMsgHeader defines secondary sp migration bucket sign doc header.
	GnfdSecondarySPMigrationBucketMsgHeader = "X-Gnfd-Secondary-Migration-Bucket-Msg"
	// GnfdSecondarySPMigrationBucketApprovalHeader defines secondary sp migration bucket bls approval header.
	GnfdSecondarySPMigrationBucketApprovalHeader = "X-Gnfd-Secondary-Migration-Bucket-Approval"
	// MigrateQueryBucketQuotaPath defines query bucket quota from src sp
	MigrateQueryBucketQuotaPath = "/greenfield/migrate/v1/migrate-query-bucket-quota"
	// MigrateQueryBucketQuotaHasEnoughQuotaPath defines query bucket quota from src sp
	MigrateQueryBucketQuotaHasEnoughQuotaPath = "/greenfield/migrate/v1/migrate-query-bucket-has-enough-quota"
	// PreMigrateBucketPath defines pre migrate bucket, can lock quota for migrate bucket
	PreMigrateBucketPath = "/greenfield/migrate/v1/pre-migrate-bucket"
	// PostMigrateBucketPath defines notifying the source sp about the completion of migration bucket
	PostMigrateBucketPath = "/greenfield/migrate/v1/post-migrate-bucket"
	// GnfdMigrateBucketMsgHeader defines migrate bucket msg header
	GnfdMigrateBucketMsgHeader = "X-Gnfd-Migrate-Bucket-Msg"
	// GnfdUnsignedApprovalMsgHeader defines unsigned msg, which is used by get-approval
	GnfdUnsignedApprovalMsgHeader = "X-Gnfd-Unsigned-Msg"
	// GnfdSignedApprovalMsgHeader defines signed msg, which is used by get-approval
	GnfdSignedApprovalMsgHeader = "X-Gnfd-Signed-Msg"
)

func (s *GfSpClient) ReplicatePieceToSecondary(ctx context.Context, endpoint string, receive coretask.ReceivePieceTask,
	data []byte) error {
	req, err := http.NewRequest(http.MethodPut, endpoint+ReplicateObjectPiecePath, bytes.NewReader(data))
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect to gateway", "endpoint", endpoint, "error", err)
		return err
	}

	receiveTask := receive.(*gfsptask.GfSpReceivePieceTask)
	receiveMsg, err := json.Marshal(receiveTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to replicate piece to secondary sp due to marshal error", "error", err)
		return err
	}
	receiveHeader := hex.EncodeToString(receiveMsg)
	req.Header.Add(GnfdReceiveMsgHeader, receiveHeader)
	resp, err := s.HTTPClient(ctx).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to replicate piece, status_code(%d) endpoint(%s)", resp.StatusCode, endpoint)
	}
	return nil
}

func (s *GfSpClient) GetPieceFromECChunks(ctx context.Context, endpoint string, task coretask.RecoveryPieceTask) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint+RecoveryObjectPiecePath, nil)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect to gateway", "endpoint", endpoint, "error", err)
		return nil, err
	}

	recoveryTask := task.(*gfsptask.GfSpRecoverPieceTask)
	recoveryMsg, err := json.Marshal(recoveryTask)
	if err != nil {
		return nil, err
	}
	recoveryHeader := hex.EncodeToString(recoveryMsg)
	req.Header.Add(GnfdRecoveryMsgHeader, recoveryHeader)

	resp, err := s.HTTPClient(ctx).Do(req)
	if err != nil {
		log.CtxErrorw(ctx, "client do recovery request to SPs", "error", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get recovery piece, status_code(%d) endpoint(%s)", resp.StatusCode, endpoint)
	}

	return resp.Body, nil
}

func (s *GfSpClient) DoneReplicatePieceToSecondary(ctx context.Context, endpoint string,
	receive coretask.ReceivePieceTask) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPut, endpoint+ReplicateObjectPiecePath, nil)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect to gateway", "endpoint", endpoint, "error", err)
		return nil, err
	}

	receiveTask := receive.(*gfsptask.GfSpReceivePieceTask)
	receiveMsg, err := json.Marshal(receiveTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to done replicate to secondary sp due to marshal error", "error", err)
		return nil, err
	}
	receiveHeader := hex.EncodeToString(receiveMsg)
	req.Header.Add(GnfdReceiveMsgHeader, receiveHeader)
	resp, err := s.HTTPClient(ctx).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to done replicate piece, status_code(%d) endpoint(%s)", resp.StatusCode, endpoint)
	}
	signature, err := hex.DecodeString(resp.Header.Get(GnfdIntegrityHashSignatureHeader))
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func (s *GfSpClient) MigratePiece(ctx context.Context, gvgTask *gfsptask.GfSpMigrateGVGTask, pieceTask *gfsptask.GfSpMigratePieceTask) ([]byte, error) {
	endpoint := pieceTask.GetSrcSpEndpoint()
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", endpoint, MigratePiecePath), nil)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect to gateway", "endpoint", endpoint, "error", err)
		return nil, err
	}

	msg, err := json.Marshal(gvgTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to marshal gvg task", "gvg_task", gvgTask, "piece_task", pieceTask)
		return nil, err
	}
	req.Header.Add(GnfdMigrateGVGMsgHeader, hex.EncodeToString(msg))
	msg, err = json.Marshal(pieceTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to marshal piece task", "gvg_task", gvgTask, "piece_task", pieceTask)
		return nil, err
	}
	req.Header.Add(GnfdMigratePieceMsgHeader, hex.EncodeToString(msg))
	resp, err := s.HTTPClient(ctx).Do(req)
	if err != nil {
		log.Errorw("failed to send requests to migrate pieces", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to migrate pieces, status_code(%d), endpoint(%s)", resp.StatusCode, endpoint)
	}
	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Errorw("failed to get resp body", "error", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

// NotifyDestSPMigrateSwapOut is used to notify dest sp start migrate swap out task.
func (s *GfSpClient) NotifyDestSPMigrateSwapOut(ctx context.Context, destEndpoint string, swapOut *virtualgrouptypes.MsgSwapOut) error {
	req, err := http.NewRequest(http.MethodPost, destEndpoint+NotifyMigrateSwapOutTaskPath, nil)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect to gateway", "endpoint", destEndpoint, "error", err)
		return err
	}
	marshalSwapOut, err := json.Marshal(swapOut)
	if err != nil {
		return err
	}
	req.Header.Add(GnfdMigrateSwapOutMsgHeader, hex.EncodeToString(marshalSwapOut))
	resp, err := s.HTTPClient(ctx).Do(req)
	if err != nil {
		log.Errorw("failed to notify migrate swap out msg", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to notify migrate swap out, status_code(%d), endpoint(%s)", resp.StatusCode, destEndpoint)
	}
	return nil
}

// QueryLatestBucketQuota is used to query src sp bucket quota before send CompleteMigrateBucket Tx
func (s *GfSpClient) QueryLatestBucketQuota(ctx context.Context, endpoint string, queryMsg *gfsptask.GfSpBucketMigrationInfo) (gfsptask.GfSpBucketQuotaInfo, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint+MigrateQueryBucketQuotaPath, nil)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect to gateway", "endpoint", endpoint, "error", err)
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}
	msg, err := json.Marshal(queryMsg)
	if err != nil {
		log.CtxErrorw(ctx, "failed to query latest bucket quota due to marshal error", "query_msg", queryMsg, "error", err)
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}
	req.Header.Add(GnfdMigrateBucketMsgHeader, hex.EncodeToString(msg))
	resp, err := s.HTTPClient(ctx).Do(req)
	if err != nil {
		log.Errorw("failed to query latest migrate bucket quota msg", "query_msg", queryMsg, "error", err)
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return gfsptask.GfSpBucketQuotaInfo{}, fmt.Errorf("failed to query latest bucket quota, bucket(%s), status_code(%d), endpoint(%s)", queryMsg, resp.StatusCode, endpoint)
	}

	signedMsg, err := hex.DecodeString(resp.Header.Get(GnfdSignedApprovalMsgHeader))
	if err != nil {
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}

	quotaResult := gfsptask.GfSpBucketQuotaInfo{}
	if err = proto.Unmarshal(signedMsg, &quotaResult); err != nil {
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}

	return quotaResult, nil
}

// PreMigrateBucket is used to notify src sp and deduct bucket quota before send dest sp migrate gvg task
func (s *GfSpClient) PreMigrateBucket(ctx context.Context, srcSPEndpoint string, preMsg *gfsptask.GfSpBucketMigrationInfo) (gfsptask.GfSpBucketQuotaInfo, error) {
	req, err := http.NewRequest(http.MethodGet, srcSPEndpoint+PreMigrateBucketPath, nil)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect to gateway", "endpoint", srcSPEndpoint, "error", err)
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}
	msg, err := json.Marshal(preMsg)
	if err != nil {
		log.CtxErrorw(ctx, "failed to pre migrate bucket due to marshal error", "error", err)
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}
	req.Header.Add(GnfdMigrateBucketMsgHeader, hex.EncodeToString(msg))
	resp, err := s.HTTPClient(ctx).Do(req)
	if err != nil {
		log.Errorw("failed to pre migrate bucket msg", "bucket_migration_info", preMsg, "error", err)
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return gfsptask.GfSpBucketQuotaInfo{}, fmt.Errorf("failed to pre migrate bucket, bucket_migration_info(%s), status_code(%d), endpoint(%s)", preMsg, resp.StatusCode, srcSPEndpoint)
	}

	signedMsg, err := hex.DecodeString(resp.Header.Get(GnfdSignedApprovalMsgHeader))
	if err != nil {
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}

	quotaResult := gfsptask.GfSpBucketQuotaInfo{}
	if err = proto.Unmarshal(signedMsg, &quotaResult); err != nil {
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}

	return quotaResult, nil
}

// PostMigrateBucket is used to notify src sp the completion of bucket migrate before dest sp send CompleteMigrateBucket Tx
func (s *GfSpClient) PostMigrateBucket(ctx context.Context, srcSPEndpoint string, postMsg *gfsptask.GfSpBucketMigrationInfo) (gfsptask.GfSpBucketQuotaInfo, error) {
	bucketID := postMsg.GetBucketId()
	req, err := http.NewRequest(http.MethodGet, srcSPEndpoint+PostMigrateBucketPath, nil)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect to gateway", "endpoint", srcSPEndpoint, "error", err)
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}

	msg, err := json.Marshal(postMsg)
	if err != nil {
		log.CtxErrorw(ctx, "failed to post migrate bucket to src sp due to marshal error", "error", err)
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}
	req.Header.Add(GnfdMigrateBucketMsgHeader, hex.EncodeToString(msg))
	resp, err := s.HTTPClient(ctx).Do(req)
	if err != nil {
		log.Errorw("failed to post migrate bucket to src sp", "bucket_id", bucketID, "error", err)
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return gfsptask.GfSpBucketQuotaInfo{}, fmt.Errorf("failed to post migrate bucket to src sp, bucket(%d), status_code(%d), endpoint(%s)", bucketID, resp.StatusCode, srcSPEndpoint)
	}

	signedMsg, err := hex.DecodeString(resp.Header.Get(GnfdSignedApprovalMsgHeader))
	if err != nil {
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}

	quotaResult := gfsptask.GfSpBucketQuotaInfo{}
	if err = proto.Unmarshal(signedMsg, &quotaResult); err != nil {
		return gfsptask.GfSpBucketQuotaInfo{}, err
	}

	return quotaResult, nil
}

// QuerySPHasEnoughQuotaForMigrateBucket is used to query src sp bucket quota at approval phase
func (s *GfSpClient) QuerySPHasEnoughQuotaForMigrateBucket(ctx context.Context, srcSPEndpoint string, queryMsg *gfsptask.GfSpBucketMigrationInfo) error {
	req, err := http.NewRequest(http.MethodGet, srcSPEndpoint+MigrateQueryBucketQuotaHasEnoughQuotaPath, nil)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect to gateway", "endpoint", srcSPEndpoint, "error", err)
		return err
	}
	bucketID := queryMsg.GetBucketId()
	msg, err := json.Marshal(queryMsg)
	if err != nil {
		log.CtxErrorw(ctx, "failed to check sp has enough quota for migrate bucket due to marshal error", "error", err)
		return err
	}
	req.Header.Add(GnfdMigrateBucketMsgHeader, hex.EncodeToString(msg))
	resp, err := s.HTTPClient(ctx).Do(req)
	if err != nil {
		log.Errorw("failed to check sp has enough quota for migrate bucket", "bucket_id", bucketID, "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to check sp has enough quota for migrate bucket, bucket(%d), status_code(%d), endpoint(%s)",
			bucketID, resp.StatusCode, srcSPEndpoint)
	}

	signedMsg, err := hex.DecodeString(resp.Header.Get(GnfdSignedApprovalMsgHeader))
	if err != nil {
		return err
	}

	quotaResult := gfsptask.GfSpBucketQuotaInfo{}
	if err = proto.Unmarshal(signedMsg, &quotaResult); err != nil {
		return err
	}

	if quotaResult.GetAllowMigrate() {
		return nil
	} else {
		return fmt.Errorf("failed to check src sp has enough bucket quota, bucket(%d), status_code(%d), endpoint(%s), quota(%s)",
			bucketID, resp.StatusCode, srcSPEndpoint, quotaResult.String())
	}
}

func (s *GfSpClient) GetSecondarySPMigrationBucketApproval(ctx context.Context, secondarySPEndpoint string,
	signDoc *storagetypes.SecondarySpMigrationBucketSignDoc) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, secondarySPEndpoint+SecondarySPMigrationBucketApprovalPath, nil)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect to gateway", "secondary_sp_endpoint", secondarySPEndpoint, "error", err)
		return nil, err
	}
	msg, err := storagetypes.ModuleCdc.MarshalJSON(signDoc)
	if err != nil {
		return nil, err
	}
	req.Header.Add(GnfdSecondarySPMigrationBucketMsgHeader, hex.EncodeToString(msg))
	resp, err := s.HTTPClient(ctx).Do(req)
	if err != nil {
		log.Errorw("failed to send requests to get secondary sp migration bucket approval", "secondary_sp_endpoint",
			secondarySPEndpoint, "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get resp body, status_code(%d), endpoint(%s)", resp.StatusCode, secondarySPEndpoint)
	}
	signature, err := hex.DecodeString(resp.Header.Get(GnfdSecondarySPMigrationBucketApprovalHeader))
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func (s *GfSpClient) GetSwapOutApproval(ctx context.Context, destSPEndpoint string, swapOutApproval *virtualgrouptypes.MsgSwapOut) (
	*virtualgrouptypes.MsgSwapOut, error) {
	req, err := http.NewRequest(http.MethodGet, destSPEndpoint+SwapOutApprovalPath, nil)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect to gateway", "dest_sp_endpoint", destSPEndpoint, "error", err)
		return nil, err
	}
	msg, err := virtualgrouptypes.ModuleCdc.MarshalJSON(swapOutApproval)
	if err != nil {
		return nil, err
	}
	req.Header.Add(GnfdUnsignedApprovalMsgHeader, hex.EncodeToString(msg))
	resp, err := s.HTTPClient(ctx).Do(req)
	if err != nil {
		log.Errorw("failed to send requests to get swap out approval", "dest_sp_endpoint",
			destSPEndpoint, "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get resp body, statue_code(%d), endpoint(%s)", resp.StatusCode, destSPEndpoint)
	}
	signedMsg, err := hex.DecodeString(resp.Header.Get(GnfdSignedApprovalMsgHeader))
	if err != nil {
		return nil, err
	}
	swapOut := &virtualgrouptypes.MsgSwapOut{}
	if err = virtualgrouptypes.ModuleCdc.UnmarshalJSON(signedMsg, swapOut); err != nil {
		return nil, err
	}
	return swapOut, nil
}

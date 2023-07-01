package gfspclient

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
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
	// NotifyMigrateGVGTaskPath defines dispatch migrate gvg task from src sp to dest sp.
	NotifyMigrateGVGTaskPath = "/greenfield/migrate/v1/notify-migrate-gvg-task"
	// GnfdMigrateGVGMsgHeader defines migrate gvg msg header
	GnfdMigrateGVGMsgHeader = "X-Gnfd-Migrate-GVG-Msg"
)

func (s *GfSpClient) ReplicatePieceToSecondary(ctx context.Context, endpoint string, receive coretask.ReceivePieceTask, data []byte) error {
	req, err := http.NewRequest(http.MethodPut, endpoint+ReplicateObjectPiecePath, bytes.NewReader(data))
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect gateway", "endpoint", endpoint, "error", err)
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
		return fmt.Errorf("failed to replicate piece, StatusCode(%d) Endpoint(%s)", resp.StatusCode, endpoint)
	}
	return nil
}

func (s *GfSpClient) GetPieceFromECChunks(ctx context.Context, endpoint string, task coretask.RecoveryPieceTask) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint+RecoveryObjectPiecePath, nil)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect gateway", "endpoint", endpoint, "error", err)
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
		return nil, fmt.Errorf("failed to get recovery piece, StatusCode(%d) Endpoint(%s)", resp.StatusCode, endpoint)
	}

	return resp.Body, nil
}

func (s *GfSpClient) DoneReplicatePieceToSecondary(ctx context.Context, endpoint string,
	receive coretask.ReceivePieceTask) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPut, endpoint+ReplicateObjectPiecePath, nil)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect gateway", "endpoint", endpoint, "error", err)
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
		return nil, fmt.Errorf("failed to replicate piece, StatusCode(%d) Endpoint(%s)", resp.StatusCode, endpoint)
	}
	signature, err := hex.DecodeString(resp.Header.Get(GnfdIntegrityHashSignatureHeader))
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func (s *GfSpClient) MigratePiece(ctx context.Context, mp gfspserver.GfSpMigratePiece, endpoint string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", endpoint, MigratePiecePath), nil)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect gateway", "endpoint", endpoint, "error", err)
		return nil, err
	}

	msg, err := json.Marshal(mp)
	if err != nil {
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
		return nil, fmt.Errorf("failed to migrate pieces, StatusCode(%d), Endpoint(%s)", resp.StatusCode, endpoint)
	}
	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Errorw("failed to get resp body", "error", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

// NotifyDestSPMigrateGVG is used to notify dest sp start migrate gvg task.
// TODO: maybe need a approval.
func (s *GfSpClient) NotifyDestSPMigrateGVG(ctx context.Context, destEndpoint string, migrateTask coretask.MigrateGVGTask) error {
	req, err := http.NewRequest(http.MethodPost, destEndpoint+NotifyMigrateGVGTaskPath, nil)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect gateway", "endpoint", destEndpoint, "error", err)
		return err
	}
	msg, err := json.Marshal(migrateTask)
	if err != nil {
		return err
	}
	req.Header.Add(GnfdMigrateGVGMsgHeader, hex.EncodeToString(msg))
	resp, err := s.HTTPClient(ctx).Do(req)
	if err != nil {
		log.Errorw("failed to notify migrate gvg msg", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to notify migrate gvg, StatusCode(%d), Endpoint(%s)", resp.StatusCode, destEndpoint)
	}
	return nil
}

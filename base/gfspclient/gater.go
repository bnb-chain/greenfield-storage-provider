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
)

// spilt server and client const definition avoids circular references
// TODO:: extract the common parts of http to the gfsp app layer
const (
	// ReplicateObjectPiecePath defines replicate-object path style
	ReplicateObjectPiecePath = "/greenfield/receiver/v1/replicate-piece"
	// MigrateObjectPath defines migrate-data path which is used in SP exiting case
	MigrateObjectPath = "greenfield/admin/v1/migrate-data"
	// GnfdReplicatePieceApprovalHeader defines secondary approved msg for replicating piece
	GnfdReplicatePieceApprovalHeader = "X-Gnfd-Replicate-Piece-Approval-Msg"
	// GnfdReceiveMsgHeader defines receive piece data meta
	GnfdReceiveMsgHeader = "X-Gnfd-Receive-Msg"
	// GnfdIntegrityHashHeader defines integrity hash, which is used by challenge and receiver
	GnfdIntegrityHashHeader = "X-Gnfd-Integrity-Hash"
	// GnfdIntegrityHashSignatureHeader defines integrity hash signature, which is used by receiver
	GnfdIntegrityHashSignatureHeader = "X-Gnfd-Integrity-Hash-Signature"
	// GnfdMigratePieceMsgHeader defines migrate piece msg header
	GnfdMigratePieceMsgHeader = "X-Gnfd-Migrate-Piece-Msg"
	// GnfdIsPrimaryHeader defines response header which is used to indicate migrated data whether belongs to PrimarySP
	GnfdIsPrimaryHeader = "X-Gnfd-Is-Primary"
	//RecoveryObjectPiecePath defines recovery-object path style
	RecoveryObjectPiecePath = "/greenfield/recovery/v1/get-piece"
	// GnfdRecoveryMsgHeader defines receive piece data meta
	GnfdRecoveryMsgHeader = "X-Gnfd-Recovery-Msg"
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

func (s *GfSpClient) MigratePieceBetweenSPs(ctx context.Context, endpoint string, task coretask.MigratePieceTask) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint+MigrateObjectPath, nil)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect gateway", "endpoint", endpoint, "error", err)
		return nil, err
	}
	migratePieceTask := task.(*gfsptask.GfSpMigratePieceTask)
	msg, err := json.Marshal(migratePieceTask)
	if err != nil {
		return nil, err
	}
	req.Header.Add(GnfdMigratePieceMsgHeader, hex.EncodeToString(msg))
	req.Header.Add(GnfdIsPrimaryHeader, "true")
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

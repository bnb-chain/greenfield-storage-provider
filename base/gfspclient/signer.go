package gfspclient

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func (s *GfSpClient) SignCreateBucketApproval(ctx context.Context, bucket *storagetypes.MsgCreateBucket) (
	[]byte, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect to signer", "error", connErr)
		return nil, ErrRpcUnknown
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_CreateBucketInfo{
			CreateBucketInfo: bucket,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to sign create bucket approval", "error", err)
		return nil, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetSignature(), nil
}

func (s *GfSpClient) SignMigrateBucketApproval(ctx context.Context, bucket *storagetypes.MsgMigrateBucket) (
	[]byte, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect to signer", "error", connErr)
		return nil, ErrRpcUnknown
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_MigrateBucketInfo{
			MigrateBucketInfo: bucket,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to sign migrate bucket approval", "error", err)
		return nil, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetSignature(), nil
}

func (s *GfSpClient) SignCreateObjectApproval(ctx context.Context, object *storagetypes.MsgCreateObject) ([]byte, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect to signer", "error", connErr)
		return nil, ErrRpcUnknown
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_CreateObjectInfo{
			CreateObjectInfo: object,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to sign create object approval", "error", err)
		return nil, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetSignature(), nil
}

func (s *GfSpClient) SealObject(ctx context.Context, object *storagetypes.MsgSealObject) (string, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect to signer", "error", connErr)
		return "", ErrRpcUnknown
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_SealObjectInfo{
			SealObjectInfo: object,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to seal object approval", "error", err)
		return "", ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return "", resp.GetErr()
	}
	return resp.GetTxHash(), nil
}

func (s *GfSpClient) UpdateSPPrice(ctx context.Context, price *sptypes.MsgUpdateSpStoragePrice) (string, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect to signer", "error", connErr)
		return "", ErrRpcUnknown
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_SpStoragePrice{
			SpStoragePrice: price,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to seal object approval", "error", err)
		return "", ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return "", resp.GetErr()
	}
	return resp.GetTxHash(), nil
}

func (s *GfSpClient) CreateGlobalVirtualGroup(ctx context.Context, group *gfspserver.GfSpCreateGlobalVirtualGroup) error {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect to signer", "error", connErr)
		return ErrRpcUnknown
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_CreateGlobalVirtualGroup{
			CreateGlobalVirtualGroup: group,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to create global virtual group", "error", err)
		return ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return resp.GetErr()
	}
	return nil
}

func (s *GfSpClient) RejectUnSealObject(ctx context.Context, object *storagetypes.MsgRejectSealObject) (string, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect to signer", "error", connErr)
		return "", ErrRpcUnknown
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_RejectObjectInfo{
			RejectObjectInfo: object,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to reject unseal object approval", "error", err)
		return "", ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return "", resp.GetErr()
	}
	return resp.GetTxHash(), nil
}

func (s *GfSpClient) DiscontinueBucket(ctx context.Context, bucket *storagetypes.MsgDiscontinueBucket) (string, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect signer", "error", connErr)
		return "", ErrRpcUnknown
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_DiscontinueBucketInfo{
			DiscontinueBucketInfo: bucket,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to discontinue bucket", "error", err)
		return "", ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return "", resp.GetErr()
	}
	return resp.GetTxHash(), nil
}

func (s *GfSpClient) SignReplicatePieceApproval(ctx context.Context, task coretask.ApprovalReplicatePieceTask) ([]byte, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect to signer", "error", connErr)
		return nil, ErrRpcUnknown
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_GfspReplicatePieceApprovalTask{
			GfspReplicatePieceApprovalTask: task.(*gfsptask.GfSpReplicatePieceApprovalTask),
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to sign replicate piece approval", "error", err)
		return nil, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetSignature(), nil
}

func (s *GfSpClient) SignSecondarySealBls(ctx context.Context, objectID uint64, gvgId uint32, checksums [][]byte) ([]byte, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect to signer", "error", connErr)
		return nil, ErrRpcUnknown
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_SignSecondarySealBls{
			SignSecondarySealBls: &gfspserver.GfSpSignSecondarySealBls{
				ObjectId:             objectID,
				GlobalVirtualGroupId: gvgId,
				Checksums:            checksums,
			},
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to sign secondary bls signature", "error", err)
		return nil, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetSignature(), nil
}

func (s *GfSpClient) SignReceiveTask(ctx context.Context, receiveTask coretask.ReceivePieceTask) ([]byte, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect to signer", "error", connErr)
		return nil, ErrRpcUnknown
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_GfspReceivePieceTask{
			GfspReceivePieceTask: receiveTask.(*gfsptask.GfSpReceivePieceTask),
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to sign receive task", "error", err)
		return nil, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetSignature(), nil
}

func (s *GfSpClient) SignRecoveryTask(ctx context.Context, recoveryTask coretask.RecoveryPieceTask) ([]byte, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect to signer", "error", connErr)
		return nil, ErrRpcUnknown
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_GfspRecoverPieceTask{
			GfspRecoverPieceTask: recoveryTask.(*gfsptask.GfSpRecoverPieceTask),
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to sign recovery task", "object name", recoveryTask.GetObjectInfo().ObjectName, "error", err)
		return nil, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetSignature(), nil
}

func (s *GfSpClient) SignP2PPingMsg(ctx context.Context, ping *gfspp2p.GfSpPing) ([]byte, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect to signer", "error", connErr)
		return nil, ErrRpcUnknown
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_PingMsg{
			PingMsg: ping,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to sign p2p ping msg", "error", err)
		return nil, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetSignature(), nil
}

func (s *GfSpClient) SignP2PPongMsg(ctx context.Context, pong *gfspp2p.GfSpPong) ([]byte, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect to signer", "error", connErr)
		return nil, ErrRpcUnknown
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_PongMsg{
			PongMsg: pong,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to sign p2p pong msg", "error", err)
		return nil, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetSignature(), nil
}

func (s *GfSpClient) SignMigratePiece(ctx context.Context, task *gfsptask.GfSpMigratePieceTask) ([]byte, error) {
	conn, err := s.SignerConn(ctx)
	if err != nil {
		log.Errorw("client failed to connect to signer", "error", err)
		return nil, err
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_GfspMigratePieceTask{
			GfspMigratePieceTask: task,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to sign migrate piece", "object_name",
			task.GetObjectInfo().GetObjectName(), "error", err)
		return nil, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetSignature(), nil
}

func (s *GfSpClient) CompleteMigrateBucket(ctx context.Context, migrateBucket *storagetypes.MsgCompleteMigrateBucket) (string, error) {
	conn, err := s.SignerConn(ctx)
	if err != nil {
		log.Errorw("failed to connect to signer", "error", err)
		return "", err
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_CompleteMigrateBucket{
			CompleteMigrateBucket: migrateBucket,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to sign complete migrate bucket", "error", err)
		return "", ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return "", resp.GetErr()
	}
	return resp.GetTxHash(), nil
}

func (s *GfSpClient) SignSecondarySPMigrationBucket(ctx context.Context, signDoc *storagetypes.SecondarySpMigrationBucketSignDoc) ([]byte, error) {
	conn, err := s.SignerConn(ctx)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to connect to signer", "error", err)
		return nil, ErrRpcUnknown
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_SignSecondarySpMigrationBucket{
			SignSecondarySpMigrationBucket: signDoc,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "failed to sign secondary sp bls migration bucket", "error", err)
		return nil, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetSignature(), nil
}

func (s *GfSpClient) SignSwapOut(ctx context.Context, swapOut *virtualgrouptypes.MsgSwapOut) ([]byte, error) {
	conn, err := s.SignerConn(ctx)
	if err != nil {
		log.Errorw("failed to connect signer", "error", err)
		return nil, err
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_SignSwapOut{
			SignSwapOut: swapOut,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to sign swap out", "error", err)
		return nil, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetSignature(), nil
}

func (s *GfSpClient) SwapOut(ctx context.Context, swapOut *virtualgrouptypes.MsgSwapOut) (string, error) {
	conn, err := s.SignerConn(ctx)
	if err != nil {
		log.Errorw("failed to connect to signer", "error", err)
		return "", err
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_SwapOut{
			SwapOut: swapOut,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to sign swap out", "error", err)
		return "", ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return "", resp.GetErr()
	}
	return resp.GetTxHash(), nil
}

func (s *GfSpClient) CompleteSwapOut(ctx context.Context, completeSwapOut *virtualgrouptypes.MsgCompleteSwapOut) (string, error) {
	conn, err := s.SignerConn(ctx)
	if err != nil {
		log.Errorw("failed to connect to signer", "error", err)
		return "", err
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_CompleteSwapOut{
			CompleteSwapOut: completeSwapOut,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to sign complete swap out", "error", err)
		return "", ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return "", resp.GetErr()
	}
	return resp.GetTxHash(), nil
}

func (s *GfSpClient) SPExit(ctx context.Context, spExit *virtualgrouptypes.MsgStorageProviderExit) (string, error) {
	conn, err := s.SignerConn(ctx)
	if err != nil {
		log.Errorw("failed to connect to signer", "error", err)
		return "", err
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_SpExit{
			SpExit: spExit,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to sign sp exit", "error", err)
		return "", ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return "", resp.GetErr()
	}
	return resp.GetTxHash(), nil
}

func (s *GfSpClient) CompleteSPExit(ctx context.Context, completeSPExit *virtualgrouptypes.MsgCompleteStorageProviderExit) (string, error) {
	conn, err := s.SignerConn(ctx)
	if err != nil {
		log.Errorw("failed to connect to signer", "error", err)
		return "", err
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_CompleteSpExit{
			CompleteSpExit: completeSPExit,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to sign complete sp exit", "error", err)
		return "", ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return "", resp.GetErr()
	}
	return resp.GetTxHash(), nil
}

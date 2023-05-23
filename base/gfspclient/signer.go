package gfspclient

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func (s *GfSpClient) SignCreateBucketApproval(
	ctx context.Context,
	bucket *storagetypes.MsgCreateBucket) (
	[]byte, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect signer", "error", connErr)
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

func (s *GfSpClient) SignCreateObjectApproval(
	ctx context.Context,
	object *storagetypes.MsgCreateObject) (
	[]byte, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect signer", "error", connErr)
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

func (s *GfSpClient) SealObject(
	ctx context.Context,
	object *storagetypes.MsgSealObject) error {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect signer", "error", connErr)
		return ErrRpcUnknown
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_SealObjectInfo{
			SealObjectInfo: object,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to seal object approval", "error", err)
		return ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return resp.GetErr()
	}
	return nil
}

func (s *GfSpClient) SignReplicatePieceApproval(
	ctx context.Context,
	task coretask.ApprovalReplicatePieceTask) (
	[]byte, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect signer", "error", connErr)
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

func (s *GfSpClient) SignIntegrityHash(
	ctx context.Context,
	objectID uint64,
	checksums [][]byte) (
	[]byte, []byte, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect signer", "error", connErr)
		return nil, nil, ErrRpcUnknown
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_SignIntegrity{
			SignIntegrity: &gfspserver.GfSpSignIntegrityHash{
				ObjectId:  objectID,
				Checksums: checksums,
			},
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to sign integrity hash", "error", err)
		return nil, nil, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return nil, nil, resp.GetErr()
	}
	return resp.GetSignature(), resp.GetIntegrityHash(), nil
}

func (s *GfSpClient) SignReceiveTask(
	ctx context.Context,
	receiveTask coretask.ReceivePieceTask) (
	[]byte, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect signer", "error", connErr)
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

func (s *GfSpClient) SignP2PPingMsg(
	ctx context.Context,
	ping *gfspp2p.GfSpPing) (
	[]byte, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect signer", "error", connErr)
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

func (s *GfSpClient) SignP2PPongMsg(
	ctx context.Context,
	pong *gfspp2p.GfSpPong) (
	[]byte, error) {
	conn, connErr := s.SignerConn(ctx)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect signer", "error", connErr)
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

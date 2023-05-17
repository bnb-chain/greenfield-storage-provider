package gfspclient

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
)

func (s *GfSpClient) SignCreateBucketApproval(
	ctx context.Context,
	bucket *storagetypes.MsgCreateBucket) (
	[]byte, error) {
	conn, err := s.SignerConn(ctx)
	if err != nil {
		return nil, err
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_CreateBucketInfo{
			CreateBucketInfo: bucket,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		return nil, err
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
	conn, err := s.SignerConn(ctx)
	if err != nil {
		return nil, err
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_CreateObjectInfo{
			CreateObjectInfo: object,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetSignature(), nil
}

func (s *GfSpClient) SealObject(
	ctx context.Context,
	object *storagetypes.MsgSealObject) error {
	conn, err := s.SignerConn(ctx)
	if err != nil {
		return err
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_SealObjectInfo{
			SealObjectInfo: object,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		return err
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
	conn, err := s.SignerConn(ctx)
	if err != nil {
		return nil, err
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_GfspReplicatePieceApprovalTask{
			GfspReplicatePieceApprovalTask: task.(*gfsptask.GfSpReplicatePieceApprovalTask),
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		return nil, err
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
	conn, err := s.SignerConn(ctx)
	if err != nil {
		return nil, nil, err
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_SignIntegrity{
			SignIntegrity: &gfspserver.GfSpSignIntegrityHash{
				OnjectId:  objectID,
				Checksums: checksums,
			},
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		return nil, nil, err
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
	conn, err := s.SignerConn(ctx)
	if err != nil {
		return nil, err
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_GfspReceivePieceTask{
			GfspReceivePieceTask: receiveTask.(*gfsptask.GfSpReceivePieceTask),
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		return nil, err
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
	conn, err := s.SignerConn(ctx)
	if err != nil {
		return nil, err
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_PingMsg{
			PingMsg: ping,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		return nil, err
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
	conn, err := s.SignerConn(ctx)
	if err != nil {
		return nil, err
	}
	req := &gfspserver.GfSpSignRequest{
		Request: &gfspserver.GfSpSignRequest_PongMsg{
			PongMsg: pong,
		},
	}
	resp, err := gfspserver.NewGfSpSignServiceClient(conn).GfSpSign(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.GetErr() != nil {
		return nil, resp.GetErr()
	}
	return resp.GetSignature(), nil
}

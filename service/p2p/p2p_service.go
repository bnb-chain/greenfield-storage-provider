package p2p

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	p2ptypes "github.com/bnb-chain/greenfield-storage-provider/service/p2p/types"
)

var _ p2ptypes.P2PServiceServer = &P2PServer{}

// GetApproval asks the approval to other SP.
func (p *P2PServer) GetApproval(ctx context.Context, req *p2ptypes.GetApprovalRequest) (*p2ptypes.GetApprovalResponse, error) {
	ctx = log.Context(ctx, req)
	objectInfo := req.GetApproval().GetObjectInfo()
	accept, refuse, err := p.node.GetApproval(objectInfo, int(req.GetExpectAccept()), req.GetTimeOut())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get approval", "object_id", objectInfo.Id.Uint64(), "error", err)
		return nil, err
	}
	resp := &p2ptypes.GetApprovalResponse{
		Accept: accept,
		Refuse: refuse,
	}
	log.CtxInfow(ctx, "success to get approval", "object_id", objectInfo.Id.Uint64(), "accept", len(accept), "refuse", len(refuse))
	return resp, nil
}

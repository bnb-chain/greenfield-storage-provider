package p2pnode

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/network"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/types"
)

// pattern: /protocol-name/request-or-response-message/version
const GetApprovalRequest = "/approval/request/0.0.1"
const GetApprovalResponse = "/approval/response/0.0.1"

// ResponseChannelSize defines the approval response size
const ResponseChannelSize = 12

// ApprovalProtocol define the approval protocol and callback
// maintains requests for getting approvals in memory
type ApprovalProtocol struct {
	node     *Node
	response map[uint64]chan coretask.ApprovalReplicatePieceTask
	mux      sync.RWMutex
}

// NewApprovalProtocol return an instance of ApprovalProtocol
func NewApprovalProtocol(host *Node) *ApprovalProtocol {
	approval := &ApprovalProtocol{
		node:     host,
		response: make(map[uint64]chan coretask.ApprovalReplicatePieceTask),
	}
	host.node.SetStreamHandler(GetApprovalRequest, approval.onGetApprovalRequest)
	host.node.SetStreamHandler(GetApprovalResponse, approval.onGetApprovalResponse)
	return approval
}

// hangApprovalRequest records the approval request in memory for response to router
// notice: the caller need to call cancelApprovalRequest to delete the record
func (a *ApprovalProtocol) hangApprovalRequest(id uint64) (
	chan coretask.ApprovalReplicatePieceTask, error) {
	a.mux.Lock()
	defer a.mux.Unlock()
	if _, ok := a.response[id]; ok {
		return nil, errors.New("the get approval request is running")
	}
	a.response[id] = make(chan coretask.ApprovalReplicatePieceTask, ResponseChannelSize)
	return a.response[id], nil
}

func (a *ApprovalProtocol) cancelApprovalRequest(id uint64) {
	a.mux.Lock()
	defer a.mux.Unlock()
	if _, ok := a.response[id]; !ok {
		return
	}
	ch := a.response[id]
	delete(a.response, id)
	close(ch)
}

// notifyApprovalResponse notifies the approval response by the approval related channel
func (a *ApprovalProtocol) notifyApprovalResponse(
	resp coretask.ApprovalReplicatePieceTask) error {
	a.mux.Lock()
	defer a.mux.Unlock()
	object := resp.GetObjectInfo()
	if object == nil {
		return errors.New("approval response missing object info")
	}
	id := object.Id.Uint64()
	if _, ok := a.response[id]; !ok {
		return errors.New("approval response has been canceled")
	}
	a.response[id] <- resp
	return nil
}

// onGetApprovalRequest defines the get approval request protocol callback
func (a *ApprovalProtocol) onGetApprovalRequest(s network.Stream) {
	req := &gfsptask.GfSpReplicatePieceApprovalTask{}
	buf, err := io.ReadAll(s)
	if err != nil {
		log.Errorw("failed to read replicate piece approval request msg from stream", "error", err)
		s.Reset()
		return
	}
	s.Close()
	err = proto.Unmarshal(buf, req)
	if err != nil {
		log.Errorw("failed to unmarshal replicate piece approval request msg", "error", err)
		return
	}
	log.Debugf("%s received replicate piece approval request from %s, object_id: %d",
		s.Conn().LocalPeer(), s.Conn().RemotePeer(), req.GetObjectInfo().Id.Uint64())

	if !a.node.peers.checkSP(req.GetAskSpOperatorAddress()) {
		log.Warnw("ignore invalid sp replicate piece approval request", "sp", req.GetAskSpOperatorAddress(),
			"local", s.Conn().LocalPeer(), "remote", s.Conn().RemotePeer())
		return
	}
	if strings.Compare(req.GetAskSpOperatorAddress(), a.node.baseApp.OperateAddress()) == 0 {
		log.Warnw("ignore self replicate piece approval request", "sp", req.GetAskSpOperatorAddress(),
			"local", s.Conn().LocalPeer(), "remote", s.Conn().RemotePeer())
		return
	}
	err = VerifySignature(req.GetAskSpOperatorAddress(), req.GetSignBytes(), req.GetAskSignature())
	if err != nil {
		log.Errorw("failed to verify replicate piece approval request signature",
			"local", s.Conn().LocalPeer(), "remote", s.Conn().RemotePeer(), "error", err)
		return
	}
	ctx := context.Background()
	current, err := a.node.baseApp.Consensus().CurrentHeight(ctx)
	if err != nil {
		log.Errorw("failed to consensus get current height", "local", s.Conn().LocalPeer(),
			"remote", s.Conn().RemotePeer(), "error", err)
		return
	}
	req.SetExpiredHeight(current + a.node.secondaryApprovalExpiredHeight)
	// TODO:: customized approval strategy, if refuse will fill back resp refuse field
	signature, err := a.node.baseApp.GfSpClient().SignReplicatePieceApproval(ctx, req)
	if err != nil {
		log.Errorw("failed to sign replicate piece approval", "local", s.Conn().LocalPeer(),
			"remote", s.Conn().RemotePeer(), "error", err)
		return
	}
	req.SetApprovedSignature(signature)
	req.SetApprovedSpOperatorAddress(a.node.baseApp.OperateAddress())
	err = a.node.sendToPeer(ctx, s.Conn().RemotePeer(), GetApprovalResponse, req)
	log.Infof("%s response to %s approval request, error: %s",
		s.Conn().LocalPeer(), s.Conn().RemotePeer(), err)
}

// onGetApprovalRequest defines the get approval response protocol callback
func (a *ApprovalProtocol) onGetApprovalResponse(s network.Stream) {
	resp := &gfsptask.GfSpReplicatePieceApprovalTask{}
	buf, err := io.ReadAll(s)
	if err != nil {
		log.Errorw("failed to read replicate piece approval response msg from stream", "error", err)
		s.Reset()
		return
	}
	s.Close()
	err = proto.Unmarshal(buf, resp)
	if err != nil {
		log.Errorw("failed to unmarshal replicate piece approval response msg", "error", err)
		return
	}
	log.Debugf("%s received approval response from %s, object_id: %d",
		s.Conn().LocalPeer(), s.Conn().RemotePeer(), resp.GetObjectInfo().Id.Uint64())

	err = types.VerifySignature(resp.GetApprovedSpApprovalAddress(), resp.GetSignBytes(), resp.GetApprovedSignature())
	if err != nil {
		log.Errorw("failed to verify get approval response msg signature", "local", s.Conn().LocalPeer(), "remote", s.Conn().RemotePeer(), "error", err)
		return
	}
	if !a.node.peers.checkSP(resp.GetApprovedSpOperatorAddress()) {
		log.Warnw("ignore invalid sp approval response", "sp", resp.GetApprovedSpOperatorAddress(),
			"local", s.Conn().LocalPeer(), "remote", s.Conn().RemotePeer())
		return
	}
	if strings.Compare(resp.GetApprovedSpOperatorAddress(), a.node.baseApp.OperateAddress()) == 0 {
		log.Warnw("ignore self approval response", "sp", resp.GetApprovedSpOperatorAddress(),
			"local", s.Conn().LocalPeer(), "remote", s.Conn().RemotePeer())
		return
	}
	err = a.notifyApprovalResponse(resp)
	log.Infof("%s received approval response to %s, and notify to hang request error: %v",
		s.Conn().LocalPeer(), s.Conn().RemotePeer(), err)
}

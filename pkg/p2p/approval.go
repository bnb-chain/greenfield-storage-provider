package p2p

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/network"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/types"
)

// pattern: /protocol-name/request-or-response-message/version
const GetApprovalRequest = "/approval/request/0.0.1"
const GetApprovalResponse = "/approval/response/0.0.1"

// ResponseChannelSize defines the approval response size
const ResponseChannelSize = 9

// ValidApprovalDuration defines the default approval validity period
const ValidApprovalDuration = "1h"

// ApprovalProtocol define the approval protocol and callback
// maintains requests for getting approvals in memory
type ApprovalProtocol struct {
	node     *Node
	response map[uint64]chan *types.GetApprovalResponse
	mux      sync.RWMutex
}

// NewApprovalProtocol return an instance of ApprovalProtocol
func NewApprovalProtocol(host *Node) *ApprovalProtocol {
	approval := &ApprovalProtocol{
		node:     host,
		response: make(map[uint64]chan *types.GetApprovalResponse),
	}
	host.node.SetStreamHandler(GetApprovalRequest, approval.onGetApprovalRequest)
	host.node.SetStreamHandler(GetApprovalResponse, approval.onGetApprovalResponse)
	return approval
}

// hangApprovalRequest records the approval request in memory for response to router
// notice: the caller need to call cancelApprovalRequest to delete the record
func (a *ApprovalProtocol) hangApprovalRequest(id uint64) (chan *types.GetApprovalResponse, error) {
	a.mux.Lock()
	defer a.mux.Unlock()
	if _, ok := a.response[id]; ok {
		return nil, errors.New("the get approval request is running")
	}
	a.response[id] = make(chan *types.GetApprovalResponse, ResponseChannelSize)
	return a.response[id], nil
}

// notifyApprovalResponse notifies the approval response by the approval related channel
func (a *ApprovalProtocol) notifyApprovalResponse(resp *types.GetApprovalResponse) error {
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

// cancelApprovalRequest delete the the approval record
// TODO:: ApprovalProtocol self gc the approval record
func (a *ApprovalProtocol) cancelApprovalRequest(id uint64) {
	a.mux.Lock()
	defer a.mux.Unlock()
	if _, ok := a.response[id]; !ok {
		return
	}
	delete(a.response, id)
}

// onGetApprovalRequest defines the get approval request protocol callback
func (a *ApprovalProtocol) onGetApprovalRequest(s network.Stream) {
	req := &types.GetApprovalRequest{}
	buf, err := io.ReadAll(s)
	if err != nil {
		log.Errorw("failed to read approval request msg from stream", "error", err)
		s.Reset()
		return
	}
	s.Close()
	err = proto.Unmarshal(buf, req)
	if err != nil {
		log.Errorw("failed to unmarshal approval request msg", "error", err)
		return
	}
	log.Debugf("%s received approval request from %s, object_id: %d",
		s.Conn().LocalPeer(), s.Conn().RemotePeer(), req.GetObjectInfo().Id.Uint64())

	err = types.VerifySignature(req.GetSpOperatorAddress(), req.GetSignBytes(), req.GetSignature())
	if err != nil {
		log.Errorw("failed to verify get approval request msg signature", "local", s.Conn().LocalPeer(), "remote", s.Conn().RemotePeer(), "error", err)
		return
	}
	if !a.node.peers.checkSP(req.GetSpOperatorAddress()) {
		log.Warnw("ignore invalid sp approval request", "sp", req.GetSpOperatorAddress(),
			"local", s.Conn().LocalPeer(), "remote", s.Conn().RemotePeer())
		return
	}
	if strings.Compare(req.GetSpOperatorAddress(), a.node.SpOperatorAddress) == 0 {
		log.Warnw("ignore self approval request", "sp", req.GetSpOperatorAddress(),
			"local", s.Conn().LocalPeer(), "remote", s.Conn().RemotePeer())
		return
	}
	validTime, _ := time.ParseDuration(ValidApprovalDuration)
	resp := &types.GetApprovalResponse{
		ObjectInfo:        req.GetObjectInfo(),
		SpOperatorAddress: a.node.SpOperatorAddress,
		ExpiredTime:       time.Now().Add(validTime).Unix(),
	}
	// TODO:: customized approval strategy, if refuse will fill back resp refuse field
	resp, err = a.node.signer.SignReplicateApprovalRspMsg(context.Background(), resp)
	if err != nil {
		log.Errorw("failed to sign get approval response msg", "local", s.Conn().LocalPeer(), "remote", s.Conn().RemotePeer(), "error", err)
		return
	}
	err = a.node.sendToPeer(s.Conn().RemotePeer(), GetApprovalResponse, resp)
	log.Infof("%s response to %s approval request, error: %s",
		s.Conn().LocalPeer(), s.Conn().RemotePeer(), err)
}

// onGetApprovalRequest defines the get approval response protocol callback
func (a *ApprovalProtocol) onGetApprovalResponse(s network.Stream) {
	resp := &types.GetApprovalResponse{}
	buf, err := io.ReadAll(s)
	if err != nil {
		log.Errorw("failed to read approval response msg from stream", "error", err)
		s.Reset()
		return
	}
	s.Close()
	err = proto.Unmarshal(buf, resp)
	if err != nil {
		log.Errorw("failed to unmarshal approval response msg", "error", err)
		return
	}
	log.Debugf("%s received approval response from %s, object_id: %d",
		s.Conn().LocalPeer(), s.Conn().RemotePeer(), resp.GetObjectInfo().Id.Uint64())

	err = types.VerifySignature(resp.GetSpOperatorAddress(), resp.GetSignBytes(), resp.GetSignature())
	if err != nil {
		log.Errorw("failed to verify get approval response msg signature", "local", s.Conn().LocalPeer(), "remote", s.Conn().RemotePeer(), "error", err)
		return
	}
	if !a.node.peers.checkSP(resp.GetSpOperatorAddress()) {
		log.Warnw("ignore invalid sp approval response", "sp", resp.GetSpOperatorAddress(),
			"local", s.Conn().LocalPeer(), "remote", s.Conn().RemotePeer())
		return
	}
	if strings.Compare(resp.GetSpOperatorAddress(), a.node.SpOperatorAddress) == 0 {
		log.Warnw("ignore self approval response", "sp", resp.GetSpOperatorAddress(),
			"local", s.Conn().LocalPeer(), "remote", s.Conn().RemotePeer())
		return
	}
	err = a.notifyApprovalResponse(resp)
	log.Infof("%s received approval response to %s, and notify to hang request error: %s",
		s.Conn().LocalPeer(), s.Conn().RemotePeer(), err)
}

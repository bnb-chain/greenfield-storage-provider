package p2p

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs"
	tmlog "github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/common/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/common/service"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/node/provider"
	sp "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	signer "github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

const (
	spChannelID  = 1
	inBoundSize  = 128
	outBoundSize = 128
)

var (
	_ service.Service = (*P2PReactor)(nil)
	_ libs.Wrapper    = (*sp.SPMessage)(nil)
)

var SpChannelDescriptor = &libs.ChannelDescriptor{
	ID:                  spChannelID,
	MessageType:         new(sp.SPMessage),
	Priority:            5,
	RecvMessageCapacity: 1024,
	RecvBufferCapacity:  128,
	Name:                "secondarysp",
}

// P2PReactor implements a service to communicate with storage providers.
type P2PReactor struct {
	service.BaseService
	inCh  chan *libs.Envelope
	outCh chan *libs.Envelope

	p2pChannel libs.Channel
	chCreator  libs.ChannelCreator

	peerEvents   libs.PeerEventSubscriber
	peersManager *libs.PeerManager
	peersQuerier provider.ProviderQuerier
	signer       *signer.SignerClient
}

// NewP2PReactor returns a reference to a new reactor.
func NewP2PReactor(peerManager *libs.PeerManager, chCreator libs.ChannelCreator,
	peerEvents libs.PeerEventSubscriber, peersVerifier provider.ProviderQuerier) service.Service {
	r := &P2PReactor{
		chCreator:    chCreator,
		peerEvents:   peerEvents,
		peersManager: peerManager,
		peersQuerier: peersVerifier,
		inCh:         make(chan *libs.Envelope, inBoundSize),
		outCh:        make(chan *libs.Envelope, outBoundSize),
	}

	r.BaseService = *service.NewBaseService(tmlog.NewNopLogger(), "SecondSp", r)
	return r
}

// OnStart starts separate go routines for each p2p Channel and listens for
// envelopes on each. In addition, it also listens for peer updates and handles
// messages on that p2p channel accordingly. The caller must be sure to execute
// OnStop to ensure the outbound p2p Channels are closed.
func (r *P2PReactor) OnStart(ctx context.Context) error {
	ch, err := r.chCreator(ctx, SpChannelDescriptor)
	if err != nil {
		return err
	}

	r.p2pChannel = ch
	go r.updatePeers(ctx)
	go r.receiveRequest(ctx)
	go r.broadcastRequest(ctx)
	return nil
}

// OnStop stops the reactor by signaling to all spawned goroutines to exit and
// blocking until they all exit.
func (r *P2PReactor) OnStop() {}

// SubscribeRequest supports subscribers to consume envelope.
func (r *P2PReactor) SubscribeRequest() <-chan *libs.Envelope {
	return r.outCh
}

// PublishRequest supports broadcastRequest to peers.
func (r *P2PReactor) PublishRequest(envelope *libs.Envelope) {
	r.inCh <- envelope
}

// setSigner set signer client use to sign the approval.
func (r *P2PReactor) setSigner(signer *signer.SignerClient) {
	r.signer = signer
}

// updatePeers initiates a blocking process where we listen for and handle
// PeerUpdate messages. When the reactor is stopped, we will catch the signal and
// close the p2p PeerUpdatesCh gracefully.
func (r *P2PReactor) updatePeers(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case peerUpdate := <-r.peerEvents(ctx).Updates():
			log.Debugw("received peer update", "peer", peerUpdate.NodeID, "status", peerUpdate.Status)
			if peerUpdate.Status == libs.PeerStatusUp {
				if inWhitelist := r.peersQuerier.Check(peerUpdate.NodeID); !inWhitelist {
					log.Debugw("not found in persist peers or local db", "nodeId", peerUpdate.NodeID)
					r.p2pChannel.SendError(ctx, libs.PeerError{
						NodeID: peerUpdate.NodeID,
						Err:    errors.New("not allowed to connect to"),
						Fatal:  false,
					})
				} else {
					// TODO::update peer info
					log.Debugw("update peer statue", "peer", peerUpdate.NodeID, "status", peerUpdate.Status)
				}
			}
		}
	}
}

// receiveRequest implements a blocking event loop where we listen for p2p
// Envelope messages from the channel.
func (r *P2PReactor) receiveRequest(ctx context.Context) {
	iter := r.p2pChannel.Receive(ctx)
	for iter.Next(ctx) {
		envelope := iter.Envelope()
		if err := r.handleMessage(ctx, envelope); err != nil {
			log.Errorw("failed to process message", "ch_id", envelope.ChannelID, "envelope", envelope, "err", err)
			if serr := r.p2pChannel.SendError(ctx, libs.PeerError{
				NodeID: envelope.From,
				Err:    err,
			}); serr != nil {
				return
			}
		}
	}
}

// handleMessage handles an Envelope sent from a peer on a specific p2p Channel.
// It will handle errors and any possible panics gracefully. A caller can handle
// any error returned by sending a PeerError on the respective channel.
func (r *P2PReactor) handleMessage(ctx context.Context, envelope *libs.Envelope) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic in processing message: %v", e)
			log.Errorw("recovering from processing message panic",
				"err", err, "stack", string(debug.Stack()),
			)
		}
	}()
	log.Infow("received message", "peer", envelope.From)

	switch envelope.ChannelID {
	case spChannelID:
		err = r.handleSpMessage(ctx, envelope)
	default:
		err = fmt.Errorf("unknown channel ID (%d) for envelope (%T)", envelope.ChannelID, envelope.Message)
	}
	return
}

// handleSspMessage handles envelopes sent from peers on the channel.
// TODO::add peer join and leave messages, use the sp operator address sign the messages
func (r *P2PReactor) handleSpMessage(ctx context.Context, envelope *libs.Envelope) error {
	switch envelope.Message.(type) {
	case *sp.AskApprovalRequest:
		r.signObjectApproval(ctx, envelope)
	case *sp.AckApproval:
		r.outCh <- envelope
	case *sp.RefuseApproval:
		r.outCh <- envelope
	default:
		return fmt.Errorf("unknown mesaege type channel ID (%d) for envelope (%T)", envelope.ChannelID, envelope.Message)
	}
	return nil
}

// signObjectApproval send object approval to singer
func (r *P2PReactor) signObjectApproval(ctx context.Context, envelope *libs.Envelope) {
	msg := envelope.Message.(*sp.AskApprovalRequest).GetCreateObjectMsg()
	if r.signer == nil {
		refuseEnvelope := &libs.Envelope{
			To:        envelope.From,
			Broadcast: false,
			Message: &sp.RefuseApproval{
				CreateObjectMsg: msg,
				Reason:          "signer service is preparing",
			},
		}
		log.CtxErrorw(ctx, "refuse to sign approval")
		r.inCh <- refuseEnvelope
		return
	}
	signature, err := r.signer.SignObjectApproval(ctx, msg)
	if err != nil {
		refuseEnvelope := &libs.Envelope{
			To:        envelope.From,
			Broadcast: false,
			Message: &sp.RefuseApproval{
				CreateObjectMsg: msg,
				Reason:          err.Error(),
			},
		}
		log.CtxErrorw(ctx, "fail to sign approval", "error", err)
		r.inCh <- refuseEnvelope
		return
	}
	msg.ExpectChecksums = make([][]byte, 0)
	msg.ExpectChecksums = append(msg.ExpectChecksums, signature)
	signedApproval := &libs.Envelope{
		To:        envelope.From,
		Broadcast: false,
		Message: &sp.AckApproval{
			CreateObjectMsg: msg,
		},
	}
	r.inCh <- signedApproval
}

// broadcastRequest broadcasts request to peers.
func (r *P2PReactor) broadcastRequest(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case envelope := <-r.inCh:
			if len(envelope.To) == 0 {
				envelope.Broadcast = true
			}
			envelope.ChannelID = spChannelID
			if err := r.p2pChannel.Send(ctx, *envelope); err != nil {
				log.Errorw("error to broadcast requests", "error:", err)
				return
			}
			log.Infow("success to broadcast requests", "envelope", envelope)
		}
	}
}

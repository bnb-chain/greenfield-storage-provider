package reactor

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	tmlog "github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/common/log"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/common/service"
	sp "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
)

const spChannelID = 1

var (
	_ service.Service = (*SpReactor)(nil)
	_ p2p.Wrapper     = (*sp.Message)(nil)
)

// SpReactor implements a service to communicate with storage providers.
type SpReactor struct {
	service.BaseService
	logger tmlog.Logger

	peerManager *p2p.PeerManager

	peerEvents p2p.PeerEventSubscriber
	chCreator  p2p.ChannelCreator

	providerQuerier ProviderQuerier
	providerUpdater ProviderUpdater
}

// NewSpReactor returns a reference to a new reactor.
func NewSpReactor(
	logger tmlog.Logger,
	peerManager *p2p.PeerManager,
	chCreator p2p.ChannelCreator,
	peerEvents p2p.PeerEventSubscriber,
	providerVerifier ProviderQuerier,
	providerUpdater ProviderUpdater,
) *SpReactor {
	r := &SpReactor{
		logger:          logger,
		peerManager:     peerManager,
		chCreator:       chCreator,
		peerEvents:      peerEvents,
		providerQuerier: providerVerifier,
		providerUpdater: providerUpdater,
	}

	r.BaseService = *service.NewBaseService(logger, "SecondSp", r)
	return r
}

// getChannelDescriptor produces an instance of a descriptor for this
// package's required channels.
func getChannelDescriptor() *p2p.ChannelDescriptor {
	return &p2p.ChannelDescriptor{
		ID:                  spChannelID,
		MessageType:         new(sp.Message),
		Priority:            5,
		RecvMessageCapacity: 1024,
		RecvBufferCapacity:  128,
		Name:                "secondarysp",
	}
}

// OnStart starts separate go routines for each p2p Channel and listens for
// envelopes on each. In addition, it also listens for peer updates and handles
// messages on that p2p channel accordingly. The caller must be sure to execute
// OnStop to ensure the outbound p2p Channels are closed.
func (r *SpReactor) OnStart(ctx context.Context) error {

	ch, err := r.chCreator(ctx, getChannelDescriptor())
	if err != nil {
		return err
	}

	go r.processChannel(ctx, ch)
	go r.processPeerUpdates(ctx, r.peerEvents(ctx), ch)

	go r.processStorageRequests(ctx, r.providerUpdater.SubscribeStorageRequest(), ch)
	go r.processSpJoinLefts(ctx, r.providerUpdater.SubscribeSpJoinLeft(), ch)

	return nil
}

// OnStop stops the reactor by signaling to all spawned goroutines to exit and
// blocking until they all exit.
func (r *SpReactor) OnStop() {}

// handleSspMessage handles envelopes sent from peers on the channel.
func (r *SpReactor) handleSspMessage(ctx context.Context, envelope *p2p.Envelope, channel p2p.Channel) error {
	logger := r.logger.With("peer", envelope.From)

	//TODO: handle messages
	switch msg := envelope.Message.(type) {
	case *sp.SecondSpRequest:
		r.handleSecondSpRequest(ctx, msg, channel)
	case *sp.SecondSpRefuse:
		r.handleSecondSpRefuse(ctx, msg)
	case *sp.SecondSpAck:
		r.handleSecondSpAck(ctx, msg)
	case *sp.SecondSpManifest:
		r.handleSecondManifest(ctx, msg, channel)
	default:
		logger.Error("received unknown message")
		return fmt.Errorf("received unknown message: %T", msg)
	}

	return nil
}

func (r *SpReactor) handleSecondSpRequest(ctx context.Context, msg *sp.SecondSpRequest, channel p2p.Channel) {
	fmt.Println("SecondSpRequest msg:", msg)

	if msg.ObjectId%2 == 0 {
		if err := channel.Send(ctx, p2p.Envelope{
			Broadcast: true,
			Message: &sp.SecondSpAck{
				ObjectId: msg.ObjectId,
			},
		}); err != nil {
			return
		}
	} else {
		if err := channel.Send(ctx, p2p.Envelope{
			Broadcast: true,
			Message: &sp.SecondSpRefuse{
				ObjectId: msg.ObjectId,
			},
		}); err != nil {
			return
		}
	}
}

func (r *SpReactor) handleSecondSpRefuse(ctx context.Context, msg *sp.SecondSpRefuse) {
	fmt.Println("SecondSpRefuse msg:", msg)
}

func (r *SpReactor) handleSecondSpAck(ctx context.Context, msg *sp.SecondSpAck) {
	fmt.Println("SecondSpAck msg:", msg)
}

func (r *SpReactor) handleSecondManifest(ctx context.Context, msg *sp.SecondSpManifest, channel p2p.Channel) {
	fmt.Println("SecondSpManifest msg:", msg)
}

// handleMessage handles an Envelope sent from a peer on a specific p2p Channel.
// It will handle errors and any possible panics gracefully. A caller can handle
// any error returned by sending a PeerError on the respective channel.
func (r *SpReactor) handleMessage(ctx context.Context, envelope *p2p.Envelope, channel p2p.Channel) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic in processing message: %v", e)
			r.logger.Error(
				"recovering from processing message panic",
				"err", err,
				"stack", string(debug.Stack()),
			)
		}
	}()

	r.logger.Debug("received message", "peer", envelope.From)

	switch envelope.ChannelID {
	case spChannelID:
		err = r.handleSspMessage(ctx, envelope, channel)
	default:
		err = fmt.Errorf("unknown channel ID (%d) for envelope (%T)", envelope.ChannelID, envelope.Message)
	}

	return
}

// processChannel implements a blocking event loop where we listen for p2p
// Envelope messages from the channel.
func (r *SpReactor) processChannel(ctx context.Context, channel p2p.Channel) {
	iter := channel.Receive(ctx)
	for iter.Next(ctx) {
		envelope := iter.Envelope()
		if err := r.handleMessage(ctx, envelope, channel); err != nil {
			r.logger.Error("failed to process message", "ch_id", envelope.ChannelID, "envelope", envelope, "err", err)
			if serr := channel.SendError(ctx, p2p.PeerError{
				NodeID: envelope.From,
				Err:    err,
			}); serr != nil {
				return
			}
		}
	}
}

// processPeerUpdate processes a PeerUpdate.
func (r *SpReactor) processPeerUpdate(ctx context.Context, peerUpdate p2p.PeerUpdate, channel p2p.Channel) {
	r.logger.Debug("received peer update", "peer", peerUpdate.NodeID, "status", peerUpdate.Status)
	if peerUpdate.Status == p2p.PeerStatusUp {
		if inWhitelist := r.providerQuerier.Check(peerUpdate.NodeID); !inWhitelist {
			r.logger.Error("not found in persist peers or local db", "nodeId", peerUpdate.NodeID)
			channel.SendError(ctx, p2p.PeerError{
				NodeID: peerUpdate.NodeID,
				Err:    errors.New("not allowed to connect to"),
				Fatal:  false,
			})
		}
	}
}

// processPeerUpdates initiates a blocking process where we listen for and handle
// PeerUpdate messages. When the reactor is stopped, we will catch the signal and
// close the p2p PeerUpdatesCh gracefully.
func (r *SpReactor) processPeerUpdates(ctx context.Context, peerUpdates *p2p.PeerUpdates, channel p2p.Channel) {
	for {
		select {
		case <-ctx.Done():
			return
		case peerUpdate := <-peerUpdates.Updates():
			r.processPeerUpdate(ctx, peerUpdate, channel)
		}
	}
}

func (r *SpReactor) processStorageRequests(ctx context.Context, c <-chan StorageRequest, channel p2p.Channel) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-c:
			request := sp.SecondSpRequest{
				ObjectId: event.ObjectId,
				Size_:    event.ObjectSize,
				SpIndex:  0,
			}
			r.broadcastSpRequest(ctx, &request, channel)
		}
	}
}

func (r *SpReactor) processSpJoinLefts(ctx context.Context, c <-chan SpJoinLeft, channel p2p.Channel) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-c:
			fmt.Println(event)
		}
	}
}

// broadcastSpRequest broadcasts secondary sp request to peers.
func (r *SpReactor) broadcastSpRequest(ctx context.Context, request *sp.SecondSpRequest, channel p2p.Channel) {
	if err := channel.Send(ctx, p2p.Envelope{
		Broadcast: true,
		Message:   request,
	}); err != nil {
		return
	}

	r.logger.Debug("broadcast requests to peers")
}

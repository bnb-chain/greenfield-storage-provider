package p2p

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/modular/p2p/p2pnode"
)

const (
	P2PModularName        = "p2p"
	P2PModularDescription = "p2p modular supports communicates between SPs on p2p protocol"
)

var _ module.P2P = &P2PModular{}

type P2PModular struct {
	baseApp                *gfspapp.GfSpBaseApp
	node                   *p2pnode.Node
	scope                  rcmgr.ResourceScope
	replicateApprovalQueue taskqueue.TQueueOnStrategy
}

func (p *P2PModular) Name() string {
	return P2PModularName
}

func (p *P2PModular) Start(ctx context.Context) error {
	scope, err := p.baseApp.ResourceManager().OpenService(p.Name())
	if err != nil {
		return err
	}
	p.scope = scope
	return nil
}

func (p *P2PModular) Stop(ctx context.Context) error {
	p.scope.Release()
	return nil
}

func (p *P2PModular) ReserveResource(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
	span, err := p.scope.BeginSpan()
	if err != nil {
		return nil, err
	}
	err = span.ReserveResources(state)
	if err != nil {
		return nil, err
	}
	return span, nil
}

func (p *P2PModular) ReleaseResource(ctx context.Context, span rcmgr.ResourceScopeSpan) {
	span.Done()
	return
}

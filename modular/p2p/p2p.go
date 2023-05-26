package p2p

import (
	"bytes"
	"context"
	"sort"
	"time"

	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/modular/p2p/p2pnode"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const UpdateSPDuration = 2

var _ module.P2P = &P2PModular{}

type P2PModular struct {
	baseApp                *gfspapp.GfSpBaseApp
	node                   *p2pnode.Node
	scope                  rcmgr.ResourceScope
	replicateApprovalQueue taskqueue.TQueueOnStrategy
}

func (p *P2PModular) Name() string {
	return module.P2PModularName
}

func (p *P2PModular) Start(ctx context.Context) error {
	scope, err := p.baseApp.ResourceManager().OpenService(p.Name())
	if err != nil {
		return err
	}
	p.scope = scope
	if err = p.node.Start(ctx); err != nil {
		return err
	}
	go p.eventLoop(ctx)
	return nil
}

func (p *P2PModular) Stop(ctx context.Context) error {
	p.node.Stop(ctx)
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
}

func (p *P2PModular) eventLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(UpdateSPDuration) * time.Second)
	var integrity []byte
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			spList, err := p.baseApp.GfSpDB().FetchAllSp()
			if err != nil {
				log.CtxWarnw(ctx, "failed to fetch all SPs", "error", err)
				continue
			}
			var spOpAddrs []string
			for _, sp := range spList {
				spOpAddrs = append(spOpAddrs, sp.GetOperatorAddress())
			}
			sort.Strings(spOpAddrs)
			var spOpByte [][]byte
			for _, addr := range spOpAddrs {
				spOpByte = append(spOpByte, []byte(addr))
			}
			currentIntegrity := hash.GenerateIntegrityHash(spOpByte)
			if bytes.Equal(currentIntegrity, integrity) {
				continue
			}
			integrity = currentIntegrity[:]
			p.node.PeersProvider().UpdateSp(spOpAddrs)
		}
	}
}

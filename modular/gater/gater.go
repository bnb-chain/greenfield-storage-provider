package gater

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
)

const (
	GateModularName        = "gateway"
	GateModularDescription = "approval modular supports receives the user requests"
)

var _ module.Modular = &GateModular{}

type GateModular struct {
	domain      string
	httpAddress string
	baseApp     *gfspapp.GfSpBaseApp
	scope       rcmgr.ResourceScope

	maxListReadQuota int64
}

func (g *GateModular) Name() string {
	return GateModularName
}

func (g *GateModular) Start(ctx context.Context) error {
	scope, err := g.baseApp.ResourceManager().OpenService(g.Name())
	if err != nil {
		return err
	}
	g.scope = scope
	return nil
}

func (g *GateModular) Stop(ctx context.Context) error {
	g.scope.Release()
	return nil
}

func (g *GateModular) ReserveResource(
	ctx context.Context,
	state *rcmgr.ScopeStat) (
	rcmgr.ResourceScopeSpan, error) {
	span, err := g.scope.BeginSpan()
	if err != nil {
		return nil, err
	}
	err = span.ReserveResources(state)
	if err != nil {
		return nil, err
	}
	return span, nil
}

func (g *GateModular) ReleaseResource(
	ctx context.Context,
	span rcmgr.ResourceScopeSpan) {
	span.Done()
	return
}

package p2p

import (
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/modular/p2p/p2pnode"
)

const (
	// P2PPrivateKey defines env variable for p2p protocol private key
	P2PPrivateKey                              = "P2P_PRIVATE_KEY"
	DefaultP2PProtocolAddress                  = "localhost:9933"
	DefaultAskReplicateApprovalParallelPerNode = 1024
)

func init() {
	gfspapp.RegisterModularInfo(P2PModularName, P2PModularDescription, NewP2PModular)
}

func NewP2PModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	if cfg.Customize.P2P != nil {
		app.SetP2P(cfg.Customize.P2P)
		return cfg.Customize.P2P, nil
	}
	p2p := &P2PModular{baseApp: app}
	if err := DefaultP2POptions(p2p, cfg); err != nil {
		return nil, err
	}
	app.SetP2P(p2p)
	return p2p, nil
}

func DefaultP2POptions(p2p *P2PModular, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Parallel.AskReplicateApprovalParallelPerNode == 0 {
		cfg.Parallel.AskReplicateApprovalParallelPerNode = DefaultAskReplicateApprovalParallelPerNode
	}
	p2p.replicateApprovalQueue = cfg.Customize.NewStrategyTQueueFunc(
		p2p.Name()+"-ask-replicate-piece", cfg.Parallel.AskReplicateApprovalParallelPerNode)
	if val, ok := os.LookupEnv(P2PPrivateKey); ok {
		cfg.P2P.P2PPrivateKey = val
	}
	if cfg.P2P.P2PAddress == "" {
		cfg.P2P.P2PAddress = DefaultP2PProtocolAddress
	}
	node, err := p2pnode.NewNode(p2p.baseApp, cfg.P2P.P2PPrivateKey, cfg.P2P.P2PAddress,
		cfg.P2P.P2PBootstrap, cfg.P2P.P2PPingPeriod, cfg.Approval.ReplicatePieceTimeoutHeight)
	if err != nil {
		return err
	}
	p2p.node = node
	return nil
}

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

func NewP2PModular(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig,
	opts ...gfspapp.Option) (
	coremodule.Modular, error) {
	if cfg.P2P != nil {
		app.SetP2P(cfg.P2P)
		return cfg.P2P, nil
	}
	p2p := &P2PModular{baseApp: app}
	opts = append(opts, p2p.DefaultP2POptions)
	for _, opt := range opts {
		if err := opt(app, cfg); err != nil {
			return nil, err
		}
	}
	app.SetP2P(p2p)
	return p2p, nil
}

func (p *P2PModular) DefaultP2POptions(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig) error {
	if cfg.AskReplicateApprovalParallelPerNode == 0 {
		cfg.AskReplicateApprovalParallelPerNode = DefaultAskReplicateApprovalParallelPerNode
	}
	p.replicateApprovalQueue = cfg.NewStrategyTQueueFunc(p.Name()+"-ask-replicate-piece",
		cfg.AskReplicateApprovalParallelPerNode)
	if val, ok := os.LookupEnv(P2PPrivateKey); ok {
		cfg.P2PPrivateKey = val
	}
	if cfg.P2PAddress == "" {
		cfg.P2PAddress = DefaultP2PProtocolAddress
	}
	node, err := p2pnode.NewNode(app, cfg.P2PPrivateKey, cfg.P2PAddress, cfg.P2PBootstrap,
		cfg.P2PPingPeriod, cfg.ReplicatePieceTimeoutHeight)
	if err != nil {
		return err
	}
	p.node = node
	return nil
}

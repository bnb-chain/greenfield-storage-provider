package p2p

import (
	"os"

	"github.com/zkMeLabs/mechain-storage-provider/base/gfspapp"
	"github.com/zkMeLabs/mechain-storage-provider/base/gfspconfig"
	coremodule "github.com/zkMeLabs/mechain-storage-provider/core/module"
	"github.com/zkMeLabs/mechain-storage-provider/modular/p2p/p2pnode"
)

const (
	// P2PPrivateKey defines env variable for p2p protocol private key
	P2PPrivateKey = "P2P_PRIVATE_KEY"
	// DefaultP2PProtocolAddress defines the default p2p protocol address
	DefaultP2PProtocolAddress = "localhost:9933"
	// DefaultAskReplicateApprovalParallelPerNode defines the default max ask replicate
	// piece approval parallel per p2p node
	DefaultAskReplicateApprovalParallelPerNode = 10240
)

func NewP2PModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	p2p := &P2PModular{baseApp: app}
	if err := DefaultP2POptions(p2p, cfg); err != nil {
		return nil, err
	}
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
	node, err := p2pnode.NewNode(p2p.baseApp, cfg.P2P.P2PPrivateKey,
		cfg.P2P.P2PAddress, cfg.P2P.P2PBootstrap, cfg.P2P.P2PPingPeriod,
		cfg.Approval.ReplicatePieceTimeoutHeight, cfg.P2P.P2PAntAddress)
	if err != nil {
		return err
	}
	p2p.node = node
	return nil
}

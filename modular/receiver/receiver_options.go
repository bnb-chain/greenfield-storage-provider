package receiver

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	// DefaultReceivePieceParallelPerNode defines the default max receive piece parallel
	// per receiver
	DefaultReceivePieceParallelPerNode = 1024
)

func NewReceiveModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	receiver := &ReceiveModular{baseApp: app}
	if err := DefaultReceiverOptions(receiver, cfg); err != nil {
		return nil, err
	}
	return receiver, nil
}

func DefaultReceiverOptions(receiver *ReceiveModular, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Parallel.ReceivePieceParallelPerNode == 0 {
		cfg.Parallel.ReceivePieceParallelPerNode = DefaultReceivePieceParallelPerNode
	}
	receiver.receiveQueue = cfg.Customize.NewStrategyTQueueFunc(
		receiver.Name()+"-receive-piece", cfg.Parallel.ReceivePieceParallelPerNode)
	return nil
}

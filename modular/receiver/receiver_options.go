package receiver

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	DefaultReceivePieceParallelPerNode = 1024
)

func init() {
	gfspapp.RegisterModularInfo(ReceiveModularName, ReceiveModularDescription, NewReceiveModular)
}

func NewReceiveModular(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig,
	opts ...gfspapp.Option) (
	coremodule.Modular, error) {
	if cfg.Receiver != nil {
		app.SetReceiver(cfg.Receiver)
		return cfg.Receiver, nil
	}
	receiver := &ReceiveModular{baseApp: app}
	opts = append(opts, receiver.DefaultReceiverOptions)
	for _, opt := range opts {
		if err := opt(app, cfg); err != nil {
			return nil, err
		}
	}
	app.SetReceiver(receiver)
	return receiver, nil
}

func (r *ReceiveModular) DefaultReceiverOptions(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig) error {
	if cfg.ReceivePieceParallelPerNode == 0 {
		cfg.ReceivePieceParallelPerNode = DefaultReceivePieceParallelPerNode
	}
	r.receiveQueue = cfg.NewStrategyTQueueFunc(r.Name()+"-receive-piece",
		cfg.ReceivePieceParallelPerNode)
	return nil
}

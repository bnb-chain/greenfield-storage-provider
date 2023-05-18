package gfspconfig

import (
	"errors"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"

	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretaskqueue "github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
)

func CustomizeGfSpDB(db spdb.SPDB) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.GfSpDB != nil {
			return errors.New("repeated set sp db")
		}
		cfg.Customize.GfSpDB = db
		return nil
	}
}

func CustomizeGfBsDB(db bsdb.BSDB) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.GfBsDB != nil {
			return errors.New("repeated set bs db")
		}
		cfg.Customize.GfBsDB = db
		return nil
	}
}

func CustomizePieceStore(store piecestore.PieceStore) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.PieceStore != nil {
			return errors.New("repeated set piece store")
		}
		cfg.Customize.PieceStore = store
		return nil
	}
}

func CustomizePieceOp(op piecestore.PieceOp) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.PieceOp != nil {
			return errors.New("repeated set piece op")
		}
		cfg.Customize.PieceOp = op
		return nil
	}
}

func CustomizeRcmgr(rcmgr corercmgr.ResourceManager) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.Rcmgr != nil {
			return errors.New("repeated set rcmgr")
		}
		cfg.Customize.Rcmgr = rcmgr
		return nil
	}
}

func CustomizeRcLimiter(limiter corercmgr.Limiter) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.RcLimiter != nil {
			return errors.New("repeated set rc limiter")
		}
		cfg.Customize.RcLimiter = limiter
		return nil
	}
}

func CustomizeConsensus(consensus consensus.Consensus) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.Consensus != nil {
			return errors.New("repeated set consensus")
		}
		cfg.Customize.Consensus = consensus
		return nil
	}
}

func CustomizeTQueue(newFunc coretaskqueue.NewTQueue) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.NewTQueueFunc != nil {
			return errors.New("repeated set task queue")
		}
		cfg.Customize.NewTQueueFunc = newFunc
		return nil
	}
}

func CustomizeTQueueWithLimit(newFunc coretaskqueue.NewTQueueWithLimit) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.NewTQueueWithLimit != nil {
			return errors.New("repeated set strategy task queue with limit")
		}
		cfg.Customize.NewTQueueWithLimit = newFunc
		return nil
	}
}

func CustomizeStrategyTQueue(newFunc coretaskqueue.NewTQueueOnStrategy) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.NewStrategyTQueueFunc != nil {
			return errors.New("repeated set strategy task queue")
		}
		cfg.Customize.NewStrategyTQueueFunc = newFunc
		return nil
	}
}

func CustomizeStrategyTQueueWithLimit(newFunc coretaskqueue.NewTQueueOnStrategyWithLimit) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.NewStrategyTQueueFunc != nil {
			return errors.New("repeated set strategy task queue with limit")
		}
		cfg.Customize.NewStrategyTQueueWithLimitFunc = newFunc
		return nil
	}
}

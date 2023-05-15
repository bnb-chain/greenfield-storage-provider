package gfspconfig

import (
	"errors"

	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
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

func CustomizeMetrics(metrics module.Modular) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.Metrics != nil {
			return errors.New("repeated set metrics")
		}
		cfg.Customize.Metrics = metrics
		return nil
	}
}

func CustomizePProf(pprof module.Modular) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.PProf != nil {
			return errors.New("repeated set pprof")
		}
		cfg.Customize.PProf = pprof
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

func CustomizeApprover(approver module.Approver) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.Approver != nil {
			return errors.New("repeated set approver")
		}
		cfg.Customize.Approver = approver
		return nil
	}
}

func CustomizeAuthorizer(authorizer module.Authorizer) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.Authorizer != nil {
			return errors.New("repeated set authorizer")
		}
		cfg.Customize.Authorizer = authorizer
		return nil
	}
}

func CustomizeDownloader(downloader module.Downloader) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.Downloader != nil {
			return errors.New("repeated set downloader")
		}
		cfg.Customize.Downloader = downloader
		return nil
	}
}

func CustomizeTaskExecutor(executor module.TaskExecutor) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.TaskExecutor != nil {
			return errors.New("repeated set executor")
		}
		cfg.Customize.TaskExecutor = executor
		return nil
	}
}

func CustomizeGater(gater module.Modular) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.Gater != nil {
			return errors.New("repeated set gater")
		}
		cfg.Customize.Gater = gater
		return nil
	}
}

func CustomizeManager(manager module.Manager) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.Manager != nil {
			return errors.New("repeated set manager")
		}
		cfg.Customize.Manager = manager
		return nil
	}
}

func CustomizeP2P(p2p module.P2P) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.P2P != nil {
			return errors.New("repeated set P2P")
		}
		cfg.Customize.P2P = p2p
		return nil
	}
}

func CustomizeReceiver(receiver module.Receiver) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.Receiver != nil {
			return errors.New("repeated set receiver")
		}
		cfg.Customize.Receiver = receiver
		return nil
	}
}

func CustomizeRetriever(retriever module.Modular) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.Retriever != nil {
			return errors.New("repeated set retriever")
		}
		cfg.Customize.Retriever = retriever
		return nil
	}
}

func CustomizeSigner(signer module.Signer) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.Signer != nil {
			return errors.New("repeated set signer")
		}
		cfg.Customize.Signer = signer
		return nil
	}
}

func CustomizeUploader(uploader module.Uploader) Option {
	return func(cfg *GfSpConfig) error {
		if cfg.Customize == nil {
			cfg.Customize = &Customize{}
		}
		if cfg.Customize.Uploader != nil {
			return errors.New("repeated set uploader")
		}
		cfg.Customize.Uploader = uploader
		return nil
	}
}

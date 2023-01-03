package gateway

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"

	"github.com/bnb-chain/inscription-storage-provider/util/log"
	"github.com/gorilla/mux"
)

// Gateway is the primary entry point of SP.
type Gateway struct {
	config  *GatewayConfig
	name    string
	running atomic.Bool

	httpServer *http.Server
	uploader   *uploaderClient
	downloader *downloaderClient
	chain      *chainClient
	retriever  *retrieverClient
}

// NewGatewayService return the gateway instance
func NewGatewayService(cfg *GatewayConfig) (*Gateway, error) {
	var (
		err error
		g   *Gateway
	)

	g = &Gateway{
		config: cfg,
		name:   "Gateway",
	}
	if g.uploader, err = newUploaderClient(g.config.UploaderConfig); err != nil {
		log.Warnw("failed to create uploader", "err", err)
		return nil, err
	}
	if g.downloader, err = newDownloaderClient(g.config.DownloaderConfig); err != nil {
		log.Warnw("failed to create downloader", "err", err)
		return nil, err
	}
	if g.chain, err = newChainClient(g.config.ChainConfig); err != nil {
		log.Warnw("failed to create chainer client", "err", err)
		return nil, err
	}
	g.retriever = newRetrieverClient()

	level := log.DebugLevel
	switch g.config.LogConfig.Level {
	case "debug":
		level = log.DebugLevel
	case "info":
		level = log.InfoLevel
	case "warn":
		level = log.WarnLevel
	case "error":
		level = log.ErrorLevel
	default:
		level = log.InfoLevel
	}
	log.Init(level, g.config.LogConfig.FilePath)
	log.Infow("gateway succeed to init")
	return g, nil
}

// Name implement the lifecycle interface
func (g *Gateway) Name() string {
	return g.name
}

// Start implement the lifecycle interface
func (g *Gateway) Start(ctx context.Context) error {
	if g.running.Swap(true) {
		return errors.New("gateway has started")
	}
	go g.Serve()
	log.Info("gateway succeed to start")
	return nil
}

// Serve starts http service.
func (g *Gateway) Serve() {
	router := mux.NewRouter().SkipClean(true)
	g.registerhandler(router)
	server := &http.Server{
		Addr:    g.config.Address,
		Handler: router,
	}
	g.httpServer = server
	if err := server.ListenAndServe(); err != nil {
		log.Warnw("failed to listen", "err", err)
		return
	}
}

// Stop implement the lifecycle interface
func (g *Gateway) Stop(ctx context.Context) error {
	if !g.running.Swap(false) {
		return errors.New("gateway has stopped")
	}
	_ = g.httpServer.Shutdown(ctx)
	log.Info("gateway succeed to stop")
	return nil
}

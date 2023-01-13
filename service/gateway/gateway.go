package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/mux"

	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

const (
	ServiceNameGateway string = "GatewayService"
)

// Gateway is the primary entry point of SP.
type Gateway struct {
	config  *GatewayConfig
	name    string
	running atomic.Bool

	httpServer        *http.Server
	uploadProcessor   *uploadProcessor
	downloadProcessor *downloadProcessor
	chain             *chainClient
	retriever         *retrieverClient
}

// NewGatewayService return the gateway instance
func NewGatewayService(cfg *GatewayConfig) (*Gateway, error) {
	var (
		err error
		g   *Gateway
	)

	g = &Gateway{
		config: cfg,
		name:   ServiceNameGateway,
	}
	if g.uploadProcessor, err = newUploadProcessor(g.config.UploaderConfig); err != nil {
		log.Warnw("failed to create uploader", "err", err)
		return nil, err
	}
	if g.downloadProcessor, err = newDownloadProcessor(g.config.DownloaderConfig); err != nil {
		log.Warnw("failed to create downloader", "err", err)
		return nil, err
	}
	if g.chain, err = newChainClient(g.config.ChainConfig); err != nil {
		log.Warnw("failed to create chain client", "err", err)
		return nil, err
	}
	g.retriever = newRetrieverClient()
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
	var errs []error
	if err := g.httpServer.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}
	if err := g.uploadProcessor.Close(); err != nil {
		errs = append(errs, err)
	}
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	log.Info("gateway succeed to stop")
	return nil
}

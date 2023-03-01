package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"

	dclient "github.com/bnb-chain/greenfield-storage-provider/service/downloader/client"
	sclient "github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
	uclient "github.com/bnb-chain/greenfield-storage-provider/service/uploader/client"

	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// Gateway is the primary entry point of SP.
type Gateway struct {
	config  *GatewayConfig
	name    string
	running atomic.Bool

	httpServer *http.Server
	uploader   *uclient.UploaderClient
	downloader *dclient.DownloaderClient
	//challenge  *client.ChallengeClient

	//syncer client.SyncerAPI
	chain  *gnfd.Greenfield
	signer *sclient.SignerClient
}

// NewGatewayService return the gateway instance
func NewGatewayService(cfg *GatewayConfig) (*Gateway, error) {
	var (
		err error
		g   *Gateway
	)

	g = &Gateway{
		config: cfg,
		name:   model.GatewayService,
	}
	if g.uploader, err = uclient.NewUploaderClient(cfg.UploaderServiceAddress); err != nil {
		log.Errorw("failed to uploader client", "err", err)
		return nil, err
	}
	if g.downloader, err = dclient.NewDownloaderClient(cfg.DownloaderServiceAddress); err != nil {
		log.Errorw("failed to downloader client", "err", err)
		return nil, err
	}
	//if g.challenge, err = client.NewChallengeClient(cfg.ChallengeServiceAddress); err != nil {
	//	log.Errorw("failed to challenge client", "err", err)
	//	return nil, err
	//}
	//if g.syncer, err = client.NewSyncerClient(g.config.SyncerServiceAddress); err != nil {
	//	log.Errorw("gateway inits syncer client failed", "error", err)
	//	return nil, err
	//}
	if g.chain, err = gnfd.NewGreenfield(cfg.ChainConfig); err != nil {
		log.Errorw("failed to create chain client", "err", err)
		return nil, err
	}
	if g.signer, err = sclient.NewSignerClient(cfg.SignerServiceAddress); err != nil {
		log.Errorw("failed to create signer client", "err", err)
		return nil, err
	}
	log.Debugw("gateway succeed to init")
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
	log.Debug("gateway succeed to start")
	return nil
}

// Serve starts http service.
func (g *Gateway) Serve() {
	router := mux.NewRouter().SkipClean(true)
	g.registerHandler(router)
	server := &http.Server{
		Addr:    g.config.Address,
		Handler: router,
	}
	g.httpServer = server
	if err := server.ListenAndServe(); err != nil {
		log.Errorw("failed to listen", "err", err)
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
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	log.Debug("gateway succeed to stop")
	return nil
}

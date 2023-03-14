package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"

	chainclient "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	challengeclient "github.com/bnb-chain/greenfield-storage-provider/service/challenge/client"
	downloaderclient "github.com/bnb-chain/greenfield-storage-provider/service/downloader/client"
	signerclient "github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
	syncerclient "github.com/bnb-chain/greenfield-storage-provider/service/syncer/client"
	uploaderclient "github.com/bnb-chain/greenfield-storage-provider/service/uploader/client"
)

// Gateway is the primary entry point of SP
type Gateway struct {
	config     *GatewayConfig
	running    atomic.Value
	httpServer *http.Server

	// chain is the required component, used to check authorization
	chain *chainclient.Greenfield

	// the below components are optional according to the config
	uploader   *uploaderclient.UploaderClient
	downloader *downloaderclient.DownloaderClient
	challenge  *challengeclient.ChallengeClient
	syncer     *syncerclient.SyncerClient
	signer     *signerclient.SignerClient
	metadata   *client.MetadataClient
}

// NewGatewayService return the gateway instance
func NewGatewayService(cfg *GatewayConfig) (*Gateway, error) {
	var (
		err error
		g   *Gateway
	)

	g = &Gateway{
		config: cfg,
	}
	if g.chain, err = chainclient.NewGreenfield(cfg.ChainConfig); err != nil {
		log.Errorw("failed to create chain client", "error", err)
		return nil, err
	}

	if cfg.UploaderServiceAddress != "" {
		if g.uploader, err = uploaderclient.NewUploaderClient(cfg.UploaderServiceAddress); err != nil {
			log.Errorw("failed to create uploader client", "error", err)
			return nil, err
		}
	}

	if cfg.DownloaderServiceAddress != "" {
		if g.downloader, err = downloaderclient.NewDownloaderClient(cfg.DownloaderServiceAddress); err != nil {
			log.Errorw("failed to create downloader client", "error", err)
			return nil, err
		}
	}
	if cfg.ChallengeServiceAddress != "" {
		if g.challenge, err = challengeclient.NewChallengeClient(cfg.ChallengeServiceAddress); err != nil {
			log.Errorw("failed to create challenge client", "error", err)
			return nil, err
		}
	}

	if cfg.SyncerServiceAddress != "" {
		if g.syncer, err = syncerclient.NewSyncerClient(cfg.SyncerServiceAddress); err != nil {
			log.Errorw("failed to create syncer client", "error", err)
			return nil, err
		}
	}

	if cfg.SignerServiceAddress != "" {
		if g.signer, err = signerclient.NewSignerClient(cfg.SignerServiceAddress); err != nil {
			log.Errorw("failed to create signer client", "error", err)
			return nil, err
		}
	}

	if cfg.MetadataServiceAddress != "" {
		if g.metadata, err = client.NewMetadataClient(cfg.MetadataServiceAddress); err != nil {
			log.Errorw("failed to create metadata client", "error", err)
			return nil, err
		}
	}

	log.Debugw("gateway succeed to init")
	return g, nil
}

// Name implement the lifecycle interface
func (g *Gateway) Name() string {
	return model.GatewayService
}

// Start implement the lifecycle interface
func (g *Gateway) Start(ctx context.Context) error {
	if g.running.Swap(true) == true {
		return errors.New("gateway has started")
	}
	go g.serve()
	return nil
}

// Serve starts http service.
func (g *Gateway) serve() {
	router := mux.NewRouter().SkipClean(true)
	g.registerHandler(router)
	server := &http.Server{
		Addr:    g.config.HTTPAddress,
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
	if g.running.Swap(false) == false {
		return errors.New("gateway has stopped")
	}
	var errs []error
	if err := g.httpServer.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

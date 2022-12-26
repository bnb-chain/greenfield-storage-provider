package gateway

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/naoina/toml"

	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

type gatewayConfig struct {
	Port      string
	Domain    string
	LogConfig struct {
		FilePath string
		Level    string
	}
	UploaderConfig   uploaderClientConfig
	ChainConfig      chainClientConfig
	DownloaderConfig downloaderClientConfig
}

type GatewayService struct {
	config     gatewayConfig
	httpServer *http.Server
	uploader   *uploaderClient
	downloader *downloaderClient
	chain      *chainClient
	retriever  *retrieverClient
}

func NewGatewayService() *GatewayService {
	return &GatewayService{}
}

func (g *GatewayService) initConfig(configFile string) bool {
	f, err := os.Open(configFile)
	if err != nil {
		log.Warnw("failed to open config file", "err", err)
		return false
	}
	defer f.Close()
	if err := toml.NewDecoder(f).Decode(&g.config); err != nil {
		log.Warnw("failed to parse config file", "err", err)
		return false
	}
	log.Infow("succeed to init config", "config", g.config)
	return true
}

func (g *GatewayService) Init(configFile string) bool {
	if ok := g.initConfig(configFile); !ok {
		return false
	}

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
	log.Infow("succeed to init")
	return true
}

func (g *GatewayService) Start() bool {
	var (
		err error
	)

	router := mux.NewRouter().SkipClean(true)
	g.registerhandler(router)
	var server = &http.Server{
		Addr:    ":" + g.config.Port,
		Handler: router,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Warnw("failed to listen", "err", err)
			return
		}
	}()
	g.httpServer = server
	if g.uploader, err = newUploaderClient(g.config.UploaderConfig); err != nil {
		log.Warnw("failed to create uploader", "err", err)
		return false
	}
	if g.downloader, err = newDownloaderClient(g.config.DownloaderConfig); err != nil {
		log.Warnw("failed to create downloader", "err", err)
		return false
	}
	if g.chain, err = newChainClient(g.config.ChainConfig); err != nil {
		log.Warnw("failed to create chainer client", "err", err)
		return false
	}
	g.retriever = newRetrieverClient()
	log.Info("gateway startup")
	return true
}

// todo: feat
func (g *GatewayService) Join() bool {
	for {
		time.Sleep(1 * time.Second)
	}
	return true
}

func (g *GatewayService) Stop() bool {
	return true
}

func (g *GatewayService) Description() string {
	return "GatewayService"
}

package gateway

import (
	"github.com/bnb-chain/inscription-storage-provider/util/log"
	"github.com/gorilla/mux"
	"github.com/naoina/toml"
	"net/http"
	"os"
	"time"
)

type gatewayConfig struct {
	Port      string
	Domain    string
	DebugDir  string
	LogConfig struct {
		FilePath string
		Level    string
	}
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

	if g.config.DebugDir != "" {
		if err := os.Mkdir(g.config.DebugDir, 0777); err != nil && !os.IsExist(err) {
			log.Warnw("failed to make debug dir", "err", err)
			return false
		}
	}
	log.Infow("succeed to init")
	return true
}

func (g *GatewayService) Start() bool {
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
	g.uploader = newUploaderClient()
	g.downloader = newDownloaderClient()
	g.chain = newChainClient()
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

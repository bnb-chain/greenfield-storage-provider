package uploader

import (
	"net"
	"os"
	"time"

	"github.com/naoina/toml"
	"google.golang.org/grpc"

	pbService "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

type uploaderConfig struct {
	Port            string
	StoreConfigFile string
	LogConfig       struct {
		FilePath string
		Level    string
	}
}

type UploaderService struct {
	config      uploaderConfig
	grpcServer  *grpc.Server
	signer      *signerClient
	stoneHub    *stoneHubClient
	eventWaiter *eventClient
	store       *storeClient
}

func NewUploaderService() *UploaderService {
	return &UploaderService{}
}

func (u *UploaderService) initConfig(configFile string) bool {
	f, err := os.Open(configFile)
	if err != nil {
		log.Warnw("failed to open config file", "err", err)
		return false
	}
	defer f.Close()
	if err := toml.NewDecoder(f).Decode(&u.config); err != nil {
		log.Warnw("failed to parse config file", "err", err)
		return false
	}
	log.Infow("succeed to init config", "config", u.config)
	return true
}

func (u *UploaderService) Init(configFile string) bool {
	if ok := u.initConfig(configFile); !ok {
		return false
	}

	level := log.DebugLevel
	switch u.config.LogConfig.Level {
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
	log.Init(level, u.config.LogConfig.FilePath)
	log.Info("succeed to init")
	return true
}

func (u *UploaderService) Start() bool {
	lis, err := net.Listen("tcp", ":"+u.config.Port)
	if err != nil {
		log.Warnw("failed to listen", "err", err)
		return false
	}
	u.grpcServer = grpc.NewServer()
	pbService.RegisterUploaderServiceServer(u.grpcServer, &uploaderImpl{uploader: u})
	go func() {
		if err := u.grpcServer.Serve(lis); err != nil {
			log.Warnw("failed to start grpc server", "err", err)
		}
	}()
	u.signer = newSignerClient()
	u.stoneHub = newStoneHubClient()
	u.eventWaiter = newEventClient()
	if u.store, err = newStoreClient(u.config.StoreConfigFile); err != nil {
		log.Warnw("fail to new store", "err", err)
		return false
	}
	log.Info("uploader startup")
	return true
}

// todo: feat
func (u *UploaderService) Join() bool {
	for {
		time.Sleep(1 * time.Second)
	}
	return true
}

func (u *UploaderService) Stop() bool {
	return true
}

func (u *UploaderService) Description() string {
	return "UploaderService"
}

package context

import (
	"errors"
	"sync"
)

var (
	StoneHubService   string = "StoneHubService"
	StoneNodeService  string = "StoneNodeService"
	UploaderService   string = "UploaderService"
	DownloaderService string = "DownloaderService"
	SyncerService     string = "SyncerService"
	GatewayService    string = "GateWayService"
	RootService       string = "RootService"
)

var ServiceSupport = map[string]string{
	"StoneHub":          StoneHubService,
	"StoneNode":         StoneNodeService,
	"Uploader":          UploaderService,
	"Downloader":        DownloaderService,
	"Gateway":           GatewayService,
	"stonehub":          StoneHubService,
	"stonenode":         StoneNodeService,
	"uploader":          UploaderService,
	"downloader":        DownloaderService,
	"gateway":           GatewayService,
	"StoneHubService":   StoneHubService,
	"StoneNodeService":  StoneNodeService,
	"UploaderService":   UploaderService,
	"DownloaderService": DownloaderService,
	"GatewayService":    GatewayService,
}

var ServiceUsage = map[string]string{
	StoneHubService:   "Stone Hub Service dispatch stone(job context and job fsm), such as upload payload.",
	StoneNodeService:  "Stone Node Service is the unit that executes the stone.",
	UploaderService:   "Uploader Service upload the payload to primary storage provider.",
	DownloaderService: "Downloader Service downloader the payload from storage provider and sends to the client.",
	GatewayService:    "Gateway Service is the access layer for external interaction.",
}

type Context struct {
	Cfg            *CliConf
	CurrentService string
}

var context *Context
var once sync.Once

func GetContext() *Context {
	once.Do(func() {
		context = &Context{}
		context.CurrentService = RootService
	})
	return context
}

func (ctx *Context) EnterService(service string) error {
	serviceName, ok := ServiceSupport[service]
	if !ok {
		return errors.New("service name error")
	}
	ctx.CurrentService = serviceName
	return nil
}

func (ctx *Context) OutService() {
	ctx.CurrentService = RootService
}

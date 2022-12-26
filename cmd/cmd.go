package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bnb-chain/inscription-storage-provider/model"
	"github.com/bnb-chain/inscription-storage-provider/service/gateway"
	"github.com/bnb-chain/inscription-storage-provider/service/uploader"
)

var (
	version    = flag.Bool("v", false, "print version")
	module     = flag.String("m", "", "run module name")
	configFile = flag.String("c", "", "config file path")
)

func main() {
	flag.Parse()
	if *version {
		fmt.Print(DumpLogo() + DumpVersion())
		os.Exit(0)
	}

	var service model.Server
	switch *module {
	case "gateway":
		service = gateway.NewGatewayService()
	case "uploader":
		service = uploader.NewUploaderService()
	default:
		fmt.Printf("Fatal: module mismatch: %s", *module)
		os.Exit(1)
	}
	if ok := service.Init(*configFile); !ok {
		fmt.Printf("%s init failed, configfile:%s", *module, *configFile)
		os.Exit(1)
	}
	if ok := service.Start(); !ok {
		fmt.Printf("%s start failed, configfile:%s", *module, *configFile)
		os.Exit(1)
	}

	service.Join()
}

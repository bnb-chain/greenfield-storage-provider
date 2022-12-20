package main

import (
	"flag"
	"fmt"
	"github.com/bnb-chain/inscription-storage-provider/model"
	"github.com/bnb-chain/inscription-storage-provider/service/gateway"
	"os"
)

var (
	version    = flag.Bool("v", false, "print version")
	module     = flag.String("m", "", "run module name")
	configFile = flag.String("c", "", "config file path")
)

func main() {
	flag.Parse()
	if *version {
		fmt.Print(LogoString() + VersionString())
		os.Exit(0)
	}

	var service model.Server
	switch *module {
	case "gateway":
		service = gateway.NewGatewayService()
		ok := service.Init(*configFile)
		if !ok {
			fmt.Printf("%s init failed, configfile:%s", *module, *configFile)
			os.Exit(1)
		}
		ok = service.Start()
		if !ok {
			fmt.Printf("%s start failed, configfile:%s", *module, *configFile)
			os.Exit(1)
		}
		// fmt.Println(service.Description())
	default:
		fmt.Printf("Fatal: module mismatch: %s", *module)
		os.Exit(1)
	}

	service.Join()
}

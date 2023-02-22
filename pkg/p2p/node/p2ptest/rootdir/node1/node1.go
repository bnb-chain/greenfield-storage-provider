package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/node"
)

var configFile = flag.String("config", "", "config file path")

func main() {
	//flag.Parse()
	//fmt.Println("config file path: ", *configFile)
	//var conf node.NodeConfig
	//_, err := toml.DecodeFile(*configFile, &conf)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(conf)

	conf := node.DefaultNodeConfig()
	conf.P2P.ListenAddress = "tcp://127.0.0.1:6733"
	conf.P2P.BootstrapPeers = "16ad8821210d50fc1bdecb5eab716d12746f6744@127.0.0.1:6933"
	ctx := context.Background()
	service, err := node.NewDefault(ctx, &conf)
	if err != nil {
		fmt.Println(err)
	}
	go service.Start(ctx)
	for {

	}
}

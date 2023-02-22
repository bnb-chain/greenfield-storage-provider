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
	conf.P2P.ListenAddress = "tcp://127.0.0.1:6933"
	conf.P2P.BootstrapPeers = "5b1e35c5470e4815544b6d6cd21d0603a5eab35c@127.0.0.1:6733"
	ctx := context.Background()
	service, err := node.NewDefault(ctx, &conf)
	if err != nil {
		fmt.Println(err)
	}
	go service.Start(ctx)
	for {

	}
}

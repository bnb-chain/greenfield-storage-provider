package p2p

//import (
//	"github.com/bnb-chain/greenfield-storage-provider/service/p2p/node"
//	db "github.com/bnb-chain/greenfield-storage-provider/store/config"
//	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metasql"
//)
//
//type P2PServiceConfig struct {
//	GrpcAddress        string
//	P2PListenAddress   string
//	P2PExternalAddress string
//	MetaDB             *db.SqlDBConfig
//}
//
//func (config *P2PServiceConfig) NodeConfig() *node.NodeConfig {
//	cfg := node.DefaultNodeConfig()
//	cfg.DBConfig = config.MetaDB
//	cfg.P2P.ListenAddress = config.P2PListenAddress
//	cfg.P2P.ExternalAddress = config.P2PExternalAddress
//	return &cfg
//}
//
//var DefaultP2PServiceConfig = &P2PServiceConfig{
//	GrpcAddress:        "127.0.0.1:9733",
//	P2PListenAddress:   "127.0.0.1:21303",
//	P2PExternalAddress: "127.0.0.1:21313",
//	MetaDB:             metasql.DefaultMetaSqlDBConfig,
//}

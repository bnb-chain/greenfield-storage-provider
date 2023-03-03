package bsdb

import "github.com/bnb-chain/greenfield-storage-provider/store/config"

// DefaultBSDBConfig is default conf for Block Syncer DB, Modify it according to the actual configuration.
var DefaultBSDBConfig = &config.SqlDBConfig{
	User:     "root",
	Passwd:   "test_pwd",
	Address:  "127.0.0.1:3306",
	Database: "block_syncer",
}

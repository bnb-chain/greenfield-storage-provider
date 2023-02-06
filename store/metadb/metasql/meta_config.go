package metasql

import "github.com/bnb-chain/greenfield-storage-provider/store/config"

// DefaultMetaSqlDBConfig is default conf, Modify it according to the actual configuration.
var DefaultMetaSqlDBConfig = &config.SqlDBConfig{
	User:     "root",
	Passwd:   "test_pwd",
	Address:  "127.0.0.1:3306",
	Database: "meta_db",
}

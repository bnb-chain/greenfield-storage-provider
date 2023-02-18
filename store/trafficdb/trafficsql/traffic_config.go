package trafficsql

import "github.com/bnb-chain/greenfield-storage-provider/store/config"

// DefaultTrafficSqlDBConfig is default conf, Modify it according to the actual configuration.
var DefaultTrafficSqlDBConfig = &config.SqlDBConfig{
	User:     "root",
	Passwd:   "test_pwd",
	Address:  "127.0.0.1:3306",
	Database: "traffic_db",
}

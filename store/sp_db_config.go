package store

// SqlDBConfig is sql-db config
type SqlDBConfig struct {
	User     string
	Passwd   string
	Address  string
	Database string
}

// DefaultSqlDBConfig is default conf, Modify it according to the actual configuration.
var DefaultSqlDBConfig = &SqlDBConfig{
	User:     "root",
	Passwd:   "test_pwd",
	Address:  "local:3306",
	Database: "storage_provider_db",
}

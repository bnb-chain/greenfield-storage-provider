package config

// SQLDBConfig is sql db config
type SQLDBConfig struct {
	User     string
	Passwd   string
	Address  string
	Database string
}

// DefaultSQLDBConfig is default conf, Modify it according to the actual configuration.
var DefaultSQLDBConfig = &SQLDBConfig{
	User:     "root",
	Passwd:   "test_pwd",
	Address:  "localhost:3306",
	Database: "storage_provider_db",
}

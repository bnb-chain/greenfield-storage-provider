package metasql

// DBOption is mysql config options
type DBOption struct {
	User     string
	Passwd   string
	Address  string
	Database string
}

// DefaultDBOption is default conf, Modify it according to the actual configuration.
var DefaultDBOption = &DBOption{
	User:     "root",
	Passwd:   "test_pwd",
	Address:  "127.0.0.1:3306",
	Database: "meta_db",
}

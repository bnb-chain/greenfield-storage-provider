package config

// SQLDBConfig is sql db config
type SQLDBConfig struct {
	User            string
	Passwd          string
	Address         string
	Database        string
	ConnMaxLifetime int
	ConnMaxIdleTime int
	MaxIdleConns    int
	MaxOpenConns    int
}

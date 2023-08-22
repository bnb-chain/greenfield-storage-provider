package config

// SQLDBConfig is sql db config
type SQLDBConfig struct {
	User            string `comment:"required"`
	Passwd          string `comment:"required"`
	Address         string `comment:"required"`
	Database        string `comment:"required"`
	ConnMaxLifetime int    `comment:"optional"`
	ConnMaxIdleTime int    `comment:"optional"`
	MaxIdleConns    int    `comment:"optional"`
	MaxOpenConns    int    `comment:"optional"`
	// whether enable trace PutEvent, use by spdb
	EnableTracePutEvent bool `comment:"optional"`
}

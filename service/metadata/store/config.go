package store

import "strings"

var DefaultDBConfig = DBConfig{
	Driver:           "mysql",
	ConnectionString: "root:password@/meta_db?parseTime=true",
	EnableLog:        true,
}

type DBConfig struct {
	Driver           string       `yaml:"Driver"`
	ConnectionString SecretString `yaml:"ConnectionString"`
	EnableLog        bool         `yaml:"EnableLog"`
}

type SecretString string

// String prevent secrets to be printed in log
func (s SecretString) String() string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		b.WriteString("*")
	}
	return b.String()
}

func (s SecretString) ToString() string {
	return string(s)
}

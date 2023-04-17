package blocksyncer

import (
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/model/errors"
)

type Config struct {
<<<<<<< HEAD
	Modules []string
	Dsn     string
	Workers uint
=======
	Modules        []string
	Dsn            string
	RecreateTables bool
>>>>>>> 62efe729c7ebc5b8135263847dbeb93344712154
}

func getDBConfigFromEnv(dsn string) (string, error) {
	dsnVal, ok := os.LookupEnv(dsn)
	if !ok {
		return "", errors.ErrDSNNotSet
	}
	return dsnVal, nil
}

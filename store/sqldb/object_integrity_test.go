package sqldb_test

import (
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	storeconfig "github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SqlDBTestSuite struct {
	suite.Suite

	db *sqldb.SpDBImpl
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(SqlDBTestSuite))
}

func (suite *SqlDBTestSuite) SetupTest() {
	spDBConfig := &storeconfig.SQLDBConfig{
		User:     "root",
		Passwd:   "root",
		Address:  "127.0.0.1",
		Database: "storage_provider_db",
	}

	db, err := sqldb.NewSpDB(spDBConfig)
	suite.Require().NoError(err)
	if err != nil {
	}
	suite.db = db
}

func (suite *SqlDBTestSuite) TestObjectIntegritySet() {

	integrityMeta := &corespdb.IntegrityMeta{
		ObjectID:          11,
		RedundancyIndex:   -1,
		PieceChecksumList: checksums,
		IntegrityChecksum: integrity,
	}

	db := suite.db
	db.SetObjectIntegrity(integrityMeta)
}

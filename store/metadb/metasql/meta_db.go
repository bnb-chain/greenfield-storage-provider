package metasql

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// import (
//
//	"encoding/hex"
//	"fmt"
//
//	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
//	"github.com/bnb-chain/greenfield-storage-provider/store/config"
//	"github.com/bnb-chain/greenfield-storage-provider/store/spdb"
//	"github.com/bnb-chain/greenfield-storage-provider/util"
//	"github.com/bnb-chain/greenfield-storage-provider/util/log"
//	"gorm.io/driver/mysql"
//	"gorm.io/gorm"
//
// )
//
// MetaDB is a sql implement of MetaDB interface
type MetaDB struct {
	db *gorm.DB
}

// NewMetaDB return a database instance
func NewMetaDB(config *config.SqlDBConfig) (*MetaDB, error) {
	db, err := InitDB(config)
	if err != nil {
		return nil, err
	}
	return &MetaDB{db: db}, nil
}

//	func (mdb *MetaDB) Close() error {
//		return nil
//	}
//
// InitDB init a db instance
func InitDB(config *config.SqlDBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User,
		config.Passwd,
		config.Address,
		config.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Errorw("gorm open db failed", "err", err)
		return nil, err
	}

	// create if not exist
	if err := db.AutoMigrate(&DBIntegrityMeta{}); err != nil {
		log.Errorw("failed to create integrity meta table", "err", err)
		return nil, err
	}
	return db, nil
}

//
//// SetIntegrityMeta put(overwrite) integrity hash info to db
//func (mdb *MetaDB) SetIntegrityMeta(meta *spdb.IntegrityMeta) error {
//	var (
//		result                    *gorm.DB
//		insertIntegrityMetaRecord *DBIntegrityMeta
//	)
//
//	insertIntegrityMetaRecord = &DBIntegrityMeta{
//		ObjectID:       meta.ObjectID,
//		EcIdx:          meta.EcIdx,
//		PieceCount:     meta.PieceCount,
//		IsPrimary:      meta.IsPrimary,
//		RedundancyType: uint32(meta.RedundancyType),
//		IntegrityHash:  hex.EncodeToString(meta.IntegrityHash),
//		PieceHash:      util.EncodePieceHash(meta.PieceHash),
//		Signature:      hex.EncodeToString(meta.Signature),
//	}
//	result = mdb.db.Create(insertIntegrityMetaRecord)
//	if result.Error != nil || result.RowsAffected != 1 {
//		return fmt.Errorf("insert integrity meta record failed, %s, %v", result.Error, *meta)
//	}
//	return nil
//}
//
//// GetIntegrityMeta return the integrity hash info
//func (mdb *MetaDB) GetIntegrityMeta(objectID uint64) (*spdb.IntegrityMeta, error) {
//	var (
//		result      *gorm.DB
//		queryReturn DBIntegrityMeta
//	)
//	result = mdb.db.Model(&DBIntegrityMeta{}).
//		Where("object_id = ?", objectID).
//		First(&queryReturn)
//	if result.Error != nil {
//		return nil, fmt.Errorf("select integrity meta record failed, %s", result.Error)
//	}
//	integrityHash, err := hex.DecodeString(queryReturn.IntegrityHash)
//	if err != nil {
//		return nil, err
//	}
//	signature, err := hex.DecodeString(queryReturn.Signature)
//	if err != nil {
//		return nil, err
//	}
//
//	meta := &spdb.IntegrityMeta{
//		ObjectID:       queryReturn.ObjectID,
//		EcIdx:          queryReturn.EcIdx,
//		PieceCount:     queryReturn.PieceCount,
//		IsPrimary:      queryReturn.IsPrimary,
//		RedundancyType: ptypes.RedundancyType(queryReturn.RedundancyType),
//		IntegrityHash:  integrityHash,
//		Signature:      signature,
//	}
//	meta.PieceHash, err = util.DecodePieceHash(queryReturn.PieceHash)
//	if err != nil {
//		return nil, err
//	}
//	return meta, nil
//}

package metasql

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	ptypesv1pb "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MetaDB is a sql implement of MetaDB interface
type MetaDB struct {
	db *gorm.DB
}

// NewMetaDB return a database instance
func NewMetaDB(option *DBOption) (*MetaDB, error) {
	db, err := InitDB(option)
	if err != nil {
		return nil, err
	}
	return &MetaDB{db: db}, nil
}

// InitDB init a db instance
func InitDB(opt *DBOption) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		opt.User,
		opt.Passwd,
		opt.Address,
		opt.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Warnw("gorm open db failed", "err", err)
		return nil, err
	}

	// create if not exist
	if err := db.AutoMigrate(&DBIntegrityMeta{}); err != nil {
		log.Warnw("failed to create integrity meta table", "err", err)
		return nil, err
	}
	if err := db.AutoMigrate(&DBUploadPayloadAskingMeta{}); err != nil {
		log.Warnw("failed to create upload payload asking meta table", "err", err)
		return nil, err
	}
	return db, nil
}

// SetIntegrityMeta put(overwrite) integrity hash info to db
func (mdb *MetaDB) SetIntegrityMeta(meta *metadb.IntegrityMeta) error {
	var (
		result *gorm.DB
	)

	_, err := mdb.GetIntegrityMeta(meta.ObjectID)
	if err != nil {
		// insert record
		pieceHash, err := json.Marshal(meta.PieceHash)
		if err != nil {
			return err
		}
		insertIntegrityMetaRecord := &DBIntegrityMeta{
			ObjectID:       meta.ObjectID,
			PieceIdx:       meta.PieceIdx,
			PieceCount:     meta.PieceCount,
			IsPrimary:      meta.IsPrimary,
			RedundancyType: uint32(meta.RedundancyType),
			IntegrityHash:  hex.EncodeToString(meta.IntegrityHash),
			PieceHash:      string(pieceHash),
		}
		result = mdb.db.Create(insertIntegrityMetaRecord)
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("insert integrity meta record failed, %s", result.Error)
		}
	} else {
		// update record
		queryCondition := &DBIntegrityMeta{
			ObjectID: meta.ObjectID,
		}
		pieceHash, err := json.Marshal(meta.PieceHash)
		if err != nil {
			return err
		}
		updateFields := &DBIntegrityMeta{
			PieceIdx:       meta.PieceIdx,
			PieceCount:     meta.PieceCount,
			IsPrimary:      meta.IsPrimary,
			RedundancyType: uint32(meta.RedundancyType),
			IntegrityHash:  hex.EncodeToString(meta.IntegrityHash),
			PieceHash:      string(pieceHash),
		}
		result = mdb.db.Model(queryCondition).Updates(updateFields)
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("update integrity meta record's height failed, %s", result.Error)
		}
	}
	return nil
}

// GetIntegrityMeta return the integrity hash info
func (mdb *MetaDB) GetIntegrityMeta(objectID uint64) (*metadb.IntegrityMeta, error) {
	var (
		result         *gorm.DB
		queryCondition *DBIntegrityMeta
		queryReturn    DBIntegrityMeta
	)
	// If the primary key is a number, the query will be written as follows:
	queryCondition = &DBIntegrityMeta{
		ObjectID: objectID,
	}
	result = mdb.db.Model(queryCondition).First(&queryReturn)
	if result.Error != nil {
		return nil, fmt.Errorf("select integrity meta record's failed, %s", result.Error)
	}
	integrityHash, err := hex.DecodeString(queryReturn.IntegrityHash)
	if err != nil {
		return nil, err
	}
	pieceHash := make(map[string][]byte)
	err = json.Unmarshal([]byte(queryReturn.PieceHash), &pieceHash)
	if err != nil {
		return nil, err
	}
	return &metadb.IntegrityMeta{
		ObjectID:       queryReturn.ObjectID,
		PieceIdx:       queryReturn.PieceIdx,
		PieceCount:     queryReturn.PieceCount,
		IsPrimary:      queryReturn.IsPrimary,
		RedundancyType: ptypesv1pb.RedundancyType(queryReturn.RedundancyType),
		IntegrityHash:  integrityHash,
		PieceHash:      pieceHash,
	}, nil
}

// SetUploadPayloadAskingMeta put(overwrite) payload asking info to db
func (mdb *MetaDB) SetUploadPayloadAskingMeta(meta *metadb.UploadPayloadAskingMeta) error {
	var (
		result *gorm.DB
	)

	_, err := mdb.GetUploadPayloadAskingMeta(meta.BucketName, meta.ObjectName)
	if err != nil {
		// insert record
		insertPayloadAskingMetaRecord := &DBUploadPayloadAskingMeta{
			BucketName: meta.BucketName,
			ObjectName: meta.ObjectName,
			Timeout:    meta.Timeout,
		}
		result = mdb.db.Create(insertPayloadAskingMetaRecord)
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("insert payload asking meta record failed, %s", result.Error)
		}
	} else {
		// update record
		result = mdb.db.Model(&DBUploadPayloadAskingMeta{}).
			Where("bucket_name = ? and object_name ?", meta.BucketName, meta.ObjectName).
			Update("timeout", meta.Timeout)
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("update payload asking meta record's height failed, %s", result.Error)
		}
	}
	return nil
}

// GetUploadPayloadAskingMeta return the payload asking info
func (mdb *MetaDB) GetUploadPayloadAskingMeta(bucket, object string) (*metadb.UploadPayloadAskingMeta, error) {
	var (
		result      *gorm.DB
		queryReturn DBUploadPayloadAskingMeta
	)
	// If the primary key is a string, the query will be written as follows:
	result = mdb.db.First(&queryReturn, "bucket_name = ? and object_name ?", bucket, object)
	if result.Error != nil {
		return nil, fmt.Errorf("select payload asking record's failed, %s", result.Error)
	}
	return &metadb.UploadPayloadAskingMeta{
		BucketName: queryReturn.BucketName,
		ObjectName: queryReturn.ObjectName,
		Timeout:    queryReturn.Timeout,
	}, nil
}

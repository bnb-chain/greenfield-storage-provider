package jobsql

import (
	"fmt"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// DBJobV2 table schema
type DBJobV2 struct {
	JobID      uint64 `gorm:"primary_key;autoIncrement"`
	JobType    uint32
	JobState   uint32
	JobErr     string
	CreateTime time.Time
	ModifyTime time.Time
}

// TableName is used to set Job Schema's table name in database
func (DBJobV2) TableName() string {
	return "job_v2"
}

// DBObjectV2 table schema
type DBObjectV2 struct {
	ObjectID       uint64 `gorm:"primary_key"`
	JobID          uint64 `gorm:"index:job_to_object"` // Job.JobID
	CreateHash     string
	SealHash       string
	Owner          string
	BucketName     string
	ObjectName     string
	Size           uint64
	Checksum       string
	IsPrivate      bool
	ContentType    string
	PrimarySP      string
	Height         uint64
	RedundancyType uint32
}

// TableName is used to set Object Schema's table name in database
func (DBObjectV2) TableName() string {
	return "object_v2"
}

// DBPieceJobV2 table schema
type DBPieceJobV2 struct {
	ObjectID        uint64 `gorm:"index:idx_piece_group"`
	PieceType       uint32 `gorm:"index:idx_piece_group"`
	PieceIdx        uint32
	PieceState      uint32
	Checksum        string
	StorageProvider string
	IntegrityHash   string
	Signature       string
}

// TableName is used to set PieceJob Schema's table name in database
func (DBPieceJobV2) TableName() string {
	return "piece_job_v2"
}

// InitDB is used to connect mysql and create table if not existed
func InitDB(config *config.SqlDBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User,
		config.Passwd,
		config.Address,
		config.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Warnw("gorm open db failed", "err", err)
		return nil, err
	}

	// create if not exist
	if err := db.AutoMigrate(&DBJobV2{}); err != nil {
		log.Warnw("failed to create job table v2", "err", err)
		return nil, err
	}
	if err := db.AutoMigrate(&DBObjectV2{}); err != nil {
		log.Warnw("failed to create object table v2", "err", err)
		return nil, err
	}
	if err := db.AutoMigrate(&DBPieceJobV2{}); err != nil {
		log.Warnw("failed to create piece job table v2", "err", err)
		return nil, err
	}

	return db, nil
}

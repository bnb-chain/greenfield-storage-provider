package jobsql

import (
	"fmt"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// DBJob table schema
type DBJob struct {
	JobID      uint64 `gorm:"primary_key;autoIncrement"`
	JobType    uint32
	JobState   uint32
	JobErr     string
	CreateTime time.Time
	ModifyTime time.Time
}

// TableName is used to set Job Schema's table name in database
func (DBJob) TableName() string {
	return "job"
}

// DBObject table schema
type DBObject struct {
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
func (DBObject) TableName() string {
	return "object"
}

// DBPieceJob table schema
type DBPieceJob struct {
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
func (DBPieceJob) TableName() string {
	return "piece_job"
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
		log.Errorw("gorm open db failed", "err", err)
		return nil, err
	}

	// create if not exist
	if err := db.AutoMigrate(&DBJob{}); err != nil {
		log.Errorw("failed to create job table", "err", err)
		return nil, err
	}
	if err := db.AutoMigrate(&DBObject{}); err != nil {
		log.Errorw("failed to create object table", "err", err)
		return nil, err
	}
	if err := db.AutoMigrate(&DBPieceJob{}); err != nil {
		log.Errorw("failed to create piece job table", "err", err)
		return nil, err
	}

	return db, nil
}

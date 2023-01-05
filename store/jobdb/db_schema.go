package jobdb

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/bnb-chain/inscription-storage-provider/util/log"
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
	CreateHash     string `gorm:"primary_key"`
	JobID          uint64 // Job.JobID
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
	CreateHash      string `gorm:"index:idx_piece_group"` // Object.CreateHash
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
	Database: "bfs_meta",
}

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
	if err := db.AutoMigrate(&DBJob{}); err != nil {
		log.Warnw("failed to create job table", "err", err)
		return nil, err
	}
	if err := db.AutoMigrate(&DBObject{}); err != nil {
		log.Warnw("failed to create object table", "err", err)
		return nil, err
	}
	if err := db.AutoMigrate(&DBPieceJob{}); err != nil {
		log.Warnw("failed to create piece job table", "err", err)
		return nil, err
	}
	return db, nil
}

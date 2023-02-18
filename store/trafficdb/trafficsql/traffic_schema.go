package trafficsql

import (
	"fmt"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DBBucketTraffic table schema
type DBBucketTraffic struct {
	BucketID  uint64 `gorm:"index:idx_bucket_traffic_group"`
	YearMonth string `gorm:"index:idx_bucket_traffic_group"`

	BucketName    string
	ReadCostSize  int64
	ReadQuotaSize int64
	ModifyTime    time.Time
}

// TableName is used to set BucketTraffic Schema's table name in database
func (DBBucketTraffic) TableName() string {
	return "bucket_traffic"
}

// DBReadRecord table schema
type DBReadRecord struct {
	ReadRecordID uint64 `gorm:"primary_key;autoIncrement"`

	BucketID    uint64 `gorm:"index:bucket_to_read_record"`
	ObjectID    uint64 `gorm:"index:object_to_read_record"`
	UserAddress string `gorm:"index:user_to_read_record"`
	ReadTime    int64  `gorm:"index:time_to_read_record"` // second timestamp

	BucketName string
	ObjectName string
	ReadSize   int64
}

// TableName is used to set ReadRecord Schema's table name in database
func (DBReadRecord) TableName() string {
	return "read_record"
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
	if err := db.AutoMigrate(&DBBucketTraffic{}); err != nil {
		log.Warnw("failed to create bucket traffic table", "err", err)
		return nil, err
	}
	if err := db.AutoMigrate(&DBReadRecord{}); err != nil {
		log.Warnw("failed to create read record table", "err", err)
		return nil, err
	}

	return db, nil
}

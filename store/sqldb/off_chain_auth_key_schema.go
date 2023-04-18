package sqldb

import (
	"time"
)

// OffChainAuthKeyTable table schema
type OffChainAuthKeyTable struct {
	UserAddress string `gorm:"primary_key"`
	Domain      string `gorm:"primary_key"`

	CurrentNonce     int32
	CurrentPublicKey string
	NextNonce        int32
	ExpiryDate       time.Time

	CreatedTime  time.Time
	ModifiedTime time.Time
}

// TableName is used to set JobTable Schema's table name in database
func (OffChainAuthKeyTable) TableName() string {
	return OffChainAuthKeyTableName
}

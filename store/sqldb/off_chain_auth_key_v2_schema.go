package sqldb

import (
	"time"
)

// OffChainAuthKeyV2Table table schema
type OffChainAuthKeyV2Table struct {
	UserAddress string `gorm:"primary_key"`
	Domain      string `gorm:"primary_key"`
	PublicKey   string `gorm:"primary_key"`

	ExpiryDate   time.Time
	CreatedTime  time.Time
	ModifiedTime time.Time
}

// TableName is used to set OffChainAuthKeyV2 Schema's table name in database
func (OffChainAuthKeyV2Table) TableName() string {
	return OffChainAuthKeyV2TableName
}

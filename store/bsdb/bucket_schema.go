package bsdb

import (
	"github.com/forbole/juno/v4/common"
	"github.com/shopspring/decimal"
)

// Bucket is the structure for user bucket
type Bucket struct {
	// ID defines db auto_increment id of bucket
	ID uint64 `gorm:"id"`
	// Owner is the account address of bucket creator, it is also the bucket owner.
	Owner common.Address `gorm:"owner"`
	// Operator defines the operator address of bucket
	Operator common.Address `gorm:"operator"`
	// BucketName is a globally unique name of bucket
	BucketName string `gorm:"bucket_name"`
	// Visibility defines the highest permissions for bucket. When a bucket is public, everyone can get storage obj
	Visibility string `gorm:"visibility"`
	// ID is the unique identification for bucket.
	BucketID common.Hash `gorm:"bucket_id"`
	// SourceType defines which chain the user should send the bucket management transactions to
	SourceType string `gorm:"source_type"`
	// CreateAt defines the block number when the bucket created.
	CreateAt int64 `gorm:"create_at"`
	// CreateTime defines the timestamp when the bucket created
	CreateTime int64 `gorm:"create_time"`
	// CreateTxHash defines the creation transaction hash of bucket
	CreateTxHash common.Hash `gorm:"create_tx_hash"`
	// PaymentAddress is the address of the payment account
	PaymentAddress common.Address `gorm:"payment_address"`
	// GlobalVirtualGroupFamilyID defines the unique id of gvg family
	GlobalVirtualGroupFamilyID uint32 `json:"global_virtual_group_family_id"`
	// ReadQuota defines the traffic quota for read
	ChargedReadQuota uint64 `gorm:"charged_read_quota"`
	// PaymentPriceTime defines price time of payment
	PaymentPriceTime int64 `gorm:"payment_price_time"`
	// Removed defines the bucket is deleted or not
	Removed bool `gorm:"removed"`
	// Status defines the status of bucket
	Status string `gorm:"column:status"`
	// DeleteAt defines the block number when the bucket deleted.
	DeleteAt int64 `gorm:"delete_at"`
	// DeleteReason defines the deleted reason of bucket
	DeleteReason string `gorm:"delete_reason"`
	// StorageSize storage size of bucket
	StorageSize decimal.Decimal `gorm:"column:storage_size;type:DECIMAL(65, 0);not null"`
	// ChargeSize charge size of bucket
	ChargeSize decimal.Decimal `gorm:"column:charge_size;type:DECIMAL(65, 0);not null"`
	// UpdateAt defines the block number when the bucket update.
	UpdateAt int64 `gorm:"column:update_at"`
	// UpdateTxHash defines the update transaction hash of bucket
	UpdateTxHash common.Hash `gorm:"update_tx_hash"`
	// UpdateTime defines the timestamp when the bucket update.
	UpdateTime int64 `gorm:"column:update_time"`
}

// TableName is used to set Bucket table name in database
func (b *Bucket) TableName() string {
	return BucketTableName
}

// BucketFullMeta is the structure for user bucket with its related info
type BucketFullMeta struct {
	Bucket
	StreamRecord
}

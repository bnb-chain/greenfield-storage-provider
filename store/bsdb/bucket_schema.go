package bsdb

import (
	"github.com/forbole/juno/v4/common"
)

// Bucket is the structure for user bucket
type Bucket struct {
	// ID defines db auto_increment id of bucket
	ID uint64 `gorm:"id"`
	// Owner is the account address of bucket creator, it is also the bucket owner.
	Owner common.Address `gorm:"column:owner_address"`
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
	// PaymentAddress is the address of the payment account
	PaymentAddress common.Address `gorm:"payment_address"`
	// PrimarySpAddress is the address of the primary sp. Objects belong to this bucket will never
	// leave this SP, unless you explicitly shift them to another SP.
	PrimarySpAddress common.Address `gorm:"primary_sp_address"`
	// ReadQuota defines the traffic quota for read
	ChargedReadQuota uint64 `gorm:"charged_read_quota"`
	// PaymentPriceTime defines price time of payment
	PaymentPriceTime int64 `gorm:"payment_price_time"`
	// Removed defines the bucket is deleted or not
	Removed bool `gorm:"removed"`
}

// TableName is used to set Bucket table name in database
func (b *Bucket) TableName() string {
	return BucketTableName
}

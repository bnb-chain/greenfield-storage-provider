package metadata

import "github.com/bnb-chain/greenfield-storage-provider/model"

// Bucket is the structure for user bucket
type Bucket struct {
	// Owner is the account address of bucket creator, it is also the bucket owner.
	Owner string `json:"owner"`
	// BucketName is a globally unique name of bucket
	BucketName string `json:"bucketName"`
	// IsPublic defines the highest permissions for bucket. When the bucket is public, everyone can get the object in it.
	IsPublic bool `json:"isPublic"`
	// ID is the unique identification for bucket.
	ID string `json:"id"`
	// SourceType defines which chain the user should send the bucket management transactions to
	SourceType int `json:"sourceType"`
	// CreateAt defines the block number when the bucket created.
	CreateAt int64 `json:"createAt"`
	// PaymentAddress is the address of the payment account
	PaymentAddress string `json:"paymentAddress"`
	// PrimarySpAddress is the address of the primary sp. Objects belong to this bucket will never
	// leave this SP, unless you explicitly shift them to another SP.
	PrimarySpAddress string `json:"primarySpAddress"`
	// ReadQuota defines the traffic quota for read
	ReadQuota int `json:"readQuota"`
	// PaymentPriceTime defines price time of payment
	PaymentPriceTime int64 `json:"paymentPriceTime"`
}

func (a *Bucket) TableName() string {
	return model.BucketTableName
}

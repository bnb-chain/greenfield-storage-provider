package metadata

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/ethereum/go-ethereum/common"
)

// Object is the structure for user object
type Object struct {
	// ID defines db auto_increment id of object
	ID uint64 `gorm:"id"`
	// Creator defines the account address of object creator
	Creator common.Address `gorm:"creator"`
	// Owner defines the account address of object owner
	Owner common.Address `gorm:"owner"`
	// BucketName is the name of the bucket
	BucketName string `gorm:"bucket_name"`
	// ObjectName is the name of object
	ObjectName string `gorm:"object_name"`
	// ObjectID is the unique identifier of object
	ObjectID int64 `gorm:"object_id"`
	// BucketID is the unique identifier of bucket
	BucketID int64 `gorm:"bucket_id"`
	// PayloadSize is the total size of the object payload
	PayloadSize uint64 `gorm:"payload_size"`
	// IsPublic defines the highest permissions for object. When the object is public, everyone can access it
	IsPublic bool `gorm:"is_public"`
	// ContentType defines the format of the object which should be a standard MIME type
	ContentType string `gorm:"content_type"`
	// CreateAt defines the block number when the object created
	CreateAt int64 `gorm:"create_at"`
	// ObjectStatus defines the upload status of the object.
	ObjectStatus string `gorm:"object_status"`
	// RedundancyType defines the type of the redundancy which can be multi-replication or EC
	RedundancyType string `gorm:"redundancy_type"`
	// SourceType defines the source of the object.
	SourceType string `gorm:"source_type"`
	// Checksums defines the root hash of the pieces which stored in a SP
	CheckSums CheckSums `gorm:"check_sums;type:text"`
	// SecondarySpAddresses defines the addresses of secondary_sps
	SecondarySpAddresses SecondarySpAddresses `gorm:"secondary_sp_addresses;type:text"`
	// LockedBalance defines locked balance of object
	LockedBalance string `gorm:"locked_balance"`
	// Removed defines the object is deleted or not
	Removed bool `gorm:"removed"`
	// UpdateTime defines the time when the object updated
	UpdateTime int64 `gorm:"update_time"`
	// UpdateAt defines the block number when the object updated
	UpdateAt int64 `gorm:"update_at"`
}

// TableName is used to set Object table name in database
func (a *Object) TableName() string {
	return model.ObjectTableName
}

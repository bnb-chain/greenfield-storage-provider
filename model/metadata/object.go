package metadata

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
)

// Object is the structure for user object
type Object struct {
	// Owner is the account address of object creator, it is also the object owner.
	Owner string `json:"owner"`
	// BucketName is the name of the bucket
	BucketName string `json:"bucketName"`
	// ObjectName is the name of object
	ObjectName string `json:"objectName"`
	// ID is the unique identifier of object
	ID string `json:"id"`
	// PayloadSize is the total size of the object payload
	PayloadSize uint64 `json:"payloadSize"`
	// IsPublic defines the highest permissions for object. When the object is public, everyone can access it.
	IsPublic bool `json:"isPublic"`
	// ContentType defines the format of the object which should be a standard MIME type.
	ContentType string `json:"contentType"`
	// CreateAt defines the block number when the object created
	CreateAt int64 `json:"createAt"`
	// ObjectStatus defines the upload status of the object.
	ObjectStatus int `json:"objectStatus"`
	// RedundancyType defines the type of the redundancy which can be multi-replication or EC.
	RedundancyType int `json:"redundancyType"`
	// SourceType defines the source of the object.
	SourceType int `json:"sourceType"`
	// SecondarySpAddresses defines the addresses of secondary_sps
	SecondarySpAddresses []string `json:"secondarySpAddresses"`
	// LockedBalance defines locked balance of object
	LockedBalance string `json:"lockedBalance"`
}

// TableName is used to set Object table name in database
func (a *Object) TableName() string {
	return model.ObjectTableName
}

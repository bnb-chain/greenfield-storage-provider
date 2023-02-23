package model

type Object struct {
	Owner string `json:"owner"`
	// bucket_name is the name of the bucket
	BucketName string `json:"bucketName"`
	// object_name is the name of object
	ObjectName string `json:"objectName"`
	// id is the unique identifier of object
	Id string `json:"id"`
	// payloadSize is the total size of the object payload
	PayloadSize uint64 `json:"payloadSize"`
	// is_public define the highest permissions for object. When the object is public, everyone can access it.
	IsPublic bool `json:"isPublic"`
	// content_type define the format of the object which should be a standard MIME type.
	ContentType string `json:"contentType"`
	// create_at define the block number when the object created
	CreateAt int64 `json:"createAt"`
	// object_status define the upload status of the object.
	ObjectStatus string `json:"objectStatus"`
	// redundancy_type define the type of the redundancy which can be multi-replication or EC.
	RedundancyType string `json:"redundancyType"`
	// source_type define the source of the object.
	SourceType string `json:"sourceType"`
	// checksums define the root hash of the pieces which stored in a SP.
	Checksums []byte `json:"checksums"`
	// secondary_sp_addresses define the addresses of secondary_sps
	SecondarySpAddresses []string `json:"secondarySpAddresses"`
	// lockedBalance
	LockedBalance string `json:"lockedBalance"`
}

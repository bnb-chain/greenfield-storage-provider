package bsdb

import "github.com/forbole/juno/v4/common"

type EventMigrateBucket struct {
	Operator   common.Address `json:"operator,omitempty"`
	BucketName string         `json:"bucket_name,omitempty"`
	// bucket_id define an u256 id for object
	BucketId       common.Hash `json:"bucket_id,omitempty"`
	DstPrimarySpId uint32      `json:"dst_primary_sp_id,omitempty"`
}

// TableName is used to set EventMigrateBucket table name in database
func (g *EventMigrateBucket) TableName() string {
	return EventMigrateBucketTableName
}

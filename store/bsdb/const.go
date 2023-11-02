package bsdb

// define metadata query statement
const (
	// DeletedObjectsDefaultSize defines the default size of ListDeletedObjectsByBlockNumberRange response
	DeletedObjectsDefaultSize = 1000
	// ExpiredBucketsDefaultSize defines the default size of ListExpiredBucketsBySp response
	ExpiredBucketsDefaultSize = 1000
	// ListObjectsDefaultMaxKeys defines the default size of ListObjectsByBucketName response
	ListObjectsDefaultMaxKeys = 50
	// ListGroupsDefaultLimit defines the default size of GetUserGroups response
	ListGroupsDefaultLimit = 50
	// GetUserBucketsLimitSize defines the default limit for the number of buckets in any given account is 100
	GetUserBucketsLimitSize = 100
	// ListObjectsLimitSize defines the default limit of ListObjectsByBucketName response
	ListObjectsLimitSize = 1000
	// ListGroupsLimitSize defines the default limit of GetUserGroups response
	ListGroupsLimitSize = 1000
	// 	LisPoliciesDefaultLimit defines the default size of policies response
	LisPoliciesDefaultLimit = 50
	// LisPoliciesLimitSize defines the max limit of policies response
	LisPoliciesLimitSize = 1000
)

// define table name constant of block syncer db
const (
	// BucketTableName defines the name of bucket table
	BucketTableName = "buckets"
	// ObjectTableName defines the name of object table
	ObjectTableName = "objects"
	// EpochTableName defines the name of epoch table
	EpochTableName = "epoch"
	// PermissionTableName defines the name of permission table
	PermissionTableName = "permission"
	// StatementTableName defines the name of statement table
	StatementTableName = "statements"
	// GroupTableName defines the name of group table
	GroupTableName = "groups"
	// MasterDBTableName defines the name of master db table
	MasterDBTableName = "master_db"
	// PrefixTreeTableName defines the name of prefix tree node table
	PrefixTreeTableName = "slash_prefix_tree_nodes"
	// GlobalVirtualGroupFamilyTableName defines the name of global virtual group family table
	GlobalVirtualGroupFamilyTableName = "global_virtual_group_families"
	// LocalVirtualGroupTableName defines the name of local virtual group table
	LocalVirtualGroupTableName = "local_virtual_groups"
	// GlobalVirtualGroupTableName defines the name of global virtual group table
	GlobalVirtualGroupTableName = "global_virtual_groups"
	// EventMigrationTableName defines the name of event migrate bucket table
	EventMigrationTableName = "event_migration_bucket"
	// EventCompleteMigrationTableName defines the name of event complete migrate bucket table
	EventCompleteMigrationTableName = "event_complete_migration_bucket"
	// EventCancelMigrationTableName defines the name of event cancel migrate bucket table
	EventCancelMigrationTableName = "event_cancel_migration_bucket"
	// EventRejectMigrateTableName defines the name of event reject migrate bucket table
	EventRejectMigrateTableName = "event_reject_migrate_bucket"
	// EventStorageProviderExitTableName defines the name of event sp exit table
	EventStorageProviderExitTableName = "event_sp_exit"
	// EventCompleteStorageProviderExitTableName defines the name of event sp exit complete table
	EventCompleteStorageProviderExitTableName = "event_sp_exit_complete"
	// EventSwapOutTableName defines the name of event swap out table
	EventSwapOutTableName = "event_swap_out"
	// ObjectIDMapTableName defines the name of object id map table
	ObjectIDMapTableName = "object_id_map"
	// EventCompleteSwapOutTableName defines the name of event swap out complete table
	EventCompleteSwapOutTableName = "event_swap_out_complete"
	// EventCancelSwapOutTableName defines the name of event swap out cancel table
	EventCancelSwapOutTableName = "event_cancel_swap_out"
	// StorageProviderTableName defines the name of storage providers table
	StorageProviderTableName = "storage_providers"
	// StreamRecordTableName stream records of payment info
	StreamRecordTableName = "stream_records"
	// PaymentAccountTableName defines payment account info
	PaymentAccountTableName = "payment_accounts"
)

// define the list objects const
const (
	ObjectName   = "object"
	CommonPrefix = "common_prefix"
	GroupAddress = "0x0000000000000000000000000000000000000000"
)

// define the metrics const
const (
	DatabaseSuccess = "success"
	DatabaseFailure = "failure"
	DatabaseLevel   = "database"
)

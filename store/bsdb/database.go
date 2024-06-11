package bsdb

import (
	"github.com/evmos/evmos/v12/x/permission/types"
	"github.com/forbole/juno/v4/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Metadata contains all the methods required by block syncer db database
type Metadata interface {
	// GetUserBuckets get buckets info by a user address
	GetUserBuckets(accountID common.Address, includeRemoved bool) ([]*Bucket, error)
	// GetUserBucketsCount get buckets count by a user address
	GetUserBucketsCount(accountID common.Address, includeRemoved bool) (int64, error)
	// GetBucketByName get buckets info by a bucket name
	GetBucketByName(bucketName string, includePrivate bool) (*Bucket, error)
	// GetBucketInfoByBucketName get buckets info by a bucket name, including deleted and private bucket
	GetBucketInfoByBucketName(bucketName string) (*Bucket, error)
	// GetBucketByID get buckets info by by a bucket id
	GetBucketByID(bucketID int64, includePrivate bool) (*Bucket, error)
	// GetLatestBlockNumber get current latest block number
	GetLatestBlockNumber() (uint64, error)
	// GetPaymentByBucketName get bucket payment info by a bucket name
	GetPaymentByBucketName(bucketName string, includePrivate bool) (*StreamRecord, error)
	// GetPaymentByBucketID get bucket payment info by a bucket id
	GetPaymentByBucketID(bucketID int64, includePrivate bool) (*StreamRecord, error)
	// GetPaymentByPaymentAddress get bucket payment info by a payment address
	GetPaymentByPaymentAddress(address common.Address) (*StreamRecord, error)
	// GetPermissionByResourceAndPrincipal get permission info by resource type & id, principal type & value
	GetPermissionByResourceAndPrincipal(resourceType, principalType, principalValue string, resourceID common.Hash) (*Permission, error)
	// GetStatementsByPolicyID get statements info by a policy id
	GetStatementsByPolicyID(policyIDList []common.Hash, includeRemoved bool) ([]*Statement, error)
	// GetPermissionsByResourceAndPrincipleType get permissions info by resource type & id, principal type
	GetPermissionsByResourceAndPrincipleType(resourceType, principalType string, resourceID common.Hash, includeRemoved bool) ([]*Permission, error)
	// GetGroupsByGroupIDAndAccount get groups info by group id list and account id
	GetGroupsByGroupIDAndAccount(groupIDList []common.Hash, account common.Address, includeRemoved bool) ([]*Group, error)
	// ListObjectsByBucketName list objects info by a bucket name
	ListObjectsByBucketName(bucketName, continuationToken, prefix, delimiter string, maxKeys int, includeRemoved bool) ([]*ListObjectsResult, error)
	// ListDeletedObjectsByBlockNumberRange list deleted objects info by a block number range
	ListDeletedObjectsByBlockNumberRange(startBlockNumber uint64, endBlockNumber uint64, includePrivate bool) ([]*Object, error)
	// ListExpiredBucketsBySp list expired buckets by sp
	ListExpiredBucketsBySp(createAt int64, primarySpID uint32, limit int64) ([]*Bucket, error)
	// GetObjectByName get object info by an object name
	GetObjectByName(objectName string, bucketName string, includePrivate bool) (*Object, error)
	// GetObjectByID get object info by an object id
	GetObjectByID(objectID int64, includeRemoved bool) (*Object, error)
	// GetLatestObjectID get latest object id
	GetLatestObjectID() (uint64, error)
	// GetSwitchDBSignal check if there is a signal to switch the database
	GetSwitchDBSignal() (*MasterDB, error)
	// GetBucketMetaByName get bucket info with its related info
	GetBucketMetaByName(bucketName string, includePrivate bool) (*BucketFullMeta, error)
	// ListGroupsByNameAndSourceType get groups list by specific parameters
	ListGroupsByNameAndSourceType(name, prefix, sourceType string, limit, offset int, includeRemoved bool) ([]*Group, int64, error)
	// ListObjectsByIDs list objects by object ids
	ListObjectsByIDs(ids []common.Hash, includeRemoved bool) ([]*Object, error)
	// ListBucketsByIDs list buckets by bucket ids
	ListBucketsByIDs(ids []common.Hash, includeRemoved bool) ([]*Bucket, error)
	// GetGroupByID get group info by an object id
	GetGroupByID(groupID int64, includeRemoved bool) (*Group, error)
	// ListVirtualGroupFamiliesBySpID list virtual group families by sp id
	ListVirtualGroupFamiliesBySpID(spID uint32) ([]*GlobalVirtualGroupFamily, error)
	// GetVirtualGroupFamiliesByVgfID get virtual group families by vgf id
	GetVirtualGroupFamiliesByVgfID(vgfID uint32) (*GlobalVirtualGroupFamily, error)
	// GetGlobalVirtualGroupByGvgID get global virtual group by gvg id
	GetGlobalVirtualGroupByGvgID(gvgID uint32) (*GlobalVirtualGroup, error)
	// ListObjectsInGVGAndBucket list objects by gvg and bucket id
	ListObjectsInGVGAndBucket(bucketID common.Hash, gvgID uint32, startAfter common.Hash, limit int) ([]*Object, *Bucket, error)
	// ListObjectsByGVGAndBucketForGC list objects by gvg and bucket for gc
	ListObjectsByGVGAndBucketForGC(bucketID common.Hash, gvgID uint32, startAfter common.Hash, limit int) ([]*Object, *Bucket, error)
	// ListObjectsInGVG list objects by gvg
	ListObjectsInGVG(gvgID uint32, startAfter common.Hash, limit int) ([]*Object, []*Bucket, error)
	// ListGvgByPrimarySpID list gvg by primary sp id
	ListGvgByPrimarySpID(spID uint32) ([]*GlobalVirtualGroup, error)
	// ListGvgBySecondarySpID list gvg by secondary sp id
	ListGvgBySecondarySpID(spID uint32) ([]*GlobalVirtualGroup, error)
	// ListGvgByBucketID list global virtual group by bucket id
	ListGvgByBucketID(bucketID common.Hash) ([]*GlobalVirtualGroup, error)
	// ListVgfByGvgID list vgf by gvg ids
	ListVgfByGvgID(gvgIDs []uint32) ([]*GlobalVirtualGroupFamily, error)
	// ListLvgByGvgAndBucketID list lvg by gvg and bucket ids
	ListLvgByGvgAndBucketID(bucketID common.Hash, gvgIDs []uint32) ([]*LocalVirtualGroup, error)
	// ListLvgByGvgID list lvg by gvg ids
	ListLvgByGvgID(gvgIDs []uint32) ([]*LocalVirtualGroup, error)
	// ListBucketsByVgfID list buckets by vgf ids
	ListBucketsByVgfID(vgfIDs []uint32, startAfter common.Hash, limit int) ([]*Bucket, error)
	// ListObjectsByLVGID list objects by lvg id
	ListObjectsByLVGID(lvgIDs []uint32, bucketID common.Hash, startAfter common.Hash, limit int, filters ...func(*gorm.DB) *gorm.DB) ([]*Object, *Bucket, error)
	// GetGvgByBucketAndLvgID get global virtual group by lvg id and bucket id
	GetGvgByBucketAndLvgID(bucketID common.Hash, lvgID uint32) (*GlobalVirtualGroup, error)
	// GetLvgByBucketAndLvgID get global virtual group by lvg id and bucket id
	GetLvgByBucketAndLvgID(bucketID common.Hash, lvgID uint32) (*LocalVirtualGroup, error)
	// ListMigrateBucketEvents list migrate bucket events
	ListMigrateBucketEvents(spID uint32, filters ...func(*gorm.DB) *gorm.DB) ([]*EventMigrationBucket, []*EventCompleteMigrationBucket, []*EventCancelMigrationBucket, []*EventRejectMigrateBucket, error)
	// ListCompleteMigrationBucket list complete migrate bucket events
	ListCompleteMigrationBucket(srcSpID uint32, filters ...func(*gorm.DB) *gorm.DB) ([]*EventCompleteMigrationBucket, error)
	// GetMigrateBucketEventByBucketID get migrate bucket event by bucket id
	GetMigrateBucketEventByBucketID(bucketID common.Hash) (*EventCompleteMigrationBucket, error)
	// ListSwapOutEvents list swap out events
	ListSwapOutEvents(blockID uint64, spID uint32) ([]*EventSwapOut, []*EventCompleteSwapOut, []*EventCancelSwapOut, error)
	// ListSpExitEvents list sp exit events
	ListSpExitEvents(blockID uint64, spID uint32) (*EventStorageProviderExit, *EventCompleteStorageProviderExit, error)
	// GetSPByAddress get sp info by operator address
	GetSPByAddress(operatorAddress common.Address) (*StorageProvider, error)
	// GetMysqlVersion get currently running mysql version of block syncer
	GetMysqlVersion() (string, error)
	// GetEpoch get current epoch of block syncer
	GetEpoch() (*Epoch, error)
	// GetSpVersion get the system version info of SP
	GetSpVersion() *SpVersion
	// GetDefaultCharacterSet get the current mysql default character set
	GetDefaultCharacterSet() (string, error)
	// GetDefaultCollationName get the current mysql default collation name
	GetDefaultCollationName() (string, error)
	// GetUserGroups get groups info by a user address
	GetUserGroups(accountID common.Address, startAfter common.Hash, limit int) ([]*GroupMemberMeta, error)
	// GetGroupMembers get group members by group id
	GetGroupMembers(groupID common.Hash, startAfter common.Address, limit int) ([]*GroupMemberMeta, error)
	// GetUserOwnedGroups retrieve groups where the user is the owner
	GetUserOwnedGroups(accountID common.Address, startAfter common.Hash, limit int) ([]*Group, error)
	// ListObjectPolicies list policies by object info
	ListObjectPolicies(objectID common.Hash, actionType types.ActionType, startAfter common.Hash, limit int) ([]*PermissionWithStatement, error)
	// GetGroupMembersCount get the count of group members
	GetGroupMembersCount(groupIDs []common.Hash) ([]*GroupCount, error)
	// ListVirtualGroupFamiliesByVgfIDs list virtual group families by vgf ids
	ListVirtualGroupFamiliesByVgfIDs(vgfIDs []uint32) ([]*GlobalVirtualGroupFamily, error)
	// ListUserPaymentAccounts list payment accounts by owner address
	ListUserPaymentAccounts(accountID common.Address) ([]*StreamRecordPaymentAccount, error)
	// ListPaymentAccountStreams list payment account streams
	ListPaymentAccountStreams(paymentAccount common.Address) ([]*Bucket, error)
	// ListGroupsByIDs list groups by ids
	ListGroupsByIDs(ids []common.Hash) ([]*Group, error)
	//GetEventMigrationBucketByBucketID get migration bucket event by bucket id
	GetEventMigrationBucketByBucketID(bucketID common.Hash) (*EventMigrationBucket, error)
	//GetEventSwapOutByGvgID get swap out event by gvg id
	GetEventSwapOutByGvgID(gvgID uint32) (*EventSwapOut, error)
	// GetPrimarySPStreamRecordBySpID return primary SP's stream records
	GetPrimarySPStreamRecordBySpID(spID uint32) ([]*PrimarySpIncomeMeta, error)
	// GetSecondarySPStreamRecordBySpID return secondary SP's stream records
	GetSecondarySPStreamRecordBySpID(spID uint32) ([]*SecondarySpIncomeMeta, error)
	// GetBucketSizeByID get bucket size info by a bucket id
	GetBucketSizeByID(bucketID uint64) (decimal.Decimal, error)
	// GetDataMigrationRecordByProcessKey  get the record of data migration by the given process key
	GetDataMigrationRecordByProcessKey(processKey string) (*DataMigrationRecord, error)
}

// BSDB contains all the methods required by block syncer database
type BSDB interface {
	Metadata
}

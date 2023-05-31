package bsdb

import (
	"github.com/forbole/juno/v4/common"
)

// Metadata contains all the methods required by block syncer db database
type Metadata interface {
	// GetUserBuckets get buckets info by a user address
	GetUserBuckets(accountID common.Address) ([]*Bucket, error)
	// GetUserBucketsCount get buckets count by a user address
	GetUserBucketsCount(accountID common.Address) (int64, error)
	// GetBucketByName get buckets info by a bucket name
	GetBucketByName(bucketName string, isFullList bool) (*Bucket, error)
	// GetBucketByID get buckets info by by a bucket id
	GetBucketByID(bucketID int64, isFullList bool) (*Bucket, error)
	// GetLatestBlockNumber get current latest block number
	GetLatestBlockNumber() (int64, error)
	// GetPaymentByBucketName get bucket payment info by a bucket name
	GetPaymentByBucketName(bucketName string, isFullList bool) (*StreamRecord, error)
	// GetPaymentByBucketID get bucket payment info by a bucket id
	GetPaymentByBucketID(bucketID int64, isFullList bool) (*StreamRecord, error)
	// GetPaymentByPaymentAddress get bucket payment info by a payment address
	GetPaymentByPaymentAddress(address common.Address) (*StreamRecord, error)
	// GetPermissionByResourceAndPrincipal get permission info by resource type & id, principal type & value
	GetPermissionByResourceAndPrincipal(resourceType, principalType, principalValue string, resourceID common.Hash) (*Permission, error)
	// GetStatementsByPolicyID get statements info by a policy id
	GetStatementsByPolicyID(policyIDList []common.Hash) ([]*Statement, error)
	// GetPermissionsByResourceAndPrincipleType get permissions info by resource type & id, principal type
	GetPermissionsByResourceAndPrincipleType(resourceType, principalType string, resourceID common.Hash) ([]*Permission, error)
	// GetGroupsByGroupIDAndAccount get groups info by group id list and account id
	GetGroupsByGroupIDAndAccount(groupIDList []common.Hash, account common.Address) ([]*Group, error)
	// ListObjectsByBucketName list objects info by a bucket name
	ListObjectsByBucketName(bucketName, continuationToken, prefix, delimiter string, maxKeys int) ([]*ListObjectsResult, error)
	// ListDeletedObjectsByBlockNumberRange list deleted objects info by a block number range
	ListDeletedObjectsByBlockNumberRange(startBlockNumber int64, endBlockNumber int64, isFullList bool) ([]*Object, error)
	// ListExpiredBucketsBySp list expired buckets by sp
	ListExpiredBucketsBySp(createAt int64, primarySpAddress string, limit int64) ([]*Bucket, error)
	// GetObjectByName get object info by an object name
	GetObjectByName(objectName string, bucketName string, isFullList bool) (*Object, error)
	// GetSwitchDBSignal check if there is a signal to switch the database
	GetSwitchDBSignal() (*MasterDB, error)
	// GetBucketMetaByName get bucket info with its related info
	GetBucketMetaByName(bucketName string, isFullList bool) (*BucketFullMeta, error)
}

// BSDB contains all the methods required by block syncer database
type BSDB interface {
	Metadata
}

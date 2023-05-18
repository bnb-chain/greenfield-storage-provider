package bsdb

import (
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	"github.com/ethereum/go-ethereum/common"
)

// MetadataDB contains all the methods required by block syncer db database
type MetadataDB interface {
	// GetUserBuckets get buckets info by a user address
	GetUserBuckets(accountID common.Address) ([]*bsdb.Bucket, error)
	// GetUserBucketsCount get buckets count by a user address
	GetUserBucketsCount(accountID common.Address) (int64, error)
	// GetBucketByName get buckets info by a bucket name
	GetBucketByName(bucketName string, isFullList bool) (*bsdb.Bucket, error)
	// GetBucketByID get buckets info by by a bucket id
	GetBucketByID(bucketID int64, isFullList bool) (*bsdb.Bucket, error)
	// GetLatestBlockNumber get current latest block number
	GetLatestBlockNumber() (int64, error)
	// GetPaymentByBucketName get bucket payment info by a bucket name
	GetPaymentByBucketName(bucketName string, isFullList bool) (*bsdb.StreamRecord, error)
	// GetPaymentByBucketID get bucket payment info by a bucket id
	GetPaymentByBucketID(bucketID int64, isFullList bool) (*bsdb.StreamRecord, error)
	// GetPaymentByPaymentAddress get bucket payment info by a payment address
	GetPaymentByPaymentAddress(address common.Address) (*bsdb.StreamRecord, error)
	// GetPermissionByResourceAndPrincipal get permission info by resource type & id, principal type & value
	GetPermissionByResourceAndPrincipal(resourceType, resourceID, principalType, principalValue string) (*bsdb.Permission, error)
	// GetStatementsByPolicyID get statements info by a policy id
	GetStatementsByPolicyID(policyIDList []common.Hash) ([]*bsdb.Statement, error)
	// GetPermissionsByResourceAndPrincipleType get permissions info by resource type & id, principal type
	GetPermissionsByResourceAndPrincipleType(resourceType, resourceID, principalType string) ([]*bsdb.Permission, error)
	// GetGroupsByGroupIDAndAccount get groups info by group id list and account id
	GetGroupsByGroupIDAndAccount(groupIDList []common.Hash, account common.Hash) ([]*bsdb.Group, error)
	// ListObjectsByBucketName list objects info by a bucket name
	ListObjectsByBucketName(bucketName, continuationToken, prefix, delimiter string, maxKeys int) ([]*bsdb.ListObjectsResult, error)
	// ListDeletedObjectsByBlockNumberRange list deleted objects info by a block number range
	ListDeletedObjectsByBlockNumberRange(startBlockNumber int64, endBlockNumber int64, isFullList bool) ([]*bsdb.Object, error)
	// ListExpiredBucketsBySp list expired buckets by sp
	ListExpiredBucketsBySp(createAt int64, primarySpAddress string, limit int64) ([]*bsdb.Bucket, error)
	// GetObjectByName get object info by an object name
	GetObjectByName(objectName string, bucketName string, isFullList bool) (*bsdb.Object, error)
	// GetSwitchDBSignal check if there is a signal to switch the database
	GetSwitchDBSignal() (*bsdb.MasterDB, error)
	// GetBucketMetaByName get bucket info with its related info
	GetBucketMetaByName(bucketName string, isFullList bool) (*bsdb.BucketFullMeta, error)
}

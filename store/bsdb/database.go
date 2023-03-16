package bsdb

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/bnb-chain/greenfield-storage-provider/model/metadata"
)

// Metadata contains all the methods required by block syncer db database
type Metadata interface {
	// GetUserBuckets get buckets info by a user address
	GetUserBuckets(accountID common.Address) ([]*metadata.Bucket, error)
	// GetUserBucketsCount get buckets count by a user address
	GetUserBucketsCount(accountID common.Address) (int64, error)
	// GetBucketByName get buckets info by a bucket name
	GetBucketByName(bucketName string, isFullList bool) (*metadata.Bucket, error)
	// GetBucketByID get buckets info by by a bucket id
	GetBucketByID(bucketID int64, isFullList bool) (*metadata.Bucket, error)
	// GetLatestBlockNumber get current latest block number
	GetLatestBlockNumber() (int64, error)
	// ListObjectsByBucketName list objects info by a bucket name
	ListObjectsByBucketName(bucketName string) ([]*metadata.Object, error)
	// ListDeletedObjectsByBlockNumberRange list deleted objects info by a block number range
	ListDeletedObjectsByBlockNumberRange(startBlockNumber int64, endBlockNumber int64, isFullList bool) ([]*metadata.Object, error)
}

// BSDB contains all the methods required by block syncer database
type BSDB interface {
	Metadata
}

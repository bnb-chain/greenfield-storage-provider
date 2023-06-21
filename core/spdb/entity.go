package spdb

import (
	"time"

	storetypes "github.com/bnb-chain/greenfield-storage-provider/store/types"
)

// SpAddressType identify address type of SP.
type SpAddressType int32

const (
	OperatorAddressType SpAddressType = iota + 1
	FundingAddressType
	SealAddressType
	ApprovalAddressType
)

// UploadObjectMeta defines the upload object state and related seal info, etc.
type UploadObjectMeta struct {
	ObjectID            uint64
	TaskState           storetypes.TaskState
	SecondaryAddresses  []string
	SecondarySignatures [][]byte
	ErrorDescription    string
}

// GCObjectMeta defines the gc object range progress info.
type GCObjectMeta struct {
	TaskKey             string
	StartBlockHeight    uint64
	EndBlockHeight      uint64
	CurrentBlockHeight  uint64
	LastDeletedObjectID uint64
}

// IntegrityMeta defines the payload integrity hash and piece checksum with objectID.
type IntegrityMeta struct {
	ObjectID          uint64
	IntegrityChecksum []byte
	PieceChecksumList [][]byte
}

// ReadRecord defines a read request record, will decrease the bucket read quota.
type ReadRecord struct {
	BucketID        uint64
	ObjectID        uint64
	UserAddress     string
	BucketName      string
	ObjectName      string
	ReadSize        uint64
	ReadTimestampUs int64
}

// BucketQuota defines read quota of a bucket.
type BucketQuota struct {
	ReadQuotaSize uint64
}

// BucketTraffic is record traffic by year and month.
type BucketTraffic struct {
	BucketID         uint64
	YearMonth        string // YearMonth is traffic's YearMonth, format "2023-02".
	BucketName       string
	ReadConsumedSize uint64
	ReadQuotaSize    uint64
	ModifyTime       int64
}

// TrafficTimeRange is used by query, return records in [StartTimestampUs, EndTimestampUs).
type TrafficTimeRange struct {
	StartTimestampUs int64
	EndTimestampUs   int64
	LimitNum         int // is unlimited if LimitNum <= 0.
}

type OffChainAuthKey struct {
	UserAddress string
	Domain      string

	CurrentNonce     int32
	CurrentPublicKey string
	NextNonce        int32
	ExpiryDate       time.Time

	CreatedTime  time.Time
	ModifiedTime time.Time
}

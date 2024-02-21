package spdb

import (
	"time"

	storetypes "github.com/bnb-chain/greenfield-storage-provider/store/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
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
	ObjectID              uint64
	TaskState             storetypes.TaskState
	GlobalVirtualGroupID  uint32
	SecondaryEndpoints    []string
	SecondarySignatures   [][]byte
	ErrorDescription      string
	CreateTimeStampSecond int64
}

// GCObjectMeta defines the gc object range progress info.
type GCObjectMeta struct {
	TaskKey             string
	StartBlockHeight    uint64
	EndBlockHeight      uint64
	CurrentBlockHeight  uint64
	LastDeletedObjectID uint64
}

// GCPieceMeta defines the gc piece with segment index and piece checksum.
type GCPieceMeta struct {
	ObjectID        uint64
	SegmentIndex    uint32
	RedundancyIndex int32
	PieceChecksum   string
	Version         int64 // the version is used to locate the correct piece in Piece Store.  Default is 0.
}

// IntegrityMeta defines the payload integrity hash and piece checksum with objectID.
type IntegrityMeta struct {
	ObjectID          uint64
	RedundancyIndex   int32
	IntegrityChecksum []byte
	PieceChecksumList [][]byte
}

// ShadowIntegrityMeta defines the payload integrity hash and piece checksum with objectID. It is used for storing user's
// updated object meta. When the SP seals the object successful, this temporary record will be persisted to IntegrityMeta
// and removed from ShadowIntegrityMeta.
type ShadowIntegrityMeta struct {
	ObjectID          uint64
	RedundancyIndex   int32
	IntegrityChecksum []byte
	PieceChecksumList [][]byte
	Version           int64
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
	ChargedQuotaSize uint64 // the charged quota of bucket on greenfield chain meta
	FreeQuotaSize    uint64 // the free quota of SP on greenfield chain
}

// BucketFreeQuota defines free quota of a bucket.
type BucketFreeQuota struct {
	QuotaSize uint64
}

// BucketTraffic is record traffic by year and month.
type BucketTraffic struct {
	BucketID              uint64
	YearMonth             string
	BucketName            string
	ReadConsumedSize      uint64
	FreeQuotaSize         uint64 // the total free quota size of SP price meta
	FreeQuotaConsumedSize uint64 // the consumed free quota size
	ChargedQuotaSize      uint64 // the total charged quota of bucket
	ModifyTime            int64
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

// MigrateGVGUnitMeta is used to record migrate type/meta/status/progress.
type MigrateGVGUnitMeta struct {
	MigrateGVGKey            string // as primary key
	SwapOutKey               string
	GlobalVirtualGroupID     uint32 // is used by sp exit/bucket migrate
	DestGlobalVirtualGroupID uint32 // is used by bucket migrate
	VirtualGroupFamilyID     uint32 // is used by sp exit
	RedundancyIndex          int32  // is used by sp exit
	BucketID                 uint64 // is used by bucket migrate
	SrcSPID                  uint32
	DestSPID                 uint32
	LastMigratedObjectID     uint64
	MigrateStatus            int    // scheduler assign unit status.
	RetryTime                int    //
	MigratedBytesSize        uint64 // migrated bytes
}

// SwapOutMeta is used to record swap out meta.
type SwapOutMeta struct {
	SwapOutKey    string // as primary key
	IsDestSP      bool
	SwapOutMsg    *virtualgrouptypes.MsgSwapOut
	CompletedGVGs []uint32
}

type RecoverStatus int

const (
	Processing RecoverStatus = 0
	Processed  RecoverStatus = 1
	Completed  RecoverStatus = 2
)

type RecoverGVGStats struct {
	VirtualGroupID       uint32
	VirtualGroupFamilyID uint32
	RedundancyIndex      int32
	StartAfter           uint64
	NextStartAfter       uint64
	Limit                uint64
	Status               RecoverStatus
	ObjectCount          uint64
}

type RecoverFailedObject struct {
	ObjectID        uint64
	VirtualGroupID  uint32
	RedundancyIndex int32
	RetryTime       int
}

// MigrateBucketProgressMeta is used to record migrate bucket progress meta.
type MigrateBucketProgressMeta struct {
	BucketID              uint64 // as primary key
	SubscribedBlockHeight uint64
	MigrateState          int

	TotalGvgNum            uint32 // Total number of GVGs that need to be migrated
	MigratedFinishedGvgNum uint32 // Number of successfully migrated GVGs
	GcFinishedGvgNum       uint32 // Number of successfully gc finished GVGs

	PreDeductedQuota uint64 // Quota pre-deducted by the source sp in the pre-migrate bucket phase
	RecoupQuota      uint64 // In case of migration failure, the dest sp recoup the quota for the source sp

	LastGcObjectID uint64 // After bucket migration is complete, the progress of GC, up to which object is GC performed.
	LastGcGvgID    uint64 // which GVG is GC performed.
}

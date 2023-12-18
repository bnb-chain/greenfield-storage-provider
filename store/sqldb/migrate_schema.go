package sqldb

// MigrateSubscribeProgressTable table schema.
type MigrateSubscribeProgressTable struct {
	EventName                 string `gorm:"primary_key"`
	LastSubscribedBlockHeight uint64
}

// TableName is used to set MigrateSubscribeEventProgressTable Schema's table name in database.
func (MigrateSubscribeProgressTable) TableName() string {
	return MigrateSubscribeProgressTableName
}

// SwapOutTable table schema.
type SwapOutTable struct {
	SwapOutKey       string `gorm:"primary_key"`
	IsDestSP         bool   `gorm:"primary_key"`
	SwapOutMsg       string
	CompletedGVGList string
}

func (SwapOutTable) TableName() string {
	return SwapOutTableName
}

// MigrateGVGTable table schema.
// sp exit, bucket migrate
type MigrateGVGTable struct {
	MigrateKey               string `gorm:"primary_key"`
	SwapOutKey               string `gorm:"index:swap_out_index"`
	GlobalVirtualGroupID     uint32 `gorm:"index:gvg_index"`        // is used by sp exit/bucket migrate
	DestGlobalVirtualGroupID uint32 `gorm:"index:dest_gvg_index"`   // is used by bucket migrate
	VirtualGroupFamilyID     uint32 `gorm:"index:vgf_index"`        // is used by sp exit
	BucketID                 uint64 `gorm:"index:bucket_index"`     // is used by bucket migrate
	RedundancyIndex          int32  `gorm:"index:redundancy_index"` // is used by sp exit
	SrcSPID                  uint32
	DestSPID                 uint32
	LastMigratedObjectID     uint64
	MigrateStatus            int `gorm:"index:migrate_status_index"`
	RetryTime                int `gorm:"comment:retry_time"`
	MigratedBytesSize        uint64
}

// TableName is used to set MigrateGVGTable Schema's table name in database.
func (MigrateGVGTable) TableName() string {
	return MigrateGVGTableName
}

// MigrateBucketProgressTable table schema.
// used by persist bucket migration progress and meta
type MigrateBucketProgressTable struct {
	BucketID              uint64 `gorm:"primary_key"`
	SubscribedBlockHeight uint64 `gorm:"primary_key"`
	MigrationState        int    `gorm:"migration_state"`

	TotalGvgNum            uint32 // Total number of GVGs that need to be migrated
	MigratedFinishedGvgNum uint32 // Number of successfully migrated GVGs
	GcFinishedGvgNum       uint32 // Number of successfully gc finished GVGs

	PreDeductedQuota uint64 // Quota pre-deducted by the source sp in the pre-migrate bucket phase
	RecoupQuota      uint64 // In case of migration failure, the dest sp recoup the quota for the source sp

	LastGcObjectID uint64 // After bucket migration is complete, the progress of GC, up to which object is GC performed.
	LastGcGvgID    uint64 // which GVG is GC performed.
}

// TableName is used to set MigrateBucketProgressTable Schema's table name in database.
func (MigrateBucketProgressTable) TableName() string {
	return MigrateBucketProgressTableName
}

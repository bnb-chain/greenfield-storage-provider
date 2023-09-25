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
	RetryTime                int
}

// TableName is used to set MigrateGVGTable Schema's table name in database.
func (MigrateGVGTable) TableName() string {
	return MigrateGVGTableName
}

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

// MigrateGVGTable table schema.
// sp exit, bucket migrate
type MigrateGVGTable struct {
	AutoID                 uint64 `gorm:"primary_key;autoIncrement"`
	GlobalVirtualGroupID   uint32 `gorm:"index:gvg_index"`
	VirtualGroupFamilyID   uint32 `gorm:"index:vgf_index"`
	MigrateRedundancyIndex int32
	BucketID               uint64 `gorm:"index:bucket_index"`
	IsSecondaryGVG         bool   `gorm:"index:secondary_index"`
	SrcSPID                uint32
	DestSPID               uint32
	LastMigrateObjectID    uint64
	MigrateStatus          int
	CheckStatus            int
}

// TableName is used to set MigrateGVGTable Schema's table name in database.
func (MigrateGVGTable) TableName() string {
	return MigrateGVGTableName
}

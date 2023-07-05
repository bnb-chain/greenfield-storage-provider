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
	MigrateKey           string `gorm:"primary_key"`
	GlobalVirtualGroupID uint32 `gorm:"index:gvg_index"`        // is used by sp exit/bucket migrate
	VirtualGroupFamilyID uint32 `gorm:"index:vgf_index"`        // is used by sp exit
	BucketID             uint64 `gorm:"index:bucket_index"`     // is used by bucket migrate
	IsSecondary          bool   `gorm:"index:secondary_index"`  // is used by sp exit
	IsConflicted         bool   `gorm:"index:conflicted_index"` // is used by sp exit
	IsRemoted            bool   `gorm:"index:remoted_index"`    // is used by sp exit
	RedundancyIndex      int32  `gorm:"index:redundancy_index"` // is used by sp exit
	SrcSPID              uint32
	DestSPID             uint32
	LastMigrateObjectID  uint64
	MigrateStatus        int
	CheckStatus          int
}

// TableName is used to set MigrateGVGTable Schema's table name in database.
func (MigrateGVGTable) TableName() string {
	return MigrateGVGTableName
}

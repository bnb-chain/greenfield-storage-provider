package sqldb

// RecoverGVGStatsTable is for a successor primary SP that keep track of data recover in a virtual group family,
type RecoverGVGStatsTable struct {
	VirtualGroupID       uint32 `gorm:"primary_key"`
	VirtualGroupFamilyID uint32 // only need to record it when SP is the successor Primary SP
	ExitingSPID          uint32
	RedundancyIndex      int32 `gorm:"index:redundancy_index"`
	StartAfterObjectID   uint64
	Limit                uint32
	Status               int //  NotDone, Processed, Completed
}

// TableName is used to set MigrateGVGTable Schema's table name in database.
func (RecoverGVGStatsTable) TableName() string {
	return RecoverGVGStatsTableName
}

type RecoverFailedObjectTable struct {
	ObjectID        uint64 `gorm:"primary_key"`
	VirtualGroupID  uint32
	RedundancyIndex int32 `gorm:"index:redundancy_index"`
	Retry           int
}

// TableName is used to set MigrateGVGTable Schema's table name in database.
func (RecoverFailedObjectTable) TableName() string {
	return RecoverFailedObjectTableName
}

package sqldb

type RecoverGVGStatsTable struct {
	VirtualGroupID       uint32 `gorm:"primary_key"`
	VirtualGroupFamilyID uint32
	RedundancyIndex      int32 `gorm:"index:redundancy_index"`
	StartAfter           uint64
	Limit                uint32
	Status               int
	ObjectCount          uint64
}

func (RecoverGVGStatsTable) TableName() string {
	return RecoverGVGStatsTableName
}

type RecoverFailedObjectTable struct {
	ObjectID        uint64 `gorm:"primary_key"`
	VirtualGroupID  uint32 `gorm:"index:idx_gvg"`
	RedundancyIndex int32
	Retry           int `gorm:"index:retry"`
}

func (RecoverFailedObjectTable) TableName() string {
	return RecoverFailedObjectTableName
}

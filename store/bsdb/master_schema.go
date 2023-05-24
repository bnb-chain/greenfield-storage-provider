package bsdb

// MasterDB stores current master DB
type MasterDB struct {
	OneRowId bool `gorm:"column:one_row_id;not null;primaryKey"`
	// IsMaster defines if current DB is master DB
	IsMaster bool `gorm:"column:is_master;not null;"`
}

// TableName is used to set Master table name in database
func (m *MasterDB) TableName() string {
	return MasterDBTableName
}

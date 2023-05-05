package bsdb

// MasterDB stores current master DB
type MasterDB struct {
	OneRowId bool `gorm:"one_row_id"`
	// IsMaster defines if current DB is master DB
	IsMaster bool `gorm:"column:is_master"`
}

// TableName is used to set Master table name in database
func (m *MasterDB) TableName() string {
	return "master_db"
}

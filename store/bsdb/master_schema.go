package bsdb

// Master stores current master DB
type Master struct {
	// IsMaster defines if current DB is master DB
	IsMaster bool `gorm:"column:is_master;"`
}

// TableName is used to set Epoch table name in database
func (e *Master) TableName() string {
	return "master"
}

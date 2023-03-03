package sqldb

// ServiceConfigTable table schema
type ServiceConfigTable struct {
	ConfigVersion string `gorm:"primary_key"`
	ServiceConfig string
}

// TableName is used to set ServiceConfigTable Schema's table name in database
func (ServiceConfigTable) TableName() string {
	return JobTableName
}

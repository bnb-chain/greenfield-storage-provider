package sqldb

// PutObjectSuccessTable table schema.
type PutObjectSuccessTable struct {
	ID         uint64 `gorm:"primary_key;autoIncrement"`
	UpdateTime string `gorm:"index:update_time_index"`
	ObjectID   uint64 `gorm:"index:object_id_index"`
	Bucket     string `gorm:"index:bucket_index"`
	Object     string `gorm:"index:object_index"`
	State      string
	Error      string
	Logs       string
}

// TableName is used to set UploadObjectProgressTable Schema's table name in database.
func (PutObjectSuccessTable) TableName() string {
	return PutObjectSuccessTableName
}

// PutObjectEventTable table schema.
type PutObjectEventTable struct {
	ID         uint64 `gorm:"primary_key;autoIncrement"`
	UpdateTime string `gorm:"index:update_time_index"`
	ObjectID   uint64 `gorm:"index:object_id_index"`
	Bucket     string `gorm:"index:bucket_index"`
	Object     string `gorm:"index:object_index"`
	State      string
	Error      string
	Logs       string
}

// TableName is used to set UploadObjectProgressTable Schema's table name in database.
func (PutObjectEventTable) TableName() string {
	return PutObjectEventTableName
}

// UploadTimeoutTable table schema.
type UploadTimeoutTable struct {
	ID         uint64 `gorm:"primary_key;autoIncrement"`
	UpdateTime string `gorm:"index:update_time_index"`
	ObjectID   uint64 `gorm:"index:object_id_index"`
	Bucket     string `gorm:"index:bucket_index"`
	Object     string `gorm:"index:object_index"`
	Error      string
	Logs       string
}

// TableName is used to set UploadTimeoutTable Schema's table name in database.
func (UploadTimeoutTable) TableName() string {
	return UploadTimeoutTableName
}

// ReplicateTimeoutTable table schema.
type ReplicateTimeoutTable struct {
	ID         uint64 `gorm:"primary_key;autoIncrement"`
	UpdateTime string `gorm:"index:update_time_index"`
	ObjectID   uint64 `gorm:"index:object_id_index"`
	Bucket     string `gorm:"index:bucket_index"`
	Object     string `gorm:"index:object_index"`
	Error      string
	Logs       string
}

// TableName is used to set ReplicateTimeoutTable Schema's table name in database.
func (ReplicateTimeoutTable) TableName() string {
	return ReplicateTimeoutTableName
}

// SealTimeoutTable table schema.
type SealTimeoutTable struct {
	ID         uint64 `gorm:"primary_key;autoIncrement"`
	UpdateTime string `gorm:"index:update_time_index"`
	ObjectID   uint64 `gorm:"index:object_id_index"`
	Bucket     string `gorm:"index:bucket_index"`
	Object     string `gorm:"index:object_index"`
	Error      string
	Logs       string
}

// TableName is used to set SealTimeoutTable Schema's table name in database.
func (SealTimeoutTable) TableName() string {
	return SealTimeoutTableName
}

// UploadFailedTable table schema.
type UploadFailedTable struct {
	ID         uint64 `gorm:"primary_key;autoIncrement"`
	UpdateTime string `gorm:"index:update_time_index"`
	ObjectID   uint64 `gorm:"index:object_id_index"`
	Bucket     string `gorm:"index:bucket_index"`
	Object     string `gorm:"index:object_index"`
	Error      string
	Logs       string
}

// TableName is used to set UploadTimeoutTable Schema's table name in database.
func (UploadFailedTable) TableName() string {
	return UploadFailedTableName
}

// ReplicateFailedTable table schema.
type ReplicateFailedTable struct {
	ID         uint64 `gorm:"primary_key;autoIncrement"`
	UpdateTime string `gorm:"index:update_time_index"`
	ObjectID   uint64 `gorm:"index:object_id_index"`
	Bucket     string `gorm:"index:bucket_index"`
	Object     string `gorm:"index:object_index"`
	Error      string
	Logs       string
}

// TableName is used to set ReplicateTimeoutTable Schema's table name in database.
func (ReplicateFailedTable) TableName() string {
	return ReplicateFailedTableName
}

// SealFailedTable table schema.
type SealFailedTable struct {
	ID         uint64 `gorm:"primary_key;autoIncrement"`
	UpdateTime string `gorm:"index:update_time_index"`
	ObjectID   uint64 `gorm:"index:object_id_index"`
	Bucket     string `gorm:"index:bucket_index"`
	Object     string `gorm:"index:object_index"`
	Error      string
	Logs       string
}

// TableName is used to set SealTimeoutTable Schema's table name in database.
func (SealFailedTable) TableName() string {
	return SealFailedTableName
}

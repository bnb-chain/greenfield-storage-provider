package sqldb

// SPInfoTable table schema
type SPInfoTable struct {
	OperatorAddress string `gorm:"primary_key"`
	IsOwn           bool   `gorm:"primary_key"`
	FundingAddress  string
	SealAddress     string
	ApprovalAddress string
	TotalDeposit    int64
	Status          int32
	Endpoint        string
	Moniker         string
	Identity        string
	Website         string
	SecurityContact string
	Details         string
}

// TableName is used to set StorageProviderInfoTable Schema's table name in database
func (SPInfoTable) TableName() string {
	return SPInfoTableName
}

// StorageParamsTable table schema
type StorageParamsTable struct {
	ID                      int64 `gorm:"primary_key;autoIncrement"`
	MaxSegmentSize          uint64
	RedundantDataChunkNum   uint32
	RedundantParityChunkNum uint32
	MaxPayloadSize          uint64
}

// TableName is used to set StorageParamsTable Schema's table name in database
func (StorageParamsTable) TableName() string {
	return StorageParamsTableName
}

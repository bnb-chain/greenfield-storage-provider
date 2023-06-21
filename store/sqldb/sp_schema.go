package sqldb

// SpInfoTable table schema
type SpInfoTable struct {
	OperatorAddress string `gorm:"primary_key"`
	IsOwn           bool   `gorm:"primary_key"`
	ID              uint32
	FundingAddress  string
	SealAddress     string
	ApprovalAddress string
	TotalDeposit    string
	Status          int32
	Endpoint        string
	Moniker         string
	Identity        string
	Website         string
	SecurityContact string
	Details         string
}

// TableName is used to set SpInfoTable Schema's table name in database
func (SpInfoTable) TableName() string {
	return SpInfoTableName
}

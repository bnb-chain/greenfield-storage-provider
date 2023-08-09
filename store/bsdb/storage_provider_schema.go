package bsdb

import (
	"github.com/forbole/juno/v4/common"
)

type StorageProvider struct {
	ID uint64 `gorm:"column:id;primaryKey"`

	SpId            uint32         `gorm:"column:sp_id;index:idx_sp_id"`
	OperatorAddress common.Address `gorm:"column:operator_address;type:BINARY(20);uniqueIndex:idx_operator_address"`
	FundingAddress  common.Address `gorm:"column:funding_address;type:BINARY(20)"`
	SealAddress     common.Address `gorm:"column:seal_address;;type:BINARY(20)"`
	ApprovalAddress common.Address `gorm:"column:approval_address;;type:BINARY(20)"`
	GcAddress       common.Address `gorm:"column:gc_address;type:BINARY(20)"`
	TotalDeposit    *common.Big    `gorm:"column:total_deposit"`
	Status          string         `gorm:"column:status;type:VARCHAR(50)"`
	Endpoint        string         `gorm:"column:endpoint;type:VARCHAR(256)"`
	Moniker         string         `gorm:"column:moniker;type:VARCHAR(128)"`
	Identity        string         `gorm:"column:identity;type:VARCHAR(256)"`
	Website         string         `gorm:"column:website;type:VARCHAR(128)"`
	SecurityContact string         `gorm:"column:security_contact;type:VARCHAR(128)"`
	Details         string         `gorm:"column:details;type:VARCHAR(256)"`
	BlsKey          string         `gorm:"column:bls_key;type:VARCHAR(64)"`

	UpdateTimeSec int64       `gorm:"column:update_time_sec"`
	ReadPrice     *common.Big `gorm:"column:read_price"`
	FreeReadQuota uint64      `gorm:"column:free_read_quota"`
	StorePrice    *common.Big `gorm:"column:store_price"`

	CreateAt     int64       `gorm:"column:create_at"`
	CreateTxHash common.Hash `gorm:"column:create_tx_hash;type:BINARY(32);not null"`
	UpdateAt     int64       `gorm:"column:update_at"`
	UpdateTxHash common.Hash `gorm:"column:update_tx_hash;type:BINARY(32);not null"`
	Removed      bool        `gorm:"column:removed;default:false"`
}

func (*StorageProvider) TableName() string {
	return StorageProviderTableName
}

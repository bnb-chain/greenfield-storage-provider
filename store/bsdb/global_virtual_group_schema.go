package bsdb

import (
	"github.com/forbole/juno/v4/common"
	"github.com/lib/pq"
)

// GlobalVirtualGroup is a global virtual group consists of one primary SP (SP) and multiple secondary SP.
type GlobalVirtualGroup struct {
	ID                    uint64         `gorm:"column:id;primaryKey"`
	GlobalVirtualGroupId  uint32         `gorm:"column:global_virtual_group_id;index:idx_gvg_id"`
	FamilyId              uint32         `gorm:"column:family_id"`
	PrimarySpId           uint32         `gorm:"column:primary_sp_id;index:idx_primary_sp_id"`
	SecondarySpIds        pq.StringArray `gorm:"column:secondary_sp_ids;type:TEXT"`
	StoredSize            uint64         `gorm:"column:stored_size"`
	VirtualPaymentAddress common.Address `gorm:"column:virtual_payment_address;type:BINARY(20)"`
	TotalDeposit          *common.Big    `gorm:"column:total_deposit"`

	CreateAt     int64       `gorm:"column:create_at"`
	CreateTxHash common.Hash `gorm:"column:create_tx_hash;type:BINARY(32);not null"`
	CreateTime   int64       `gorm:"column:create_time"` // seconds
	UpdateAt     int64       `gorm:"column:update_at"`
	UpdateTxHash common.Hash `gorm:"column:update_tx_hash;type:BINARY(32);not null"`
	UpdateTime   int64       `gorm:"column:update_time"` // seconds
	Removed      bool        `gorm:"column:removed;default:false"`
}

// TableName is used to set GlobalVirtualGroup table name in database
func (g *GlobalVirtualGroup) TableName() string {
	return GlobalVirtualGroupTableName
}

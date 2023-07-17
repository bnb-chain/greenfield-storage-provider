package bsdb

import "github.com/forbole/juno/v4/common"

type LocalVirtualGroup struct {
	ID                   uint64      `gorm:"column:id;primaryKey"`
	LocalVirtualGroupId  uint32      `gorm:"column:local_virtual_group_id;index:idx_lvg_id"`
	GlobalVirtualGroupId uint32      `gorm:"column:global_virtual_group_id;index:idx_gvg_id"`
	BucketID             common.Hash `gorm:"column:bucket_id;type:BINARY(32);index:idx_bucket_id"`
	StoredSize           uint64      `gorm:"column:stored_size"`

	CreateAt     int64       `gorm:"column:create_at"`
	CreateTxHash common.Hash `gorm:"column:create_tx_hash;type:BINARY(32);not null"`
	CreateTime   int64       `gorm:"column:create_time"` // seconds
	UpdateAt     int64       `gorm:"column:update_at"`
	UpdateTxHash common.Hash `gorm:"column:update_tx_hash;type:BINARY(32);not null"`
	UpdateTime   int64       `gorm:"column:update_time"` // seconds
	Removed      bool        `gorm:"column:removed;default:false"`
}

// TableName is used to set VirtualGroupFamily table name in database
func (g *LocalVirtualGroup) TableName() string {
	return LocalVirtualGroupFamilyTableName
}

package bsdb

import "github.com/forbole/juno/v4/common"

// Epoch stores current information of the latest block
type Epoch struct {
	// OneRowID defines if the table only has one row
	OneRowID bool `gorm:"one_row_id;not null;default:true;primaryKey"`
	// BlockHeight defines the latest block number
	BlockHeight int64 `gorm:"block_height;type:bigint(64)"`
	// BlockHash defines the latest block hash
	BlockHash common.Hash `gorm:"block_hash;type:BINARY(32)"`
	// UpdateTime defines the update time of the latest block
	UpdateTime int64 `gorm:"update_time;type:bigint(64)"`
}

// TableName is used to set Epoch table name in database
func (*Epoch) TableName() string {
	return EpochTableName
}

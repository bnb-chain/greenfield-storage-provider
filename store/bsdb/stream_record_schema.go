package bsdb

import (
	"github.com/forbole/juno/v4/common"
)

type StreamRecord struct {
	// ID defines db auto_increment id of stream record
	ID uint64 `gorm:"id"`
	// Account defines the account address
	Account common.Address `gorm:"account"`
	// CrudTimestamp defines the latest update timestamp of the stream record
	CrudTimestamp int64 `gorm:"crud_timestamp"`
	// NetflowRate defines the per-second rate that an account's balance is changing.
	// It is the sum of the account's inbound and outbound flow rates.
	NetflowRate *common.Big `gorm:"netflow_rate"`
	// StaticBalance defines the balance of the stream account at the latest CRUD timestamp.
	StaticBalance *common.Big `gorm:"static_balance"`
	// BufferBalance defines reserved balance of the stream account
	// If the netflow rate is negative, the reserved balance is `netflow_rate * reserve_time`
	BufferBalance *common.Big `gorm:"buffer_balance"`
	// LockBalance defines the locked balance of the stream account after it puts a new object and before the object is sealed
	LockBalance *common.Big `gorm:"lock_balance"`
	// Status defines the status of the stream account
	Status string `gorm:"status"`
	// SettleTimestamp defines the unix timestamp when the stream account will be settled
	SettleTimestamp int64 `gorm:"column:settle_timestamp"`
	// OutFlows defines the accumulated outflow rates of the stream account
	OutFlows []byte `gorm:"out_flows;type:longblob"`
}

// TableName is used to set StreamRecord table name in database
func (s *StreamRecord) TableName() string {
	return "stream_records"
}

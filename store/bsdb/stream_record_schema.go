package bsdb

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

type StreamRecord struct {
	// ID defines db auto_increment id of stream record
	ID uint64 `gorm:"id"`
	// Account defines the account address
	Account common.Address `gorm:"account"`
	// UpdateTime defines the latest update timestamp of the stream record
	UpdateTime int64 `gorm:"update_time"`
	// NetflowRate defines the per-second rate that an account's balance is changing.
	// It is the sum of the account's inbound and outbound flow rates.
	NetflowRate decimal.Decimal `gorm:"netflow_rate"`
	// StaticBalance defines the balance of the stream account at the latest CRUD timestamp.
	StaticBalance decimal.Decimal `gorm:"static_balance"`
	// BufferBalance defines reserved balance of the stream account
	// If the netflow rate is negative, the reserved balance is `netflow_rate * reserve_time`
	BufferBalance decimal.Decimal `gorm:"buffer_balance"`
	// LockBalance defines the locked balance of the stream account after it puts a new object and before the object is sealed
	LockBalance decimal.Decimal `gorm:"lock_balance"`
	// Status defines the status of the stream account
	Status string `gorm:"status"`
	// SettleTimestamp defines the unix timestamp when the stream account will be settled
	SettleTimestamp int64 `gorm:"column:settle_time"`
	// OutFlows defines the accumulated outflow rates of the stream account
	OutFlows OutFlows `gorm:"out_flows;type:json"`
}

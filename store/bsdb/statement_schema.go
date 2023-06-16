package bsdb

import (
	"github.com/forbole/juno/v4/common"
	"github.com/lib/pq"
)

// Statement defines the details content of the permission, include effect/actions/sub-resources
type Statement struct {
	// ID defines db auto_increment id of statement
	ID uint64 `gorm:"id"`
	// PolicyID is a unique u256 sequence for each policy. It also is used as NFT tokenID
	PolicyID common.Hash `gorm:"policy_id"`
	// Effect define the impact of permissions, which is Allow/Deny
	Effect string `gorm:"effect"`
	// ActionValue define the operation type you can act. greenfield defines a set of permission
	// that you can specify in a permissionInfo
	ActionValue int `gorm:"action_value"`
	// Resources CAN ONLY BE USED IN bucket level. Support fuzzy match and limit to 5
	// If no sub-resource is specified in a statement, then all objects in the bucket are accessible by the principal.
	// However, if the sub-resource is defined as 'bucket/test_*,' in the statement, then only objects with a 'test_'
	// prefix can be accessed by the principal.
	Resources pq.StringArray `gorm:"resources;type:text"`
	// ExpirationTime defines how long the permission is valid. If not explicitly specified, it means it will not expire.
	ExpirationTime int64 `gorm:"expiration_time"`
	// LimitSize defines the total data size that is allowed to operate. If not explicitly specified, it means it will not limit.
	LimitSize uint64 `gorm:"limit_size"`
	// Removed defines the statement is deleted or not
	Removed bool `gorm:"removed"`
}

// TableName is used to set Statements table name in database
func (s *Statement) TableName() string {
	return StatementTableName
}

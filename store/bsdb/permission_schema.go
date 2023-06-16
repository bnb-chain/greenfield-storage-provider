package bsdb

//
//import (
//	"github.com/forbole/juno/v4/common"
//)
//
//// Permission is the structure to verify action permission
//type Permission struct {
//	// ID defines db auto_increment id of permission
//	ID uint64 `gorm:"id"`
//	// PrincipalType defines the type of principal
//	PrincipalType int32 `gorm:"principal_type"`
//	// PrincipalValue defines the value of principal
//	// When the type is an account, its value is sdk.AccAddress().String();
//	// when the type is a group, its value is math.Uint().String()
//	PrincipalValue string `gorm:"principal_value"`
//	// ResourceType defines the type of resource that grants permission for
//	ResourceType string `gorm:"resource_type"`
//	// ResourceID defines the bucket/object/group id of the resource that grants permission for
//	ResourceID common.Hash `gorm:"resource_id"`
//	// PolicyID is a unique u256 sequence for each policy. It also is used as NFT tokenID
//	PolicyID common.Hash `gorm:"policy_id"`
//	// CreateTimestamp defines the create time of permission
//	CreateTimestamp int64 `gorm:"create_timestamp"`
//	// UpdateTimestamp defines the update time of permission
//	UpdateTimestamp int64 `gorm:"update_timestamp"`
//	// ExpirationTime defines the expiration time of permission
//	ExpirationTime int64 `gorm:"expiration_time;type:bigint(64)"`
//	// Removed defines the permission is deleted or not
//	Removed bool `gorm:"removed"`
//}
//
//// TableName is used to set Permission table name in database
//func (p *Permission) TableName() string {
//	return PermissionTableName
//}

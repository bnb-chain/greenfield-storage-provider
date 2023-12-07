package bsdb

import (
	"encoding/json"

	"github.com/forbole/juno/v4/common"
	"gorm.io/datatypes"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// Group is the structure for group information
type Group struct {
	// ID defines db auto_increment id of group
	ID uint64 `gorm:"column:id"`
	// Owner is the account address of group creator
	Owner common.Address `gorm:"column:owner"`
	// GroupID is the unique identification for bucket.
	GroupID common.Hash `gorm:"column:group_id"`
	// GroupName defines the name of the group
	GroupName string `gorm:"column:group_name"`
	// SourceType defines which chain the user should send the bucket management transactions to
	SourceType string `gorm:"column:source_type"`
	// Extra defines the extra info for the group to update
	Extra string `gorm:"column:extra"`
	// AccountID defines the group user address
	AccountID common.Address `gorm:"column:account_id"`
	// Operator defines operator address of group
	Operator common.Address `gorm:"column:operator"`
	// ExpirationTime defines the expiration time of group member
	ExpirationTime int64 `gorm:"column:expiration_time"`
	// CreateAt defines the block number when the group created
	CreateAt int64 `gorm:"column:create_at"`
	// CreateTime defines the timestamp when the group created
	CreateTime int64 `gorm:"column:create_time"`
	// UpdateAt defines the block number when the group updated
	UpdateAt int64 `gorm:"column:update_at"`
	// UpdateTime defines the timestamp when the group updated
	UpdateTime int64 `gorm:"column:update_time"`
	// Removed defines the group is deleted or not
	Removed bool `gorm:"column:removed"`
	// Tags
	Tags datatypes.JSON `gorm:"column:tags;TYPE:json"` // tags
}

type GroupMeta struct {
	// AccountID defines the group user address
	AccountID common.Address `gorm:"column:account_id"`
	// ExpirationTime defines the expiration time of group member
	ExpirationTime int64 `gorm:"column:expiration_time"`
}

type GroupCount struct {
	GroupID common.Hash `gorm:"column:group_id"`
	Count   int64       `gorm:"column:count"`
}

// TableName is used to set Group table name in database
func (g *Group) TableName() string {
	return GroupTableName
}

// GetResourceTags is used to convert the db tags string to *storage_types.ResourceTags type
func (g *Group) GetResourceTags() *storagetypes.ResourceTags {
	tags := &storagetypes.ResourceTags{}
	if g.Tags != nil {
		tagUnmarshalErr := json.Unmarshal([]byte(g.Tags), tags)
		if tagUnmarshalErr != nil {
			log.Warnw("failed to Unmarshal group tags", "error", tagUnmarshalErr)
		}
	}
	return tags
}

package bsdb

type EventSwapOut struct {
	StorageProviderId          uint32   `json:"storage_provider_id,omitempty"`
	GlobalVirtualGroupFamilyId uint32   `json:"global_virtual_group_family_id,omitempty"`
	GlobalVirtualGroupIds      []uint32 `json:"global_virtual_group_ids,omitempty"`
	SuccessorSpId              uint32   `json:"successor_sp_id,omitempty"`
}

// TableName is used to set EventSwapOut table name in database
func (g *EventSwapOut) TableName() string {
	return EventSwapOutTableName
}

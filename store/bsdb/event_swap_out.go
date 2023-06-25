package bsdb

// ListSwapOutEvents list swap out events
func (b *BsDBImpl) ListSwapOutEvents(blockID uint64, spID uint32) ([]*EventSwapOut, error) {
	var (
		events []*EventSwapOut
		err    error
	)

	err = b.db.Table((&EventSwapOut{}).TableName()).
		Select("*").
		Where("successor_sp_id = ? and create_at = ?", spID, blockID).
		Find(&events).Error
	return events, err
}

package bsdb

// ListSwapOutEvents list swap out events
func (b *BsDBImpl) ListSwapOutEvents(blockID uint64, spID uint32) ([]*EventSwapOut, []*EventCompleteSwapOut, []*EventCancelSwapOut, error) {
	var (
		events         []*EventSwapOut
		completeEvents []*EventCompleteSwapOut
		cancelEvents   []*EventCancelSwapOut
		err            error
	)

	err = b.db.Table((&EventSwapOut{}).TableName()).
		Select("*").
		Where("storage_provider_id = ? and create_at <= ?", spID, blockID).
		Find(&events).Error
	if err != nil {
		return nil, nil, nil, err
	}

	err = b.db.Table((&EventCompleteSwapOut{}).TableName()).
		Select("*").
		Where("storage_provider_id = ? and create_at <= ?", spID, blockID).
		Find(&completeEvents).Error
	if err != nil {
		return nil, nil, nil, err
	}

	err = b.db.Table((&EventCancelSwapOut{}).TableName()).
		Select("*").
		Where("storage_provider_id = ? and create_at <= ?", spID, blockID).
		Find(&cancelEvents).Error
	if err != nil {
		return nil, nil, nil, err
	}

	return events, completeEvents, cancelEvents, err
}

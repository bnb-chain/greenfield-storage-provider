package bsdb

// ListMigrateBucketEvents list migrate bucket events
func (b *BsDBImpl) ListMigrateBucketEvents(blockID uint64, spID uint32) ([]*EventMigrationBucket, []*EventCompleteMigrationBucket, error) {
	var (
		events         []*EventMigrationBucket
		completeEvents []*EventCompleteMigrationBucket
		err            error
	)

	err = b.db.Table((&EventMigrationBucket{}).TableName()).
		Select("*").
		Where("dst_primary_sp_id = ? and create_at = ?", spID, blockID).
		Find(&events).Error
	if err != nil {
		return nil, nil, err
	}

	err = b.db.Table((&EventCompleteMigrationBucket{}).TableName()).
		Select("*").
		Where("dst_primary_sp_id = ? and create_at = ?", spID, blockID).
		Find(&completeEvents).Error
	if err != nil {
		return events, nil, err
	}

	return events, completeEvents, err
}

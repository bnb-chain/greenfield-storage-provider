package bsdb

// ListMigrateBucketEvents list migrate bucket events
func (b *BsDBImpl) ListMigrateBucketEvents(blockID uint64, spID uint32) ([]*EventMigrationBucket, error) {
	var (
		events []*EventMigrationBucket
		err    error
	)

	err = b.db.Table((&EventMigrationBucket{}).TableName()).
		Select("*").
		Where("dst_primary_sp_id = ? and create_at = ?", spID, blockID).
		Find(&events).Error
	return events, err
}

package bsdb

import "time"

// ListMigrateBucketEvents list migrate bucket events
func (b *BsDBImpl) ListMigrateBucketEvents(blockID uint64, spID uint32) ([]*EventMigrationBucket, []*EventCompleteMigrationBucket, []*EventCancelMigrationBucket, error) {
	var (
		events         []*EventMigrationBucket
		completeEvents []*EventCompleteMigrationBucket
		cancelEvents   []*EventCancelMigrationBucket
		err            error
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	err = b.db.Table((&EventMigrationBucket{}).TableName()).
		Select("*").
		Where("dst_primary_sp_id = ? and create_at <= ?", spID, blockID).
		Find(&events).Error
	if err != nil {
		return nil, nil, nil, err
	}

	err = b.db.Table((&EventCompleteMigrationBucket{}).TableName()).
		Select("*").
		Where("create_at <= ?", blockID).
		Find(&completeEvents).Error
	if err != nil {
		return events, nil, nil, err
	}

	err = b.db.Table((&EventCancelMigrationBucket{}).TableName()).
		Select("*").
		Where("create_at <= ?", blockID).
		Find(&cancelEvents).Error
	if err != nil {
		return events, nil, nil, err
	}

	return events, completeEvents, cancelEvents, err
}

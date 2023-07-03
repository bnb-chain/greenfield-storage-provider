package bsdb

import "github.com/forbole/juno/v4/common"

// ListMigrateBucketEvents list migrate bucket events
func (b *BsDBImpl) ListMigrateBucketEvents(blockID uint64, spID uint32) ([]*EventMigrationBucket, []*EventCompleteMigrationBucket, error) {
	var (
		events         []*EventMigrationBucket
		res            []*EventMigrationBucket
		eventsIDMap    map[common.Hash]*EventMigrationBucket
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
		return nil, nil, err
	}

	eventsIDMap = make(map[common.Hash]*EventMigrationBucket)
	for _, event := range events {
		eventsIDMap[event.BucketID] = event
	}

	res = make([]*EventMigrationBucket, 0)
	for _, event := range completeEvents {
		if e, ok := eventsIDMap[event.BucketID]; ok {
			res = append(res, e)
		}
	}

	return res, completeEvents, err
}

package bsdb

import (
	"errors"
	"time"

	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"
)

// ListMigrateBucketEvents list migrate bucket events
func (b *BsDBImpl) ListMigrateBucketEvents(spID uint32, filters ...func(*gorm.DB) *gorm.DB) ([]*EventMigrationBucket, []*EventCompleteMigrationBucket, []*EventCancelMigrationBucket, []*EventRejectMigrateBucket, error) {
	var (
		events         []*EventMigrationBucket
		completeEvents []*EventCompleteMigrationBucket
		cancelEvents   []*EventCancelMigrationBucket
		rejectEvents   []*EventRejectMigrateBucket
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
		Where("dst_primary_sp_id = ?", spID).
		Scopes(filters...).
		Find(&events).Error
	if err != nil {
		return nil, nil, nil, nil, err
	}

	err = b.db.Table((&EventCompleteMigrationBucket{}).TableName()).
		Select("*").
		Scopes(filters...).
		Find(&completeEvents).Error
	if err != nil {
		return events, nil, nil, nil, err
	}

	err = b.db.Table((&EventCancelMigrationBucket{}).TableName()).
		Select("*").
		Scopes(filters...).
		Find(&cancelEvents).Error
	if err != nil {
		return events, nil, nil, nil, err
	}

	err = b.db.Table((&EventRejectMigrateBucket{}).TableName()).
		Select("*").
		Scopes(filters...).
		Find(&rejectEvents).Error
	if err != nil {
		return events, nil, nil, nil, err
	}

	return events, completeEvents, cancelEvents, rejectEvents, err
}

// GetMigrateBucketEventByBucketID get migrate bucket event by bucket id
func (b *BsDBImpl) GetMigrateBucketEventByBucketID(bucketID common.Hash) (*EventCompleteMigrationBucket, error) {
	var (
		completeEvents *EventCompleteMigrationBucket
		err            error
	)

	err = b.db.Table((&EventCompleteMigrationBucket{}).TableName()).
		Select("*").
		Where("bucket_id = ?", bucketID).
		Order("create_time desc").
		Take(&completeEvents).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return completeEvents, err
}

// GetEventMigrationBucketByBucketID get migration bucket event by bucket id
func (b *BsDBImpl) GetEventMigrationBucketByBucketID(bucketID common.Hash) (*EventMigrationBucket, error) {
	var (
		event       *EventMigrationBucket
		cancelEvent *EventCancelMigrationBucket
		err         error
	)

	err = b.db.Table((&EventMigrationBucket{}).TableName()).
		Select("*").
		Where("bucket_id = ?", bucketID).
		Order("create_time desc").
		Take(&event).Error
	if err != nil {
		return nil, err
	}

	err = b.db.Table((&EventCancelMigrationBucket{}).TableName()).
		Select("*").
		Where("bucket_id = ?", bucketID).
		Order("create_time desc").
		Take(&cancelEvent).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return event, nil
		}
		return nil, err
	}

	// check if the latest cancel event create at is larger than migration bucket
	// it means the bucket is not in migration status
	if cancelEvent.CreateAt > event.CreateAt {
		return nil, errors.New("the bucket is not in migration status")
	}

	return event, err
}

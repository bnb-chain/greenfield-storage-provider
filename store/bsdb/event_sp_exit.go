package bsdb

import (
	"time"

	"gorm.io/gorm"
)

// ListSpExitEvents list sp exit events
func (b *BsDBImpl) ListSpExitEvents(blockID uint64, spID uint32) (*EventStorageProviderExit, *EventCompleteStorageProviderExit, error) {
	var (
		event         *EventStorageProviderExit
		completeEvent *EventCompleteStorageProviderExit
		err           error
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

	err = b.db.Table((&EventStorageProviderExit{}).TableName()).
		Select("*").
		Where("storage_provider_id = ? and create_at <= ?", spID, blockID).
		Order("create_at desc").
		Take(&event).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	err = b.db.Table((&EventCompleteStorageProviderExit{}).TableName()).
		Select("*").
		Where("storage_provider_id = ? and create_at <= ?", spID, blockID).
		Order("create_at desc").
		Take(&completeEvent).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return event, nil, nil
		}
		return nil, nil, err
	}

	return event, completeEvent, nil
}

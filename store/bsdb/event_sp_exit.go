package bsdb

import (
	"time"

	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"
)

// ListSpExitEvents list sp exit events
func (b *BsDBImpl) ListSpExitEvents(blockID uint64, operatorAddress common.Address) (*EventStorageProviderExit, *EventCompleteStorageProviderExit, error) {
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
		Where("operator_address = ? and create_at <= ?", operatorAddress, blockID).
		Take(&event).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	err = b.db.Table((&EventCompleteStorageProviderExit{}).TableName()).
		Select("*").
		Where("operator_address = ? and create_at <= ?", operatorAddress, blockID).
		Take(&completeEvent).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return event, nil, nil
		}
		return nil, nil, err
	}

	return event, completeEvent, nil
}

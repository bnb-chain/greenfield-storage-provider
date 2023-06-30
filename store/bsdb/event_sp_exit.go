package bsdb

import "github.com/forbole/juno/v4/common"

// ListSpExitEvents list sp exit events
func (b *BsDBImpl) ListSpExitEvents(blockID uint64, operatorAddress common.Address) ([]*EventStorageProviderExit, []*EventCompleteStorageProviderExit, error) {
	var (
		events         []*EventStorageProviderExit
		res            []*EventStorageProviderExit
		eventsIDMap    map[uint32]*EventStorageProviderExit
		completeEvents []*EventCompleteStorageProviderExit
		err            error
	)

	err = b.db.Table((&EventStorageProviderExit{}).TableName()).
		Select("*").
		Where("operator_address = ? and create_at = ?", operatorAddress, blockID).
		Find(&events).Error

	err = b.db.Table((&EventCompleteStorageProviderExit{}).TableName()).
		Select("*").
		Where("operator_address = ? and create_at = ?", operatorAddress, blockID).
		Find(&completeEvents).Error

	for _, event := range events {
		eventsIDMap[event.StorageProviderId] = event
	}

	res = make([]*EventStorageProviderExit, 0)
	for _, event := range completeEvents {
		if e, ok := eventsIDMap[event.StorageProviderId]; ok {
			res = append(res, e)
		}
	}

	return res, completeEvents, err
}

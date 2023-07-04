package bsdb

import "github.com/forbole/juno/v4/common"

// ListSpExitEvents list sp exit events
func (b *BsDBImpl) ListSpExitEvents(blockID uint64, operatorAddress common.Address) ([]*EventStorageProviderExit, []*EventCompleteStorageProviderExit, error) {
	var (
		events         []*EventStorageProviderExit
		completeEvents []*EventCompleteStorageProviderExit
		err            error
	)

	err = b.db.Table((&EventStorageProviderExit{}).TableName()).
		Select("*").
		Where("operator_address = ? and create_at = ?", operatorAddress, blockID).
		Find(&events).Error
	if err != nil {
		return nil, nil, err
	}

	err = b.db.Table((&EventCompleteStorageProviderExit{}).TableName()).
		Select("*").
		Where("operator_address = ? and create_at = ?", operatorAddress, blockID).
		Find(&completeEvents).Error
	if err != nil {
		return nil, nil, err
	}

	return events, completeEvents, err
}

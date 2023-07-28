package bsdb

import (
	"fmt"
	"time"
)

// ListSwapOutEvents list swap out events
func (b *BsDBImpl) ListSwapOutEvents(blockID uint64, spID uint32) ([]*EventSwapOut, []*EventCompleteSwapOut, []*EventCancelSwapOut, error) {
	var (
		events         []*EventSwapOut
		completeEvents []*EventCompleteSwapOut
		cancelEvents   []*EventCancelSwapOut
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

	err = b.db.Table((&EventSwapOut{}).TableName()).
		Select("*").
		Where("storage_provider_id = ? and create_at <= ?", spID, blockID).
		Find(&events).Error
	if err != nil {
		return nil, nil, nil, err
	}

	err = b.db.Table((&EventCompleteSwapOut{}).TableName()).
		Select("*").
		Where("src_storage_provider_id = ? and create_at <= ?", spID, blockID).
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

func CreateSwapOutIdx(vgfID uint32, spID uint32, gvgIDS []uint32) string {
	var (
		idx string
		ids string
	)

	// StorageProviderId and GlobalVirtualGroupFamilyId ensure event continuity when GlobalVirtualGroupFamilyId != 0.
	// If it's 0, 'StorageProviderId' with GlobalVirtualGroupIds serve the same purpose.
	if vgfID == 0 {
		for j, id := range gvgIDS {
			if j != 0 {
				ids += "+"
			}
			ids += fmt.Sprintf("%d", id)
		}
		idx = fmt.Sprintf("%d+%s", spID, ids)
	} else {
		idx = fmt.Sprintf("%d+%d", spID, vgfID)
	}
	return idx
}

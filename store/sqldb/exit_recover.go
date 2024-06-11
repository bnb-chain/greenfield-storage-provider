package sqldb

import (
	"fmt"

	"github.com/zkMeLabs/mechain-storage-provider/core/spdb"
)

func (s *SpDBImpl) GetRecoverGVGStats(gvgID uint32) (*spdb.RecoverGVGStats, error) {
	var queryReturn RecoverGVGStatsTable
	if err := s.db.Model(&RecoverGVGStatsTable{}).
		Where("virtual_group_id = ?", gvgID).
		First(&queryReturn).Error; err != nil {
		return nil, err
	}
	return &spdb.RecoverGVGStats{
		VirtualGroupFamilyID: queryReturn.VirtualGroupFamilyID,
		VirtualGroupID:       queryReturn.VirtualGroupID,
		RedundancyIndex:      queryReturn.RedundancyIndex,
		Status:               spdb.RecoverStatus(queryReturn.Status),
		StartAfter:           queryReturn.StartAfter,
		NextStartAfter:       queryReturn.NextStartAfter,
		Limit:                uint64(queryReturn.Limit),
		ObjectCount:          queryReturn.ObjectCount,
	}, nil
}

func (s *SpDBImpl) BatchGetRecoverGVGStats(gvgIDs []uint32) ([]*spdb.RecoverGVGStats, error) {
	var queryReturn []*RecoverGVGStatsTable
	if err := s.db.Model(&RecoverGVGStatsTable{}).
		Where("virtual_group_id IN ?", gvgIDs).
		Find(&queryReturn).Error; err != nil {
		return nil, err
	}
	res := make([]*spdb.RecoverGVGStats, 0, len(queryReturn))
	for _, ret := range queryReturn {
		res = append(res, &spdb.RecoverGVGStats{
			VirtualGroupFamilyID: ret.VirtualGroupFamilyID,
			VirtualGroupID:       ret.VirtualGroupID,
			RedundancyIndex:      ret.RedundancyIndex,
			StartAfter:           ret.StartAfter,
			NextStartAfter:       ret.NextStartAfter,
			Status:               spdb.RecoverStatus(ret.Status),
			Limit:                uint64(ret.Limit),
			ObjectCount:          ret.ObjectCount,
		})
	}
	return res, nil
}

func (s *SpDBImpl) SetRecoverGVGStats(stats []*spdb.RecoverGVGStats) error {
	saveGVG := make([]*RecoverGVGStatsTable, 0)
	for _, g := range stats {
		gt := &RecoverGVGStatsTable{
			VirtualGroupFamilyID: g.VirtualGroupFamilyID,
			VirtualGroupID:       g.VirtualGroupID,
			RedundancyIndex:      g.RedundancyIndex,
			StartAfter:           g.StartAfter,
			NextStartAfter:       g.NextStartAfter,
			Limit:                uint32(g.Limit),
			Status:               int(g.Status),
		}
		saveGVG = append(saveGVG, gt)
	}

	err := s.db.CreateInBatches(saveGVG, len(saveGVG)).Error
	if err != nil && MysqlErrCode(err) == ErrDuplicateEntryCode {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to set gvg recover stats record: %s", err.Error())
	}
	return nil
}

func (s *SpDBImpl) UpdateRecoverGVGStats(stats *spdb.RecoverGVGStats) (err error) {
	result := s.db.Table(RecoverGVGStatsTableName).Where("virtual_group_id = ?", stats.VirtualGroupID).
		Updates(&RecoverGVGStatsTable{
			Status:         int(stats.Status),
			StartAfter:     stats.StartAfter,
			NextStartAfter: stats.NextStartAfter,
			ObjectCount:    stats.ObjectCount,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to update the GVG status for recover_stats table: %s", result.Error)
	}
	return nil
}

func (s *SpDBImpl) DeleteRecoverGVGStats(gvgID uint32) (err error) {
	err = s.db.Table(RecoverGVGStatsTableName).Delete(&RecoverGVGStatsTable{
		VirtualGroupID: gvgID,
	}).Error
	return err
}

func (s *SpDBImpl) InsertRecoverFailedObject(object *spdb.RecoverFailedObject) error {
	result := s.db.Create(&RecoverFailedObjectTable{
		ObjectID:        object.ObjectID,
		VirtualGroupID:  object.VirtualGroupID,
		RedundancyIndex: object.RedundancyIndex,
	})
	if result.Error != nil && MysqlErrCode(result.Error) == ErrDuplicateEntryCode {
		return nil
	}
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("failed to set recover object record: %s", result.Error)
	}
	return nil
}

func (s *SpDBImpl) DeleteRecoverFailedObject(objectID uint64) (err error) {
	err = s.db.Table(RecoverFailedObjectTableName).Delete(&RecoverFailedObjectTable{
		ObjectID: objectID,
	}).Error
	return err
}

func (s *SpDBImpl) GetRecoverFailedObject(objectID uint64) (*spdb.RecoverFailedObject, error) {
	var queryReturn RecoverFailedObjectTable
	if err := s.db.Model(&RecoverFailedObjectTable{}).
		Where("object_id = ?", objectID).
		First(&queryReturn).Error; err != nil {
		return nil, err
	}
	return &spdb.RecoverFailedObject{
		ObjectID:        objectID,
		VirtualGroupID:  queryReturn.VirtualGroupID,
		RedundancyIndex: queryReturn.RedundancyIndex,
		RetryTime:       queryReturn.Retry,
	}, nil
}

func (s *SpDBImpl) GetRecoverFailedObjects(retry, limit uint32) ([]*spdb.RecoverFailedObject, error) {
	var (
		recoverObjects []*RecoverFailedObjectTable
		returnObjects  []*spdb.RecoverFailedObject
		err            error
	)

	err = s.db.Table(RecoverFailedObjectTableName).
		Select("*").
		Where("retry <= ?", retry).
		Order("object_id asc").
		Limit(int(limit)).
		Find(&recoverObjects).Error
	if err != nil {
		return nil, err
	}
	for _, object := range recoverObjects {
		returnObjects = append(returnObjects, &spdb.RecoverFailedObject{
			ObjectID:        object.ObjectID,
			VirtualGroupID:  object.VirtualGroupID,
			RedundancyIndex: object.RedundancyIndex,
			RetryTime:       object.Retry,
		})
	}
	return returnObjects, nil
}

func (s *SpDBImpl) GetRecoverFailedObjectsByRetryTime(retry uint32) ([]*spdb.RecoverFailedObject, error) {
	var (
		recoverObjects []*RecoverFailedObjectTable
		returnObjects  []*spdb.RecoverFailedObject
		err            error
	)

	err = s.db.Table(RecoverFailedObjectTableName).
		Select("*").
		Where("retry >= ?", retry).
		Order("object_id asc").
		Find(&recoverObjects).Error
	if err != nil {
		return nil, err
	}
	for _, object := range recoverObjects {
		returnObjects = append(returnObjects, &spdb.RecoverFailedObject{
			ObjectID:        object.ObjectID,
			VirtualGroupID:  object.VirtualGroupID,
			RedundancyIndex: object.RedundancyIndex,
			RetryTime:       object.Retry,
		})
	}
	return returnObjects, nil
}

func (s *SpDBImpl) UpdateRecoverFailedObject(object *spdb.RecoverFailedObject) (err error) {
	result := s.db.Table(RecoverFailedObjectTableName).Where("object_id = ?", object.ObjectID).
		Updates(&RecoverFailedObjectTable{
			Retry: object.RetryTime,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to VerifyGVGProgress%s", result.Error)
	}
	return nil
}

func (s *SpDBImpl) CountRecoverFailedObject() (count int64, err error) {
	result := s.db.Table(RecoverFailedObjectTableName).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count recover failed object%s", result.Error)
	}
	return
}

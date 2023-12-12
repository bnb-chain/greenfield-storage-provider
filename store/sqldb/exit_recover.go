package sqldb

import (
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
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
		Limit:                uint64(queryReturn.Limit),
	}, nil
}

func (s *SpDBImpl) SetRecoverGVGStats(stats []*spdb.RecoverGVGStats) error {
	saveGVG := make([]*RecoverGVGStatsTable, 0)
	for _, g := range stats {
		gt := &RecoverGVGStatsTable{
			VirtualGroupFamilyID: g.VirtualGroupFamilyID,
			VirtualGroupID:       g.VirtualGroupID,
			RedundancyIndex:      g.RedundancyIndex,
			StartAfterObjectID:   g.StartAfter,
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
		return fmt.Errorf("failed to set gvg recover stats record: %s", err.Error)
	}
	return nil
}

func (s *SpDBImpl) UpdateRecoverGVGStats(stats *spdb.RecoverGVGStats) (err error) {
	result := s.db.Table(RecoverGVGStatsTableName).Where("virtual_group_id = ?", stats.VirtualGroupID).
		Updates(&RecoverGVGStatsTable{
			Status:             int(stats.Status),
			StartAfterObjectID: stats.StartAfter,
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

func (s *SpDBImpl) SetVerifyGVGProgress(gvgProgress []*spdb.VerifyGVGProgress) error {
	saveGVGProgress := make([]*VerifyGVGProgressTable, 0)
	for _, g := range gvgProgress {
		gt := &VerifyGVGProgressTable{
			VirtualGroupID:  g.VirtualGroupID,
			RedundancyIndex: g.RedundancyIndex,
			StartAfter:      g.StartAfter,
			Limit:           uint32(g.Limit),
		}
		saveGVGProgress = append(saveGVGProgress, gt)
	}

	err := s.db.CreateInBatches(saveGVGProgress, len(saveGVGProgress)).Error
	if err != nil && MysqlErrCode(err) == ErrDuplicateEntryCode {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to set gvg recover stats record: %s", err.Error)
	}
	return nil
}

func (s *SpDBImpl) GetVerifyGVGProgress(gvgID uint32) (*spdb.VerifyGVGProgress, error) {
	var queryReturn VerifyGVGProgressTable
	if err := s.db.Model(&VerifyGVGProgressTable{}).
		Where("virtual_group_id = ?", gvgID).
		First(&queryReturn).Error; err != nil {
		return nil, err
	}
	return &spdb.VerifyGVGProgress{
		VirtualGroupID:  queryReturn.VirtualGroupID,
		RedundancyIndex: queryReturn.RedundancyIndex,
		StartAfter:      queryReturn.StartAfter,
		Limit:           uint64(queryReturn.Limit),
	}, nil
}

func (s *SpDBImpl) UpdateVerifyGVGProgress(gvgProgress *spdb.VerifyGVGProgress) (err error) {
	result := s.db.Table(VerifyGVGProgressTableName).Where("virtual_group_id = ?", gvgProgress.VirtualGroupID).
		Updates(&VerifyGVGProgressTable{
			StartAfter: gvgProgress.StartAfter,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to VerifyGVGProgress%s", result.Error)
	}
	return nil
}

func (s *SpDBImpl) DeleteVerifyGVGProgress(gvgID uint64) (err error) {
	err = s.db.Table(VerifyGVGProgressTableName).Delete(&VerifyGVGProgressTable{
		VirtualGroupID: uint32(gvgID),
	}).Error
	return err
}

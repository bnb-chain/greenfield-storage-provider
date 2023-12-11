package sqldb

import (
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
)

const ListRecoverObjectLimit = 50

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
		RedundancyIndex:      uint32(queryReturn.RedundancyIndex),
		Status:               queryReturn.Status,
	}, nil
}

func (s *SpDBImpl) SetRecoverGVGStats(stats []*spdb.RecoverGVGStats) error {
	result := s.db.Create(&RecoverGVGStatsTable{
		VirtualGroupFamilyID: stats.VirtualGroupFamilyID,
		VirtualGroupID:       stats.VirtualGroupID,
		ExitingSPID:          stats.ExitingSPID,
		RedundancyIndex:      stats.RedundancyIndex,
		StartAfterObjectID:   stats.StartAfterObjectID,
		Limit:                uint32(stats.Limit),
		Status:               stats.Status,
	})
	if result.Error != nil && MysqlErrCode(result.Error) == ErrDuplicateEntryCode {
		return nil
	}
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("failed to set gvg recover stats record: %s", result.Error)
	}
	return nil
}

func (s *SpDBImpl) UpdateRecoverGVGStats(stats *spdb.RecoverGVGStats) (err error) {
	result := s.db.Table(RecoverGVGStatsTableName).Where("virtual_group_id = ?", stats.VirtualGroupID).
		Updates(&RecoverGVGStatsTable{
			Status:             stats.Status,
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

func (s *SpDBImpl) GetRecoverGVGStatsByFamilyIDAndStatus(familyID uint32, status int) ([]*spdb.RecoverGVGStats, error) {
	var (
		recoverGVGstats []*RecoverGVGStatsTable
		returnGVGstats  []*spdb.RecoverGVGStats
		err             error
	)

	err = s.db.Table(RecoverGVGStatsTableName).
		Select("*").
		Where("virtual_group_family_id = ? and status = ?", familyID, status).
		Order("virtual_group_id asc").
		Find(&recoverGVGstats).Error
	if err != nil {
		return nil, err
	}
	for _, gvgStats := range recoverGVGstats {
		returnGVGstats = append(returnGVGstats, &spdb.RecoverGVGStats{
			VirtualGroupFamilyID: gvgStats.VirtualGroupFamilyID,
			VirtualGroupID:       gvgStats.VirtualGroupID,
			RedundancyIndex:      uint32(gvgStats.RedundancyIndex),
			Status:               gvgStats.Status,
		})
	}
	return returnGVGstats, nil
}

func (s *SpDBImpl) InsertRecoverFailedObject(object *spdb.RecoverFailedObject) error {
	result := s.db.Create(&RecoverFailedObjectTable{
		ObjectID:        object.ObjectID,
		VirtualGroupID:  object.VirtualGroupID,
		RedundancyIndex: int32(object.RedundancyIndex),
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

func (s *SpDBImpl) GetRecoverObject(objectID uint64) (*spdb.RecoverFailedObject, error) {
	var queryReturn RecoverFailedObjectTable
	if err := s.db.Model(&RecoverFailedObjectTable{}).
		Where("object_id = ?", objectID).
		First(&queryReturn).Error; err != nil {
		return nil, err
	}
	return &spdb.RecoverFailedObject{
		ObjectID:        objectID,
		VirtualGroupID:  queryReturn.VirtualGroupID,
		RedundancyIndex: uint32(queryReturn.RedundancyIndex),
	}, nil
}

func (s *SpDBImpl) GetRecoverObjectsByGVGID(gvgID uint32) ([]*spdb.RecoverFailedObject, error) {
	var (
		recoverObjects []*RecoverFailedObjectTable
		returnObjects  []*spdb.RecoverFailedObject
		err            error
	)

	err = s.db.Table(RecoverFailedObjectTableName).
		Select("*").
		Where("virtual_group_id = ?", gvgID).
		Order("object_id asc").
		Limit(ListRecoverObjectLimit).
		Find(&recoverObjects).Error
	if err != nil {
		return nil, err
	}
	for _, object := range recoverObjects {
		returnObjects = append(returnObjects, &spdb.RecoverFailedObject{
			ObjectID:        object.ObjectID,
			VirtualGroupID:  object.VirtualGroupID,
			RedundancyIndex: uint32(object.RedundancyIndex),
		})
	}
	return returnObjects, nil
}

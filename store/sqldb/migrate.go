package sqldb

import (
	"errors"
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"gorm.io/gorm"
)

const (
	SPExitProgressKey        = "sp_exit_progress"
	SwapOutProgressKey       = "swap_out_progress"
	BucketMigrateProgressKey = "bucket_migrate_progress"
)

// UpdateSPExitSubscribeProgress is used to update progress.
// insert a new one if it is not found in db.
func (s *SpDBImpl) UpdateSPExitSubscribeProgress(blockHeight uint64) error {
	var (
		result       *gorm.DB
		queryReturn  *MigrateSubscribeProgressTable
		needInsert   bool
		updateRecord *MigrateSubscribeProgressTable
	)
	queryReturn = &MigrateSubscribeProgressTable{}
	result = s.db.First(queryReturn, "event_name = ?", SPExitProgressKey)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	if result.Error != nil {
		needInsert = errors.Is(result.Error, gorm.ErrRecordNotFound)
	}
	updateRecord = &MigrateSubscribeProgressTable{
		EventName:                 SPExitProgressKey,
		LastSubscribedBlockHeight: blockHeight,
	}
	if needInsert {
		result = s.db.Create(updateRecord)
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("failed to insert record in subscribe progress table: %s", result.Error)
		}

	} else { // update
		result = s.db.Model(&MigrateSubscribeProgressTable{}).
			Where("event_name = ?", SPExitProgressKey).Updates(updateRecord)
		if result.Error != nil {
			return fmt.Errorf("failed to detele record in subscribe progress table: %s", result.Error)
		}
	}
	return nil
}

func (s *SpDBImpl) QuerySPExitSubscribeProgress() (uint64, error) {
	var (
		result      *gorm.DB
		queryReturn *MigrateSubscribeProgressTable
	)
	queryReturn = &MigrateSubscribeProgressTable{}
	result = s.db.First(queryReturn, "event_name = ?", SPExitProgressKey)
	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if result.Error != nil {
		return 0, result.Error
	}
	return queryReturn.LastSubscribedBlockHeight, nil
}

func (s *SpDBImpl) UpdateSwapOutSubscribeProgress(blockHeight uint64) error {
	// TODO:
	return nil
}

func (s *SpDBImpl) QuerySwapOutSubscribeProgress() (uint64, error) {
	// TODO:
	return 0, nil
}
func (s *SpDBImpl) UpdateBucketMigrateSubscribeProgress(blockHeight uint64) error {
	// TODO:

	return nil
}

func (s *SpDBImpl) QueryBucketMigrateSubscribeProgress() (uint64, error) {
	var (
		result      *gorm.DB
		queryReturn *MigrateSubscribeProgressTable
	)
	queryReturn = &MigrateSubscribeProgressTable{}
	result = s.db.First(queryReturn, "event_name = ?", BucketMigrateProgressKey)
	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if result.Error != nil {
		return 0, result.Error
	}
	return queryReturn.LastSubscribedBlockHeight, nil
}

func (s *SpDBImpl) InsertMigrateGVGUnit(meta *spdb.MigrateGVGUnitMeta) error {
	var (
		err              error
		result           *gorm.DB
		insertMigrateGVG *MigrateGVGTable
	)
	insertMigrateGVG = &MigrateGVGTable{
		GlobalVirtualGroupID:   meta.GlobalVirtualGroupID,
		VirtualGroupFamilyID:   meta.VirtualGroupFamilyID,
		MigrateRedundancyIndex: meta.MigrateRedundancyIndex,
		BucketID:               meta.BucketID,
		IsSecondaryGVG:         meta.IsSecondaryGVG,
		MigrateStatus:          meta.MigrateStatus,
	}
	insertMigrateGVG.MigrateKey = MigrateGVGPrimaryKey(insertMigrateGVG)
	result = s.db.Create(insertMigrateGVG)
	if result.Error != nil || result.RowsAffected != 1 {
		err = fmt.Errorf("failed to insert migrate gvg table: %s", result.Error)
		return err
	}
	return nil
}

func (s *SpDBImpl) DeleteMigrateGVGUnit(meta *spdb.MigrateGVGUnitMeta) error {
	// TODO:
	return nil
}

func (s *SpDBImpl) UpdateMigrateGVGUnit(meta *spdb.MigrateGVGUnitMeta) error {
	// TODO:
	return nil
}

func (s *SpDBImpl) UpdateMigrateGVGUnitStatus(meta *spdb.MigrateGVGUnitMeta, migrateStatus int) error {
	var (
		result     *gorm.DB
		queryMeta  *MigrateGVGTable
		migrateKey string
	)
	queryMeta = &MigrateGVGTable{
		GlobalVirtualGroupID:   meta.GlobalVirtualGroupID,
		VirtualGroupFamilyID:   meta.VirtualGroupFamilyID,
		MigrateRedundancyIndex: meta.MigrateRedundancyIndex,
		BucketID:               meta.BucketID,
		IsSecondaryGVG:         meta.IsSecondaryGVG,
	}
	migrateKey = MigrateGVGPrimaryKey(queryMeta)
	if result = s.db.Model(&MigrateGVGTable{}).Where("migrate_key = ?", migrateKey).Updates(&MigrateGVGTable{
		MigrateStatus: migrateStatus,
	}); result.Error != nil {
		return fmt.Errorf("failed to update migrate gvg record: %s", result.Error)
	}
	return nil
}

func (s *SpDBImpl) QueryMigrateGVGUnit(meta *spdb.MigrateGVGUnitMeta) (*spdb.MigrateGVGUnitMeta, error) {
	// TODO:
	return nil, nil
}

// ListMigrateGVGUnitsByFamilyID is used to src sp load to build execute plan.
func (s *SpDBImpl) ListMigrateGVGUnitsByFamilyID(familyID uint32, srcSP uint32) ([]*spdb.MigrateGVGUnitMeta, error) {
	// TODO:
	return nil, nil
}

func (s *SpDBImpl) ListConflictsMigrateGVGUnitsByFamilyID(familyID uint32) ([]*spdb.MigrateGVGUnitMeta, error) {
	return nil, nil
}

// ListSecondaryMigrateGVGUnits is used to src sp load to build execute plan.
func (s *SpDBImpl) ListSecondaryMigrateGVGUnits(srcSP uint32) ([]*spdb.MigrateGVGUnitMeta, error) {
	// TODO:
	return nil, nil
}

func (s *SpDBImpl) ListMigrateGVGUnitsByBucketID(bucketID uint32, destSP uint32) ([]*spdb.MigrateGVGUnitMeta, error) {
	// TODO:
	return nil, nil
}

func (s *SpDBImpl) ListSecondaryMigrateGVGUnitsBySPID(destSP uint32) ([]*spdb.MigrateGVGUnitMeta, error) {
	// TODO:
	return nil, nil
}

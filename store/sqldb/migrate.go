package sqldb

import (
	"errors"

	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"gorm.io/gorm"
)

const (
	SPExitProgressKey        = "sp_exit_progress"
	SwapOutProgressKey       = "swap_out_progress"
	BucketMigrateProgressKey = "bucket_migrate_progress"
)

func (s *SpDBImpl) UpdateSPExitSubscribeProgress(blockHeight uint64) error {
	// TODO:
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
	// TODO:
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
	return 0, nil
}

func (s *SpDBImpl) InsertMigrateGVGUnit(meta *spdb.MigrateGVGUnitMeta) error {
	// TODO:
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

func (s *SpDBImpl) ListMigrateGVGUnitsByFamilyID(familyID uint32, srcSP uint32) ([]*spdb.MigrateGVGUnitMeta, error) {
	// TODO:
	return nil, nil
}

func (s *SpDBImpl) QuerySecondaryMigrateGVGUnit(gvgID uint32, srcSP uint32) (*spdb.MigrateGVGUnitMeta, error) {
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

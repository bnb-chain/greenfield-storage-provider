package sqldb

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
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
			return fmt.Errorf("failed to update record in subscribe progress table: %s", result.Error)
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
	var (
		result       *gorm.DB
		queryReturn  *MigrateSubscribeProgressTable
		needInsert   bool
		updateRecord *MigrateSubscribeProgressTable
	)
	queryReturn = &MigrateSubscribeProgressTable{}
	result = s.db.First(queryReturn, "event_name = ?", SwapOutProgressKey)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	if result.Error != nil {
		needInsert = errors.Is(result.Error, gorm.ErrRecordNotFound)
	}
	updateRecord = &MigrateSubscribeProgressTable{
		EventName:                 SwapOutProgressKey,
		LastSubscribedBlockHeight: blockHeight,
	}
	if needInsert {
		result = s.db.Create(updateRecord)
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("failed to insert record in subscribe progress table: %s", result.Error)
		}

	} else { // update
		result = s.db.Model(&MigrateSubscribeProgressTable{}).
			Where("event_name = ?", SwapOutProgressKey).Updates(updateRecord)
		if result.Error != nil {
			return fmt.Errorf("failed to update record in subscribe progress table: %s", result.Error)
		}
	}
	return nil
}

func (s *SpDBImpl) QuerySwapOutSubscribeProgress() (uint64, error) {
	var (
		result      *gorm.DB
		queryReturn *MigrateSubscribeProgressTable
	)
	queryReturn = &MigrateSubscribeProgressTable{}
	result = s.db.First(queryReturn, "event_name = ?", SwapOutProgressKey)
	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if result.Error != nil {
		return 0, result.Error
	}
	return queryReturn.LastSubscribedBlockHeight, nil
}

func (s *SpDBImpl) UpdateBucketMigrateSubscribeProgress(blockHeight uint64) error {
	var (
		result       *gorm.DB
		queryReturn  *MigrateSubscribeProgressTable
		needInsert   bool
		updateRecord *MigrateSubscribeProgressTable
	)
	queryReturn = &MigrateSubscribeProgressTable{}
	result = s.db.First(queryReturn, "event_name = ?", BucketMigrateProgressKey)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	if result.Error != nil {
		needInsert = errors.Is(result.Error, gorm.ErrRecordNotFound)
	}
	updateRecord = &MigrateSubscribeProgressTable{
		EventName:                 BucketMigrateProgressKey,
		LastSubscribedBlockHeight: blockHeight,
	}
	if needInsert {
		result = s.db.Create(updateRecord)
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("failed to insert record in subscribe progress table: %s", result.Error)
		}

	} else { // update
		result = s.db.Model(&MigrateSubscribeProgressTable{}).
			Where("event_name = ?", BucketMigrateProgressKey).Updates(updateRecord)
		if result.Error != nil {
			return fmt.Errorf("failed to update record in subscribe progress table: %s", result.Error)
		}
	}
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

func (s *SpDBImpl) InsertSwapOutUnit(meta *spdb.SwapOutMeta) error {
	var (
		err            error
		result         *gorm.DB
		insertSwapOut  *SwapOutTable
		swapOutMarshal []byte
	)
	if swapOutMarshal, err = json.Marshal(meta.SwapOutMsg); err != nil {
		return err
	}

	insertSwapOut = &SwapOutTable{
		SwapOutKey:       meta.SwapOutKey,
		IsDestSP:         meta.IsDestSP,
		SwapOutMsg:       hex.EncodeToString(swapOutMarshal),
		CompletedGVGList: util.Uint32SliceToString(meta.CompletedGVGs),
	}
	result = s.db.Create(insertSwapOut)
	if result.Error != nil || result.RowsAffected != 1 {
		err = fmt.Errorf("failed to insert swap out table: %s", result.Error)
		return err
	}
	return nil
}

func (s *SpDBImpl) UpdateSwapOutUnitCompletedGVGList(swapOutKey string, completedGVGList []uint32) error {
	result := s.db.Model(&SwapOutTable{}).
		Where("swap_out_key = ? and is_dest_sp = 1", swapOutKey).
		Update("completed_gvg_list", util.Uint32SliceToString(completedGVGList))
	return result.Error
}

func (s *SpDBImpl) QuerySwapOutUnitInSrcSP(swapOutKey string) (*spdb.SwapOutMeta, error) {
	var (
		err            error
		result         *gorm.DB
		queryReturn    *SwapOutTable
		swapOutMarshal []byte
		completedGVGs  []uint32
	)
	queryReturn = &SwapOutTable{}
	result = s.db.First(queryReturn, "is_dest_sp = false and swap_out_key = ?", swapOutKey)
	if result.Error != nil {
		return nil, result.Error
	}

	if swapOutMarshal, err = hex.DecodeString(queryReturn.SwapOutMsg); err != nil {
		return nil, err
	}
	swapOut := virtualgrouptypes.MsgSwapOut{}
	if err = json.Unmarshal(swapOutMarshal, &swapOut); err != nil {
		return nil, err
	}
	if completedGVGs, err = util.StringToUint32Slice(queryReturn.CompletedGVGList); err != nil {
		return nil, err
	}

	return &spdb.SwapOutMeta{
		SwapOutKey:    queryReturn.SwapOutKey,
		IsDestSP:      queryReturn.IsDestSP,
		SwapOutMsg:    &swapOut,
		CompletedGVGs: completedGVGs,
	}, nil
}

func (s *SpDBImpl) ListDestSPSwapOutUnits() ([]*spdb.SwapOutMeta, error) {
	var queryReturns []SwapOutTable
	result := s.db.Where("is_dest_sp = true").Find(&queryReturns)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list swap out table: %s", result.Error)
	}
	returns := make([]*spdb.SwapOutMeta, 0)
	for _, queryReturn := range queryReturns {
		var (
			err            error
			swapOutMarshal []byte
			completedGVGs  []uint32
		)
		if swapOutMarshal, err = hex.DecodeString(queryReturn.SwapOutMsg); err != nil {
			return nil, err
		}
		swapOut := virtualgrouptypes.MsgSwapOut{}
		if err = json.Unmarshal(swapOutMarshal, &swapOut); err != nil {
			return nil, err
		}
		if completedGVGs, err = util.StringToUint32Slice(queryReturn.CompletedGVGList); err != nil {
			return nil, err
		}
		returns = append(returns, &spdb.SwapOutMeta{
			SwapOutKey:    queryReturn.SwapOutKey,
			IsDestSP:      queryReturn.IsDestSP,
			SwapOutMsg:    &swapOut,
			CompletedGVGs: completedGVGs,
		})
	}
	return returns, nil
}

func (s *SpDBImpl) InsertMigrateGVGUnit(meta *spdb.MigrateGVGUnitMeta) error {
	var (
		err              error
		result           *gorm.DB
		insertMigrateGVG *MigrateGVGTable
	)
	insertMigrateGVG = &MigrateGVGTable{
		MigrateKey:               meta.MigrateGVGKey,
		SwapOutKey:               meta.SwapOutKey,
		GlobalVirtualGroupID:     meta.GlobalVirtualGroupID,
		DestGlobalVirtualGroupID: meta.DestGlobalVirtualGroupID,
		VirtualGroupFamilyID:     meta.VirtualGroupFamilyID,
		BucketID:                 meta.BucketID,
		RedundancyIndex:          meta.RedundancyIndex,

		SrcSPID:              meta.SrcSPID,
		DestSPID:             meta.DestSPID,
		LastMigratedObjectID: meta.LastMigratedObjectID,
		MigrateStatus:        meta.MigrateStatus,
	}
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

func (s *SpDBImpl) UpdateMigrateGVGUnitStatus(migrateKey string, migrateStatus int) error {
	if result := s.db.Model(&MigrateGVGTable{}).Where("migrate_key = ?", migrateKey).Updates(&MigrateGVGTable{
		MigrateStatus: migrateStatus,
	}); result.Error != nil {
		return fmt.Errorf("failed to update migrate gvg status: %s", result.Error)
	}
	return nil
}

func (s *SpDBImpl) UpdateMigrateGVGUnitLastMigrateObjectID(migrateKey string, lastMigratedObjectID uint64) error {
	if result := s.db.Model(&MigrateGVGTable{}).Where("migrate_key = ?", migrateKey).Updates(&MigrateGVGTable{
		LastMigratedObjectID: lastMigratedObjectID,
	}); result.Error != nil {
		return fmt.Errorf("failed to update migrate gvg progress: %s", result.Error)
	}
	return nil
}

func (s *SpDBImpl) QueryMigrateGVGUnit(migrateKey string) (*spdb.MigrateGVGUnitMeta, error) {
	var (
		result      *gorm.DB
		queryReturn *MigrateGVGTable
	)
	queryReturn = &MigrateGVGTable{}
	result = s.db.First(queryReturn, "migrate_key = ?", migrateKey)
	if result.Error != nil {
		return nil, result.Error
	}
	return &spdb.MigrateGVGUnitMeta{
		GlobalVirtualGroupID:     queryReturn.GlobalVirtualGroupID,
		DestGlobalVirtualGroupID: queryReturn.DestGlobalVirtualGroupID,
		VirtualGroupFamilyID:     queryReturn.VirtualGroupFamilyID,
		RedundancyIndex:          queryReturn.RedundancyIndex,
		BucketID:                 queryReturn.BucketID,
		SrcSPID:                  queryReturn.SrcSPID,
		DestSPID:                 queryReturn.DestSPID,
		LastMigratedObjectID:     queryReturn.LastMigratedObjectID,
		MigrateStatus:            queryReturn.MigrateStatus,
	}, nil
}

func (s *SpDBImpl) ListMigrateGVGUnitsByBucketID(bucketID uint64) ([]*spdb.MigrateGVGUnitMeta, error) {
	var queryReturns []MigrateGVGTable
	result := s.db.Where("bucket_id = ?", bucketID).Find(&queryReturns)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query migrate gvg table: %s", result.Error)
	}
	returns := make([]*spdb.MigrateGVGUnitMeta, 0)
	for _, queryReturn := range queryReturns {
		returns = append(returns, &spdb.MigrateGVGUnitMeta{
			GlobalVirtualGroupID:     queryReturn.GlobalVirtualGroupID,
			DestGlobalVirtualGroupID: queryReturn.DestGlobalVirtualGroupID,
			VirtualGroupFamilyID:     queryReturn.VirtualGroupFamilyID,
			RedundancyIndex:          queryReturn.RedundancyIndex,
			BucketID:                 queryReturn.BucketID,
			SrcSPID:                  queryReturn.SrcSPID,
			DestSPID:                 queryReturn.DestSPID,
			LastMigratedObjectID:     queryReturn.LastMigratedObjectID,
			MigrateStatus:            queryReturn.MigrateStatus,
		})
	}
	return returns, nil
}

func (s *SpDBImpl) DeleteMigrateGVGUnitsByBucketID(bucketID uint64) error {
	var results []MigrateGVGTable

	result := s.db.Where("bucket_id = ?", bucketID).Find(&results).Delete(&results)
	if result.Error != nil {
		return fmt.Errorf("failed to delete migrate gvg table: %s", result.Error)
	}

	return nil
}

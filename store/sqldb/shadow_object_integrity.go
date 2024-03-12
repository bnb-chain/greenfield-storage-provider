package sqldb

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"gorm.io/gorm"

	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

const ListShadowIntegrityMetaDefaultSize = 5

// GetShadowObjectIntegrity returns the integrity hash info
func (s *SpDBImpl) GetShadowObjectIntegrity(objectID uint64, redundancyIndex int32) (meta *corespdb.ShadowIntegrityMeta, err error) {
	queryReturn := &ShadowIntegrityMetaTable{}
	result := s.db.Table(ShadowIntegrityMetaTableName).Model(&ShadowIntegrityMetaTable{}).
		Where("object_id = ? and redundancy_index = ?", objectID, redundancyIndex).
		First(queryReturn)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		err = result.Error
		return nil, err
	}
	if result.Error != nil {
		err = fmt.Errorf("failed to query integrity meta record: %s", result.Error)
		return nil, err
	}
	integrityChecksum, err := hex.DecodeString(queryReturn.IntegrityChecksum)
	if err != nil {
		return nil, err
	}

	meta = &corespdb.ShadowIntegrityMeta{
		ObjectID:          queryReturn.ObjectID,
		RedundancyIndex:   queryReturn.RedundancyIndex,
		IntegrityChecksum: integrityChecksum,
		Version:           queryReturn.Version,
		PieceSize:         queryReturn.PieceSize,
	}
	meta.PieceChecksumList, err = util.StringToBytesSlice(queryReturn.PieceChecksumList)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

// SetShadowObjectIntegrity puts(overwrites) integrity hash info to db
func (s *SpDBImpl) SetShadowObjectIntegrity(meta *corespdb.ShadowIntegrityMeta) (err error) {
	insertIntegrityMetaRecord := &ShadowIntegrityMetaTable{
		ObjectID:          meta.ObjectID,
		RedundancyIndex:   meta.RedundancyIndex,
		PieceChecksumList: util.BytesSliceToString(meta.PieceChecksumList),
		IntegrityChecksum: hex.EncodeToString(meta.IntegrityChecksum),
		Version:           meta.Version,
		PieceSize:         meta.PieceSize,
	}
	result := s.db.Table(ShadowIntegrityMetaTableName).Create(insertIntegrityMetaRecord)
	if result.Error != nil && MysqlErrCode(result.Error) == ErrDuplicateEntryCode {
		return nil
	}
	if result.Error != nil || result.RowsAffected != 1 {
		err = fmt.Errorf("failed to insert integrity meta record: %s", result.Error)
		return err
	}
	return nil
}

func (s *SpDBImpl) UpdateShadowIntegrityChecksum(meta *corespdb.ShadowIntegrityMeta) (err error) {
	result := s.db.Table(ShadowIntegrityMetaTableName).Where("object_id = ? and redundancy_index = ?", meta.ObjectID, meta.RedundancyIndex).
		Updates(&ShadowIntegrityMetaTable{
			IntegrityChecksum: hex.EncodeToString(meta.IntegrityChecksum),
		})
	if result.Error != nil {
		return fmt.Errorf("failed to update integrity checksum for integrity meta table: %s", result.Error)
	}
	return nil
}

func (s *SpDBImpl) DeleteShadowObjectIntegrity(objectID uint64, redundancyIndex int32) (err error) {
	err = s.db.Table(ShadowIntegrityMetaTableName).Delete(&ShadowIntegrityMetaTable{
		ObjectID:        objectID, // should be the primary key
		RedundancyIndex: redundancyIndex,
	}).Error
	return err
}

func (s *SpDBImpl) ListShadowIntegrityMeta() ([]*corespdb.ShadowIntegrityMeta, error) {
	var (
		shadowIntegrityMetas []*ShadowIntegrityMetaTable
		resIntegrityMetas    []*corespdb.ShadowIntegrityMeta
		err                  error
	)
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureListObjectIntegrity).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureListObjectIntegrity).Observe(
				time.Since(startTime).Seconds())
		} else {
			metrics.SPDBCounter.WithLabelValues(SPDBSuccessListObjectIntegrity).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBSuccessListObjectIntegrity).Observe(
				time.Since(startTime).Seconds())
		}
	}()

	err = s.db.Table(ShadowIntegrityMetaTableName).
		Select("*").
		Limit(ListShadowIntegrityMetaDefaultSize).
		Order("object_id asc").
		Find(&shadowIntegrityMetas).Error

	for _, sim := range shadowIntegrityMetas {
		integrityChecksum, err := hex.DecodeString(sim.IntegrityChecksum)
		if err != nil {
			return nil, err
		}
		meta := &corespdb.ShadowIntegrityMeta{
			ObjectID:          sim.ObjectID,
			RedundancyIndex:   sim.RedundancyIndex,
			IntegrityChecksum: integrityChecksum,
			Version:           sim.Version,
		}
		meta.PieceChecksumList, err = util.StringToBytesSlice(sim.PieceChecksumList)
		if err != nil {
			return nil, err
		}
		resIntegrityMetas = append(resIntegrityMetas, meta)
	}
	return resIntegrityMetas, err
}

// UpdateShadowPieceChecksum 1) If the ShadowIntegrityMetaTable does not exist, it will be created.
// 2) If the ShadowIntegrityMetaTable already exists, it will be appended to the existing PieceChecksumList.
func (s *SpDBImpl) UpdateShadowPieceChecksum(objectID uint64, redundancyIndex int32, checksum []byte, version int64, dataLength uint64) (err error) {
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureUpdatePieceChecksum).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureUpdatePieceChecksum).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessUpdatePieceChecksum).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessUpdatePieceChecksum).Observe(
			time.Since(startTime).Seconds())
	}()

	integrityMeta, err := s.GetShadowObjectIntegrity(objectID, redundancyIndex)
	var checksums [][]byte
	var integrity []byte
	if err == gorm.ErrRecordNotFound {
		integrityMetaNew := &corespdb.ShadowIntegrityMeta{
			ObjectID:          objectID,
			RedundancyIndex:   redundancyIndex,
			PieceChecksumList: append(checksums, checksum),
			IntegrityChecksum: integrity,
			Version:           version,
			PieceSize:         dataLength,
		}
		err = s.SetShadowObjectIntegrity(integrityMetaNew)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		newChecksums := append(integrityMeta.PieceChecksumList, checksum)
		integrityMeta.PieceChecksumList = newChecksums
		result := s.db.Table(ShadowIntegrityMetaTableName).Where("object_id = ? and redundancy_index = ?", objectID, redundancyIndex).
			Updates(&ShadowIntegrityMetaTable{
				PieceChecksumList: util.BytesSliceToString(newChecksums),
				PieceSize:         integrityMeta.PieceSize + dataLength,
			})
		if result.Error != nil {
			return fmt.Errorf("failed to update integrity meta table: %s", result.Error)
		}
	}
	return nil
}

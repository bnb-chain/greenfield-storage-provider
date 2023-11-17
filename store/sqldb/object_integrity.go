package sqldb

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

const (
	// SPDBSuccessGetObjectIntegrity defines the metrics label of successfully get object integrity
	SPDBSuccessGetObjectIntegrity = "get_object_integrity_meta_success"
	// SPDBFailureGetObjectIntegrity defines the metrics label of unsuccessfully get object integrity
	SPDBFailureGetObjectIntegrity = "get_object_integrity_meta_failure"
	// SPDBSuccessSetObjectIntegrity defines the metrics label of successfully set object integrity
	SPDBSuccessSetObjectIntegrity = "set_object_integrity_meta_success"
	// SPDBFailureSetObjectIntegrity defines the metrics label of unsuccessfully set object integrity
	SPDBFailureSetObjectIntegrity = "set_object_integrity_meta_failure"
	// SPDBSuccessDelObjectIntegrity defines the metrics label of successfully del object integrity
	SPDBSuccessDelObjectIntegrity = "del_object_integrity_meta_success"
	// SPDBFailureDelObjectIntegrity defines the metrics label of unsuccessfully del object integrity
	SPDBFailureDelObjectIntegrity = "del_object_integrity_meta_failure"

	// SPDBSuccessUpdatePieceChecksum defines the metrics label of successfully update object piece checksum
	SPDBSuccessUpdatePieceChecksum = "append_object_checksum_integrity_success"
	// SPDBFailureUpdatePieceChecksum defines the metrics label of unsuccessfully update object piece checksum
	SPDBFailureUpdatePieceChecksum = "append_object_checksum_integrity_failure"
	// SPDBSuccessUpdateIntegrityChecksum defines the metrics label of successfully update object integrity checksum
	SPDBSuccessUpdateIntegrityChecksum = "append_object_checksum_integrity_success"
	// SPDBFailureUpdateIntegrityChecksum defines the metrics label of unsuccessfully update object integrity checksum
	SPDBFailureUpdateIntegrityChecksum = "append_object_checksum_integrity_failure"
	// SPDBSuccessGetReplicatePieceChecksum defines the metrics label of successfully get replicate piece checksum
	SPDBSuccessGetReplicatePieceChecksum = "get_replicate_piece_checksum_success"
	// SPDBFailureGetReplicatePieceChecksum defines the metrics label of unsuccessfully get replicate piece checksum
	SPDBFailureGetReplicatePieceChecksum = "get_replicate_piece_checksum_failure"
	// SPDBSuccessSetReplicatePieceChecksum defines the metrics label of successfully set replicate piece checksum
	SPDBSuccessSetReplicatePieceChecksum = "set_replicate_piece_checksum_success"
	// SPDBFailureSetReplicatePieceChecksum defines the metrics label of unsuccessfully set replicate piece checksum
	SPDBFailureSetReplicatePieceChecksum = "set_replicate_piece_checksum_failure"
	// SPDBSuccessDelReplicatePieceChecksum defines the metrics label of successfully del replicate piece checksum
	SPDBSuccessDelReplicatePieceChecksum = "del_replicate_piece_checksum_success"
	// SPDBFailureDelReplicatePieceChecksum defines the metrics label of unsuccessfully del replicate piece checksum
	SPDBFailureDelReplicatePieceChecksum = "del_replicate_piece_checksum_failure"

	// SPDBSuccessGetAllReplicatePieceChecksum defines the metrics label of successfully get all replicate piece checksum
	SPDBSuccessGetAllReplicatePieceChecksum = "get_all_replicate_piece_checksum_success"
	// SPDBFailureGetAllReplicatePieceChecksum defines the metrics label of unsuccessfully get all replicate piece checksum
	SPDBFailureGetAllReplicatePieceChecksum = "get_all_replicate_piece_checksum_failure"
	// SPDBSuccessDelAllReplicatePieceChecksum defines the metrics label of successfully del all replicate piece checksum
	SPDBSuccessDelAllReplicatePieceChecksum = "del_all_replicate_piece_checksum_success"
	// SPDBFailureDelAllReplicatePieceChecksum defines the metrics label of unsuccessfully del all replicate piece checksum
	SPDBFailureDelAllReplicatePieceChecksum = "del_all_replicate_piece_checksum_failure"

	// SPDBSuccessListReplicatePieceChecksumByObjectID defines the metrics label of successfully list replicate piece checksum by object id
	SPDBSuccessListReplicatePieceChecksumByObjectID = "list_replicate_piece_checksum_by_object_id_success"
	// SPDBFailureListReplicatePieceChecksumByObjectID defines the metrics label of unsuccessfully list replicate piece checksum by object id
	SPDBFailureListReplicatePieceChecksumByObjectID = "list_replicate_piece_checksum_by_object_id_failure"
)

// GetObjectIntegrity returns the integrity hash info
func (s *SpDBImpl) GetObjectIntegrity(objectID uint64, redundancyIndex int32) (meta *corespdb.IntegrityMeta, err error) {
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureGetObjectIntegrity).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureGetObjectIntegrity).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessGetObjectIntegrity).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessGetObjectIntegrity).Observe(
			time.Since(startTime).Seconds())
	}()

	queryReturn := &IntegrityMetaTable{}
	shardTableName := GetIntegrityMetasTableName(objectID)
	result := s.db.Table(shardTableName).Model(&IntegrityMetaTable{}).
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

	meta = &corespdb.IntegrityMeta{
		ObjectID:          queryReturn.ObjectID,
		RedundancyIndex:   queryReturn.RedundancyIndex,
		IntegrityChecksum: integrityChecksum,
	}
	meta.PieceChecksumList, err = util.StringToBytesSlice(queryReturn.PieceChecksumList)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

func MysqlErrCode(err error) int {
	mysqlErr, ok := err.(*mysql.MySQLError)
	if !ok {
		return 0
	}
	return int(mysqlErr.Number)
}

var (
	ErrDuplicateEntryCode = 1062
)

// SetObjectIntegrity puts(overwrites) integrity hash info to db
func (s *SpDBImpl) SetObjectIntegrity(meta *corespdb.IntegrityMeta) (err error) {
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureSetObjectIntegrity).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureSetObjectIntegrity).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessSetObjectIntegrity).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessSetObjectIntegrity).Observe(
			time.Since(startTime).Seconds())
	}()

	insertIntegrityMetaRecord := &IntegrityMetaTable{
		ObjectID:          meta.ObjectID,
		RedundancyIndex:   meta.RedundancyIndex,
		PieceChecksumList: util.BytesSliceToString(meta.PieceChecksumList),
		IntegrityChecksum: hex.EncodeToString(meta.IntegrityChecksum),
	}
	shardTableName := GetIntegrityMetasTableName(meta.ObjectID)
	result := s.db.Table(shardTableName).Create(insertIntegrityMetaRecord)
	if result.Error != nil && MysqlErrCode(result.Error) == ErrDuplicateEntryCode {
		return nil
	}
	if result.Error != nil || result.RowsAffected != 1 {
		err = fmt.Errorf("failed to insert integrity meta record: %s", result.Error)
		return err
	}
	return nil
}

// UpdateIntegrityChecksum update integrity hash info to db, TODO: if not exit, whether create it
func (s *SpDBImpl) UpdateIntegrityChecksum(meta *corespdb.IntegrityMeta) (err error) {
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureUpdateIntegrityChecksum).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureUpdateIntegrityChecksum).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessUpdateIntegrityChecksum).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessUpdateIntegrityChecksum).Observe(
			time.Since(startTime).Seconds())
	}()

	shardTableName := GetIntegrityMetasTableName(meta.ObjectID)
	result := s.db.Table(shardTableName).Where("object_id = ? and redundancy_index = ?", meta.ObjectID, meta.RedundancyIndex).
		Updates(&IntegrityMetaTable{
			IntegrityChecksum: hex.EncodeToString(meta.IntegrityChecksum),
		})
	if result.Error != nil {
		return fmt.Errorf("failed to update integrity checksum for integrity meta table: %s", result.Error)
	}

	return nil
}

// DeleteObjectIntegrity deletes integrity meta info.
func (s *SpDBImpl) DeleteObjectIntegrity(objectID uint64, redundancyIndex int32) (err error) {
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureDelObjectIntegrity).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureDelObjectIntegrity).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessDelObjectIntegrity).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessDelObjectIntegrity).Observe(
			time.Since(startTime).Seconds())
	}()

	shardTableName := GetIntegrityMetasTableName(objectID)
	err = s.db.Table(shardTableName).Delete(&IntegrityMetaTable{
		ObjectID:        objectID, // should be the primary key
		RedundancyIndex: redundancyIndex,
	}).Error
	return err
}

type ByRedundancyIndexAndObjectID []*corespdb.IntegrityMeta

func (b ByRedundancyIndexAndObjectID) Len() int { return len(b) }

// Less we want to sort as ascending here
func (b ByRedundancyIndexAndObjectID) Less(i, j int) bool {
	if b[i].ObjectID == b[j].ObjectID {
		return b[i].RedundancyIndex < b[j].RedundancyIndex
	}
	return b[i].ObjectID < b[j].ObjectID
}
func (b ByRedundancyIndexAndObjectID) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

// ListIntegrityMetaByObjectIDRange list objects info by a block number range
func (s *SpDBImpl) ListIntegrityMetaByObjectIDRange(startObjectID int64, endObjectID int64, includePrivate bool) ([]*corespdb.IntegrityMeta, error) {
	var (
		totalObjects []*corespdb.IntegrityMeta
		objects      []*corespdb.IntegrityMeta
		err          error
	)
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureGetObjectIntegrity).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureGetObjectIntegrity).Observe(
				time.Since(startTime).Seconds())
		} else {
			metrics.SPDBCounter.WithLabelValues(SPDBSuccessGetObjectIntegrity).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBSuccessGetObjectIntegrity).Observe(
				time.Since(startTime).Seconds())
		}
	}()

	if includePrivate {
		for i := 0; i < IntegrityMetasNumberOfShards; i++ {
			err = s.db.Table(GetIntegrityMetasTableNameByShardNumber(i)).
				Select("*").
				Where("object_id >= ? and object_id <= ?", startObjectID, endObjectID).
				Limit(ListObjectsDefaultSize).
				Order("object_id,redundancy_index asc").
				Find(&objects).Error
			totalObjects = append(totalObjects, objects...)
		}
	} else {
		for i := 0; i < IntegrityMetasNumberOfShards; i++ {
			integrityMetasTableName := GetIntegrityMetasTableNameByShardNumber(i)
			joins := fmt.Sprintf("right join buckets on buckets.bucket_id = %s.bucket_id", integrityMetasTableName)
			order := fmt.Sprintf("%s.object_id, %s.redundancy_index asc", integrityMetasTableName, integrityMetasTableName)
			where := fmt.Sprintf("%s.object_id >= ? and %s.object_id <= ? and "+
				"((%s.visibility='VISIBILITY_TYPE_PUBLIC_READ') or "+
				"(%s.visibility='VISIBILITY_TYPE_INHERIT' and buckets.visibility='VISIBILITY_TYPE_PUBLIC_READ'))",
				integrityMetasTableName, integrityMetasTableName, integrityMetasTableName, integrityMetasTableName)

			err = s.db.Table(integrityMetasTableName).
				Select(integrityMetasTableName+".*").
				Joins(joins).
				Where(where, startObjectID, endObjectID).
				Limit(ListObjectsDefaultSize).
				Order(order).
				Find(&objects).Error
			totalObjects = append(totalObjects, objects...)
		}
	}

	sort.Sort(ByRedundancyIndexAndObjectID(totalObjects))

	if len(totalObjects) > ListObjectsDefaultSize {
		totalObjects = totalObjects[0:ListObjectsDefaultSize]
	}
	return totalObjects, err
}

// UpdatePieceChecksum 1) If the IntegrityMetaTable does not exist, it will be created.
// 2) If the IntegrityMetaTable already exists, it will be appended to the existing PieceChecksumList.
func (s *SpDBImpl) UpdatePieceChecksum(objectID uint64, redundancyIndex int32, checksum []byte) (err error) {
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

	integrityMeta, err := s.GetObjectIntegrity(objectID, redundancyIndex)
	shardTableName := GetIntegrityMetasTableName(objectID)
	var checksums [][]byte
	var integrity []byte
	if err == gorm.ErrRecordNotFound {
		integrityMetaNew := &corespdb.IntegrityMeta{
			ObjectID:          objectID,
			RedundancyIndex:   redundancyIndex,
			PieceChecksumList: append(checksums, checksum),
			IntegrityChecksum: integrity,
		}
		err = s.SetObjectIntegrity(integrityMetaNew)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		newChecksums := append(integrityMeta.PieceChecksumList, checksum)
		integrityMeta.PieceChecksumList = newChecksums
		result := s.db.Table(shardTableName).Where("object_id = ? and redundancy_index = ?", objectID, redundancyIndex).
			Updates(&IntegrityMetaTable{
				PieceChecksumList: util.BytesSliceToString(newChecksums),
			})
		if result.Error != nil {
			return fmt.Errorf("failed to update integrity meta table: %s", result.Error)
		}
	}
	return nil
}

// GetReplicatePieceChecksum gets replicate piece checksum.
func (s *SpDBImpl) GetReplicatePieceChecksum(objectID uint64, segmentIdx uint32, redundancyIdx int32) ([]byte, error) {
	var (
		err           error
		queryReturn   PieceHashTable
		pieceChecksum []byte
	)
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureGetReplicatePieceChecksum).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureGetReplicatePieceChecksum).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessGetReplicatePieceChecksum).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessGetReplicatePieceChecksum).Observe(
			time.Since(startTime).Seconds())
	}()

	if err = s.db.Model(&PieceHashTable{}).
		Where("object_id = ? and segment_index = ? and redundancy_index = ?", objectID, segmentIdx, redundancyIdx).
		First(&queryReturn).Error; err != nil {
		return nil, err
	}
	if pieceChecksum, err = hex.DecodeString(queryReturn.PieceChecksum); err != nil {
		return nil, err
	}
	return pieceChecksum, nil
}

// SetReplicatePieceChecksum sets replicate checksum.
func (s *SpDBImpl) SetReplicatePieceChecksum(objectID uint64, segmentIdx uint32, redundancyIdx int32, checksum []byte) error {
	var (
		result *gorm.DB
		err    error
	)
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureSetReplicatePieceChecksum).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureSetReplicatePieceChecksum).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessSetReplicatePieceChecksum).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessSetReplicatePieceChecksum).Observe(
			time.Since(startTime).Seconds())
	}()

	insertPieceHash := &PieceHashTable{
		ObjectID:        objectID,
		SegmentIndex:    segmentIdx,
		RedundancyIndex: redundancyIdx,
		PieceChecksum:   hex.EncodeToString(checksum),
	}
	result = s.db.Create(insertPieceHash)
	if result.Error != nil && MysqlErrCode(result.Error) == ErrDuplicateEntryCode {
		return nil
	}
	if result.Error != nil || result.RowsAffected != 1 {
		err = fmt.Errorf("failed to insert piece hash record: %s", result.Error)
		return err
	}
	return nil
}

// DeleteReplicatePieceChecksum deletes piece checksum.
func (s *SpDBImpl) DeleteReplicatePieceChecksum(objectID uint64, segmentIdx uint32, redundancyIdx int32) (err error) {
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureDelReplicatePieceChecksum).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureDelReplicatePieceChecksum).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessDelReplicatePieceChecksum).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessDelReplicatePieceChecksum).Observe(
			time.Since(startTime).Seconds())
	}()

	err = s.db.Delete(&PieceHashTable{
		ObjectID:        objectID,
		SegmentIndex:    segmentIdx,
		RedundancyIndex: redundancyIdx,
	}).Error
	return err
}

// GetAllReplicatePieceChecksum gets all the piece checksums.
func (s *SpDBImpl) GetAllReplicatePieceChecksum(objectID uint64, redundancyIdx int32, pieceCount uint32) ([][]byte, error) {
	var (
		segmentIdx        uint32
		err               error
		pieceChecksum     []byte
		pieceChecksumList = make([][]byte, 0)
	)
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureGetAllReplicatePieceChecksum).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureGetAllReplicatePieceChecksum).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessGetAllReplicatePieceChecksum).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessGetAllReplicatePieceChecksum).Observe(
			time.Since(startTime).Seconds())
	}()

	for segmentIdx < pieceCount {
		pieceChecksum, err = s.GetReplicatePieceChecksum(objectID, segmentIdx, redundancyIdx)
		if err != nil {
			return nil, err
		}
		pieceChecksumList = append(pieceChecksumList, pieceChecksum)
		segmentIdx++
	}
	return pieceChecksumList, nil
}

// GetAllReplicatePieceChecksumOptimized gets all replicate piece checksums for a given objectID and redundancyIdx.
func (s *SpDBImpl) GetAllReplicatePieceChecksumOptimized(objectID uint64, redundancyIdx int32, pieceCount uint32) ([][]byte, error) {
	var (
		err            error
		queryReturns   []PieceHashTable
		pieceChecksums [][]byte
	)
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureGetAllReplicatePieceChecksum).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureGetAllReplicatePieceChecksum).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessGetAllReplicatePieceChecksum).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessGetAllReplicatePieceChecksum).Observe(
			time.Since(startTime).Seconds())
	}()

	if err = s.db.Model(&PieceHashTable{}).
		Where("object_id = ? and redundancy_index = ?", objectID, redundancyIdx).
		Limit(int(pieceCount)).
		Find(&queryReturns).Error; err != nil {
		return nil, err
	}

	pieceChecksums = make([][]byte, 0, len(queryReturns))
	for _, queryReturn := range queryReturns {
		pieceChecksum, err := hex.DecodeString(queryReturn.PieceChecksum)
		if err != nil {
			return nil, err
		}
		pieceChecksums = append(pieceChecksums, pieceChecksum)
	}

	return pieceChecksums, nil
}

// DeleteAllReplicatePieceChecksum deletes all the piece checksum.
func (s *SpDBImpl) DeleteAllReplicatePieceChecksum(objectID uint64, redundancyIdx int32, pieceCount uint32) error {
	var (
		segmentIdx uint32
		err        error
	)
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureDelAllReplicatePieceChecksum).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureDelAllReplicatePieceChecksum).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessDelAllReplicatePieceChecksum).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessDelAllReplicatePieceChecksum).Observe(
			time.Since(startTime).Seconds())
	}()

	for segmentIdx < pieceCount {
		err = s.DeleteReplicatePieceChecksum(objectID, segmentIdx, redundancyIdx)
		if err != nil {
			return err
		}
		segmentIdx++
	}
	return nil
}

// DeleteAllReplicatePieceChecksumOptimized deletes all piece checksums for a given objectID.
func (s *SpDBImpl) DeleteAllReplicatePieceChecksumOptimized(objectID uint64, redundancyIdx int32) (err error) {
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureDelAllReplicatePieceChecksum).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureDelAllReplicatePieceChecksum).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessDelAllReplicatePieceChecksum).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessDelAllReplicatePieceChecksum).Observe(
			time.Since(startTime).Seconds())
	}()

	err = s.db.Delete(&PieceHashTable{
		ObjectID:        objectID,
		RedundancyIndex: redundancyIdx,
	}).Error
	return err
}

// ListReplicatePieceChecksumByObjectIDRange gets all replicate piece checksums for a given objectID and redundancyIdx.
func (s *SpDBImpl) ListReplicatePieceChecksumByObjectIDRange(startObjectID int64, endObjectID int64) ([]*corespdb.GCPieceMeta, error) {
	var (
		err          error
		queryReturns []PieceHashTable
	)
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureListReplicatePieceChecksumByObjectID).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureListReplicatePieceChecksumByObjectID).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessListReplicatePieceChecksumByObjectID).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessListReplicatePieceChecksumByObjectID).Observe(
			time.Since(startTime).Seconds())
	}()

	if err = s.db.Model(&PieceHashTable{}).
		Where("object_id >= ? and object_id <= ?", startObjectID, endObjectID).
		Limit(ListObjectsDefaultSize).
		Find(&queryReturns).Error; err != nil {
		return nil, err
	}

	var pieceMetas []*corespdb.GCPieceMeta

	for _, queryReturn := range queryReturns {
		pieceMeta := &corespdb.GCPieceMeta{
			ObjectID:        queryReturn.ObjectID,
			SegmentIndex:    queryReturn.SegmentIndex,
			RedundancyIndex: queryReturn.RedundancyIndex,
			PieceChecksum:   queryReturn.PieceChecksum,
		}
		pieceMetas = append(pieceMetas, pieceMeta)
	}

	return pieceMetas, nil
}

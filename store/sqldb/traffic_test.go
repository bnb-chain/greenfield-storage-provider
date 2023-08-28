package sqldb

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
)

const noRecordInTrafficErr = "failed to query bucket traffic table: record not found"

func TestSpDBImpl_CheckQuotaAndAddReadRecordSuccess1(t *testing.T) {
	record := &corespdb.ReadRecord{
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
		ReadTimestampUs: 1,
	}

	newRecord := &corespdb.ReadRecord{
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        2,
		ReadTimestampUs: 2,
	}

	quota := &corespdb.BucketQuota{
		ChargedQuotaSize: 20,
		FreeQuotaSize:    10,
	}

	b := BucketTrafficTable{
		BucketID:              2,
		BucketName:            "mockBucketName",
		ReadConsumedSize:      10,
		FreeQuotaConsumedSize: 100,
		FreeQuotaSize:         25,
		ChargedQuotaSize:      30,
		ModifiedTime:          time.Now(),
	}

	yearMonth := TimestampYearMonth(b.ModifiedTime.Unix())
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT * FROM `bucket_traffic` WHERE bucket_id = ?  and month = ? FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"bucket_id", "year_month", "bucket_name", "read_consumed_size", "free_quota_consumed_size",
		"free_quota_size", "charged_quota_size", "modified_time"}).AddRow(b.BucketID, yearMonth, b.BucketName, b.ReadConsumedSize,
		b.FreeQuotaConsumedSize, b.FreeQuotaSize, b.ChargedQuotaSize, b.ModifiedTime))

	mock.ExpectExec("UPDATE `bucket_traffic` SET `charged_quota_size`=?,`modified_time`=? WHERE `bucket_id` = ? ").
		WithArgs(20, sqlmock.AnyArg(), record.BucketID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE `bucket_traffic` SET `read_consumed_size`=?,`free_quota_consumed_size`=?,`modified_time`=? WHERE `bucket_id` = ? ").
		WithArgs(sqlmock.AnyArg(), 12, sqlmock.AnyArg(), record.BucketID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `read_record` (`bucket_id`,`object_id`,`user_address`,`read_timestamp_us`,`bucket_name`,`object_name`,`read_size`) VALUES (?,?,?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := s.CheckQuotaAndAddReadRecord(newRecord, quota)

	assert.Nil(t, err)
}

func TestSpDBImpl_CheckQuotaAndAddReadRecordSuccess2(t *testing.T) {
	t.Log("Success case description: check quota, no error")
	record := &corespdb.ReadRecord{
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
		ReadTimestampUs: 1,
	}
	quota := &corespdb.BucketQuota{
		ChargedQuotaSize: 30,
		FreeQuotaSize:    30,
	}
	b := BucketTrafficTable{
		BucketID:              2,
		BucketName:            "mockBucketName",
		ReadConsumedSize:      29,
		FreeQuotaConsumedSize: 24,
		FreeQuotaSize:         25,
		ChargedQuotaSize:      30,
		ModifiedTime:          time.Now(),
	}

	yearMonth := TimestampYearMonth(b.ModifiedTime.Unix())
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT * FROM `bucket_traffic` WHERE bucket_id = ?  and month = ? FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"bucket_id", "year_month", "bucket_name", "read_consumed_size", "free_quota_consumed_size",
		"free_quota_size", "charged_quota_size", "modified_time"}).AddRow(b.BucketID, yearMonth, b.BucketName, b.ReadConsumedSize,
		b.FreeQuotaConsumedSize, b.FreeQuotaSize, b.ChargedQuotaSize, b.ModifiedTime))

	mock.ExpectExec("UPDATE `bucket_traffic` SET `read_consumed_size`=?,`free_quota_consumed_size`=?,`modified_time`=? WHERE `bucket_id` = ?").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `read_record` (`bucket_id`,`object_id`,`user_address`,`read_timestamp_us`,`bucket_name`,`object_name`,`read_size`) VALUES (?,?,?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.CheckQuotaAndAddReadRecord(record, quota)
	assert.Nil(t, err)
}

func TestSpDBImpl_CheckQuotaAndAddReadRecordFailure1(t *testing.T) {
	t.Log("Failure case description: mock get bucket traffic returns error")
	record := &corespdb.ReadRecord{
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
		ReadTimestampUs: 1,
	}
	quota := &corespdb.BucketQuota{
		ChargedQuotaSize: 20,
		FreeQuotaSize:    10,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `bucket_traffic` WHERE bucket_id = ? and month = ? ORDER BY `bucket_traffic`.`bucket_id` LIMIT 1").
		WillReturnError(mockDBInternalError)
	err := s.CheckQuotaAndAddReadRecord(record, quota)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_CheckQuotaAndAddReadRecordFailure2(t *testing.T) {
	t.Log("Failure case description: mock get bucket traffic no record")
	record := &corespdb.ReadRecord{
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
		ReadTimestampUs: 1,
	}
	quota := &corespdb.BucketQuota{
		ChargedQuotaSize: 20,
		FreeQuotaSize:    10,
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT * FROM `bucket_traffic` WHERE bucket_id = ?  and month = ? FOR UPDATE").WillReturnError(gorm.ErrRecordNotFound)

	err := s.CheckQuotaAndAddReadRecord(record, quota)
	assert.Contains(t, err.Error(), noRecordInTrafficErr)
}

func TestSpDBImpl_CheckQuotaAndAddReadRecordFailure3(t *testing.T) {
	t.Log("Failure case description: ChargedQuotaSize mock update bucket traffic table returns error")
	record := &corespdb.ReadRecord{
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
		ReadTimestampUs: 1,
	}
	quota := &corespdb.BucketQuota{
		ChargedQuotaSize: 20,
		FreeQuotaSize:    10,
	}
	b := BucketTrafficTable{
		BucketID:              2,
		BucketName:            "mockBucketName",
		ReadConsumedSize:      10,
		FreeQuotaConsumedSize: 100,
		FreeQuotaSize:         25,
		ChargedQuotaSize:      30,
		ModifiedTime:          time.Now(),
	}

	yearMonth := TimestampYearMonth(b.ModifiedTime.Unix())
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT * FROM `bucket_traffic` WHERE bucket_id = ?  and month = ? FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"bucket_id", "year_month", "bucket_name", "read_consumed_size", "free_quota_consumed_size",
		"free_quota_size", "charged_quota_size", "modified_time"}).AddRow(b.BucketID, yearMonth, b.BucketName, b.ReadConsumedSize,
		b.FreeQuotaConsumedSize, b.FreeQuotaSize, b.ChargedQuotaSize, b.ModifiedTime))

	mock.ExpectExec("UPDATE `bucket_traffic` SET `charged_quota_size`=?,`modified_time`=?  WHERE `bucket_id` = ?").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.CheckQuotaAndAddReadRecord(record, quota)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_CheckQuotaAndAddReadRecordFailure4(t *testing.T) {
	t.Log("Failure case description:  mock update bucket traffic table returns error")
	record := &corespdb.ReadRecord{
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
		ReadTimestampUs: 1,
	}
	quota := &corespdb.BucketQuota{
		ChargedQuotaSize: 20,
		FreeQuotaSize:    10,
	}
	b := BucketTrafficTable{
		BucketID:              2,
		BucketName:            "mockBucketName",
		ReadConsumedSize:      10,
		FreeQuotaConsumedSize: 100,
		FreeQuotaSize:         25,
		ChargedQuotaSize:      30,
		ModifiedTime:          time.Now(),
	}
	yearMonth := TimestampYearMonth(b.ModifiedTime.Unix())
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT * FROM `bucket_traffic` WHERE bucket_id = ?  and month = ? FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"bucket_id", "year_month", "bucket_name", "read_consumed_size", "free_quota_consumed_size",
		"free_quota_size", "charged_quota_size", "modified_time"}).AddRow(b.BucketID, yearMonth, b.BucketName, b.ReadConsumedSize,
		b.FreeQuotaConsumedSize, b.FreeQuotaSize, b.ChargedQuotaSize, b.ModifiedTime))

	mock.ExpectExec("UPDATE `bucket_traffic` SET `charged_quota_size`=?,`modified_time`=? WHERE `bucket_id` = ?").
		WithArgs(20, sqlmock.AnyArg(), record.BucketID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE `bucket_traffic` SET `read_consumed_size`=?,`free_quota_consumed_size`=?,`modified_time`=? WHERE `bucket_id` = ?").
		WithArgs(sqlmock.AnyArg(), 11, sqlmock.AnyArg(), record.BucketID).
		WillReturnError(mockDBInternalError)

	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.CheckQuotaAndAddReadRecord(record, quota)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_CheckQuotaAndAddReadRecordFailure5(t *testing.T) {
	t.Log("Failure case description:  mock update bucket traffic table returns error")
	record := &corespdb.ReadRecord{
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
		ReadTimestampUs: 1,
	}
	quota := &corespdb.BucketQuota{
		ChargedQuotaSize: 20,
		FreeQuotaSize:    10,
	}
	b := BucketTrafficTable{
		BucketID:              2,
		BucketName:            "mockBucketName",
		ReadConsumedSize:      10,
		FreeQuotaConsumedSize: 100,
		FreeQuotaSize:         25,
		ChargedQuotaSize:      30,
		ModifiedTime:          time.Now(),
	}
	yearMonth := TimestampYearMonth(b.ModifiedTime.Unix())
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT * FROM `bucket_traffic` WHERE bucket_id = ?  and month = ? FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"bucket_id", "year_month", "bucket_name", "read_consumed_size", "free_quota_consumed_size",
		"free_quota_size", "charged_quota_size", "modified_time"}).AddRow(b.BucketID, yearMonth, b.BucketName, b.ReadConsumedSize,
		b.FreeQuotaConsumedSize, b.FreeQuotaSize, b.ChargedQuotaSize, b.ModifiedTime))

	mock.ExpectExec("UPDATE `bucket_traffic` SET `charged_quota_size`=?,`modified_time`=? WHERE `bucket_id` = ?").
		WithArgs(20, sqlmock.AnyArg(), record.BucketID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE `bucket_traffic` SET `read_consumed_size`=?,`free_quota_consumed_size`=?,`modified_time`=? WHERE `bucket_id` = ?").
		WithArgs(sqlmock.AnyArg(), 11, sqlmock.AnyArg(), record.BucketID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `read_record` (`bucket_id`,`object_id`,`user_address`,`read_timestamp_us`,`bucket_name`,`object_name`,`read_size`) VALUES (?,?,?,?,?,?,?)").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.CheckQuotaAndAddReadRecord(record, quota)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_CheckQuotaAndAddReadRecordFailure6(t *testing.T) {
	t.Log("Failure case description: check quota not enough")
	record := &corespdb.ReadRecord{
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
		ReadTimestampUs: 1,
	}
	quota := &corespdb.BucketQuota{
		ChargedQuotaSize: 20,
		FreeQuotaSize:    30,
	}
	b := BucketTrafficTable{
		BucketID:              2,
		BucketName:            "mockBucketName",
		ReadConsumedSize:      60,
		FreeQuotaConsumedSize: 25,
		FreeQuotaSize:         25,
		ChargedQuotaSize:      30,
		ModifiedTime:          time.Now(),
	}
	yearMonth := TimestampYearMonth(b.ModifiedTime.Unix())
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT * FROM `bucket_traffic` WHERE bucket_id = ?  and month = ? FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"bucket_id", "year_month", "bucket_name", "read_consumed_size", "free_quota_consumed_size",
		"free_quota_size", "charged_quota_size", "modified_time"}).AddRow(b.BucketID, yearMonth, b.BucketName, b.ReadConsumedSize,
		b.FreeQuotaConsumedSize, b.FreeQuotaSize, b.ChargedQuotaSize, b.ModifiedTime))

	mock.ExpectExec("UPDATE `bucket_traffic` SET `charged_quota_size`=?,`modified_time`=? WHERE `bucket_id` = ?").
		WithArgs(20, sqlmock.AnyArg(), record.BucketID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()
	err := s.CheckQuotaAndAddReadRecord(record, quota)
	assert.Equal(t, ErrCheckQuotaEnough, err)
}

func TestSpDBImpl_CheckQuotaAndAddReadRecordFailure7(t *testing.T) {
	t.Log("Failure case description: ChargedQuotaSize mock update bucket traffic table returns error")
	record := &corespdb.ReadRecord{
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
		ReadTimestampUs: 1,
	}
	quota := &corespdb.BucketQuota{
		ChargedQuotaSize: 20,
		FreeQuotaSize:    10,
	}
	b := BucketTrafficTable{
		BucketID:              2,
		BucketName:            "mockBucketName",
		ReadConsumedSize:      10,
		FreeQuotaConsumedSize: 100,
		FreeQuotaSize:         25,
		ChargedQuotaSize:      30,
		ModifiedTime:          time.Now(),
	}
	yearMonth := TimestampYearMonth(b.ModifiedTime.Unix())
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT * FROM `bucket_traffic` WHERE bucket_id = ?  and month = ? FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"bucket_id", "year_month", "bucket_name", "read_consumed_size", "free_quota_consumed_size",
		"free_quota_size", "charged_quota_size", "modified_time"}).AddRow(b.BucketID, yearMonth, b.BucketName, b.ReadConsumedSize,
		b.FreeQuotaConsumedSize, b.FreeQuotaSize, b.ChargedQuotaSize, b.ModifiedTime))

	mock.ExpectExec("UPDATE `bucket_traffic` SET `charged_quota_size`=?,`modified_time`=?  WHERE `bucket_id` = ?").
		WithArgs(20, sqlmock.AnyArg(), record.BucketID).
		WillReturnResult(sqlmock.NewResult(2, 2))
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.CheckQuotaAndAddReadRecord(record, quota)
	assert.Contains(t, err.Error(), "has affected more than one rows")
}

func TestSpDBImpl_InitBucketTrafficSuccess(t *testing.T) {
	quota := &corespdb.BucketQuota{
		ChargedQuotaSize: 20,
		FreeQuotaSize:    10,
	}

	record := &corespdb.ReadRecord{
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
		ReadTimestampUs: time.Now().Unix(),
	}
	b := BucketTrafficTable{
		BucketID:              2,
		BucketName:            "mockBucketName",
		ReadConsumedSize:      10,
		FreeQuotaConsumedSize: 100,
		FreeQuotaSize:         25,
		ChargedQuotaSize:      30,
		ModifiedTime:          time.Now(),
	}

	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT * FROM `bucket_traffic` WHERE bucket_id = ? ORDER BY `bucket_traffic`.`bucket_id` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"bucket_id", "bucket_name", "read_consumed_size", "free_quota_consumed_size",
			"free_quota_size", "charged_quota_size", "modified_time"}).AddRow(b.BucketID, b.BucketName, b.ReadConsumedSize,
			b.FreeQuotaConsumedSize, b.FreeQuotaSize, b.ChargedQuotaSize, b.ModifiedTime))

	mock.ExpectQuery("SELECT * FROM `bucket_traffic` WHERE bucket_id = ? ORDER BY month DESC LIMIT 1").WillReturnRows(sqlmock.NewRows([]string{"bucket_id", "bucket_name", "read_consumed_size", "free_quota_consumed_size",
		"free_quota_size", "charged_quota_size", "modified_time"}).AddRow(b.BucketID, b.BucketName, b.ReadConsumedSize,
		b.FreeQuotaConsumedSize, b.FreeQuotaSize, b.ChargedQuotaSize, b.ModifiedTime))

	mock.ExpectExec("INSERT INTO `bucket_traffic` (`bucket_id`,`month`,`bucket_name`,`read_consumed_size`,`free_quota_consumed_size`,`free_quota_size`,`charged_quota_size`,`modified_time`) VALUES (?,?,?,?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := s.InitBucketTraffic(record, quota)
	assert.Nil(t, err)
}

func TestSpDBImpl_InitBucketTrafficFailure(t *testing.T) {
	quota := &corespdb.BucketQuota{
		ChargedQuotaSize: 20,
		FreeQuotaSize:    10,
	}

	record := &corespdb.ReadRecord{
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
		ReadTimestampUs: time.Now().Unix(),
	}

	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT * FROM `bucket_traffic` WHERE bucket_id = ? ORDER BY `bucket_traffic`.`bucket_id` LIMIT 1").
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectExec("INSERT INTO `bucket_traffic` (`bucket_id`,`month`,`bucket_name`,`read_consumed_size`,`free_quota_consumed_size`,`free_quota_size`,`charged_quota_size`,`modified_time`) VALUES (?,?,?,?,?,?,?,?)").
		WillReturnError(mockDBInternalError)

	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.InitBucketTraffic(record, quota)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_InitBucketTrafficFailure2(t *testing.T) {
	quota := &corespdb.BucketQuota{
		ChargedQuotaSize: 20,
		FreeQuotaSize:    10,
	}
	record := &corespdb.ReadRecord{
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
		ReadTimestampUs: time.Now().Unix(),
	}

	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT * FROM `bucket_traffic` WHERE bucket_id = ? ORDER BY `bucket_traffic`.`bucket_id` LIMIT 1").
		WillReturnError(gorm.ErrNotImplemented)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.InitBucketTraffic(record, quota)
	assert.Contains(t, err.Error(), gorm.ErrNotImplemented.Error())
}

func TestSpDBImpl_GetBucketTrafficSuccess(t *testing.T) {
	bucketID := uint64(2)
	yearMonth := TimestampYearMonth(time.Now().Unix())
	b := BucketTrafficTable{
		BucketID:              bucketID,
		Month:                 yearMonth,
		BucketName:            "mockBucketName",
		ReadConsumedSize:      10,
		FreeQuotaConsumedSize: 100,
		FreeQuotaSize:         25,
		ChargedQuotaSize:      30,
		ModifiedTime:          time.Now(),
	}

	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `bucket_traffic` WHERE bucket_id = ? and month = ? ORDER BY `bucket_traffic`.`bucket_id` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"bucket_id", "year_month", "bucket_name", "read_consumed_size", "free_quota_consumed_size",
			"free_quota_size", "charged_quota_size", "modified_time"}).AddRow(b.BucketID, b.Month, b.BucketName, b.ReadConsumedSize,
			b.FreeQuotaConsumedSize, b.FreeQuotaSize, b.ChargedQuotaSize, b.ModifiedTime))
	result, err := s.GetBucketTraffic(bucketID, yearMonth)
	assert.Nil(t, err)
	assert.Equal(t, "mockBucketName", result.BucketName)
}

func TestSpDBImpl_GetBucketTrafficFailure1(t *testing.T) {
	t.Log("Failure case description: no record")
	bucketID := uint64(2)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `bucket_traffic` WHERE bucket_id = ? and month = ? ORDER BY `bucket_traffic`.`bucket_id` LIMIT 1").
		WillReturnError(gorm.ErrRecordNotFound)
	result, err := s.GetBucketTraffic(bucketID, TimestampYearMonth(time.Now().Unix()))
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	assert.Nil(t, result)
}

func TestSpDBImpl_GetBucketTrafficFailure2(t *testing.T) {
	t.Log("Failure case description: query db returns error")
	bucketID := uint64(2)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `bucket_traffic` WHERE bucket_id = ? and month = ? ORDER BY `bucket_traffic`.`bucket_id` LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.GetBucketTraffic(bucketID, TimestampYearMonth(time.Now().Unix()))
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_GetReadRecordSuccess1(t *testing.T) {
	t.Log("Success case description: limit num is less than or equal to 0")
	timeRange := &corespdb.TrafficTimeRange{
		StartTimestampUs: 1,
		EndTimestampUs:   2,
		LimitNum:         0,
	}
	ta := &ReadRecordTable{
		ReadRecordID:    1,
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		ReadTimestampUs: 1,
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `read_record` WHERE read_timestamp_us >= ? and read_timestamp_us < ?").
		WillReturnRows(sqlmock.NewRows([]string{"read_record_id", "bucket_id", "object_id", "user_address", "read_timestamp_us",
			"bucket_name", "object_name", "read_size"}).AddRow(ta.ReadRecordID, ta.BucketID, ta.ObjectID, ta.UserAddress,
			ta.ReadTimestampUs, ta.BucketName, ta.ObjectName, ta.ReadSize))
	result, err := s.GetReadRecord(timeRange)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}

func TestSpDBImpl_GetReadRecordSuccess2(t *testing.T) {
	t.Log("Success case description: limit num is greater than 0")
	timeRange := &corespdb.TrafficTimeRange{
		StartTimestampUs: 1,
		EndTimestampUs:   2,
		LimitNum:         1,
	}
	ta := &ReadRecordTable{
		ReadRecordID:    1,
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		ReadTimestampUs: 1,
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `read_record` WHERE read_timestamp_us >= ? and read_timestamp_us < ? LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"read_record_id", "bucket_id", "object_id", "user_address", "read_timestamp_us",
			"bucket_name", "object_name", "read_size"}).AddRow(ta.ReadRecordID, ta.BucketID, ta.ObjectID, ta.UserAddress,
			ta.ReadTimestampUs, ta.BucketName, ta.ObjectName, ta.ReadSize))
	result, err := s.GetReadRecord(timeRange)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}

func TestSpDBImpl_GetReadRecordFailure1(t *testing.T) {
	t.Log("Failure case description: no record")
	timeRange := &corespdb.TrafficTimeRange{
		StartTimestampUs: 1,
		EndTimestampUs:   2,
		LimitNum:         1,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `read_record` WHERE read_timestamp_us >= ? and read_timestamp_us < ? LIMIT 1").
		WillReturnError(gorm.ErrRecordNotFound)
	result, err := s.GetReadRecord(timeRange)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	assert.Nil(t, result)
}

func TestSpDBImpl_GetReadRecordFailure2(t *testing.T) {
	t.Log("Failure case description: mock query db returns error")
	timeRange := &corespdb.TrafficTimeRange{
		StartTimestampUs: 1,
		EndTimestampUs:   2,
		LimitNum:         1,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `read_record` WHERE read_timestamp_us >= ? and read_timestamp_us < ? LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.GetReadRecord(timeRange)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_GetBucketReadRecordSuccess1(t *testing.T) {
	t.Log("Success case description: limit num is less than or equal to 0")
	bucketID := uint64(2)
	timeRange := &corespdb.TrafficTimeRange{
		StartTimestampUs: 1,
		EndTimestampUs:   2,
		LimitNum:         0,
	}
	ta := &ReadRecordTable{
		ReadRecordID:    1,
		BucketID:        bucketID,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		ReadTimestampUs: 1,
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `read_record` WHERE read_timestamp_us >= ? and read_timestamp_us < ? and bucket_id = ?").
		WillReturnRows(sqlmock.NewRows([]string{"read_record_id", "bucket_id", "object_id", "user_address", "read_timestamp_us",
			"bucket_name", "object_name", "read_size"}).AddRow(ta.ReadRecordID, ta.BucketID, ta.ObjectID, ta.UserAddress,
			ta.ReadTimestampUs, ta.BucketName, ta.ObjectName, ta.ReadSize))
	result, err := s.GetBucketReadRecord(bucketID, timeRange)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}

func TestSpDBImpl_GetBucketReadRecordSuccess2(t *testing.T) {
	t.Log("Success case description: limit num is greater than 0")
	bucketID := uint64(2)
	timeRange := &corespdb.TrafficTimeRange{
		StartTimestampUs: 1,
		EndTimestampUs:   2,
		LimitNum:         1,
	}
	ta := &ReadRecordTable{
		ReadRecordID:    1,
		BucketID:        bucketID,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		ReadTimestampUs: 1,
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `read_record` WHERE read_timestamp_us >= ? and read_timestamp_us < ? and bucket_id = ? LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"read_record_id", "bucket_id", "object_id", "user_address", "read_timestamp_us",
			"bucket_name", "object_name", "read_size"}).AddRow(ta.ReadRecordID, ta.BucketID, ta.ObjectID, ta.UserAddress,
			ta.ReadTimestampUs, ta.BucketName, ta.ObjectName, ta.ReadSize))
	result, err := s.GetBucketReadRecord(bucketID, timeRange)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}

func TestSpDBImpl_GetBucketReadRecordFailure(t *testing.T) {
	t.Log("Failure case description: mock query db returns error")
	objectID := uint64(3)
	timeRange := &corespdb.TrafficTimeRange{
		StartTimestampUs: 1,
		EndTimestampUs:   2,
		LimitNum:         1,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `read_record` WHERE read_timestamp_us >= ? and read_timestamp_us < ? and bucket_id = ? LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.GetBucketReadRecord(objectID, timeRange)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Equal(t, []*corespdb.ReadRecord(nil), result)
}

func TestSpDBImpl_GetObjectReadRecordSuccess1(t *testing.T) {
	t.Log("Success case description: limit num is less than or equal to 0")
	objectID := uint64(3)
	timeRange := &corespdb.TrafficTimeRange{
		StartTimestampUs: 1,
		EndTimestampUs:   2,
		LimitNum:         0,
	}
	ta := &ReadRecordTable{
		ReadRecordID:    1,
		BucketID:        2,
		ObjectID:        objectID,
		UserAddress:     "mockUserAddress",
		ReadTimestampUs: 1,
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `read_record` WHERE read_timestamp_us >= ? and read_timestamp_us < ? and object_id = ?").
		WillReturnRows(sqlmock.NewRows([]string{"read_record_id", "bucket_id", "object_id", "user_address", "read_timestamp_us",
			"bucket_name", "object_name", "read_size"}).AddRow(ta.ReadRecordID, ta.BucketID, ta.ObjectID, ta.UserAddress,
			ta.ReadTimestampUs, ta.BucketName, ta.ObjectName, ta.ReadSize))
	result, err := s.GetObjectReadRecord(objectID, timeRange)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}

func TestSpDBImpl_GetObjectReadRecordSuccess2(t *testing.T) {
	t.Log("Success case description: limit num is greater than 0")
	objectID := uint64(3)
	timeRange := &corespdb.TrafficTimeRange{
		StartTimestampUs: 1,
		EndTimestampUs:   2,
		LimitNum:         1,
	}
	ta := &ReadRecordTable{
		ReadRecordID:    1,
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     "mockUserAddress",
		ReadTimestampUs: 1,
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `read_record` WHERE read_timestamp_us >= ? and read_timestamp_us < ? and object_id = ? LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"read_record_id", "bucket_id", "object_id", "user_address", "read_timestamp_us",
			"bucket_name", "object_name", "read_size"}).AddRow(ta.ReadRecordID, ta.BucketID, ta.ObjectID, ta.UserAddress,
			ta.ReadTimestampUs, ta.BucketName, ta.ObjectName, ta.ReadSize))
	result, err := s.GetObjectReadRecord(objectID, timeRange)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}

func TestSpDBImpl_GetObjectReadRecordFailure(t *testing.T) {
	t.Log("Failure case description: mock query db returns error")
	objectID := uint64(3)
	timeRange := &corespdb.TrafficTimeRange{
		StartTimestampUs: 1,
		EndTimestampUs:   2,
		LimitNum:         1,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `read_record` WHERE read_timestamp_us >= ? and read_timestamp_us < ? and object_id = ? LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.GetObjectReadRecord(objectID, timeRange)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Equal(t, []*corespdb.ReadRecord(nil), result)
}

func TestSpDBImpl_GetUserReadRecordSuccess1(t *testing.T) {
	t.Log("Success case description: limit num is less than or equal to 0")
	userAddress := "mockUserAddress"
	timeRange := &corespdb.TrafficTimeRange{
		StartTimestampUs: 1,
		EndTimestampUs:   2,
		LimitNum:         0,
	}
	ta := &ReadRecordTable{
		ReadRecordID:    1,
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     userAddress,
		ReadTimestampUs: 1,
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `read_record` WHERE read_timestamp_us >= ? and read_timestamp_us < ? and user_address = ?").
		WillReturnRows(sqlmock.NewRows([]string{"read_record_id", "bucket_id", "object_id", "user_address", "read_timestamp_us",
			"bucket_name", "object_name", "read_size"}).AddRow(ta.ReadRecordID, ta.BucketID, ta.ObjectID, ta.UserAddress,
			ta.ReadTimestampUs, ta.BucketName, ta.ObjectName, ta.ReadSize))
	result, err := s.GetUserReadRecord(userAddress, timeRange)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}

func TestSpDBImpl_GetUserReadRecordSuccess2(t *testing.T) {
	t.Log("Success case description: limit num is greater than 0")
	userAddress := "mockUserAddress"
	timeRange := &corespdb.TrafficTimeRange{
		StartTimestampUs: 1,
		EndTimestampUs:   2,
		LimitNum:         1,
	}
	ta := &ReadRecordTable{
		ReadRecordID:    1,
		BucketID:        2,
		ObjectID:        3,
		UserAddress:     userAddress,
		ReadTimestampUs: 1,
		BucketName:      "mockBucketName",
		ObjectName:      "mockObjectName",
		ReadSize:        1,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `read_record` WHERE read_timestamp_us >= ? and read_timestamp_us < ? and user_address = ? LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"read_record_id", "bucket_id", "object_id", "user_address", "read_timestamp_us",
			"bucket_name", "object_name", "read_size"}).AddRow(ta.ReadRecordID, ta.BucketID, ta.ObjectID, ta.UserAddress,
			ta.ReadTimestampUs, ta.BucketName, ta.ObjectName, ta.ReadSize))
	result, err := s.GetUserReadRecord(userAddress, timeRange)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}

func TestSpDBImpl_GetUserReadRecordFailure(t *testing.T) {
	t.Log("Failure case description: mock query db returns error")
	userAddress := "mockUserAddress"
	timeRange := &corespdb.TrafficTimeRange{
		StartTimestampUs: 1,
		EndTimestampUs:   2,
		LimitNum:         1,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `read_record` WHERE read_timestamp_us >= ? and read_timestamp_us < ? and user_address = ? LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.GetUserReadRecord(userAddress, timeRange)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Equal(t, []*corespdb.ReadRecord(nil), result)
}

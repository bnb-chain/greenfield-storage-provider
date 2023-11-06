package bsdb

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

// GetDataMigrationRecordByProcessKey get the record of data migration by the given process key
func (b *BsDBImpl) GetDataMigrationRecordByProcessKey(processKey string) (*DataMigrationRecord, error) {
	var (
		dataRecord *DataMigrationRecord
		err        error
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

	err = b.db.Take(&dataRecord, "process_key = ?", processKey).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return dataRecord, err
}

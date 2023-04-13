package sqldb

import (
	"strings"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	errorstypes "github.com/bnb-chain/greenfield-storage-provider/pkg/errors/types"
	"gorm.io/gorm"
)

// GetAllServiceConfigs query service config table to get all service configs
func (s *SpDBImpl) GetAllServiceConfigs() (string, string, error) {
	queryReturn := &ServiceConfigTable{}
	result := s.db.Last(&queryReturn)
	if result.Error != nil {
		return "", "", errorstypes.Error(merrors.DBQueryInServiceConfigTableErrCode, result.Error.Error())
	}
	return queryReturn.ConfigVersion, queryReturn.ServiceConfig, nil
}

// SetAllServiceConfigs set service configs to db; if there is no data in db, insert a new record
// otherwise update data in db
func (s *SpDBImpl) SetAllServiceConfigs(version, config string) error {
	configVersion, _, err := s.GetAllServiceConfigs()
	if err != nil && !strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
		return err
	}

	newRecord := &ServiceConfigTable{
		ConfigVersion: version,
		ServiceConfig: config,
	}
	// if there is no record in db, insert a new record
	if configVersion == "" {
		if err := s.insertNewRecordIntoSvcCfgTable(newRecord); err != nil {
			return err
		}
	} else {
		queryReturn := &ServiceConfigTable{}
		result := s.db.Where("config_version = ?", configVersion).Delete(queryReturn)
		if result.Error != nil {
			return errorstypes.Error(merrors.DBDeleteInServiceConfigTableErrCode, result.Error.Error())
		}
		if err := s.insertNewRecordIntoSvcCfgTable(newRecord); err != nil {
			return err
		}
	}
	return nil
}

// insertNewRecordIntoSvcCfgTable insert a new record into service config table
func (s *SpDBImpl) insertNewRecordIntoSvcCfgTable(newRecord *ServiceConfigTable) error {
	result := s.db.Create(newRecord)
	if result.Error != nil || result.RowsAffected != 1 {
		return errorstypes.Error(merrors.DBInsertInServiceConfigTableErrCode, result.Error.Error())
	}
	return nil
}

package sqldb

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// GetAllServiceConfigs query service config table to get all service configs
func (s *SpDBImpl) GetAllServiceConfigs() (string, string, error) {
	queryReturn := &ServiceConfigTable{}
	result := s.db.Last(&queryReturn)
	if result.Error != nil {
		return "", "", fmt.Errorf("failed to query service config table: %s", result.Error)
	}
	return queryReturn.ConfigVersion, queryReturn.ServiceConfig, nil
}

// SetAllServiceConfigs set service configs to db; if there is no data in db, insert a new record
// otherwise update data in db
func (s *SpDBImpl) SetAllServiceConfigs(version, config string) error {
	configVersion, _, err := s.GetAllServiceConfigs()
	isNotFound := strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error())
	if err != nil && !isNotFound {
		return fmt.Errorf("failed to query service config table: %s", err)
	}

	newRecord := &ServiceConfigTable{
		ConfigVersion: version,
		ServiceConfig: config,
	}
	if isNotFound {
		if err := s.insertNewRecordIntoSvcCfgTable(newRecord); err != nil {
			return err
		}
	} else {
		queryReturn := &ServiceConfigTable{}
		result := s.db.Where("config_version = ?", configVersion).Delete(queryReturn)
		if result.Error != nil {
			return fmt.Errorf("failed to detele record in service config table: %s", result.Error)
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
		return fmt.Errorf("failed to insert record in service config table: %s", result.Error)
	}
	return nil
}

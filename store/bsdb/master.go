package bsdb

// GetSwitchDBSignal get signal to switch db
func (b *BsDBImpl) GetSwitchDBSignal() (*MasterDB, error) {
	var (
		signal *MasterDB
		err    error
	)

	err = b.db.Table((&MasterDB{}).TableName()).
		Select("*").
		Take(&signal).Error

	return signal, err
}

// GetMysqlVersion get the current mysql version
func (b *BsDBImpl) GetMysqlVersion() (string, error) {
	var (
		query   string
		version string
		err     error
	)

	query = "SELECT VERSION();"
	err = b.db.Raw(query).Find(&version).Error

	return version, err
}

// GetDefaultCharacterSet get the current mysql default character set
func (b *BsDBImpl) GetDefaultCharacterSet() (string, error) {
	var (
		query               string
		defaultCharacterSet string
		err                 error
	)

	query = "SELECT DEFAULT_CHARACTER_SET_NAME FROM INFORMATION_SCHEMA.SCHEMATA where SCHEMA_NAME in ('block_syncer');"
	err = b.db.Raw(query).Find(&defaultCharacterSet).Error

	return defaultCharacterSet, err
}

// GetDefaultCollationName get the current mysql default collation name
func (b *BsDBImpl) GetDefaultCollationName() (string, error) {
	var (
		query                string
		defaultCollationName string
		err                  error
	)

	query = "SELECT DEFAULT_COLLATION_NAME FROM INFORMATION_SCHEMA.SCHEMATA where SCHEMA_NAME in ('block_syncer');"
	err = b.db.Raw(query).Find(&defaultCollationName).Error

	return defaultCollationName, err
}

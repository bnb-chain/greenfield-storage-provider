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

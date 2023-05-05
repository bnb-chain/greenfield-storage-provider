package bsdb

// GetSwitchDBSignal get signal to switch db
func (b *BsDBImpl) GetSwitchDBSignal() (*Master, error) {
	var (
		master *Master
		err    error
	)

	err = b.db.Table((&Master{}).TableName()).
		Select("*").
		Take(&master).Error

	return master, err
}

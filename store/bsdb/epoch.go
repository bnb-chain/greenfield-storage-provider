package bsdb

// GetSwitchDBSignal get signal to switch db
func (b *BsDBImpl) GetSwitchDBSignal() (bool, error) {
	var (
		epoch  *Epoch
		signal bool
		err    error
	)

	err = b.db.Table((&Epoch{}).TableName()).
		Select("*").
		Take(&epoch).Error
	if epoch.OneRowID {
		signal = true
	}
	return signal, err
}

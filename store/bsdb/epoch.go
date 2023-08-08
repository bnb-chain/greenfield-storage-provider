package bsdb

// GetEpoch get the current epoch
func (b *BsDBImpl) GetEpoch() (*Epoch, error) {
	var (
		query string
		epoch *Epoch
		err   error
	)

	query = "SELECT * FROM epoch LIMIT 1;"
	err = b.db.Raw(query).Find(&epoch).Error

	return epoch, err
}

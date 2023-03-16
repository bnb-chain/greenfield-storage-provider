package bsdb

// GetLatestBlockNumber get current latest block number
func (b *BsDBImpl) GetLatestBlockNumber() (int64, error) {
	var (
		latestBlockNumber int64
		err               error
	)

	err = b.db.Table((&Block{}).TableName()).Select("MAX(height)").Take(&latestBlockNumber).Error
	return latestBlockNumber, err
}

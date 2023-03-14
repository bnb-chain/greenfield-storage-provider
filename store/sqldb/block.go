package sqldb

import "github.com/bnb-chain/greenfield-storage-provider/model/metadata"

// GetLatestBlockNumber get current latest block number
func (s *SpDBImpl) GetLatestBlockNumber() (int64, error) {
	var (
		latestBlockNumber int64
		err               error
	)

	err = s.db.Table((&metadata.Block{}).TableName()).Select("MAX(height)").Take(&latestBlockNumber).Error
	return latestBlockNumber, err
}

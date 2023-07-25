package bsdb

import (
	"github.com/forbole/juno/v4/common"
)

// GetSPByAddress get sp info by operator address
func (b *BsDBImpl) GetSPByAddress(operatorAddress common.Address) (*StorageProvider, error) {
	var (
		sp  *StorageProvider
		err error
	)

	err = b.db.Table((&StorageProvider{}).TableName()).
		Select("*").
		Where("operator_address = ? and removed = false", operatorAddress).
		Take(&sp).Error

	return sp, err
}

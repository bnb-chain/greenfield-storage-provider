package sqldb

import (
	"fmt"
	"time"

	"gorm.io/gorm/clause"

	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// InsertAuthKeyV2 insert a new record into OffChainAuthKeyV2
func (s *SpDBImpl) InsertAuthKeyV2(newRecord *corespdb.OffChainAuthKeyV2) error {
	result := &OffChainAuthKeyV2Table{
		UserAddress:  newRecord.UserAddress,
		Domain:       newRecord.Domain,
		PublicKey:    newRecord.PublicKey,
		ExpiryDate:   newRecord.ExpiryDate,
		CreatedTime:  newRecord.CreatedTime,
		ModifiedTime: newRecord.ModifiedTime,
	}

	err := s.db.Table((&OffChainAuthKeyV2Table{}).TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_address"}, {Name: "domain"}, {Name: "public_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"expiry_date", "modified_time"}),
	}).Create(result).Error
	if err != nil {
		return fmt.Errorf("failed to insert record in OffChainAuthKeyV2Table: %s", err)
	}
	return nil
}

// GetAuthKeyV2 get OffChainAuthKeyV2 from OffChainAuthKeyV2Table
func (s *SpDBImpl) GetAuthKeyV2(userAddress string, domain string, publicKey string) (*corespdb.OffChainAuthKeyV2, error) {
	queryKeyReturn := &OffChainAuthKeyV2Table{}
	result := s.db.First(queryKeyReturn, "user_address = ? and domain =? and public_key=?", userAddress, domain, publicKey)

	if result.Error != nil {
		if errIsNotFound(result.Error) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query OffChainAuthKey table: %s", result.Error)
	}
	return &corespdb.OffChainAuthKeyV2{
		UserAddress:  queryKeyReturn.UserAddress,
		Domain:       queryKeyReturn.Domain,
		PublicKey:    queryKeyReturn.PublicKey,
		ExpiryDate:   queryKeyReturn.ExpiryDate,
		CreatedTime:  queryKeyReturn.CreatedTime,
		ModifiedTime: queryKeyReturn.ModifiedTime,
	}, nil
}

// ClearExpiredOffChainAuthKeys will clear those expired off chain auth keys from OffChainAuthKeyV2Table
func (s *SpDBImpl) ClearExpiredOffChainAuthKeys() error {
	result := s.db.Table(OffChainAuthKeyV2TableName).Where("expiry_date < ? ", time.Now()).Delete(&OffChainAuthKeyV2Table{})
	if result.Error != nil {
		return result.Error
	}
	log.Infow("ClearExpiredOffChainAuthKeys successfully.", "removed rows are ", result.RowsAffected)

	return nil
}

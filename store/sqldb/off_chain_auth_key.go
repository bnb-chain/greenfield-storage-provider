package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
)

// InsertAuthKey insert a new record into OffChainAuthKeyTable
func (s *SpDBImpl) InsertAuthKey(newRecord *corespdb.OffChainAuthKey) error {
	result := s.db.Create(&OffChainAuthKeyTable{
		UserAddress:      newRecord.UserAddress,
		Domain:           newRecord.Domain,
		CurrentNonce:     newRecord.CurrentNonce,
		CurrentPublicKey: newRecord.CurrentPublicKey,
		NextNonce:        newRecord.NextNonce,
		ExpiryDate:       newRecord.ExpiryDate,
		CreatedTime:      newRecord.CreatedTime,
		ModifiedTime:     newRecord.ModifiedTime,
	})
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("failed to insert record in service config table: %s", result.Error)
	}
	return nil
}

// UpdateAuthKey update OffChainAuthKey from OffChainAuthKeyTable
func (s *SpDBImpl) UpdateAuthKey(userAddress string, domain string, oldNonce int32, newNonce int32, newPublicKey string, newExpiryDate time.Time) error {
	queryKeyReturn := &OffChainAuthKeyTable{}
	result := s.db.First(queryKeyReturn, "user_address = ? and domain =? and current_nonce=?", userAddress, domain, oldNonce)
	if result.Error != nil {
		return fmt.Errorf("failed to query OffChainAuthKey table: %s", result.Error)
	}
	queryCondition := &OffChainAuthKeyTable{
		UserAddress:  queryKeyReturn.UserAddress,
		Domain:       queryKeyReturn.Domain,
		CurrentNonce: queryKeyReturn.CurrentNonce,
	}
	updateFields := &OffChainAuthKeyTable{
		CurrentPublicKey: newPublicKey,
		CurrentNonce:     newNonce,
		NextNonce:        newNonce + 1, // increase the Nonce for future use
		ExpiryDate:       newExpiryDate,
		ModifiedTime:     time.Now(),
	}
	result = s.db.Model(queryCondition).Updates(updateFields)
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("failed to update OffChainAuthKeyTable record's state: %s", result.Error)
	}
	return nil
}

func errIsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows) || errors.Is(err, gorm.ErrRecordNotFound)
}

// GetAuthKey get OffChainAuthKey from OffChainAuthKeyTable
func (s *SpDBImpl) GetAuthKey(userAddress string, domain string) (*corespdb.OffChainAuthKey, error) {
	if userAddress == "" || domain == "" {
		return nil, fmt.Errorf("failed to GetAuthKey: userAddress or domain can't be null")
	}

	queryKeyReturn := &OffChainAuthKeyTable{}
	result := s.db.First(queryKeyReturn, "user_address = ? and domain =?", userAddress, domain)

	if result.Error != nil {
		if errIsNotFound(result.Error) {
			// this is a new initial record, not containing any public key but just generate the first nonce as 1
			newRecord := &corespdb.OffChainAuthKey{
				UserAddress:      userAddress,
				Domain:           domain,
				CurrentNonce:     0,
				CurrentPublicKey: "",
				NextNonce:        1,
				ExpiryDate:       time.Now(),
				CreatedTime:      time.Now(),
				ModifiedTime:     time.Now(),
			}
			insertError := s.InsertAuthKey(newRecord)
			if insertError != nil {
				return nil, fmt.Errorf("failed to InsertAuthKey: %s", insertError)
			}
			return newRecord, nil
		}
		return nil, fmt.Errorf("failed to query OffChainAuthKey table: %s", result.Error)
	}
	return &corespdb.OffChainAuthKey{
		UserAddress:      queryKeyReturn.UserAddress,
		Domain:           queryKeyReturn.Domain,
		CurrentNonce:     queryKeyReturn.CurrentNonce,
		CurrentPublicKey: queryKeyReturn.CurrentPublicKey,
		NextNonce:        queryKeyReturn.NextNonce,
		ExpiryDate:       queryKeyReturn.ExpiryDate,
		CreatedTime:      queryKeyReturn.CreatedTime,
		ModifiedTime:     queryKeyReturn.ModifiedTime,
	}, nil
}

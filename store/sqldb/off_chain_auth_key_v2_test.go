package sqldb

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
)

func TestSpDBImpl_InsertAuthKeyV2Success(t *testing.T) {
	o := &corespdb.OffChainAuthKeyV2{
		UserAddress:  "mockUserAddress",
		Domain:       "mockDomain",
		PublicKey:    "mockPublicKey",
		ExpiryDate:   time.Now(),
		CreatedTime:  time.Now(),
		ModifiedTime: time.Now(),
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("\t\t\tINSERT INTO `off_chain_auth_key_v2` (`user_address`,`domain`,`public_key`,`expiry_date`,`created_time`,`modified_time`) VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `expiry_date`=VALUES(`expiry_date`),`modified_time`=VALUES(`modified_time`)\n").
		WithArgs(o.UserAddress, o.Domain, o.PublicKey, o.ExpiryDate, o.CreatedTime, o.ModifiedTime).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.InsertAuthKeyV2(o)
	assert.Nil(t, err)
}

func TestSpDBImpl_InsertAuthKeyV2Failure(t *testing.T) {
	o := &corespdb.OffChainAuthKeyV2{
		UserAddress:  "mockUserAddress",
		Domain:       "mockDomain",
		PublicKey:    " CurrentPublicKey",
		ExpiryDate:   time.Now(),
		CreatedTime:  time.Now(),
		ModifiedTime: time.Now(),
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `off_chain_auth_key_v2` (`user_address`,`domain`,`public_key`,`expiry_date`,`created_time`,`modified_time`) VALUES (?,?,?,?,?,?)").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.InsertAuthKeyV2(o)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_GetAuthKeyV2Success1(t *testing.T) {
	t.Log("Success case description: query db and has data")
	var (
		userAddress = "mockUserAddress"
		domain      = "mockDomain"
		publicKey   = "mockPublicKey"
	)
	o := &corespdb.OffChainAuthKeyV2{
		UserAddress:  userAddress,
		Domain:       domain,
		PublicKey:    publicKey,
		ExpiryDate:   time.Now(),
		CreatedTime:  time.Now(),
		ModifiedTime: time.Now(),
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `off_chain_auth_key_v2` WHERE user_address = ? and domain =? and public_key=? ORDER BY `off_chain_auth_key_v2`.`user_address` LIMIT 1").
		WithArgs(userAddress, domain, publicKey).WillReturnRows(sqlmock.NewRows([]string{"user_address", "domain", "public_key", "expiry_date", "created_time", "modified_time"}).AddRow(o.UserAddress,
		o.Domain, o.PublicKey, o.ExpiryDate, o.CreatedTime, o.ModifiedTime))
	result, err := s.GetAuthKeyV2(userAddress, domain, publicKey)
	assert.Nil(t, err)
	assert.Equal(t, publicKey, result.PublicKey)
}

func TestSpDBImpl_GetAuthKeyV2Failure1(t *testing.T) {
	t.Log("failure case description: query db and return db error")
	var (
		userAddress = "mockUserAddress"
		domain      = "mockDomain"
		publicKey   = "mockPublicKey"
	)

	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `off_chain_auth_key_v2` WHERE user_address = ? and domain =? and public_key=? ORDER BY `off_chain_auth_key_v2`.`user_address` LIMIT 1").
		WithArgs(userAddress, domain, publicKey).WillReturnError(mockDBInternalError)
	result, err := s.GetAuthKeyV2(userAddress, domain, publicKey)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_GetAuthKeyV2Failure2(t *testing.T) {
	t.Log("failure case description: query db and has no data")
	var (
		userAddress = "mockUserAddress"
		domain      = "mockDomain"
		publicKey   = "mockPublicKey"
	)

	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `off_chain_auth_key_v2` WHERE user_address = ? and domain =? and public_key=? ORDER BY `off_chain_auth_key_v2`.`user_address` LIMIT 1").
		WithArgs(userAddress, domain, publicKey).WillReturnError(gorm.ErrRecordNotFound)
	result, err := s.GetAuthKeyV2(userAddress, domain, publicKey)
	assert.Nil(t, result)
	assert.Nil(t, err)
}

func TestSpDBImpl_ClearExpiredOffChainAuthKeysSuccess(t *testing.T) {
	t.Log("Success case description: delete data")

	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `off_chain_auth_key_v2` WHERE expiry_date < ?").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	err := s.ClearExpiredOffChainAuthKeys()
	//assert.Nil(t, result)
	assert.Nil(t, err)
}

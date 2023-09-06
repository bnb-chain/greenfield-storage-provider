package sqldb

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
)

func TestSpDBImpl_InsertAuthKeySuccess(t *testing.T) {
	o := &corespdb.OffChainAuthKey{
		UserAddress:      "mockUserAddress",
		Domain:           "mockDomain",
		CurrentNonce:     1,
		CurrentPublicKey: "mockCurrentPublicKey",
		NextNonce:        2,
		ExpiryDate:       time.Now(),
		CreatedTime:      time.Now(),
		ModifiedTime:     time.Now(),
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `off_chain_auth_key` (`user_address`,`domain`,`current_nonce`,`current_public_key`,`next_nonce`,`expiry_date`,`created_time`,`modified_time`) VALUES (?,?,?,?,?,?,?,?)").
		WithArgs(o.UserAddress, o.Domain, o.CurrentNonce, o.CurrentPublicKey, o.NextNonce, o.ExpiryDate, o.CreatedTime, o.ModifiedTime).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.InsertAuthKey(o)
	assert.Nil(t, err)
}

func TestSpDBImpl_InsertAuthKeyFailure(t *testing.T) {
	o := &corespdb.OffChainAuthKey{
		UserAddress:      "mockUserAddress",
		Domain:           "mockDomain",
		CurrentNonce:     1,
		CurrentPublicKey: "mockCurrentPublicKey",
		NextNonce:        2,
		ExpiryDate:       time.Now(),
		CreatedTime:      time.Now(),
		ModifiedTime:     time.Now(),
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `off_chain_auth_key` (`user_address`,`domain`,`current_nonce`,`current_public_key`,`next_nonce`,`expiry_date`,`created_time`,`modified_time`) VALUES (?,?,?,?,?,?,?,?)").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.InsertAuthKey(o)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_UpdateAuthKeySuccess(t *testing.T) {
	var (
		userAddress   = "mockUserAddress"
		domain        = "mockDomain"
		oldNonce      = int32(1)
		newNonce      = int32(3)
		newPublicKey  = "mockCurrentPublicKey"
		newExpiryDate = time.Now()
	)
	o := &corespdb.OffChainAuthKey{
		UserAddress:      userAddress,
		Domain:           domain,
		CurrentNonce:     oldNonce,
		CurrentPublicKey: newPublicKey,
		NextNonce:        3,
		ExpiryDate:       newExpiryDate,
		CreatedTime:      time.Now(),
		ModifiedTime:     time.Now(),
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `off_chain_auth_key` WHERE user_address = ? and domain =? and current_nonce=? ORDER BY `off_chain_auth_key`.`user_address` LIMIT 1").
		WithArgs(userAddress, domain, oldNonce).WillReturnRows(sqlmock.NewRows([]string{"user_address", "domain", "current_nonce",
		"current_public_key", "next_nonce", "expiry_date", "created_time", "modified_time"}).AddRow(
		o.UserAddress, o.Domain, o.CurrentNonce, o.CurrentPublicKey, o.NextNonce, o.ExpiryDate, o.CreatedTime, o.ModifiedTime))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `off_chain_auth_key` SET `current_nonce`=?,`current_public_key`=?,`next_nonce`=?,`expiry_date`=?,`modified_time`=? WHERE `user_address` = ? AND `domain` = ?").
		WithArgs(newNonce, newPublicKey, newNonce+1, newExpiryDate, AnyTime{}, userAddress, domain).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateAuthKey(userAddress, domain, oldNonce, newNonce, newPublicKey, newExpiryDate)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateAuthKeySuccess2(t *testing.T) {
	var (
		userAddress   = "mockUserAddress"
		domain        = "mockDomain"
		oldNonce      = int32(0)
		newNonce      = int32(1)
		newPublicKey  = "mockCurrentPublicKey"
		newExpiryDate = time.Now()
	)
	o := &corespdb.OffChainAuthKey{
		UserAddress:      userAddress,
		Domain:           domain,
		CurrentNonce:     oldNonce,
		CurrentPublicKey: newPublicKey,
		NextNonce:        newNonce,
		ExpiryDate:       newExpiryDate,
		CreatedTime:      time.Now(),
		ModifiedTime:     time.Now(),
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `off_chain_auth_key` WHERE user_address = ? and domain =? and current_nonce=? ORDER BY `off_chain_auth_key`.`user_address` LIMIT 1").
		WithArgs(userAddress, domain, oldNonce).WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `off_chain_auth_key` (`user_address`,`domain`,`current_nonce`,`current_public_key`,`next_nonce`,`expiry_date`,`created_time`,`modified_time`) VALUES (?,?,?,?,?,?,?,?)").
		WithArgs(o.UserAddress, o.Domain, 0, "", 1, AnyTime{}, AnyTime{}, AnyTime{}).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `off_chain_auth_key` SET `current_nonce`=?,`current_public_key`=?,`next_nonce`=?,`expiry_date`=?,`modified_time`=? WHERE `user_address` = ? AND `domain` = ?").
		WithArgs(newNonce, newPublicKey, newNonce+1, newExpiryDate, AnyTime{}, userAddress, domain).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateAuthKey(userAddress, domain, oldNonce, newNonce, newPublicKey, newExpiryDate)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateAuthKeyFailure1(t *testing.T) {
	t.Log("Failure case description: mock query db returns error")
	var (
		userAddress   = "mockUserAddress"
		domain        = "mockDomain"
		oldNonce      = int32(1)
		newNonce      = int32(3)
		newPublicKey  = "mockCurrentPublicKey"
		newExpiryDate = time.Now()
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `off_chain_auth_key` WHERE user_address = ? and domain =? and current_nonce=? ORDER BY `off_chain_auth_key`.`user_address` LIMIT 1").
		WillReturnError(mockDBInternalError)
	err := s.UpdateAuthKey(userAddress, domain, oldNonce, newNonce, newPublicKey, newExpiryDate)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_UpdateAuthKeyFailure2(t *testing.T) {
	t.Log("Failure case description: mock db update returns error")
	var (
		userAddress   = "mockUserAddress"
		domain        = "mockDomain"
		oldNonce      = int32(1)
		newNonce      = int32(3)
		newPublicKey  = "mockCurrentPublicKey"
		newExpiryDate = time.Now()
	)
	o := &corespdb.OffChainAuthKey{
		UserAddress:      userAddress,
		Domain:           domain,
		CurrentNonce:     oldNonce,
		CurrentPublicKey: newPublicKey,
		NextNonce:        3,
		ExpiryDate:       newExpiryDate,
		CreatedTime:      time.Now(),
		ModifiedTime:     time.Now(),
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `off_chain_auth_key` WHERE user_address = ? and domain =? and current_nonce=? ORDER BY `off_chain_auth_key`.`user_address` LIMIT 1").
		WithArgs(userAddress, domain, oldNonce).WillReturnRows(sqlmock.NewRows([]string{"user_address", "domain", "current_nonce",
		"current_public_key", "next_nonce", "expiry_date", "created_time", "modified_time"}).AddRow(
		o.UserAddress, o.Domain, o.CurrentNonce, o.CurrentPublicKey, o.NextNonce, o.ExpiryDate, o.CreatedTime, o.ModifiedTime))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `off_chain_auth_key` SET `current_nonce`=?,`current_public_key`=?,`next_nonce`=?,`expiry_date`=?,`modified_time`=? WHERE `user_address` = ? AND `domain` = ?").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.UpdateAuthKey(userAddress, domain, oldNonce, newNonce, newPublicKey, newExpiryDate)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_UpdateAuthKeyFailure3(t *testing.T) {
	t.Log("Failure case description: mock query db returns error")
	var (
		userAddress   = "mockUserAddress"
		domain        = "mockDomain"
		oldNonce      = int32(1)
		newNonce      = int32(3)
		newPublicKey  = "mockCurrentPublicKey"
		newExpiryDate = time.Now()
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `off_chain_auth_key` WHERE user_address = ? and domain =? and current_nonce=? ORDER BY `off_chain_auth_key`.`user_address` LIMIT 1").
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()

	mock.ExpectExec("INSERT INTO `off_chain_auth_key` (`user_address`,`domain`,`current_nonce`,`current_public_key`,`next_nonce`,`expiry_date`,`created_time`,`modified_time`) VALUES (?,?,?,?,?,?,?,?)").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()

	err := s.UpdateAuthKey(userAddress, domain, oldNonce, newNonce, newPublicKey, newExpiryDate)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func Test_errIsNotFound(t *testing.T) {
	ok := errIsNotFound(mockDBInternalError)
	assert.Equal(t, false, ok)
}

func TestSpDBImpl_GetAuthKeySuccess1(t *testing.T) {
	t.Log("Success case description: query db and has data")
	var (
		userAddress      = "mockUserAddress"
		domain           = "mockDomain"
		currentPublicKey = "mockCurrentPublicKey"
	)
	o := &corespdb.OffChainAuthKey{
		UserAddress:      userAddress,
		Domain:           domain,
		CurrentNonce:     1,
		CurrentPublicKey: currentPublicKey,
		NextNonce:        2,
		ExpiryDate:       time.Now(),
		CreatedTime:      time.Now(),
		ModifiedTime:     time.Now(),
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `off_chain_auth_key` WHERE user_address = ? and domain =? ORDER BY `off_chain_auth_key`.`user_address` LIMIT 1").
		WithArgs(userAddress, domain).WillReturnRows(sqlmock.NewRows([]string{"user_address", "domain", "current_nonce",
		"current_public_key", "next_nonce", "expiry_date", "created_time", "modified_time"}).AddRow(o.UserAddress,
		o.Domain, o.CurrentNonce, o.CurrentPublicKey, o.NextNonce, o.ExpiryDate, o.CreatedTime, o.ModifiedTime))
	result, err := s.GetAuthKey(userAddress, domain)
	assert.Nil(t, err)
	assert.Equal(t, currentPublicKey, result.CurrentPublicKey)
}

func TestSpDBImpl_GetAuthKeySuccess2(t *testing.T) {
	t.Log("Success case description: query db no data and inert a new row")
	var (
		userAddress = "mockUserAddress"
		domain      = "mockDomain"
	)

	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `off_chain_auth_key` WHERE user_address = ? and domain =? ORDER BY `off_chain_auth_key`.`user_address` LIMIT 1").
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := s.GetAuthKey(userAddress, domain)
	assert.Nil(t, err)
	assert.Equal(t, int32(1), result.NextNonce)
}

func TestSpDBImpl_GetAuthKeyFailure1(t *testing.T) {
	t.Log("Failure case description: empty userAddress returns error")
	s, _ := setupDB(t)
	result, err := s.GetAuthKey("", "mock")
	assert.Equal(t, errors.New("failed to GetAuthKey: userAddress or domain can't be null"), err)
	assert.Nil(t, result)
}

func TestSpDBImpl_GetAuthKeyFailure2(t *testing.T) {
	t.Log("Failure case description: db insert returns error")
	var (
		userAddress = "mockUserAddress"
		domain      = "mockDomain"
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `off_chain_auth_key` WHERE user_address = ? and domain =? ORDER BY `off_chain_auth_key`.`user_address` LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.GetAuthKey(userAddress, domain)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

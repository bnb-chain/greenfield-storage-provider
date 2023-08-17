package sqldb

import (
	"errors"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

func TestSpDBImpl_UpdateAllSpSuccess1(t *testing.T) {
	t.Log("Success case description: query db no result and insert a row into db")
	spList := []*sptypes.StorageProvider{
		{
			Id:                 1,
			OperatorAddress:    "mockOperatorAddress",
			FundingAddress:     "mockFundingAddress",
			SealAddress:        "mockSealAddress",
			ApprovalAddress:    "mockApprovalAddress",
			GcAddress:          "mockGcAddress",
			MaintenanceAddress: "mockMaintenanceAddress",
			TotalDeposit:       sdkmath.NewInt(100),
			Status:             1,
			Endpoint:           "mockEndpoint",
			Description: sptypes.Description{
				Moniker:         "mockMoniker",
				Identity:        "mockIdentity",
				Website:         "mockWebsite",
				SecurityContact: "mockSecurityContact",
				Details:         "mockDetails",
			},
			BlsKey: []byte("mockBlsKey"),
		},
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE operator_address = ? and is_own = false ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `sp_info` (`operator_address`,`is_own`,`id`,`funding_address`,`seal_address`,`approval_address`,`total_deposit`,`status`,`endpoint`,`moniker`,`identity`,`website`,`security_contact`,`details`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateAllSp(spList)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateAllSpSuccess2(t *testing.T) {
	t.Log("Success case description: query db get result and update db")
	spList := []*sptypes.StorageProvider{
		{
			Id:                 1,
			OperatorAddress:    "mockOperatorAddress",
			FundingAddress:     "mockFundingAddress",
			SealAddress:        "mockSealAddress",
			ApprovalAddress:    "mockApprovalAddress",
			GcAddress:          "mockGcAddress",
			MaintenanceAddress: "mockMaintenanceAddress",
			TotalDeposit:       sdkmath.NewInt(100),
			Status:             1,
			Endpoint:           "mockEndpoint",
			Description: sptypes.Description{
				Moniker:         "mockMoniker",
				Identity:        "mockIdentity",
				Website:         "mockWebsite",
				SecurityContact: "mockSecurityContact",
				Details:         "mockDetails",
			},
			BlsKey: []byte("mockBlsKey"),
		},
	}
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE operator_address = ? and is_own = false ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(sp.OperatorAddress, sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit,
				sp.Status, sp.Endpoint, sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `sp_info` SET `operator_address`=?,`id`=?,`funding_address`=?,`seal_address`=?,`approval_address`=?,`total_deposit`=?,`status`=?,`endpoint`=?,`moniker`=?,`identity`=?,`website`=?,`security_contact`=?,`details`=? WHERE operator_address = ? and is_own = false").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateAllSp(spList)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateAllSpFailure1(t *testing.T) {
	t.Log("Failure case description: query db returns error")
	spList := []*sptypes.StorageProvider{
		{
			Id:                 1,
			OperatorAddress:    "mockOperatorAddress",
			FundingAddress:     "mockFundingAddress",
			SealAddress:        "mockSealAddress",
			ApprovalAddress:    "mockApprovalAddress",
			GcAddress:          "mockGcAddress",
			MaintenanceAddress: "mockMaintenanceAddress",
			TotalDeposit:       sdkmath.NewInt(100),
			Status:             1,
			Endpoint:           "mockEndpoint",
			Description: sptypes.Description{
				Moniker:         "mockMoniker",
				Identity:        "mockIdentity",
				Website:         "mockWebsite",
				SecurityContact: "mockSecurityContact",
				Details:         "mockDetails",
			},
			BlsKey: []byte("mockBlsKey"),
		},
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE operator_address = ? and is_own = false ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnError(mockDBInternalError)
	err := s.UpdateAllSp(spList)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_UpdateAllSpFailure2(t *testing.T) {
	t.Log("Failure case description: db inserts a row returns error")
	spList := []*sptypes.StorageProvider{
		{
			Id:                 1,
			OperatorAddress:    "mockOperatorAddress",
			FundingAddress:     "mockFundingAddress",
			SealAddress:        "mockSealAddress",
			ApprovalAddress:    "mockApprovalAddress",
			GcAddress:          "mockGcAddress",
			MaintenanceAddress: "mockMaintenanceAddress",
			TotalDeposit:       sdkmath.NewInt(100),
			Status:             1,
			Endpoint:           "mockEndpoint",
			Description: sptypes.Description{
				Moniker:         "mockMoniker",
				Identity:        "mockIdentity",
				Website:         "mockWebsite",
				SecurityContact: "mockSecurityContact",
				Details:         "mockDetails",
			},
			BlsKey: []byte("mockBlsKey"),
		},
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE operator_address = ? and is_own = false ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `sp_info` (`operator_address`,`is_own`,`id`,`funding_address`,`seal_address`,`approval_address`,`total_deposit`,`status`,`endpoint`,`moniker`,`identity`,`website`,`security_contact`,`details`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.UpdateAllSp(spList)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_UpdateAllSpFailure3(t *testing.T) {
	t.Log("Failure case description: db updates a row returns error")
	spList := []*sptypes.StorageProvider{
		{
			Id:                 1,
			OperatorAddress:    "mockOperatorAddress",
			FundingAddress:     "mockFundingAddress",
			SealAddress:        "mockSealAddress",
			ApprovalAddress:    "mockApprovalAddress",
			GcAddress:          "mockGcAddress",
			MaintenanceAddress: "mockMaintenanceAddress",
			TotalDeposit:       sdkmath.NewInt(100),
			Status:             1,
			Endpoint:           "mockEndpoint",
			Description: sptypes.Description{
				Moniker:         "mockMoniker",
				Identity:        "mockIdentity",
				Website:         "mockWebsite",
				SecurityContact: "mockSecurityContact",
				Details:         "mockDetails",
			},
			BlsKey: []byte("mockBlsKey"),
		},
	}
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE operator_address = ? and is_own = false ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(sp.OperatorAddress, sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit,
				sp.Status, sp.Endpoint, sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `sp_info` SET `operator_address`=?,`id`=?,`funding_address`=?,`seal_address`=?,`approval_address`=?,`total_deposit`=?,`status`=?,`endpoint`=?,`moniker`=?,`identity`=?,`website`=?,`security_contact`=?,`details`=? WHERE operator_address = ? and is_own = false").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.UpdateAllSp(spList)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_FetchAllSpSuccess1(t *testing.T) {
	t.Log("Success case description: status length is 0")
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "123",
		Status:          1,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = false").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(sp.OperatorAddress, sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit,
				sp.Status, sp.Endpoint, sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	result, err := s.FetchAllSp()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}

func TestSpDBImpl_FetchAllSpSuccess2(t *testing.T) {
	t.Log("Success case description: status length is 1")
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "123",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = false and status = ?").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(sp.OperatorAddress, sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit,
				sp.Status, sp.Endpoint, sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	result, err := s.FetchAllSp(sptypes.STATUS_IN_SERVICE)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}

func TestSpDBImpl_FetchAllSpFailure1(t *testing.T) {
	t.Log("Failure case description: status length is 0 and query db returns error")
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = false").
		WillReturnError(mockDBInternalError)
	result, err := s.FetchAllSp()
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_FetchAllSpFailure2(t *testing.T) {
	t.Log("Failure case description: status length is 1 and query db returns error")
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = false and status = ?").
		WillReturnError(mockDBInternalError)
	result, err := s.FetchAllSp(sptypes.STATUS_IN_SERVICE)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_FetchAllSpFailure3(t *testing.T) {
	t.Log("Failure case description: convert string to int returns error")
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "mockTotalDeposit",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = false and status = ?").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(sp.OperatorAddress, sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit,
				sp.Status, sp.Endpoint, sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	result, err := s.FetchAllSp(sptypes.STATUS_IN_SERVICE)
	assert.Equal(t, errors.New("failed to parse int"), err)
	assert.Equal(t, []*sptypes.StorageProvider{}, result)
}

func TestSpDBImpl_FetchAllSpWithoutOwnSpSuccess1(t *testing.T) {
	t.Log("Success case description: status length is 0")
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "123",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = true ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(sp.OperatorAddress, sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit,
				sp.Status, sp.Endpoint, sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE operator_address != ?").WillReturnRows(sqlmock.NewRows(
		[]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address", "total_deposit",
			"status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).AddRow(sp.OperatorAddress,
		sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit, sp.Status, sp.Endpoint,
		sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	result, err := s.FetchAllSpWithoutOwnSp()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}

func TestSpDBImpl_FetchAllSpWithoutOwnSpSuccess2(t *testing.T) {
	t.Log("Success case description: status length is not 0")
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "123",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = true ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(sp.OperatorAddress, sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit,
				sp.Status, sp.Endpoint, sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE status = ? and operator_address != ?").WillReturnRows(sqlmock.NewRows(
		[]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address", "total_deposit",
			"status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).AddRow(sp.OperatorAddress,
		sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit, sp.Status, sp.Endpoint,
		sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	result, err := s.FetchAllSpWithoutOwnSp(sptypes.STATUS_IN_SERVICE)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}

func TestSpDBImpl_FetchAllSpWithoutOwnSpFailure1(t *testing.T) {
	t.Log("Failure case description: get own sp info returns error")
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = true ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.FetchAllSpWithoutOwnSp()
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_FetchAllSpWithoutOwnSpFailure2(t *testing.T) {
	t.Log("Failure case description: status length is 0, query db returns error")
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "123",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	s, mock := setupDB(t)

	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = true ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(sp.OperatorAddress, sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit,
				sp.Status, sp.Endpoint, sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE operator_address != ?").WillReturnError(mockDBInternalError)
	result, err := s.FetchAllSpWithoutOwnSp()
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_FetchAllSpWithoutOwnSpFailure3(t *testing.T) {
	t.Log("Success case description: status length is not 0, query db returns error")
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "123",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = true ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(sp.OperatorAddress, sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit,
				sp.Status, sp.Endpoint, sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE status = ? and operator_address != ?").
		WillReturnError(mockDBInternalError)
	result, err := s.FetchAllSpWithoutOwnSp(sptypes.STATUS_IN_SERVICE)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_GetSpByAddressSuccess(t *testing.T) {
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "123",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE operator_address = ? and is_own = false ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(sp.OperatorAddress, sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit,
				sp.Status, sp.Endpoint, sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	result, err := s.GetSpByAddress(sp.OperatorAddress, corespdb.OperatorAddressType)
	assert.Nil(t, err)
	assert.Equal(t, "mockSealAddress", result.SealAddress)
}

func TestSpDBImpl_GetSpByAddressFailure1(t *testing.T) {
	t.Log("Failure case description: unknown address type")
	var unknownAddressType corespdb.SpAddressType = -1
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "mockTotalDeposit",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	s, _ := setupDB(t)
	result, err := s.GetSpByAddress(sp.OperatorAddress, unknownAddressType)
	assert.Equal(t, errors.New("unknown address type"), err)
	assert.Nil(t, result)
}

func TestSpDBImpl_GetSpByAddressFailure2(t *testing.T) {
	t.Log("Failure case description: mock query db returns error")
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE operator_address = ? and is_own = false ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.GetSpByAddress("mockOperatorAddress", corespdb.OperatorAddressType)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_GetSpByAddressFailure3(t *testing.T) {
	t.Log("Failure case description: convert string to int returns error")
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "mockTotalDeposit",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE operator_address = ? and is_own = false ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(sp.OperatorAddress, sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit,
				sp.Status, sp.Endpoint, sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	result, err := s.GetSpByAddress(sp.OperatorAddress, corespdb.OperatorAddressType)
	assert.Equal(t, errors.New("failed to parse int"), err)
	assert.Nil(t, result)
}

func Test_getAddressCondition(t *testing.T) {
	var unknownAddressType corespdb.SpAddressType = -1
	cases := []struct {
		name         string
		addressType  corespdb.SpAddressType
		wantedResult string
		wantedErr    error
	}{
		{
			name:         "operator address",
			addressType:  corespdb.OperatorAddressType,
			wantedResult: "operator_address = ? and is_own = false",
			wantedErr:    nil,
		},
		{
			name:         "funding address",
			addressType:  corespdb.FundingAddressType,
			wantedResult: "funding_address = ? and is_own = false",
			wantedErr:    nil,
		},
		{
			name:         "seal address",
			addressType:  corespdb.SealAddressType,
			wantedResult: "seal_address = ? and is_own = false",
			wantedErr:    nil,
		},
		{
			name:         "approval address",
			addressType:  corespdb.ApprovalAddressType,
			wantedResult: "approval_address = ? and is_own = false",
			wantedErr:    nil,
		},
		{
			name:         "unknown address type",
			addressType:  unknownAddressType,
			wantedResult: "",
			wantedErr:    errors.New("unknown address type"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getAddressCondition(tt.addressType)
			assert.Equal(t, tt.wantedResult, result)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestSpDBImpl_GetSpByEndpointSuccess(t *testing.T) {
	endpoint := "mockEndpoint"
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "123",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE endpoint = ? and is_own = false ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(sp.OperatorAddress, sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit,
				sp.Status, sp.Endpoint, sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	result, err := s.GetSpByEndpoint(endpoint)
	assert.Nil(t, err)
	assert.Equal(t, "mockSealAddress", result.SealAddress)
}

func TestSpDBImpl_GetSpByEndpointFailure1(t *testing.T) {
	t.Log("Failure case description: mock query db returns error")
	endpoint := "mockEndpoint"
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE endpoint = ? and is_own = false ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.GetSpByEndpoint(endpoint)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_GetSpByEndpointFailure2(t *testing.T) {
	t.Log("Failure case description: convert string to int returns error")
	endpoint := "mockEndpoint"
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "mockTotalDeposit",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE endpoint = ? and is_own = false ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(sp.OperatorAddress, sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit,
				sp.Status, sp.Endpoint, sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	result, err := s.GetSpByEndpoint(endpoint)
	assert.Equal(t, errors.New("failed to parse int"), err)
	assert.Nil(t, result)
}

func TestSpDBImpl_GetSpByIDSuccess(t *testing.T) {
	id := uint32(1)
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "123",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE id = ? and is_own = false ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(sp.OperatorAddress, sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit,
				sp.Status, sp.Endpoint, sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	result, err := s.GetSpByID(id)
	assert.Nil(t, err)
	assert.Equal(t, "mockSealAddress", result.SealAddress)
}

func TestSpDBImpl_GetSpByIDFailure1(t *testing.T) {
	t.Log("Failure case description: query db returns error")
	id := uint32(1)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE id = ? and is_own = false ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.GetSpByID(id)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_GetSpByIDFailure2(t *testing.T) {
	t.Log("Failure case description: convert string to int returns error")
	id := uint32(1)
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "mockTotalDeposit",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE id = ? and is_own = false ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(sp.OperatorAddress, sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit,
				sp.Status, sp.Endpoint, sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	result, err := s.GetSpByID(id)
	assert.Equal(t, err, errors.New("failed to parse int"))
	assert.Nil(t, result)
}

func TestSpDBImpl_GetOwnSpInfoSuccess(t *testing.T) {
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           true,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "123",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = true ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(sp.OperatorAddress, sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit,
				sp.Status, sp.Endpoint, sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	result, err := s.GetOwnSpInfo()
	assert.Nil(t, err)
	assert.Equal(t, "mockSealAddress", result.SealAddress)
}

func TestSpDBImpl_GetOwnSpInfoFailure1(t *testing.T) {
	t.Log("Success case description: mock query db returns error")
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = true ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.GetOwnSpInfo()
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_GetOwnSpInfoFailure2(t *testing.T) {
	t.Log("Success case description: convert string to int return error")
	sp := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           true,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "mockTotalDeposit",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = true ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(sp.OperatorAddress, sp.IsOwn, sp.ID, sp.FundingAddress, sp.SealAddress, sp.ApprovalAddress, sp.TotalDeposit,
				sp.Status, sp.Endpoint, sp.Moniker, sp.Identity, sp.Website, sp.SecurityContact, sp.Details))
	result, err := s.GetOwnSpInfo()
	assert.Equal(t, err, errors.New("failed to parse int"))
	assert.Nil(t, result)
}

func TestSpDBImpl_SetOwnSpInfoSuccess1(t *testing.T) {
	t.Log("Success case description: query db has record, update it")
	ta := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "123",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	sp := &sptypes.StorageProvider{
		Id:                 ta.ID,
		OperatorAddress:    ta.OperatorAddress,
		FundingAddress:     ta.FundingAddress,
		SealAddress:        ta.SealAddress,
		ApprovalAddress:    ta.ApprovalAddress,
		GcAddress:          "mockGCAddress",
		MaintenanceAddress: "mockMaintenanceAddress",
		TotalDeposit:       sdkmath.NewInt(123),
		Status:             0,
		Endpoint:           ta.Endpoint,
		Description: sptypes.Description{
			Moniker:         ta.Moniker,
			Identity:        ta.Identity,
			Website:         ta.Website,
			SecurityContact: ta.Website,
			Details:         ta.Details,
		},
		BlsKey: []byte("mockBlsKey"),
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = true ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `sp_info` (`operator_address`,`is_own`,`id`,`funding_address`,`seal_address`,`approval_address`,`total_deposit`,`status`,`endpoint`,`moniker`,`identity`,`website`,`security_contact`,`details`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.SetOwnSpInfo(sp)
	assert.Nil(t, err)
}

func TestSpDBImpl_SetOwnSpInfoSuccess2(t *testing.T) {
	t.Log("Success case description: query db no record, insert a row")
	ta := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "123",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	sp := &sptypes.StorageProvider{
		Id:                 ta.ID,
		OperatorAddress:    ta.OperatorAddress,
		FundingAddress:     ta.FundingAddress,
		SealAddress:        ta.SealAddress,
		ApprovalAddress:    ta.ApprovalAddress,
		GcAddress:          "mockGCAddress",
		MaintenanceAddress: "mockMaintenanceAddress",
		TotalDeposit:       sdkmath.NewInt(123),
		Status:             0,
		Endpoint:           ta.Endpoint,
		Description: sptypes.Description{
			Moniker:         ta.Moniker,
			Identity:        ta.Identity,
			Website:         ta.Website,
			SecurityContact: ta.Website,
			Details:         ta.Details,
		},
		BlsKey: []byte("mockBlsKey"),
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = true ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(ta.OperatorAddress, ta.IsOwn, ta.ID, ta.FundingAddress, ta.SealAddress, ta.ApprovalAddress, ta.TotalDeposit,
				ta.Status, ta.Endpoint, ta.Moniker, ta.Identity, ta.Website, ta.SecurityContact, ta.Details))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `sp_info` SET `operator_address`=?,`is_own`=?,`id`=?,`funding_address`=?,`seal_address`=?,`approval_address`=?,`total_deposit`=?,`endpoint`=?,`moniker`=?,`identity`=?,`website`=?,`security_contact`=?,`details`=? WHERE is_own = true").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.SetOwnSpInfo(sp)
	assert.Nil(t, err)
}

func TestSpDBImpl_SetOwnSpInfoFailure1(t *testing.T) {
	t.Log("Failure case description: query db returns error")
	ta := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "123",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	sp := &sptypes.StorageProvider{
		Id:                 ta.ID,
		OperatorAddress:    ta.OperatorAddress,
		FundingAddress:     ta.FundingAddress,
		SealAddress:        ta.SealAddress,
		ApprovalAddress:    ta.ApprovalAddress,
		GcAddress:          "mockGCAddress",
		MaintenanceAddress: "mockMaintenanceAddress",
		TotalDeposit:       sdkmath.NewInt(123),
		Status:             0,
		Endpoint:           ta.Endpoint,
		Description: sptypes.Description{
			Moniker:         ta.Moniker,
			Identity:        ta.Identity,
			Website:         ta.Website,
			SecurityContact: ta.Website,
			Details:         ta.Details,
		},
		BlsKey: []byte("mockBlsKey"),
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = true ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnError(mockDBInternalError)
	err := s.SetOwnSpInfo(sp)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_SetOwnSpInfoFailure2(t *testing.T) {
	t.Log("Failure case description: db inserts a row returns error")
	ta := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "123",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	sp := &sptypes.StorageProvider{
		Id:                 ta.ID,
		OperatorAddress:    ta.OperatorAddress,
		FundingAddress:     ta.FundingAddress,
		SealAddress:        ta.SealAddress,
		ApprovalAddress:    ta.ApprovalAddress,
		GcAddress:          "mockGCAddress",
		MaintenanceAddress: "mockMaintenanceAddress",
		TotalDeposit:       sdkmath.NewInt(123),
		Status:             0,
		Endpoint:           ta.Endpoint,
		Description: sptypes.Description{
			Moniker:         ta.Moniker,
			Identity:        ta.Identity,
			Website:         ta.Website,
			SecurityContact: ta.Website,
			Details:         ta.Details,
		},
		BlsKey: []byte("mockBlsKey"),
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = true ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `sp_info` (`operator_address`,`is_own`,`id`,`funding_address`,`seal_address`,`approval_address`,`total_deposit`,`status`,`endpoint`,`moniker`,`identity`,`website`,`security_contact`,`details`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.SetOwnSpInfo(sp)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_SetOwnSpInfoFailure3(t *testing.T) {
	t.Log("Failure case description: db update a row returns error")
	ta := &SpInfoTable{
		OperatorAddress: "mockOperatorAddress",
		IsOwn:           false,
		ID:              1,
		FundingAddress:  "mockFundingAddress",
		SealAddress:     "mockSealAddress",
		ApprovalAddress: "mockApprovalAddress",
		TotalDeposit:    "123",
		Status:          0,
		Endpoint:        "mockEndpoint",
		Moniker:         "mockMoniker",
		Identity:        "mockIdentity",
		Website:         "mockWebsite",
		SecurityContact: "mockSecurityContact",
		Details:         "mockDetails",
	}
	sp := &sptypes.StorageProvider{
		Id:                 ta.ID,
		OperatorAddress:    ta.OperatorAddress,
		FundingAddress:     ta.FundingAddress,
		SealAddress:        ta.SealAddress,
		ApprovalAddress:    ta.ApprovalAddress,
		GcAddress:          "mockGCAddress",
		MaintenanceAddress: "mockMaintenanceAddress",
		TotalDeposit:       sdkmath.NewInt(123),
		Status:             0,
		Endpoint:           ta.Endpoint,
		Description: sptypes.Description{
			Moniker:         ta.Moniker,
			Identity:        ta.Identity,
			Website:         ta.Website,
			SecurityContact: ta.Website,
			Details:         ta.Details,
		},
		BlsKey: []byte("mockBlsKey"),
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `sp_info` WHERE is_own = true ORDER BY `sp_info`.`operator_address` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"operator_address", "is_own", "id", "funding_address", "seal_address", "approval_address",
			"total_deposit", "status", "endpoint", "moniker", "identity", "website", "security_contract", "details"}).
			AddRow(ta.OperatorAddress, ta.IsOwn, ta.ID, ta.FundingAddress, ta.SealAddress, ta.ApprovalAddress, ta.TotalDeposit,
				ta.Status, ta.Endpoint, ta.Moniker, ta.Identity, ta.Website, ta.SecurityContact, ta.Details))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `sp_info` SET `operator_address`=?,`is_own`=?,`id`=?,`funding_address`=?,`seal_address`=?,`approval_address`=?,`total_deposit`=?,`endpoint`=?,`moniker`=?,`identity`=?,`website`=?,`security_contact`=?,`details`=? WHERE is_own = true").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.SetOwnSpInfo(sp)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

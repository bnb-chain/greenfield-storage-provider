package sqldb

import (
	"errors"
	"fmt"

	"cosmossdk.io/math"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"gorm.io/gorm"
)

// UpdateAllSP implements SPInfo interface
func (s *SQLDB) UpdateAllSP(spList []*sptypes.StorageProvider) error {
	for _, value := range spList {
		queryReturn := &SPInfoTable{}
		// 1. check record whether exists
		result := s.db.Where("operator_address = ? and is_own = false", value.GetOperatorAddress()).First(queryReturn)
		sameError := errors.Is(result.Error, gorm.ErrRecordNotFound)
		if result.Error != nil && !sameError {
			return fmt.Errorf("failed to query record in sp info table: %s", result.Error)
		}
		// 2. if there is no record, insert new record; otherwise delete old record, then insert new record
		if queryReturn != nil && sameError {
			if err := s.insertNewRecordInSPInfoTable(value); err != nil {
				return err
			}
		} else {
			result = s.db.Where("operator_address = ? and is_own = false", value.GetOperatorAddress()).Delete(queryReturn)
			if result.Error != nil {
				return fmt.Errorf("failed to detele record in sp info table: %s", result.Error)
			}
			if err := s.insertNewRecordInSPInfoTable(value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SQLDB) insertNewRecordInSPInfoTable(sp *sptypes.StorageProvider) error {
	insertRecord := &SPInfoTable{
		OperatorAddress: sp.GetOperatorAddress(),
		IsOwn:           false,
		FundingAddress:  sp.GetFundingAddress(),
		SealAddress:     sp.GetSealAddress(),
		ApprovalAddress: sp.GetApprovalAddress(),
		TotalDeposit:    sp.TotalDeposit.Int64(),
		Status:          int32(sp.Status),
		Endpoint:        sp.GetEndpoint(),
		Moniker:         sp.GetDescription().Moniker,
		Identity:        sp.GetDescription().Identity,
		Website:         sp.GetDescription().Website,
		SecurityContact: sp.GetDescription().SecurityContact,
		Details:         sp.GetDescription().Identity,
	}
	result := s.db.Create(insertRecord)
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("failed to insert record in sp info table: %s", result.Error)
	}
	return nil
}

// FetchAllSP implements SPInfo interface
func (s *SQLDB) FetchAllSP(status ...sptypes.Status) ([]*sptypes.StorageProvider, error) {
	queryReturn := []SPInfoTable{}
	if len(status) == 0 {
		result := s.db.Where("operator_address = ? and is_own = false").Find(&queryReturn)
		if result.Error != nil {
			return nil, fmt.Errorf("failed to query sp info table: %s", result.Error)
		}
	} else {
		for _, val := range status {
			temp := []SPInfoTable{}
			result := s.db.Where("operator_address = ? and is_own = false and status = ?", int32(val)).Find(&temp)
			if result.Error != nil {
				return nil, fmt.Errorf("failed to query sp info table: %s", result.Error)
			}
			queryReturn = append(queryReturn, temp...)
		}
	}
	records := []*sptypes.StorageProvider{}
	for _, value := range queryReturn {
		records = append(records, &sptypes.StorageProvider{
			OperatorAddress: value.OperatorAddress,
			FundingAddress:  value.FundingAddress,
			SealAddress:     value.SealAddress,
			ApprovalAddress: value.ApprovalAddress,
			TotalDeposit:    math.NewInt(value.TotalDeposit),
			Status:          sptypes.Status(value.Status),
			Endpoint:        value.Endpoint,
			Description: sptypes.Description{
				Moniker:         value.Moniker,
				Identity:        value.Identity,
				Website:         value.Website,
				SecurityContact: value.SecurityContact,
				Details:         value.Details,
			},
		})
	}
	return records, nil
}

// FetchAllSPWithoutOwnSP implements SPInfo interface
func (s *SQLDB) FetchAllSPWithoutOwnSP(status ...sptypes.Status) ([]*sptypes.StorageProvider, error) {
	ownSP, err := s.GetOwnSPInfo()
	if err != nil {
		return nil, err
	}
	queryReturn := []SPInfoTable{}
	if len(status) == 0 {
		result := s.db.Where("operator_address != ?", ownSP.GetOperatorAddress()).Find(&queryReturn)
		if result.Error != nil {
			return nil, fmt.Errorf("failed to query sp info table: %s", result.Error)
		}
	} else {
		for _, val := range status {
			temp := []SPInfoTable{}
			result := s.db.Where("status = ? and operator_address != ?", int32(val), ownSP.GetOperatorAddress()).Find(&temp)
			if result.Error != nil {
				return nil, fmt.Errorf("failed to query sp info table: %s", result.Error)
			}
			queryReturn = append(queryReturn, temp...)
		}
	}

	records := []*sptypes.StorageProvider{}
	for _, value := range queryReturn {
		records = append(records, &sptypes.StorageProvider{
			OperatorAddress: value.OperatorAddress,
			FundingAddress:  value.FundingAddress,
			SealAddress:     value.SealAddress,
			ApprovalAddress: value.ApprovalAddress,
			TotalDeposit:    math.NewInt(value.TotalDeposit),
			Status:          sptypes.Status(value.Status),
			Endpoint:        value.Endpoint,
			Description: sptypes.Description{
				Moniker:         value.Moniker,
				Identity:        value.Identity,
				Website:         value.Website,
				SecurityContact: value.SecurityContact,
				Details:         value.Details,
			},
		})
	}
	return records, nil
}

// GetSPByAddress implement SPInfo interface
func (s *SQLDB) GetSPByAddress(address string, addressType SPAddressType) (*sptypes.StorageProvider, error) {
	condition, err := getAddressCondition(addressType)
	if err != nil {
		return nil, err
	}
	queryReturn := &SPInfoTable{}
	result := s.db.First(queryReturn, condition, address)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query sp info table: %s", result.Error)
	}
	return &sptypes.StorageProvider{
		OperatorAddress: queryReturn.OperatorAddress,
		FundingAddress:  queryReturn.FundingAddress,
		SealAddress:     queryReturn.SealAddress,
		ApprovalAddress: queryReturn.ApprovalAddress,
		TotalDeposit:    math.NewInt(queryReturn.TotalDeposit),
		Status:          sptypes.Status(queryReturn.Status),
		Endpoint:        queryReturn.Endpoint,
		Description: sptypes.Description{
			Moniker:         queryReturn.Moniker,
			Identity:        queryReturn.Identity,
			Website:         queryReturn.Website,
			SecurityContact: queryReturn.SecurityContact,
			Details:         queryReturn.Details,
		},
	}, nil
}

func getAddressCondition(addressType SPAddressType) (string, error) {
	var condition string
	switch addressType {
	case OperatorAddressType:
		condition = "operator_address = ?"
	case FundingAddressType:
		condition = "funding_address = ?"
	case SealAddressType:
		condition = "seal_address = ?"
	case ApprovalAddressType:
		condition = "approval_address = ?"
	default:
		return "", fmt.Errorf("unknown address type")
	}
	return condition, nil
}

// // GetSPByEndpoint implement SPInfo interface
func (s *SQLDB) GetSPByEndpoint(endpoint string) (*sptypes.StorageProvider, error) {
	queryReturn := &SPInfoTable{}
	result := s.db.First(queryReturn, "endpoint = ?", endpoint)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query sp info table: %s", result.Error)
	}
	return &sptypes.StorageProvider{
		OperatorAddress: queryReturn.OperatorAddress,
		FundingAddress:  queryReturn.FundingAddress,
		SealAddress:     queryReturn.SealAddress,
		ApprovalAddress: queryReturn.ApprovalAddress,
		TotalDeposit:    math.NewInt(queryReturn.TotalDeposit),
		Status:          sptypes.Status(queryReturn.Status),
		Endpoint:        queryReturn.Endpoint,
		Description: sptypes.Description{
			Moniker:         queryReturn.Moniker,
			Identity:        queryReturn.Identity,
			Website:         queryReturn.Website,
			SecurityContact: queryReturn.SecurityContact,
			Details:         queryReturn.Details,
		},
	}, nil
}

// GetOwnSPInfo implements SPInfo interface
func (s *SQLDB) GetOwnSPInfo() (*sptypes.StorageProvider, error) {
	queryReturn := &SPInfoTable{}
	result := s.db.First(queryReturn, "is_own = true")
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query own sp record in sp info table: %s", result.Error)
	}
	return &sptypes.StorageProvider{
		OperatorAddress: queryReturn.OperatorAddress,
		FundingAddress:  queryReturn.FundingAddress,
		SealAddress:     queryReturn.SealAddress,
		ApprovalAddress: queryReturn.ApprovalAddress,
		TotalDeposit:    math.NewInt(queryReturn.TotalDeposit),
		Status:          sptypes.Status(queryReturn.Status),
		Endpoint:        queryReturn.Endpoint,
		Description: sptypes.Description{
			Moniker:         queryReturn.Moniker,
			Identity:        queryReturn.Identity,
			Website:         queryReturn.Website,
			SecurityContact: queryReturn.SecurityContact,
			Details:         queryReturn.Details,
		},
	}, nil
}

// SetOwnSPInfo implements SPInfo interface
func (s *SQLDB) SetOwnSPInfo(sp *sptypes.StorageProvider) error {
	_, err := s.GetOwnSPInfo()
	isNotFound := errors.Is(err, gorm.ErrRecordNotFound)
	if err != nil && !isNotFound {
		return err
	}

	insertRecord := &SPInfoTable{
		OperatorAddress: sp.GetOperatorAddress(),
		IsOwn:           true,
		FundingAddress:  sp.GetFundingAddress(),
		SealAddress:     sp.GetSealAddress(),
		ApprovalAddress: sp.GetApprovalAddress(),
		TotalDeposit:    sp.TotalDeposit.Int64(),
		Status:          int32(sp.GetStatus()),
		Endpoint:        sp.GetEndpoint(),
		Moniker:         sp.GetDescription().Moniker,
		Identity:        sp.GetDescription().Identity,
		Website:         sp.GetDescription().Website,
		SecurityContact: sp.GetDescription().SecurityContact,
		Details:         sp.GetDescription().Details,
	}
	// if there is no records in SPInfoTable, insert a new record
	if isNotFound {
		result := s.db.Create(insertRecord)
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("failed to insert own sp record in sp info table: %s", result.Error)
		}
		return nil
	} else {
		// if there is a record in SPInfoTable, update record
		result := s.db.Model(&SPInfoTable{}).Where("is_own = true").Updates(insertRecord)
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("failed to update own sp record in sp info table: %s", result.Error)
		}
		return nil
	}
}

// GetStorageParams implements StorageParam interface
func (s *SQLDB) GetStorageParams() (*storagetypes.Params, error) {
	queryReturn := &StorageParamsTable{}
	result := s.db.Last(queryReturn)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query storage params table: %s", result.Error)
	}
	return &storagetypes.Params{
		MaxSegmentSize:          queryReturn.MaxSegmentSize,
		RedundantDataChunkNum:   queryReturn.RedundantDataChunkNum,
		RedundantParityChunkNum: queryReturn.RedundantParityChunkNum,
		MaxPayloadSize:          queryReturn.MaxPayloadSize,
	}, nil
}

// SetStorageParams implements StorageParam interface
func (s *SQLDB) SetStorageParams(params *storagetypes.Params) error {
	queryReturn := &StorageParamsTable{}
	result := s.db.Last(queryReturn)
	isNotFound := errors.Is(result.Error, gorm.ErrRecordNotFound)
	if result.Error != nil && !isNotFound {
		return fmt.Errorf("failed to query storage params table: %s", result.Error)
	}

	insertParamsRecord := StorageParamsTable{
		MaxSegmentSize:          params.GetMaxSegmentSize(),
		RedundantDataChunkNum:   params.GetRedundantDataChunkNum(),
		RedundantParityChunkNum: params.GetRedundantParityChunkNum(),
		MaxPayloadSize:          params.GetMaxPayloadSize(),
	}
	// if there is no records in StorageParamsTable, insert a new record
	if isNotFound {
		result = s.db.Create(&insertParamsRecord)
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("failed to insert storage params table: %s", result.Error)
		}
		return nil
	} else {
		queryCondition := &StorageParamsTable{ID: queryReturn.ID}
		result = s.db.Model(queryCondition).Updates(insertParamsRecord)
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("failed to update storage params table: %s", result.Error)
		}
		return nil
	}
}

package sqldb

import (
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"gorm.io/gorm"
)

// UpdateAllSp update(maybe overwrite) all sp info in db
func (s *SpDBImpl) UpdateAllSp(spList []*sptypes.StorageProvider) error {
	for _, value := range spList {
		queryReturn := &SpInfoTable{}
		// 1. check record whether exists
		result := s.db.Where("operator_address = ? and is_own = false", value.GetOperatorAddress()).First(queryReturn)
		recordNotFound := errors.Is(result.Error, gorm.ErrRecordNotFound)
		if result.Error != nil && !recordNotFound {
			return fmt.Errorf("failed to query record in sp info table: %s", result.Error)
		}
		// 2. if there is no record, insert new record; otherwise delete old record, then insert new record
		if recordNotFound {
			if err := s.insertNewRecordInSpInfoTable(value); err != nil {
				return err
			}
		} else {
			result = s.db.Where("operator_address = ? and is_own = false", value.GetOperatorAddress()).Delete(queryReturn)
			if result.Error != nil {
				return fmt.Errorf("failed to detele record in sp info table: %s", result.Error)
			}
			if err := s.insertNewRecordInSpInfoTable(value); err != nil {
				return err
			}
		}
	}
	return nil
}

// insertNewRecordInSpInfoTable insert a new record in sp info table
func (s *SpDBImpl) insertNewRecordInSpInfoTable(sp *sptypes.StorageProvider) error {
	insertRecord := &SpInfoTable{
		OperatorAddress: sp.GetOperatorAddress(),
		IsOwn:           false,
		FundingAddress:  sp.GetFundingAddress(),
		SealAddress:     sp.GetSealAddress(),
		ApprovalAddress: sp.GetApprovalAddress(),
		TotalDeposit:    sp.GetTotalDeposit().String(),
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

// FetchAllSp get all sp info
func (s *SpDBImpl) FetchAllSp(status ...sptypes.Status) ([]*sptypes.StorageProvider, error) {
	queryReturn := []SpInfoTable{}
	if len(status) == 0 {
		result := s.db.Where("is_own = false").Find(&queryReturn)
		if result.Error != nil {
			return nil, fmt.Errorf("failed to query sp info table: %s", result.Error)
		}
	} else {
		for _, val := range status {
			temp := []SpInfoTable{}
			result := s.db.Where("is_own = false and status = ?", int32(val)).Find(&temp)
			if result.Error != nil {
				return nil, fmt.Errorf("failed to query sp info table: %s", result.Error)
			}
			queryReturn = append(queryReturn, temp...)
		}
	}
	records := []*sptypes.StorageProvider{}
	for _, value := range queryReturn {
		totalDeposit, ok := sdkmath.NewIntFromString(value.TotalDeposit)
		if !ok {
			return records, fmt.Errorf("failed to parse int")
		}
		records = append(records, &sptypes.StorageProvider{
			OperatorAddress: value.OperatorAddress,
			FundingAddress:  value.FundingAddress,
			SealAddress:     value.SealAddress,
			ApprovalAddress: value.ApprovalAddress,
			TotalDeposit:    totalDeposit,
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

// FetchAllSpWithoutOwnSp get all spp info without own sp info, own sp is identified by is_own field in db
func (s *SpDBImpl) FetchAllSpWithoutOwnSp(status ...sptypes.Status) ([]*sptypes.StorageProvider, error) {
	ownSp, err := s.GetOwnSpInfo()
	if err != nil {
		return nil, err
	}
	queryReturn := []SpInfoTable{}
	if len(status) == 0 {
		result := s.db.Where("operator_address != ?", ownSp.GetOperatorAddress()).Find(&queryReturn)
		if result.Error != nil {
			return nil, fmt.Errorf("failed to query sp info table: %s", result.Error)
		}
	} else {
		for _, val := range status {
			temp := []SpInfoTable{}
			result := s.db.Where("status = ? and operator_address != ?", int32(val), ownSp.GetOperatorAddress()).Find(&temp)
			if result.Error != nil {
				return nil, fmt.Errorf("failed to query sp info table: %s", result.Error)
			}
			queryReturn = append(queryReturn, temp...)
		}
	}

	records := []*sptypes.StorageProvider{}
	for _, value := range queryReturn {
		totalDeposit, ok := sdkmath.NewIntFromString(value.TotalDeposit)
		if !ok {
			return records, fmt.Errorf("failed to parse int")
		}
		records = append(records, &sptypes.StorageProvider{
			OperatorAddress: value.OperatorAddress,
			FundingAddress:  value.FundingAddress,
			SealAddress:     value.SealAddress,
			ApprovalAddress: value.ApprovalAddress,
			TotalDeposit:    totalDeposit,
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

// GetSpByAddress query sp info in db by address and address type
func (s *SpDBImpl) GetSpByAddress(address string, addressType SpAddressType) (*sptypes.StorageProvider, error) {
	condition, err := getAddressCondition(addressType)
	if err != nil {
		return nil, err
	}
	queryReturn := &SpInfoTable{}
	result := s.db.First(queryReturn, condition, address)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query sp info table: %s", result.Error)
	}
	totalDeposit, ok := sdkmath.NewIntFromString(queryReturn.TotalDeposit)
	if !ok {
		return nil, fmt.Errorf("failed to parse int")
	}
	return &sptypes.StorageProvider{
		OperatorAddress: queryReturn.OperatorAddress,
		FundingAddress:  queryReturn.FundingAddress,
		SealAddress:     queryReturn.SealAddress,
		ApprovalAddress: queryReturn.ApprovalAddress,
		TotalDeposit:    totalDeposit,
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

// getAddressCondition return different condition by address type
func getAddressCondition(addressType SpAddressType) (string, error) {
	var condition string
	switch addressType {
	case OperatorAddressType:
		condition = "operator_address = ? and is_own = false"
	case FundingAddressType:
		condition = "funding_address = ? and is_own = false"
	case SealAddressType:
		condition = "seal_address = ? and is_own = false"
	case ApprovalAddressType:
		condition = "approval_address = ? and is_own = false"
	default:
		return "", fmt.Errorf("unknown address type")
	}
	return condition, nil
}

// GetSpByEndpoint query sp info by endpoint
func (s *SpDBImpl) GetSpByEndpoint(endpoint string) (*sptypes.StorageProvider, error) {
	queryReturn := &SpInfoTable{}
	result := s.db.First(queryReturn, "endpoint = ? and is_own = false", endpoint)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query sp info table: %s", result.Error)
	}
	totalDeposit, ok := sdkmath.NewIntFromString(queryReturn.TotalDeposit)
	if !ok {
		return nil, fmt.Errorf("failed to parse int")
	}
	return &sptypes.StorageProvider{
		OperatorAddress: queryReturn.OperatorAddress,
		FundingAddress:  queryReturn.FundingAddress,
		SealAddress:     queryReturn.SealAddress,
		ApprovalAddress: queryReturn.ApprovalAddress,
		TotalDeposit:    totalDeposit,
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

// GetOwnSpInfo query own sp info in db
func (s *SpDBImpl) GetOwnSpInfo() (*sptypes.StorageProvider, error) {
	queryReturn := &SpInfoTable{}
	result := s.db.First(queryReturn, "is_own = true")
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query own sp record in sp info table: %s", result.Error)
	}
	totalDeposit, ok := sdkmath.NewIntFromString(queryReturn.TotalDeposit)
	if !ok {
		return nil, fmt.Errorf("failed to parse int")
	}
	return &sptypes.StorageProvider{
		OperatorAddress: queryReturn.OperatorAddress,
		FundingAddress:  queryReturn.FundingAddress,
		SealAddress:     queryReturn.SealAddress,
		ApprovalAddress: queryReturn.ApprovalAddress,
		TotalDeposit:    totalDeposit,
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

// SetOwnSpInfo set(maybe overwrite) own sp info to db
func (s *SpDBImpl) SetOwnSpInfo(sp *sptypes.StorageProvider) error {
	spInfo, err := s.GetOwnSpInfo()
	if err != nil && !strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
		return err
	}

	insertRecord := &SpInfoTable{
		OperatorAddress: sp.GetOperatorAddress(),
		IsOwn:           true,
		FundingAddress:  sp.GetFundingAddress(),
		SealAddress:     sp.GetSealAddress(),
		ApprovalAddress: sp.GetApprovalAddress(),
		TotalDeposit:    sp.GetTotalDeposit().String(),
		Status:          int32(sp.GetStatus()),
		Endpoint:        sp.GetEndpoint(),
		Moniker:         sp.GetDescription().Moniker,
		Identity:        sp.GetDescription().Identity,
		Website:         sp.GetDescription().Website,
		SecurityContact: sp.GetDescription().SecurityContact,
		Details:         sp.GetDescription().Details,
	}
	// if there is no records in SPInfoTable, insert a new record
	if spInfo == nil {
		result := s.db.Create(insertRecord)
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("failed to insert own sp record in sp info table: %s", result.Error)
		}
		return nil
	} else {
		// if there is a record in SPInfoTable, update record
		result := s.db.Model(&SpInfoTable{}).Where("is_own = true").Updates(insertRecord)
		if result.Error != nil {
			return fmt.Errorf("failed to update own sp record in sp info table: %s", result.Error)
		}
		return nil
	}
}

// GetStorageParams query storage params in db
func (s *SpDBImpl) GetStorageParams() (*storagetypes.Params, error) {
	queryReturn := &StorageParamsTable{}
	result := s.db.Last(queryReturn)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query storage params table: %s", result.Error)
	}
	return &storagetypes.Params{
		VersionedParams: storagetypes.VersionedParams{
			MaxSegmentSize:          queryReturn.MaxSegmentSize,
			RedundantDataChunkNum:   queryReturn.RedundantDataChunkNum,
			RedundantParityChunkNum: queryReturn.RedundantParityChunkNum,
		},

		MaxPayloadSize: queryReturn.MaxPayloadSize,
	}, nil
}

// SetStorageParams set(maybe overwrite) storage params to db
func (s *SpDBImpl) SetStorageParams(params *storagetypes.Params) error {
	queryReturn := &StorageParamsTable{}
	result := s.db.Last(queryReturn)
	recordNotFound := errors.Is(result.Error, gorm.ErrRecordNotFound)
	if result.Error != nil && !recordNotFound {
		return fmt.Errorf("failed to query storage params table: %s", result.Error)
	}

	insertParamsRecord := &StorageParamsTable{
		MaxSegmentSize:          params.VersionedParams.GetMaxSegmentSize(),
		RedundantDataChunkNum:   params.VersionedParams.GetRedundantDataChunkNum(),
		RedundantParityChunkNum: params.VersionedParams.GetRedundantParityChunkNum(),
		MaxPayloadSize:          params.GetMaxPayloadSize(),
	}
	// if there is no records in StorageParamsTable, insert a new record
	if recordNotFound {
		result = s.db.Create(insertParamsRecord)
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("failed to insert storage params table: %s", result.Error)
		}
		return nil
	} else {
		queryCondition := &StorageParamsTable{ID: queryReturn.ID}
		result = s.db.Model(queryCondition).Updates(insertParamsRecord)
		if result.Error != nil {
			return fmt.Errorf("failed to update storage params table: %s", result.Error)
		}
		return nil
	}
}

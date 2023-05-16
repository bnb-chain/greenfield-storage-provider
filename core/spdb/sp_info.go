package spdb

import (
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

// SpAddressType identify address type of SP
type SpAddressType int32

const (
	OperatorAddressType SpAddressType = iota + 1
	FundingAddressType
	SealAddressType
	ApprovalAddressType
)

type SPInfoDB interface {
	// UpdateAllSp update all sp info, delete old sp info
	UpdateAllSp(spList []*sptypes.StorageProvider) error
	// FetchAllSp if status is nil return all sp info; otherwise return sp info by status
	FetchAllSp(status ...sptypes.Status) ([]*sptypes.StorageProvider, error)
	// FetchAllSpWithoutOwnSp if status is nil return all sp info without own sp;
	// otherwise return sp info by status without own sp
	FetchAllSpWithoutOwnSp(status ...sptypes.Status) ([]*sptypes.StorageProvider, error)
	// GetSpByAddress return sp info by address and addressType
	GetSpByAddress(address string, addressType SpAddressType) (*sptypes.StorageProvider, error)
	// GetSpByEndpoint return sp info by endpoint
	GetSpByEndpoint(endpoint string) (*sptypes.StorageProvider, error)
	// GetOwnSpInfo return own sp info
	GetOwnSpInfo() (*sptypes.StorageProvider, error)
	// SetOwnSpInfo set(maybe overwrite) own sp info
	SetOwnSpInfo(sp *sptypes.StorageProvider) error
}

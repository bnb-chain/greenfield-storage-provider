package metadata

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/bnb-chain/greenfield-storage-provider/model"
)

// Block is the structure for Block Info
type Block struct {
	// ID defines db auto_increment id of block
	ID uint64 `gorm:"id"`
	// Hash defines the block hash
	Hash common.Hash `gorm:"hash"`
	// Height defines the block height
	Height uint64 `gorm:"height"`
	// LastCommitHash defines the latest commit hash of block
	LastCommitHash common.Hash `gorm:"last_commit_hash"`
	// DataHash defines the data hash of block
	DataHash common.Hash `gorm:"data_hash"`
	// ValidatorsHash defines the validators hash of block
	ValidatorsHash common.Hash `gorm:"validators_hash"`
	// NextValidatorsHash defines the next validators hash of block
	NextValidatorsHash common.Hash `gorm:"next_validators_hash"`
	// ConsensusHash defines the consensus hash of block
	ConsensusHash common.Hash `gorm:"consensus_hash"`
	// AppHash defines the app hash of block
	AppHash common.Hash `gorm:"app_hash"`
	// LastResultsHash defines the last results hash of block
	LastResultsHash common.Hash `gorm:"last_results_hash"`
	// EvidenceHash defines the evidence hash of block
	EvidenceHash common.Hash `gorm:"evidence_hash"`
	// ProposerAddress defines the proposer address of block
	ProposerAddress common.Address `gorm:"proposer_address"`
	// Timestamp defines the timestamp of block
	Timestamp int64 `gorm:"timestamp"`
	// NumTxs defines the number of transactions of block
	NumTxs int64 `gorm:"num_txs"`
	// TotalGas defines the total gas of block
	TotalGas int64 `gorm:"total_gas"`
}

func (a *Block) TableName() string {
	return model.BlockTableName
}

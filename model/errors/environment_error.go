package errors

import "errors"

var (
	// ErrDSNNotSet defines deny request by ip error
	ErrDSNNotSet       = errors.New("dsn config is not set in environment")
	ErrChainConfNotSet = errors.New("chain config is not set in environment")
)

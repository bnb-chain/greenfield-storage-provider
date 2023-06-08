package sqldb

import "errors"

var (
	// ErrCheckQuotaEnough defines check quota is enough
	ErrCheckQuotaEnough = errors.New("quota is not enough")
)

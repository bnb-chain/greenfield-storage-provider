package spdb

import "time"

type OffChainAuthKey struct {
	UserAddress string
	Domain      string

	CurrentNonce     int32
	CurrentPublicKey string
	NextNonce        int32
	ExpiryDate       time.Time

	CreatedTime  time.Time
	ModifiedTime time.Time
}

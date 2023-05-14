package spdb

//import "time"
//
//// OffChainAuthKeyTable table schema
//type OffChainAuthKeyTable struct {
//	UserAddress string `gorm:"primary_key"`
//	Domain      string `gorm:"primary_key"`
//
//	CurrentNonce     int32
//	CurrentPublicKey string
//	NextNonce        int32
//	ExpiryDate       time.Time
//
//	CreatedTime  time.Time
//	ModifiedTime time.Time
//}
//
//type OffChainAuthKey interface {
//	GetAuthKey(userAddress string, domain string) (*OffChainAuthKeyTable, error)
//	UpdateAuthKey(userAddress string, domain string, oldNonce int32, newNonce int32, newPublicKey string, newExpiryDate time.Time) error
//	InsertAuthKey(newRecord *OffChainAuthKeyTable) error
//}

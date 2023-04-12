package errors

const (
	// RPCErrCode defines storage provider rpc error code
	RPCErrCode = 10000
	// ErrorCodeBadRequest defines bad request error code
	ErrorCodeBadRequest = 40001
	// ErrorCodeNotFound defines not found error code
	ErrorCodeNotFound = 40004
	// ErrorCodeInternalError defines internal error code
	ErrorCodeInternalError = 50001
)

const (
	// parse error
	ParseStrToIntErrCode         = 15000
	ConvertStrToByteSliceErrCode = 15001
	// db error
	UnKnownAddressType                = 19999
	QueryInSPInfoTableErrCode         = 20000
	DeleteInSPInfoTableErrCode        = 20001
	InsertInSPInfoTableErrCode        = 20002
	QueryOwnSPInSPInfoTableErrCode    = 20003
	InsertOwnSPInSPInfoTableErrCode   = 20004
	UpdateOwnSPInSPInfoTableErrCode   = 20005
	QueryInStorageParamsTableErrCode  = 20006
	InsertInStorageParamsTableErrCode = 20007
	UpdateInStorageParamsTableErrCode = 20008
	QueryInJobTableErrCode            = 20009
	InsertInJobTableErrCode           = 20010
	UpdateInJobTableErrCode           = 20011
	QueryInObjectTableErrCode         = 20012
	InsertInObjectTableErrCode        = 20013
	UpdateInObjectTableErrCode        = 20014
	RecordNotFound                    = 20015
	QueryInIntegrityMetaTableErrCode  = 20016
	InsertInIntegrityMetaTableErrCode = 20017
)

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
	// deinfe common error code, from 10000 to 14999
	MismatchIntegrityHashErrCode    = 11000
	CacheMissedErrCode              = 11001
	DanglingPointerErrCode          = 11002
	PayloadStreamErrCode            = 11003
	ResourceMgrBeginSpanErrCode     = 11004
	ComputePieceSizeErrCode         = 11005
	ResourceMgrReserveMemoryErrCode = 11006
	ObjectNotFoundErrCode           = 11007
	HexDecodeStringErrCode          = 11008
	StringToByteSliceErrCode        = 11009
	ParseStringToIntErrCode         = 11010
	// sp database error, from 15000 to 19999
	DBRecordNotFoundErrCode             = 15000
	DBUnknownAddressTypeErrCode         = 15001
	DBQueryInSPInfoTableErrCode         = 15100
	DBDeleteInSPInfoTableErrCode        = 15101
	DBInsertInSPInfoTableErrCode        = 15102
	DBQueryOwnSPInSPInfoTableErrCode    = 15103
	DBInsertOwnSPInSPInfoTableErrCode   = 15104
	DBUpdateOwnSPInSPInfoTableErrCode   = 15105
	DBQueryInStorageParamsTableErrCode  = 15106
	DBInsertInStorageParamsTableErrCode = 15107
	DBUpdateInStorageParamsTableErrCode = 15108
	DBQueryInJobTableErrCode            = 15109
	DBInsertInJobTableErrCode           = 15110
	DBUpdateInJobTableErrCode           = 15111
	DBQueryInObjectTableErrCode         = 15112
	DBInsertInObjectTableErrCode        = 15113
	DBUpdateInObjectTableErrCode        = 15114
	DBQueryInIntegrityMetaTableErrCode  = 15117
	DBInsertInIntegrityMetaTableErrCode = 15118
	DBQueryInServiceConfigTableErrCode  = 15119
	DBDeleteInServiceConfigTableErrCode = 15120
	DBInsertInServiceConfigTableErrCode = 15121
	DBQuotaNotEnoughErrCode             = 15122
	DBQueryInBucketTrafficTableErrCode  = 15123
	DBInsertInBucketTrafficTableErrCode = 15124
	DBUpdateInBucketTrafficTableErrCode = 15125
	DBQueryInReadRecordTableErrCode     = 15126
	DBInsertInReadRecordTableErrCode    = 15127

	// uploader service error, from 20000 to 20999
	UploaderMismatchChecksumNumErrCode = 20100

	// challenge service error, from 21000 to 21999

	// downloader service error, from 22000 to 22999
	DownloaderInvalidPieceInfoParamsErrCode = 22100

	// signer service error, from 23000 to 23999
	SignerSignIntegrityHashErrCode = 23100

	// p2p service, from 24000 to 24999

	// receiver service error code, from 25000 to 25999

	// task node service error, from 26000 to 26999

	// piece store error code, from 27000 to 27999
	PieceStorePutObjectError = 27100
	PieceStoreGetObjectError = 27101
)

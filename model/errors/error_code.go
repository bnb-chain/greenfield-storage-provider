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
	SuccessCode    = 0
	UnknownErrCode = 1
	// deinfe common error code, from 10000 to 14999
	InternalErrCode                   = 10000
	MismatchIntegrityHashErrCode      = 11000
	CacheMissedErrCode                = 11001
	DanglingPointerErrCode            = 11002
	PayloadStreamErrCode              = 11003
	ResourceMgrBeginSpanErrCode       = 11004
	ComputePieceSizeErrCode           = 11005
	ResourceMgrReserveMemoryErrCode   = 11006
	NoSuchObjectErrCode               = 11007
	NoSuchBucketErrCode               = 11008
	HexDecodeStringErrCode            = 11009
	StringToByteSliceErrCode          = 11010
	ParseStringToIntErrCode           = 11011
	RouterNotFoundErrCode             = 11012
	InvalidHeaderErrCode              = 11013
	StringToInt64ErrCode              = 11014
	InvalidBucketNameErrCode          = 11015
	InvalidObjectNameErrCode          = 11016
	ZeroPayloadErrCode                = 11017
	InvalidRangeErrCode               = 11020
	InvalidAddressErrCode             = 11021
	InvalidAuthorizationFormatErrCode = 11022
	InconsistentRequestErrCode        = 11023
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

	// gateway service error code, from 20000 to 20999
	AuthorizationFormatErrCode          = 20100
	InconsistentCanonicalRequestErrCode = 20101
	SignatureConsistentErrCode          = 20102
	UnsupportedSignTypeErrCode          = 20103
	NoPermissionErrCode                 = 20104
	ObjectNotCreatedErrCode             = 20105
	ObjectNotSealedErrCode              = 20106
	CheckPaymentAccountActiveErrCode    = 20107

	// uploader service error, from 21000 to 21999
	UploaderMismatchChecksumNumErrCode = 21100

	// challenge service error, from 22000 to 22999

	// downloader service error, from 23000 to 23999
	DownloaderInvalidPieceInfoParamsErrCode = 23100

	// signer service error, from 24000 to 24999
	SignerSignIntegrityHashErrCode = 24100

	// p2p service, from 25000 to 25999

	// receiver service error code, from 26000 to 26999

	// task node service error, from 27000 to 27999

	// piece store error code, from 28000 to 28999
	PieceStorePutObjectError = 28100
	PieceStoreGetObjectError = 28101

	// Greenfield chain error code, from 29000 to 29999
	ChainGetLatestBlockErrCode           = 29100
	ChainQueryAccountErrCode             = 29101
	ChainQuerySPListErrCode              = 29102
	ChainQueryStorageParamsErrCode       = 29103
	ChainHeadBucketErrCode               = 29104
	ChainHeadObjectErrCode               = 29105
	ChainHeadObjectByIDErrCode           = 29106
	ChainSealObjectTimeoutErrCode        = 29107
	ChainQueryStreamRecordErrCode        = 29108
	ChainGetObjetVerifyPermissionErrCode = 29109
	ChainPutObjetVerifyPermissionErrCode = 29110
)

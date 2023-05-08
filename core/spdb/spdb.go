package spdb

type SPDB interface {
	JobDB
	ObjectDB
	ObjectIntegrityDB
	TrafficDB
	SPInfoDB
	GCObjectInfoDB
	StorageParamDB
	//OffChainAuthKey
}

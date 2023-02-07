package metalevel

import "encoding/binary"

// The fields below define the low level database schema prefixing.
var (
	// IntegrityMetaPrefix + objectID -> PrimaryIntegrityMetaKey
	IntegrityMetaPrefix = []byte("IntegrityMeta")
	// PayloadAskingPrefix + bucket + object -> UploadPayloadMetaKey
	PayloadAskingPrefix = "PayloadAskingInfo"
)

// IntegrityMetaKey return the integrity meta key.
func IntegrityMetaKey(prefix string, objectID uint64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, objectID)
	return append(append([]byte(prefix), IntegrityMetaPrefix...), buf...)
}

// UploadPayloadAsingKey return the payload asking info.
func UploadPayloadAsingKey(prefix, bucket, object string) []byte {
	return []byte(prefix + PayloadAskingPrefix + bucket + object)
}

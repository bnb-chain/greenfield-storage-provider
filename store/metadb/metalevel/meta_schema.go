package metalevel

import (
	"encoding/binary"

	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
)

// The fields below define the low level database schema prefixing.
var (
	// IntegrityMetaPrefix + objectID + isPrimary + redundancyType + ecIdx-> PrimaryIntegrityMetaKey
	IntegrityMetaPrefix = []byte("IntegrityMeta")
	// PayloadAskingPrefix + bucketName + objectName -> UploadPayloadMetaKey
	PayloadAskingPrefix = "PayloadAskingInfo"
)

// IntegrityMetaKey return the integrity meta key.
func IntegrityMetaKey(prefix string, objectID uint64,
	isPrimary bool, redundancyType ptypes.RedundancyType, ecIdx uint32) []byte {
	var (
		buf             = make([]byte, 8+2+4+4)
		isPrimaryUint16 uint16
	)
	if isPrimary {
		isPrimaryUint16 = 1
	}

	binary.BigEndian.PutUint64(buf, objectID)
	binary.BigEndian.AppendUint16(buf, isPrimaryUint16)
	binary.BigEndian.AppendUint32(buf, uint32(redundancyType))
	binary.BigEndian.AppendUint32(buf, ecIdx)
	return append(append([]byte(prefix), IntegrityMetaPrefix...), buf...)
}

// UploadPayloadAsingKey return the payload asking info.
func UploadPayloadAsingKey(prefix, bucketName, objectName string) []byte {
	return []byte(prefix + PayloadAskingPrefix + bucketName + objectName)
}

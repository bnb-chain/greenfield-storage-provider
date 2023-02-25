package metalevel

import (
	"encoding/binary"
)

// The fields below define the low level database schema prefixing.
var (
	// IntegrityMetaPrefix + objectID + isPrimary + redundancyType + ecIdx-> PrimaryIntegrityMetaKey
	IntegrityMetaPrefix = []byte("IntegrityMeta")
)

// IntegrityMetaKey return the integrity meta key.
func IntegrityMetaKey(prefix string, objectID uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, objectID)
	return append(append([]byte(prefix), IntegrityMetaPrefix...), buf...)
}

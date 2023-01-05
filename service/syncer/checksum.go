package syncer

import (
	"crypto/sha256"
)

func generateChecksum(pieceData []byte) []byte {
	hash := sha256.New()
	hash.Write(pieceData)
	return hash.Sum(nil)
}

func generateIntegrityHash(checksumList [][]byte, storageProviderID string) []byte {
	hash := sha256.New()
	for _, j := range checksumList {
		hash.Write(j)
	}
	hash.Write([]byte(storageProviderID))
	return hash.Sum(nil)
}

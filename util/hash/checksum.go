package hash

import (
	"bytes"
	"crypto/sha256"
)

// GenerateChecksum generates the checksum of one piece data
func GenerateChecksum(pieceData []byte) []byte {
	hash := sha256.New()
	hash.Write(pieceData)
	return hash.Sum(nil)
}

// GenerateIntegrityHash generates integrity hash of all piece data checksum
func GenerateIntegrityHash(checksumList [][]byte) []byte {
	hash := sha256.New()
	checksumBytesTotal := bytes.Join(checksumList, []byte(""))
	hash.Write(checksumBytesTotal)
	return hash.Sum(nil)
}

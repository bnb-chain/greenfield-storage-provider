package hash

import (
	"bytes"
	"crypto/sha256"
	"fmt"
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

// CheckIntegrityHash is used by challenge
func CheckIntegrityHash(integrityHash []byte, checksumList [][]byte, index int, pieceData []byte) error {
	if len(checksumList) <= index {
		return fmt.Errorf("checksum list is not correct")
	}
	if !bytes.Equal(checksumList[index], GenerateChecksum(pieceData)) {
		return fmt.Errorf("piece data checksum is not correct")
	}
	if !bytes.Equal(integrityHash, GenerateIntegrityHash(checksumList)) {
		return fmt.Errorf("piece data integrity checksum is not correct")
	}
	return nil
}

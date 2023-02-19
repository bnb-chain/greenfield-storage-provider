package hash

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// GenerateChecksum generates the checksum of piece data
func GenerateChecksum(pieceData []byte) []byte {
	hash := sha256.New()
	hash.Write(pieceData)
	return hash.Sum(nil)
}

// GenerateIntegrityHash generate integrity hash of ec data
func GenerateIntegrityHash(checksumList [][]byte) []byte {
	hash := sha256.New()
	for _, j := range checksumList {
		hash.Write(j)
	}
	return hash.Sum(nil)
}

// CheckIntegrityHash is used by challenge
func CheckIntegrityHash(integrityHash []byte, checksumList [][]byte, index int, pieceData []byte) error {
	if len(checksumList) <= index {
		return fmt.Errorf("checksum list is not correct")
	}
	if bytes.Compare(checksumList[index], GenerateChecksum(pieceData)) != 0 {
		return fmt.Errorf("piece data checksum is not correct")
	}
	if bytes.Compare(integrityHash, GenerateIntegrityHash(checksumList)) != 0 {
		return fmt.Errorf("piece data integrity checksum is not correct")
	}
	return nil
}

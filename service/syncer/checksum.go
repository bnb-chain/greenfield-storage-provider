package syncer

import (
	"crypto/sha256"
)

func generateChecksum(pieceData []byte) []byte {
	hash := sha256.New()
	hash.Write(pieceData)
	bytes := hash.Sum(nil)
	//hashCode := hex.EncodeToString(bytes)
	return bytes
}

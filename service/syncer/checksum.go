package syncer

import (
	"crypto/sha256"
	"encoding/base64"
)

func generateChecksum(pieceData []byte) string {
	sha := sha256.New()
	sha.Write(pieceData)
	checksum := base64.URLEncoding.EncodeToString(sha.Sum(nil))
	return checksum
}

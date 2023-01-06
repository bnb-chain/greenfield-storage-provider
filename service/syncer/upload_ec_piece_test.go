package syncer

import (
	"fmt"
	"testing"

	"github.com/bnb-chain/inscription-storage-provider/util/hash"
	"github.com/stretchr/testify/assert"
)

func TestGenerateChecksum(t *testing.T) {
	pieceData := "secondary service"
	p2 := "test"
	checksum1 := hash.GenerateChecksum([]byte(pieceData))
	checksum2 := hash.GenerateChecksum([]byte(p2))
	fmt.Println(string(checksum1) + string(checksum2))
	checksum3 := hash.GenerateChecksum([]byte(pieceData + p2))
	fmt.Println(checksum3)
	assert.NotEmpty(t, checksum1)
}

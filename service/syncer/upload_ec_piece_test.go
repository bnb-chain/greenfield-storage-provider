package syncer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateChecksum(t *testing.T) {
	pieceData := "secondary service"
	p2 := "test"
	checksum1 := hash.generateChecksum([]byte(pieceData))
	checksum2 := hash.generateChecksum([]byte(p2))
	fmt.Println(checksum1 + checksum2)
	checksum3 := hash.generateChecksum([]byte(pieceData + p2))
	fmt.Println(checksum3)
	assert.NotEmpty(t, checksum1)
}

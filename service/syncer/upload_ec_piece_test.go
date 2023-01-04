package syncer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateChecksum(t *testing.T) {
	pieceData := "secondary service"
	p2 := "test"
	checksum1 := generateChecksum([]byte(pieceData))
	checksum2 := generateChecksum([]byte(p2))
	fmt.Println(checksum1 + checksum2)
	checksum3 := generateChecksum([]byte(pieceData + p2))
	fmt.Println(checksum3)
	assert.NotEmpty(t, checksum1)
}

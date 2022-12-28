package syncer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateChecksum(t *testing.T) {
	pieceData := "secondary service"
	checksum := generateChecksum([]byte(pieceData))
	fmt.Println(checksum)
	assert.NotEmpty(t, checksum)
}

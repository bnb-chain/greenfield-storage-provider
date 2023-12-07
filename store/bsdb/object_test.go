package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetObjectsTableName(t *testing.T) {
	objectTableName := GetObjectsTableName("ot005test-bucket")
	assert.Equal(t, "objects_62", objectTableName)
}

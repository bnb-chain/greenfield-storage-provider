package bsdb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_GetObjectsTableName(t *testing.T) {
	objectTableName := GetObjectsTableName("ot005test-bucket")
	assert.Equal(t, "objects_62", objectTableName)
}

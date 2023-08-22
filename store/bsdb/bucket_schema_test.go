package bsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBucket_TableName(t *testing.T) {
	bucket := Bucket{ID: 0, BucketName: "mock"}
	name := bucket.TableName()
	assert.Equal(t, BucketTableName, name)
}

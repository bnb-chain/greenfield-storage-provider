package util

import (
	"testing"

	"github.com/stretchr/testify/assert"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func Test_GetTagString(t *testing.T) {
	var tags storagetypes.ResourceTags
	tags.Tags = append(tags.Tags, storagetypes.ResourceTags_Tag{Key: "key1", Value: "value1"})
	tagStr := GetTagJson(&tags)
	assert.Equal(t, "{\"tags\":[{\"key\":\"key1\",\"value\":\"value1\"}]}", tagStr)

	tagStr = GetTagJson(nil)
	assert.Equal(t, "{}", tagStr)
	var emptyTags storagetypes.ResourceTags
	tagStr = GetTagJson(&emptyTags)
	assert.Equal(t, "{}", tagStr)
}

package util

import (
	"encoding/json"

	"gorm.io/datatypes"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func GetTagJson(resourceTags *storagetypes.ResourceTags) datatypes.JSON {
	if resourceTags == nil || resourceTags.Tags == nil {
		return datatypes.JSON{}
	}
	tags, err := json.Marshal(&resourceTags)
	if err != nil {
		panic(err)
	}
	return tags
}

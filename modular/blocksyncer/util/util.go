package util

import (
	"encoding/json"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"gorm.io/datatypes"
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

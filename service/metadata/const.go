package metadata

import (
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
)

// ActionTypeMap db action value is a bitmap, traverse this map to get the corresponding action list
var ActionTypeMap = map[permtypes.ActionType]int{
	permtypes.ACTION_TYPE_ALL:            0,
	permtypes.ACTION_UPDATE_BUCKET_INFO:  1,
	permtypes.ACTION_DELETE_BUCKET:       2,
	permtypes.ACTION_CREATE_OBJECT:       3,
	permtypes.ACTION_DELETE_OBJECT:       4,
	permtypes.ACTION_COPY_OBJECT:         5,
	permtypes.ACTION_GET_OBJECT:          6,
	permtypes.ACTION_EXECUTE_OBJECT:      7,
	permtypes.ACTION_LIST_OBJECT:         8,
	permtypes.ACTION_UPDATE_GROUP_MEMBER: 9,
	permtypes.ACTION_DELETE_GROUP:        10,
	permtypes.ACTION_UPDATE_OBJECT_INFO:  11,
}

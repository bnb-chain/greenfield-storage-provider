package bsdb

import permtypes "github.com/bnb-chain/greenfield/x/permission/types"

// ListObjectsResult represents the result of a List Objects operation.
type ListObjectsResult struct {
	PathName   string
	ResultType string
	*Object
}

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

const (
	// BsDBUser defines env variable name for block syncer db username.
	BsDBUser = "BS_DB_USER"
	// BsDBPasswd defines env variable name for block syncer db user passwd.
	BsDBPasswd = "BS_DB_PASSWORD"
	// BsDBAddress defines env variable name for block syncer db address.
	BsDBAddress = "BS_DB_ADDRESS"
	// BsDBDataBase defines env variable name for block syncer db database.
	BsDBDataBase = "BS_DB_DATABASE"
	// BsDBSwitchedUser defines env variable name for switched block syncer db username.
	BsDBSwitchedUser = "BS_DB_SWITCHED_USER"
	// BsDBSwitchedPasswd defines env variable name for switched block syncer db user passwd.
	BsDBSwitchedPasswd = "BS_DB_SWITCHED_PASSWORD"
	// BsDBSwitchedAddress defines env variable name for switched block syncer db address.
	BsDBSwitchedAddress = "BS_DB_SWITCHED_ADDRESS"
	// BsDBSwitchedDataBase defines env variable name for switched block syncer db database.
	BsDBSwitchedDataBase = "BS_DB_SWITCHED_DATABASE"
)

package bsdb

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"

	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
)

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

// ObjectIDs represents the request of list object by ids
type ObjectIDs struct {
	IDs []uint64 `json:"ids"`
}

// BucketIDs represents the request of list bucket by ids
type BucketIDs struct {
	IDs []uint64 `json:"ids"`
}

type Uint32Array []uint32

func (a *Uint32Array) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan Uint32Array value: %v", value)
	}

	s := string(bytes)
	fields := strings.Split(s, ",")
	result := make([]uint32, len(fields))
	for i, field := range fields {
		v, err := strconv.ParseUint(field, 10, 32)
		if err != nil {
			return fmt.Errorf("failed to scan Uint32Array value: %v", err)
		}
		result[i] = uint32(v)
	}
	*a = result
	return nil
}

func (a Uint32Array) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}
	values := make([]string, len(a))
	for i, value := range a {
		values[i] = strconv.FormatUint(uint64(value), 10)
	}
	return strings.Join(values, ","), nil
}

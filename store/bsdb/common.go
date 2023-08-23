package bsdb

import (
	"database/sql/driver"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
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
	permtypes.ACTION_UPDATE_GROUP_EXTRA:  12,
}

const (
	// BsDBUser defines env variable name for block syncer db username.
	BsDBUser = "BS_DB_USER"
	// BsDBPasswd defines env variable name for block syncer db user passwd.
	BsDBPasswd = "BS_DB_PASSWORD"
	// BsDBAddress defines env variable name for block syncer db address.
	BsDBAddress = "BS_DB_ADDRESS"
	// BsDBDatabase defines env variable name for block syncer db database.
	BsDBDatabase = "BS_DB_DATABASE"
	// BsDBSwitchedUser defines env variable name for switched block syncer db username.
	BsDBSwitchedUser = "BS_DB_SWITCHED_USER"
	// BsDBSwitchedPasswd defines env variable name for switched block syncer db user passwd.
	BsDBSwitchedPasswd = "BS_DB_SWITCHED_PASSWORD"
	// BsDBSwitchedAddress defines env variable name for switched block syncer db address.
	BsDBSwitchedAddress = "BS_DB_SWITCHED_ADDRESS"
	// BsDBSwitchedDatabase defines env variable name for switched block syncer db database.
	BsDBSwitchedDatabase = "BS_DB_SWITCHED_DATABASE"
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

func MetadataDatabaseFailureMetrics(err error, startTime time.Time, methodName string) {
	metrics.MetadataReqTime.WithLabelValues(DatabaseFailure, DatabaseLevel, methodName, err.Error()).Observe(time.Since(startTime).Seconds())
}

func MetadataDatabaseSuccessMetrics(startTime time.Time, methodName string) {
	metrics.MetadataReqTime.WithLabelValues(DatabaseSuccess, DatabaseLevel, methodName, strconv.Itoa(0)).Observe(time.Since(startTime).Seconds())
}

func currentFunction() string {
	counter, _, _, success := runtime.Caller(1)

	if !success {
		println("functionName: runtime.Caller: failed")
	}

	fullName := runtime.FuncForPC(counter).Name()
	splitNames := strings.Split(fullName, ".")
	return splitNames[len(splitNames)-1]
}

func PossibleValuesForAction(targetAction permtypes.ActionType) []int {
	maxVal := 0
	for _, val := range ActionTypeMap {
		maxVal |= 1 << val
	}

	targetBit := 1 << targetAction

	var results []int
	for i := 0; i <= maxVal; i++ {
		if i&targetBit == targetBit {
			results = append(results, i)
		}
	}
	return results
}

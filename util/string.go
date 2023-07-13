package util

import (
	"encoding/hex"
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

var (
	// ErrIntegerOverflow defines integer overflow
	ErrIntegerOverflow = errors.New("integer overflow")
)

// StringToUint64 converts string to uint64
func StringToUint64(str string) (uint64, error) {
	ui64, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return ui64, nil
}

// StringToInt64 converts string to int64
func StringToInt64(str string) (int64, error) {
	i64, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return i64, nil
}

// StringToUint32 coverts string to uint32
func StringToUint32(str string) (uint32, error) {
	ui64, err := StringToUint64(str)
	if err != nil {
		return 0, err
	}
	if ui64 > math.MaxUint32 {
		return 0, ErrIntegerOverflow
	}
	return uint32(ui64), nil
}

// StringToInt32 converts string to int32
func StringToInt32(str string) (int32, error) {
	i64, err := StringToInt64(str)
	if err != nil {
		return 0, err
	}
	if i64 > math.MaxInt32 {
		return 0, ErrIntegerOverflow
	}
	return int32(i64), nil
}

// StringToBool coverts string to bool
func StringToBool(str string) (bool, error) {
	b, err := strconv.ParseBool(str)
	if err != nil {
		return false, err
	}
	return b, nil
}

// BoolToInt converts bool to int
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// JoinWithComma converts string slice to one string with comma
func JoinWithComma(slice []string) string {
	return strings.Join(slice, ",")
}

// SplitByComma splits string by comma
func SplitByComma(str string) []string {
	str = strings.TrimSpace(str)
	strArr := strings.Split(str, ",")
	var trimStr []string
	for _, item := range strArr {
		if len(strings.TrimSpace(item)) > 0 {
			trimStr = append(trimStr, strings.TrimSpace(item))
		}
	}
	return trimStr
}

// Uint64ToString coverts uint64 to string
func Uint64ToString(u uint64) string {
	return strconv.FormatUint(u, 10)
}

// Uint32ToString converts uint32 to string
func Uint32ToString(u uint32) string {
	return Uint64ToString(uint64(u))
}

// BytesSliceToString is used to serialize
func BytesSliceToString(bytes [][]byte) string {
	stringList := make([]string, len(bytes))
	for index, h := range bytes {
		stringList[index] = hex.EncodeToString(h)
	}
	return JoinWithComma(stringList)
}

// StringToBytesSlice is used to deserialize
func StringToBytesSlice(str string) ([][]byte, error) {
	var err error
	stringList := SplitByComma(str)
	hashList := make([][]byte, len(stringList))
	for idx := range stringList {
		if hashList[idx], err = hex.DecodeString(stringList[idx]); err != nil {
			return hashList, err
		}
	}
	return hashList, nil
}

func StringArrayToUint32Slice(arr pq.StringArray) ([]uint32, error) {
	uint32Slice := make([]uint32, len(arr))
	for i, str := range arr {
		val, err := StringToUint32(str)
		if err != nil {
			return nil, err
		}
		uint32Slice[i] = val
	}
	return uint32Slice, nil
}

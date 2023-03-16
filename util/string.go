package util

import (
	"encoding/hex"
	"strconv"
	"strings"
)

// StringToUint64 convert string to uint64
func StringToUint64(str string) (uint64, error) {
	ui64, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return ui64, nil
}

// StringToInt64 convert string to int64
func StringToInt64(str string) (int64, error) {
	i64, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return i64, nil
}

// StringToUint32 covert string to uint32
func StringToUint32(str string) (uint32, error) {
	ui64, err := StringToUint64(str)
	if err != nil {
		return 0, err
	}
	// TODO: check overflow
	return uint32(ui64), nil
}

// StringToInt32 convert string to int32
func StringToInt32(str string) (int32, error) {
	i64, err := StringToInt64(str)
	if err != nil {
		return 0, err
	}
	// TODO: check overflow
	return int32(i64), nil
}

// StringToBool covert string to bool
func StringToBool(str string) (bool, error) {
	b, err := strconv.ParseBool(str)
	if err != nil {
		return false, err
	}
	return b, nil
}

// BoolToInt convert bool to int
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// JoinWithComma con vert string slice to one string with comma
func JoinWithComma(slice []string) string {
	if len(slice) == 1 {
		return slice[0]
	}
	return strings.Join(slice, ",")
}

// SplitByComma split string by comma
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

// Uint64ToString covert uint64 to string
func Uint64ToString(u uint64) string {
	return strconv.FormatUint(u, 10)
}

// Uint32ToString convert uint32 to string
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

// StringListToBytesSlice is used to deserialize
func StringListToBytesSlice(stringList []string) [][]byte {
	hashList := make([][]byte, len(stringList))
	for idx := range stringList {
		hashList[idx] = []byte(stringList[idx])
	}
	return hashList
}

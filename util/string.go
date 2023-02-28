package util

import (
	"encoding/hex"
	"strconv"
	"strings"
)

func StringToUint64(str string) (uint64, error) {
	ui64, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return ui64, nil
}

func StringToInt64(str string) (int64, error) {
	i64, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return i64, nil
}

func StringToUint32(str string) (uint32, error) {
	ui64, err := StringToUint64(str)
	if err != nil {
		return 0, err
	}
	// TODO: check overflow
	return uint32(ui64), nil
}

func StringToInt32(str string) (int32, error) {
	i64, err := StringToInt64(str)
	if err != nil {
		return 0, err
	}
	// TODO: check overflow
	return int32(i64), nil
}

func StringToBool(str string) (bool, error) {
	b, err := strconv.ParseBool(str)
	if err != nil {
		return false, err
	}
	return b, nil
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func JoinWithComma(slice []string) string {
	if len(slice) == 1 {
		return slice[0]
	}
	return strings.Join(slice, ",")
}

func SplitByComma(str string) []string {
	return strings.Split(str, ",")
}

func Uint64ToString(u uint64) string {
	return strconv.FormatUint(u, 10)
}

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

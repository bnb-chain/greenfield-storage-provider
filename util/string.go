package util

import (
	"encoding/hex"
	"strconv"
	"strings"
)

func StringToUin64(str string) (uint64, error) {
	ui64, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return ui64, nil
}

func StringToIn64(str string) (int64, error) {
	ui64, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return ui64, nil
}

func StringToUint32(str string) (uint32, error) {
	ui64, err := StringToUin64(str)
	if err != nil {
		return 0, err
	}
	// TODO: check overflow
	return uint32(ui64), nil
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

func SliceToString(slice []string) string {
	if len(slice) == 1 {
		return slice[0]
	}
	return strings.Join(slice, ",")
}

func StringToStringSlice(str string) []string {
	return strings.Split(str, ",")
}

func Uint64ToString(u uint64) string {
	return strconv.FormatUint(u, 10)
}

// BytesSliceToString is used to serialize
func BytesSliceToString(bytes [][]byte) string {
	PieceStringList := make([]string, len(bytes))
	for index, h := range bytes {
		PieceStringList[index] = hex.EncodeToString(h)
	}
	return SliceToString(PieceStringList)
}

// StringToBytesSlice is used to deserialize
func StringToBytesSlice(str string) ([][]byte, error) {
	var err error
	pieceStringList := StringToStringSlice(str)
	hashList := make([][]byte, len(pieceStringList))
	for idx := range pieceStringList {
		if hashList[idx], err = hex.DecodeString(pieceStringList[idx]); err != nil {
			return hashList, err
		}
	}
	return hashList, nil
}

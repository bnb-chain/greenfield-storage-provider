package util

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/naoina/toml"

	"github.com/bnb-chain/inscription-storage-provider/model"
)

// TomlSettings - These settings ensure that TOML keys use the same names as Go struct fields.
var TomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://pkg.go.dev/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		_, _ = fmt.Fprintf(os.Stderr, "field '%s' is not defined in %s%s\n", field, rt.String(), link)
		return nil
	},
}

// GenerateRequestID is used to generate random requestID.
func GenerateRequestID() string {
	var uUID uuid.UUID
	var err error
	if uUID, err = uuid.NewRandom(); err != nil {
		return ""
	}
	return strings.ReplaceAll(uUID.String(), "-", "")
}

// ComputeSegmentCount return the segments counter by payload size.
func ComputeSegmentCount(size uint64) uint32 {
	segmentCount := uint32(size / model.SegmentSize)
	if (size % model.SegmentSize) > 0 {
		segmentCount++
	}
	return segmentCount
}

type MapKeySorted interface {
	map[string][]byte | map[string][][]byte
}

// SortedKeys sort keys of a map
//func GenericSortedKeys[M MapKeySorted](dataMap M) []string {
//	keys := make([]string, 0, len(dataMap))
//	for k := range dataMap {
//		keys = append(keys, k)
//	}
//	sort.Strings(keys)
//	return keys
//}

func SortedKeys(dataMap map[string][]byte) []string {
	keys := make([]string, 0, len(dataMap))
	for k := range dataMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

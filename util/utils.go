package util

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/naoina/toml"
	"golang.org/x/exp/constraints"

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
	return strconv.FormatUint(rand.Uint64(), 10)
}

// ComputeSegmentCount return the segments counter by payload size.
func ComputeSegmentCount(size uint64) uint32 {
	segmentCount := uint32(size / model.SegmentSize)
	if (size % model.SegmentSize) > 0 {
		segmentCount++
	}
	return segmentCount
}

// ReadJobState parser the job state to readable
func ReadJobState(state string) string {
	return strings.ToLower(strings.TrimPrefix(state, "JOB_STATE_"))
}

// SortedKeys sort keys of a map
func GenericSortedKeys[K constraints.Ordered, V any](dataMap map[K]V) []K {
	keys := make([]K, 0, len(dataMap))
	for k := range dataMap {
		keys = append(keys, k)
	}
	sortSlice(keys)
	return keys
}

func sortSlice[T constraints.Ordered](s []T) {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}

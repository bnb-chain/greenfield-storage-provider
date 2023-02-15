package util

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/naoina/toml"
	"google.golang.org/grpc/peer"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
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

// JobStateReadable parser the job state to readable
func JobStateReadable(state string) string {
	return strings.ToLower(strings.TrimPrefix(state, "JOB_STATE_"))
}

// SpReadable parser the storage provider to readable
func SpReadable(provider string) string {
	return provider[:8]
}

// ValidateRedundancyType validate redundancy type
func ValidateRedundancyType(redundancyType ptypes.RedundancyType) error {
	switch redundancyType {
	case ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED:
		return nil
	case ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE, ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
		return nil
	default:
		return merrors.ErrRedundancyType
	}
}

// GetIPFromGRPCContext returns a IP from grpc client
func GetIPFromGRPCContext(ctx context.Context) net.IP {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return nil
	}

	addr := strings.Split(pr.Addr.String(), ":")
	if len(addr) < 1 {
		return nil
	}

	return net.ParseIP(addr[0])
}

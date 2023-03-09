package util

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/forbole/juno/v4/types"
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

// HeaderToRedundancyType can be EC or Replica or Inline, default is EC
func HeaderToRedundancyType(header string) ptypes.RedundancyType {
	if header == model.ReplicaRedundancyTypeHeaderValue {
		return ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE
	}
	if header == model.InlineRedundancyTypeHeaderValue {
		return ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE
	}
	return ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED
}

func HeaderToUint64(header string) (uint64, error) {
	ui64, err := strconv.ParseUint(header, 10, 64)
	if err != nil {
		return 0, err
	}
	return ui64, nil
}

func HeaderToInt64(header string) (int64, error) {
	ui64, err := strconv.ParseInt(header, 10, 64)
	if err != nil {
		return 0, err
	}
	return ui64, nil
}

func HeaderToUint32(header string) (uint32, error) {
	ui64, err := HeaderToUint64(header)
	if err != nil {
		return 0, err
	}
	// TODO: check overflow
	return uint32(ui64), nil
}

func HeaderToBool(header string) (bool, error) {
	b, err := strconv.ParseBool(header)
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

func StringSliceToHeader(stringSlice []string) string {
	if len(stringSlice) == 1 {
		return stringSlice[0]
	}
	return strings.Join(stringSlice, ",")
}

func HeaderToStringSlice(header string) []string {
	return strings.Split(header, ",")
}

func Uint64ToHeader(u uint64) string {
	return strconv.FormatUint(u, 10)
}

// EncodePieceHash is used to serialize
func EncodePieceHash(pieceHash [][]byte) string {
	PieceStringList := make([]string, len(pieceHash))
	for index, h := range pieceHash {
		PieceStringList[index] = hex.EncodeToString(h)
	}
	return StringSliceToHeader(PieceStringList)
}

// DecodePieceHash is used to deserialize
func DecodePieceHash(pieceHash string) ([][]byte, error) {
	var err error
	pieceStringList := HeaderToStringSlice(pieceHash)
	hashList := make([][]byte, len(pieceStringList))
	for idx := range pieceStringList {
		if hashList[idx], err = hex.DecodeString(pieceStringList[idx]); err != nil {
			return hashList, err
		}
	}
	return hashList, nil
}

func TransferRedundancyType(redundancyType string) (ptypes.RedundancyType, error) {
	switch redundancyType {
	case ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED.String():
		return ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED, nil
	case ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE.String():
		return ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE, nil
	case ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE.String():
		return ptypes.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE, nil
	default:
		return -1, merrors.ErrRedundancyType
	}
}

// SumGasTxs returns the total gas consumed by a set of transactions.
func SumGasTxs(txs []*types.Tx) uint64 {
	var totalGas uint64

	for _, tx := range txs {
		totalGas += uint64(tx.GasUsed)
	}

	return totalGas
}

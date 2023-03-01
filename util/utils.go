package util

import (
	"context"
	"encoding/hex"
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

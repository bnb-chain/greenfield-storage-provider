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

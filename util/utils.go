package util

import (
	"context"
	"math/rand"
	"net"
	"strconv"
	"strings"

	"google.golang.org/grpc/peer"
)

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
	return net.ParseIP(addr[0])
}

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

func GetRPCRemoteAddress(ctx context.Context) string {
	var addr string
	if pr, ok := peer.FromContext(ctx); ok {
		if tcpAddr, ok := pr.Addr.(*net.TCPAddr); ok {
			addr = tcpAddr.IP.String()
		} else {
			addr = pr.Addr.String()
		}
	}
	return addr
}

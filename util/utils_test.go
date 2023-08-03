package util

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/peer"
)

func TestGenerateRequestID(t *testing.T) {
	str := GenerateRequestID()
	assert.NotEmpty(t, str)
}

type addr struct {
	ipAddress string
}

func (addr) Network() string   { return "" }
func (a *addr) String() string { return a.ipAddress }

func TestGetIPFromGRPCContext(t *testing.T) {
	cases := []struct {
		name string
		ctx  context.Context
		ip   net.IP
	}{
		{
			"Context without peer",
			context.Background(),
			nil,
		},
		{
			"Context with correct IP address",
			peer.NewContext(context.Background(), &peer.Peer{Addr: &addr{ipAddress: "127.0.0.1:9000"}}),
			net.IP{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xff, 0x7f, 0x0, 0x0, 0x1},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ip := GetIPFromGRPCContext(tt.ctx)
			assert.Equal(t, tt.ip, ip)
		})
	}
}

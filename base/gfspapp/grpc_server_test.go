package gfspapp

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/util"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

func TestGfSpBaseApp_StartRPCServerSuccess(t *testing.T) {
	g := &GfSpBaseApp{grpcAddress: "localhost:0"}
	g.server = grpc.NewServer()
	go func() {
		// make sure Serve() is called
		time.Sleep(time.Millisecond * 500)
		err := g.StopRPCServer(context.TODO())
		assert.Nil(t, err)
	}()
	err := g.StartRPCServer(context.TODO())
	assert.Nil(t, err)
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
		addr string
	}{
		{
			"Context without addr",
			context.Background(),
			"",
		},
		{
			"Context with correct IP address",
			peer.NewContext(context.Background(), &peer.Peer{Addr: &addr{ipAddress: "127.0.0.1:9000"}}),
			"127.0.0.1:9000",
		},
		{
			"Context with correct IP address",
			peer.NewContext(context.Background(), &peer.Peer{Addr: &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0}}),
			"127.0.0.1",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			addr := util.GetRPCRemoteAddress(tt.ctx)
			assert.Equal(t, tt.addr, addr)
		})
	}
}

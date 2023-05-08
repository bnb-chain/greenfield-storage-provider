package gfspclient

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	MaxCallMsgSize  = 32 * 1024 * 1024
	ClientCodeSpace = "GfSpClient"
	// HttpMaxIdleConns defines the max idle connections for HTTP server
	HttpMaxIdleConns = 20
	// HttpIdleConnTimout defines the idle time of connection for closing
	HttpIdleConnTimout = 60 * time.Second
)

var (
	ErrRpcUnknown       = gfsperrors.Register(ClientCodeSpace, http.StatusBadRequest, 98001, "server slipped away, try again later")
	ErrExceptionsStream = gfsperrors.Register(ClientCodeSpace, http.StatusBadRequest, 98002, "stream closed abnormally")
	ErrTypeMismatch     = gfsperrors.Register(ClientCodeSpace, http.StatusInternalServerError, 98101, "response type mismatch")
)

type GfSpClient struct {
	approverEndpoint   string
	managerEndpoint    string
	downloaderEndpoint string
	receiverEndpoint   string
	metadataEndpoint   string
	retrieverEndpoint  string
	uploaderEndpoint   string
	p2pEndpoint        string
	singerEndpoint     string
	authorizerEndpoint string

	mux          sync.RWMutex
	managerConn  *grpc.ClientConn
	approverConn *grpc.ClientConn
	p2pConn      *grpc.ClientConn
	singerConn   *grpc.ClientConn
	httpClient   *http.Client
}

func (s *GfSpClient) Connection(
	ctx context.Context,
	address string,
	opts ...grpc.DialOption) (
	*grpc.ClientConn, error) {
	options := append(DefaultClientOptions(), opts...)
	return grpc.DialContext(ctx, address, options...)
}

func (s *GfSpClient) ManagerConn(
	ctx context.Context,
	opts ...grpc.DialOption) (
	*grpc.ClientConn, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	options := append(DefaultClientOptions(), opts...)
	if s.managerConn == nil {
		conn, err := s.Connection(ctx, s.managerEndpoint, options...)
		if err != nil {
			log.CtxErrorw(ctx, "failed to create connection", "error", err)
			return nil, ErrRpcUnknown
		}
		s.managerConn = conn
	}
	return s.managerConn, nil
}

func (s *GfSpClient) ApproverConn(
	ctx context.Context,
	opts ...grpc.DialOption) (
	*grpc.ClientConn, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	options := append(DefaultClientOptions(), opts...)
	if s.approverConn == nil {
		conn, err := s.Connection(ctx, s.approverEndpoint, options...)
		if err != nil {
			log.CtxErrorw(ctx, "failed to create connection", "error", err)
			return nil, ErrRpcUnknown
		}
		s.approverConn = conn
	}
	return s.approverConn, nil
}

func (s *GfSpClient) P2PConn(
	ctx context.Context,
	opts ...grpc.DialOption) (
	*grpc.ClientConn, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	options := append(DefaultClientOptions(), opts...)
	if s.p2pConn == nil {
		conn, err := s.Connection(ctx, s.p2pEndpoint, options...)
		if err != nil {
			log.CtxErrorw(ctx, "failed to create connection", "error", err)
			return nil, ErrRpcUnknown
		}
		s.p2pConn = conn
	}
	return s.p2pConn, nil
}

func (s *GfSpClient) SingerConn(
	ctx context.Context,
	opts ...grpc.DialOption) (
	*grpc.ClientConn, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	options := append(DefaultClientOptions(), opts...)
	if s.singerConn == nil {
		conn, err := s.Connection(ctx, s.singerEndpoint, options...)
		if err != nil {
			log.CtxErrorw(ctx, "failed to create connection", "error", err)
			return nil, ErrRpcUnknown
		}
		s.singerConn = conn
	}
	return s.singerConn, nil
}

func (s *GfSpClient) HttpClient(ctx context.Context) *http.Client {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.httpClient == nil {
		s.httpClient = &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:    HttpMaxIdleConns,
				IdleConnTimeout: HttpIdleConnTimout,
			}}
	}
	return s.httpClient
}

func DefaultClientOptions() []grpc.DialOption {
	var options []grpc.DialOption
	options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	options = append(options, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxCallMsgSize)))
	options = append(options, grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(MaxCallMsgSize)))
	return options
}

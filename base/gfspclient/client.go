package gfspclient

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	utilgrpc "github.com/bnb-chain/greenfield-storage-provider/util/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	// MaxClientCallMsgSize defines the max message size for grpc client
	MaxClientCallMsgSize = 3 * 1024 * 1024 * 1024
	// ClientCodeSpace defines the code space for gfsp client
	ClientCodeSpace = "GfSpClient"
	// HttpMaxIdleConns defines the max idle connections for HTTP server
	HttpMaxIdleConns = 20
	// HttpIdleConnTimout defines the idle time of connection for closing
	HttpIdleConnTimout = 60 * time.Second

	// DefaultStreamBufSize defines gateway stream forward payload buf size
	DefaultStreamBufSize = 16 * 1024 * 1024
)

var (
	ErrRpcUnknown       = gfsperrors.Register(ClientCodeSpace, http.StatusNotFound, 98001, "server slipped away, try again later")
	ErrExceptionsStream = gfsperrors.Register(ClientCodeSpace, http.StatusBadRequest, 98002, "stream closed abnormally")
	ErrTypeMismatch     = gfsperrors.Register(ClientCodeSpace, http.StatusBadRequest, 98101, "response type mismatch")
	ErrNoSuchObject     = gfsperrors.Register(ClientCodeSpace, http.StatusBadRequest, 98093, "no such object from metadata")
)

type GfSpClient struct {
	approverEndpoint      string
	managerEndpoint       string
	downloaderEndpoint    string
	receiverEndpoint      string
	metadataEndpoint      string
	uploaderEndpoint      string
	p2pEndpoint           string
	signerEndpoint        string
	authenticatorEndpoint string

	mux            sync.RWMutex
	downloaderConn *grpc.ClientConn
	managerConn    *grpc.ClientConn
	approverConn   *grpc.ClientConn
	p2pConn        *grpc.ClientConn
	signerConn     *grpc.ClientConn
	httpClient     *http.Client
	metrics        bool
}

func NewGfSpClient(
	approverEndpoint string,
	managerEndpoint string,
	downloaderEndpoint string,
	receiverEndpoint string,
	metadataEndpoint string,
	uploaderEndpoint string,
	p2pEndpoint string,
	signerEndpoint string,
	authenticatorEndpoint string,
	metrics bool) *GfSpClient {
	return &GfSpClient{
		approverEndpoint:      approverEndpoint,
		managerEndpoint:       managerEndpoint,
		downloaderEndpoint:    downloaderEndpoint,
		receiverEndpoint:      receiverEndpoint,
		metadataEndpoint:      metadataEndpoint,
		uploaderEndpoint:      uploaderEndpoint,
		p2pEndpoint:           p2pEndpoint,
		signerEndpoint:        signerEndpoint,
		authenticatorEndpoint: authenticatorEndpoint,
		metrics:               metrics,
	}
}

func (s *GfSpClient) Connection(ctx context.Context, address string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	options := append(DefaultClientOptions(), opts...)
	return grpc.DialContext(ctx, address, options...)
}

func (s *GfSpClient) DownloaderConn(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	options := append(DefaultClientOptions(), opts...)
	if s.metrics {
		options = append(options, utilgrpc.GetDefaultClientInterceptor()...)
	}
	if s.downloaderConn == nil {
		conn, err := s.Connection(ctx, s.downloaderEndpoint, options...)
		if err != nil {
			log.CtxErrorw(ctx, "failed to create connection", "error", err)
			return nil, ErrRpcUnknown
		}
		s.downloaderConn = conn
	}
	return s.downloaderConn, nil
}

func (s *GfSpClient) ManagerConn(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	options := append(DefaultClientOptions(), opts...)
	if s.metrics {
		options = append(options, utilgrpc.GetDefaultClientInterceptor()...)
	}
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

func (s *GfSpClient) ApproverConn(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	options := append(DefaultClientOptions(), opts...)
	if s.metrics {
		options = append(options, utilgrpc.GetDefaultClientInterceptor()...)
	}
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

func (s *GfSpClient) P2PConn(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	options := append(DefaultClientOptions(), opts...)
	if s.metrics {
		options = append(options, utilgrpc.GetDefaultClientInterceptor()...)
	}
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

func (s *GfSpClient) SignerConn(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	options := append(DefaultClientOptions(), opts...)
	if s.metrics {
		options = append(options, utilgrpc.GetDefaultClientInterceptor()...)
	}
	if s.signerConn == nil {
		conn, err := s.Connection(ctx, s.signerEndpoint, options...)
		if err != nil {
			log.CtxErrorw(ctx, "failed to create connection", "error", err)
			return nil, ErrRpcUnknown
		}
		s.signerConn = conn
	}
	return s.signerConn, nil
}

func (s *GfSpClient) HTTPClient(ctx context.Context) *http.Client {
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

func (s *GfSpClient) Close() error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.managerConn != nil {
		s.managerConn.Close()
	}
	if s.downloaderConn != nil {
		s.downloaderConn.Close()
	}
	if s.approverConn != nil {
		s.approverConn.Close()
	}
	if s.p2pConn != nil {
		s.p2pConn.Close()
	}
	if s.signerConn != nil {
		s.signerConn.Close()
	}
	return nil
}

func DefaultClientOptions() []grpc.DialOption {
	var options []grpc.DialOption
	options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	options = append(options, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxClientCallMsgSize)))
	options = append(options, grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(MaxClientCallMsgSize)))

	//var kacp = keepalive.ClientParameters{
	//	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	//	Timeout:             10 * time.Second, // wait 1 second for ping ack before considering the connection dead
	//	PermitWithoutStream: true,             // send pings even without active streams
	//}
	//options = append(options, grpc.WithKeepaliveParams(kacp))
	return options
}

package gfspclient

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	utilgrpc "github.com/bnb-chain/greenfield-storage-provider/util/grpc"
)

const (
	// MaxClientCallMsgSize defines the max message size for grpc client
	MaxClientCallMsgSize = 3 * 1024 * 1024 * 1024
	// ClientCodeSpace defines the code space for gfsp client
	ClientCodeSpace = "GfSpClient"
	// HTTPMaxIdleConns defines the max idle connections for HTTP server
	HTTPMaxIdleConns = 20
	// HTTPIdleConnTimout defines the idle time of connection for closing
	HTTPIdleConnTimout = 60 * time.Second

	// DefaultStreamBufSize defines gateway stream forward payload buf size
	DefaultStreamBufSize = 16 * 1024 * 1024
)

var (
	ErrExceptionsStream = gfsperrors.Register(ClientCodeSpace, http.StatusBadRequest, 98002, "stream closed abnormally")
	ErrTypeMismatch     = gfsperrors.Register(ClientCodeSpace, http.StatusBadRequest, 98101, "response type mismatch")
	ErrNoSuchObject     = gfsperrors.Register(ClientCodeSpace, http.StatusBadRequest, 98093, "no such object from metadata")
)

func ErrRPCUnknownWithDetail(detail string, err error) *gfsperrors.GfSpError {
	if gfspErr := gfsperrors.MakeGfSpError(err); gfspErr != nil {
		return gfspErr
	}
	return gfsperrors.Register(ClientCodeSpace, http.StatusInternalServerError, 98001, detail+err.Error())
}

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

	mux          sync.RWMutex
	managerConn  *grpc.ClientConn
	approverConn *grpc.ClientConn
	p2pConn      *grpc.ClientConn
	signerConn   *grpc.ClientConn
	httpClient   *http.Client
	metrics      bool
}

func NewGfSpClient(approverEndpoint, managerEndpoint, downloaderEndpoint, receiverEndpoint, metadataEndpoint,
	uploaderEndpoint, p2pEndpoint, signerEndpoint, authenticatorEndpoint string, metrics bool) *GfSpClient {
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
	if s.metrics {
		options = append(options, utilgrpc.GetDefaultClientInterceptor()...)
	}
	return grpc.DialContext(ctx, address, options...)
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
			return nil, ErrRPCUnknownWithDetail("failed to create connection, error: ", err)
		}
		s.approverConn = conn
	}
	return s.approverConn, nil
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
			return nil, ErrRPCUnknownWithDetail("failed to create connection, error: ", err)
		}
		s.managerConn = conn
	}
	return s.managerConn, nil
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
			return nil, ErrRPCUnknownWithDetail("failed to create connection, error: ", err)
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
			return nil, ErrRPCUnknownWithDetail("failed to create connection, error: ", err)
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
				MaxIdleConns:    HTTPMaxIdleConns,
				IdleConnTimeout: HTTPIdleConnTimout,
				TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
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
	return options
}

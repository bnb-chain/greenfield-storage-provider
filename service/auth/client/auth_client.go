package client

/*
import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/auth/types"
	authtypes "github.com/bnb-chain/greenfield-storage-provider/service/auth/types"
)

// AuthClient is an auth server gRPC service client wrapper
type AuthClient struct {
	Address string
	Conn    *grpc.ClientConn
	Auth    types.AuthServiceClient
}

// NewAuthClient return a AuthClient instance
func NewAuthClient(address string) (*AuthClient, error) {
	options := []grpc.DialOption{}
	//if metrics.GetMetrics().Enabled() {
	//	options = append(options, utilgrpc.GetDefaultClientInterceptor()...)
	//}
	options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.DialContext(context.Background(), address, options...)
	if err != nil {
		log.Errorw("failed to dial Auth server", "error", err)
		return nil, err
	}
	client := &AuthClient{
		Address: address,
		Conn:    conn,
		Auth:    types.NewAuthServiceClient(conn),
	}
	return client, nil
}

// Close the Auth server gPRC connection
func (client *AuthClient) Close() error {
	return client.Conn.Close()
}

// GetAuthNonce get the auth nonce for which the Dapp or client can generate EDDSA key pairs.
func (client *AuthClient) GetAuthNonce(ctx context.Context, in *authtypes.GetAuthNonceRequest, opts ...grpc.CallOption) (*authtypes.GetAuthNonceResponse, error) {
	resp, err := client.Auth.GetAuthNonce(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get auth nonce rpc", "error", err)
		return nil, err
	}
	return resp, nil
}

// UpdateUserPublicKey updates the user public key once the Dapp or client generates the EDDSA key pairs.
func (client *AuthClient) UpdateUserPublicKey(ctx context.Context, in *authtypes.UpdateUserPublicKeyRequest, opts ...grpc.CallOption) (*authtypes.UpdateUserPublicKeyResponse, error) {
	resp, err := client.Auth.UpdateUserPublicKey(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to update user public key rpc", "error", err)
		return nil, err
	}
	return resp, nil
}

// VerifyOffChainSignature verifies the signature signed by user's EDDSA private key.
func (client *AuthClient) VerifyOffChainSignature(ctx context.Context, in *authtypes.VerifyOffChainSignatureRequest, opts ...grpc.CallOption) (*authtypes.VerifyOffChainSignatureResponse, error) {
	resp, err := client.Auth.VerifyOffChainSignature(ctx, in, opts...)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to verify off-chain signature rpc", "error", err)
		return nil, err
	}
	return resp, nil
}

*/

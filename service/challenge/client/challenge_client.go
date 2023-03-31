package client

import (
	"context"

	"google.golang.org/grpc"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	mwgrpc "github.com/bnb-chain/greenfield-storage-provider/pkg/middleware/grpc"
	"github.com/bnb-chain/greenfield-storage-provider/service/challenge/types"
	utilgrpc "github.com/bnb-chain/greenfield-storage-provider/util/grpc"
)

// ChallengeClient is a challenge gRPC service client wrapper
type ChallengeClient struct {
	address   string
	challenge types.ChallengeServiceClient
	conn      *grpc.ClientConn
}

// NewChallengeClient return a ChallengeClient instance
func NewChallengeClient(address string) (*ChallengeClient, error) {
	options := utilgrpc.GetDefaultClientOptions()
	if metrics.GetMetrics().Enabled() {
		options = append(options, mwgrpc.GetDefaultClientInterceptor()...)
	}
	conn, err := grpc.DialContext(context.Background(), address, options...)
	if err != nil {
		log.Errorw("failed to dial challenge", "error", err)
		return nil, err
	}
	client := &ChallengeClient{
		address:   address,
		conn:      conn,
		challenge: types.NewChallengeServiceClient(conn),
	}
	return client, nil
}

// Close the challenge gPRC connection
func (client *ChallengeClient) Close() error {
	return client.conn.Close()
}

// ChallengePiece send challenge piece request
func (client *ChallengeClient) ChallengePiece(ctx context.Context, objectID uint64, redundancyIdx int32, segmentIdx uint32,
	opts ...grpc.CallOption) ([]byte, [][]byte, []byte, error) {
	resp, err := client.challenge.ChallengePiece(ctx, &types.ChallengePieceRequest{
		ObjectId:      objectID,
		SegmentIdx:    segmentIdx,
		RedundancyIdx: redundancyIdx,
	}, opts...)
	if err != nil {
		return nil, nil, nil, merrors.GRPCErrorToInnerError(err)
	}
	return resp.GetIntegrityHash(), resp.GetPieceHash(), resp.GetPieceData(), err
}

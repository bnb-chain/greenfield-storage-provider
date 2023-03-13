package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/challenge/types"
)

// ChallengeClient is a challenge gRPC service client wrapper
type ChallengeClient struct {
	address   string
	challenge types.ChallengeServiceClient
	conn      *grpc.ClientConn
}

// NewChallengeClient return a ChallengeClient instance
func NewChallengeClient(address string) (*ChallengeClient, error) {
	conn, err := grpc.DialContext(context.Background(), address,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorw("failed to dail challenge", "error", err)
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
func (client *ChallengeClient) ChallengePiece(ctx context.Context, objectID uint64, replicaIdx int32, segmentIdx uint32,
	opts ...grpc.CallOption) ([]byte, [][]byte, []byte, error) {
	resp, err := client.challenge.ChallengePiece(ctx, &types.ChallengePieceRequest{
		ObjectId:   objectID,
		ReplicaIdx: replicaIdx,
		SegmentIdx: segmentIdx,
	}, opts...)
	return resp.IntegrityHash, resp.PieceHash, resp.PieceData, err
}

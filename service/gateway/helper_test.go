package gateway

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"google.golang.org/grpc"

	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
)

func setup(t *testing.T) *Gateway {
	return &Gateway{
		config: &GatewayConfig{
			StorageProvider:          "test1",
			Address:                  "test2",
			Domain:                   "test3",
			UploaderServiceAddress:   "test4",
			DownloaderServiceAddress: "test5",
			ChallengeServiceAddress:  "test6",
			SyncerServiceAddress:     "test7",
			ChainConfig:              greenfield.DefaultGreenfieldChainConfig,
		},
		name: model.GatewayService,
	}
}

func makeStreamMock() *StreamMock {
	return &StreamMock{
		ctx:          context.Background(),
		recvToServer: make(chan *stypes.SyncerServiceSyncPieceRequest, 10),
	}
}

type StreamMock struct {
	grpc.ClientStream
	ctx          context.Context
	recvToServer chan *stypes.SyncerServiceSyncPieceRequest
}

func (m *StreamMock) Send(resp *stypes.SyncerServiceSyncPieceRequest) error {
	m.recvToServer <- resp
	return nil
}

func (m *StreamMock) CloseAndRecv() (*stypes.SyncerServiceSyncPieceResponse, error) {
	integrityHash, _ := base64.URLEncoding.DecodeString("pgPGdR4c9_KYz6wQxl-SifyzHXlHhx5XfNa89LzdNCI=")
	return &stypes.SyncerServiceSyncPieceResponse{
		TraceId: "test_traceID",
		SecondarySpInfo: &stypes.StorageProviderSealInfo{
			StorageProviderId: "sp1",
			PieceIdx:          1,
			PieceChecksum:     [][]byte{[]byte("1"), []byte("2"), []byte("3"), []byte("4"), []byte("5"), []byte("6")},
			IntegrityHash:     integrityHash,
			Signature:         nil,
		},
		ErrMessage: &stypes.ErrMessage{
			ErrCode: stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED,
			ErrMsg:  "Success",
		},
	}, nil
}

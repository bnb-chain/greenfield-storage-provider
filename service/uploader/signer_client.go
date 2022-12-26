package uploader

import (
	"context"

	pbService "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
)

type signerClient struct {
}

func newSignerClient() *signerClient {
	return &signerClient{}
}

func (sc *signerClient) coSignTx(ctx context.Context, req *pbService.UploaderServiceCreateObjectRequest) ([]byte, error) {
	return []byte("txHash"), nil
}

package syncer

import "context"

type signerClient struct{}

func newSignerClient() *signerClient {
	return &signerClient{}
}

func (sc *signerClient) coSignTx(ctx context.Context) ([]byte, error) {
	return []byte("syncer_service"), nil
}

package uploader

type eventClient struct {
}

func newEventClient() *eventClient {
	return &eventClient{}
}

func (ec *eventClient) waitChainEvent(txHash []byte) (uint64, error) {
	return 2012, nil
}

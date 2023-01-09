package gateway

// todo: impl of call RetrieverService
type retrieverClient struct {
}

func newRetrieverClient() *retrieverClient {
	return &retrieverClient{}
}

type queryACLOption struct {
	reqCtx *requestContext
}

func (rc *retrieverClient) queryACL(name string, opt *queryACLOption) error {
	return nil
}

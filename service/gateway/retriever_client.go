package gateway

// todo: impl of call RetriverService
type retrieverClient struct {
}

func newRetrieverClient() *retrieverClient {
	return &retrieverClient{}
}

type queryACLOption struct {
	reqCtx *requestContext
	// todo:
}

func (rc *retrieverClient) queryACL(name string, opt *queryACLOption) error {
	return nil
}

package gateway

import (
	"testing"

	"github.com/gorilla/mux"
)

var (
	testDomain = "www.route-test.com"
	config     = &GatewayConfig{
		Domain: testDomain,
	}
	gw = &Gateway{
		config: config,
	}
	scheme     = "https://"
	bucketName = "test-bucket-name"
	objectName = "test-object-name"
)

func setupRouter(t *testing.T) *mux.Router {
	gwRouter := mux.NewRouter().SkipClean(true)
	gw.registerHandler(gwRouter)
	return gwRouter
}
